package commands

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/importer"
	"github.com/sumnerevans/tracktime/internal/resolver"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Import struct {
	Type          string             `arg:"positional,required" help:"importer type (e.g. tempo)"`
	File          types.Filename     `arg:"positional,required" help:"path to the file to import"`
	Customer      timeentry.Customer `arg:"--customer" help:"default customer for imported entries (overridden by importer-supplied customer)"`
	NoItemDetails bool               `arg:"--no-item-details" help:"skip seeding the item detail cache"`
	DryRun        bool               `arg:"--dry-run" help:"show what would be imported without writing anything"`
}

func (i *Import) Run(ctx context.Context, cfg *config.Config) error {
	log := zerolog.Ctx(ctx)

	var imp importer.Importer
	for _, candidate := range importer.Importers {
		if candidate.Name() == i.Type {
			imp = candidate
			break
		}
	}
	if imp == nil {
		return fmt.Errorf("unknown importer %q", i.Type)
	}

	log.Info().Str("importer", i.Type).Str("file", i.File.Expand()).Msg("starting import")

	result, err := imp.Import(ctx, cfg, i.File.Expand())
	if err != nil {
		return err
	}

	log.Info().
		Int("entries", len(result.Entries)).
		Int("item_details", len(result.ItemDetails)).
		Msg("importer returned results")

	if i.Customer != "" {
		for idx := range result.Entries {
			if result.Entries[idx].Entry.Customer == "" {
				result.Entries[idx].Entry.Customer = i.Customer
			}
		}
	}

	added, skipped, updated, err := applyImport(ctx, cfg, result, i.DryRun)
	if err != nil {
		return err
	}

	if i.DryRun {
		log.Info().Str("importer", i.Type).Msg("dry run complete")
		msg := fmt.Sprintf("dry run: would import %d entries (%d skipped", added, skipped)
		if updated > 0 {
			msg += fmt.Sprintf(", %d updated", updated)
		}
		msg += ")"
		fmt.Fprintln(os.Stderr, msg)
		return nil
	}

	if i.NoItemDetails {
		log.Info().Msg("skipping item detail cache (--no-item-details)")
		fmt.Fprintf(os.Stderr, "item details skipped (--no-item-details)\n")
	} else if len(result.ItemDetails) > 0 {
		log.Info().Int("count", len(result.ItemDetails)).Msg("seeding item detail cache")
		resolver.NewItemDetailCache(ctx, string(cfg.Directory), cfg, nil).Seed(ctx, result.ItemDetails)
	}

	log.Info().Str("importer", i.Type).Msg("import complete")
	msg := fmt.Sprintf("import complete: %d added, %d skipped", added, skipped)
	if updated > 0 {
		msg += fmt.Sprintf(", %d updated", updated)
	}
	fmt.Fprintln(os.Stderr, msg)
	return nil
}

// applyImport applies result.Entries: append-only, deduped by {start, type, project, taskID}.
// Existing entries with no customer are updated if the incoming entry has one.
// When dryRun is true, computes counts but skips writing to disk.
// Returns the total number of entries added, skipped, and updated.
func applyImport(ctx context.Context, cfg *config.Config, result *importer.ImportResult, dryRun bool) (totalAdded, totalSkipped, totalUpdated int, err error) {
	log := zerolog.Ctx(ctx)

	// Group entries by date string to avoid time.Time map key issues.
	byDate := make(map[string][]importer.ImportEntry)
	for _, ie := range result.Entries {
		key := ie.Date.Format("2006-01-02")
		byDate[key] = append(byDate[key], ie)
	}

	log.Debug().Int("dates", len(byDate)).Msg("grouped entries by date")

	for dateKey, incoming := range byDate {
		parsed, _ := parseDate(dateKey)
		el, err := timeentry.EntryListForDay(cfg, parsed)
		if err != nil {
			log.Error().Err(err).Str("date", dateKey).Msg("failed to load entry list")
			continue
		}

		log := log.With().Str("date", dateKey).Logger()

		log.Debug().
			Int("existing", len(el.Entries)).
			Int("incoming", len(incoming)).
			Msg("processing day")

		// Build map from dedup key to existing entry pointer so we can update in place.
		type entryKey struct {
			start int
			timeentry.AggregatedTimeKey
		}
		entKey := func(e *timeentry.TimeEntry) entryKey {
			return entryKey{e.Start.Minutes(), timeentry.AggregatedTimeKey{Type: e.Type, Project: e.Project, TaskID: e.TaskID}}
		}
		seen := make(map[entryKey]*timeentry.TimeEntry, len(el.Entries))
		for _, e := range el.Entries {
			seen[entKey(e)] = e
		}

		added := 0
		skipped := 0
		updated := 0
		for _, ie := range incoming {
			log := log.With().
				Str("start", ie.Entry.Start.String()).
				Str("stop", ie.Entry.Stop.String()).
				Stringer("type", ie.Entry.Type).
				Stringer("project", ie.Entry.Project).
				Stringer("taskid", ie.Entry.TaskID).
				Stringer("customer", ie.Entry.Customer).
				Logger()

			k := entKey(ie.Entry)
			if existing, ok := seen[k]; ok {
				if existing.Customer == "" && ie.Entry.Customer != "" {
					log.Debug().Msg("updating customer on existing entry")
					existing.Customer = ie.Entry.Customer
					updated++
				} else {
					log.Debug().Msg("skipping duplicate entry")
					skipped++
				}
				continue
			}
			log.Debug().Msg("adding entry")
			seen[k] = ie.Entry
			el.Entries = append(el.Entries, ie.Entry)
			added++
		}

		log.Info().Int("added", added).Int("skipped", skipped).Int("updated", updated).Msg("processed day")
		totalAdded += added
		totalSkipped += skipped
		totalUpdated += updated

		sort.Slice(el.Entries, func(a, b int) bool {
			return el.Entries[a].Start.Before(el.Entries[b].Start)
		})

		if dryRun {
			log.Debug().Msg("dry run: skipping save")
		} else if err := el.Save(); err != nil {
			log.Err(err).Msg("failed to save entry list")
		} else {
			log.Debug().Msg("saved entry list")
		}
	}

	log.Info().Int("total_added", totalAdded).Int("total_skipped", totalSkipped).Int("total_updated", totalUpdated).Msg("apply complete")
	return totalAdded, totalSkipped, totalUpdated, nil
}

func parseDate(s string) (types.Date, error) {
	var d types.Date
	err := d.UnmarshalText([]byte(s))
	return d, err
}
