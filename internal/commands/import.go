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
	Type          string         `arg:"positional,required" help:"importer type (e.g. tempo)"`
	File          types.Filename `arg:"positional,required" help:"path to the file to import"`
	NoItemDetails bool           `arg:"--no-item-details" help:"skip seeding the item detail cache"`
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

	added, skipped, err := applyImport(ctx, cfg, result)
	if err != nil {
		return err
	}

	if i.NoItemDetails {
		log.Info().Msg("skipping item detail cache (--no-item-details)")
		fmt.Fprintf(os.Stderr, "item details skipped (--no-item-details)\n")
	} else if len(result.ItemDetails) > 0 {
		log.Info().Int("count", len(result.ItemDetails)).Msg("seeding item detail cache")
		resolver.NewItemDetailCache(ctx, string(cfg.Directory), cfg, nil).Seed(ctx, result.ItemDetails)
	}

	log.Info().Str("importer", i.Type).Msg("import complete")
	fmt.Fprintf(os.Stderr, "import complete: %d added, %d skipped\n", added, skipped)
	return nil
}

// applyImport writes result.Entries to disk: append-only, deduped by {start, type, project, taskID}.
// Returns the total number of entries added and skipped.
func applyImport(ctx context.Context, cfg *config.Config, result *importer.ImportResult) (totalAdded, totalSkipped int, err error) {
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

		log.Debug().
			Str("date", dateKey).
			Int("existing", len(el.Entries)).
			Int("incoming", len(incoming)).
			Msg("processing day")

		// Build set of existing entry keys for dedup.
		type entryKey struct {
			start int
			timeentry.AggregatedTimeKey
		}
		entKey := func(e *timeentry.TimeEntry) entryKey {
			return entryKey{e.Start.Minutes(), timeentry.AggregatedTimeKey{Type: e.Type, Project: e.Project, TaskID: e.TaskID}}
		}
		seen := make(map[entryKey]bool, len(el.Entries))
		for _, e := range el.Entries {
			seen[entKey(e)] = true
		}

		added := 0
		skipped := 0
		for _, ie := range incoming {
			k := entKey(ie.Entry)
			if seen[k] {
				log.Debug().
					Str("date", dateKey).
					Str("start", ie.Entry.Start.String()).
					Str("type", string(ie.Entry.Type)).
					Str("project", string(ie.Entry.Project)).
					Str("taskid", string(ie.Entry.TaskID)).
					Msg("skipping duplicate entry")
				skipped++
				continue
			}
			log.Debug().
				Str("date", dateKey).
				Str("start", ie.Entry.Start.String()).
				Str("stop", ie.Entry.Stop.String()).
				Str("type", string(ie.Entry.Type)).
				Str("project", string(ie.Entry.Project)).
				Str("taskid", string(ie.Entry.TaskID)).
				Msg("adding entry")
			seen[k] = true
			el.Entries = append(el.Entries, ie.Entry)
			added++
		}

		log.Info().Str("date", dateKey).Int("added", added).Int("skipped", skipped).Msg("processed day")
		totalAdded += added
		totalSkipped += skipped

		sort.Slice(el.Entries, func(a, b int) bool {
			return el.Entries[a].Start.Before(el.Entries[b].Start)
		})

		if err := el.Save(); err != nil {
			log.Error().Err(err).Str("date", dateKey).Msg("failed to save entry list")
		} else {
			log.Debug().Str("date", dateKey).Msg("saved entry list")
		}
	}

	log.Info().Int("total_added", totalAdded).Int("total_skipped", totalSkipped).Msg("apply complete")
	return totalAdded, totalSkipped, nil
}

func parseDate(s string) (types.Date, error) {
	var d types.Date
	err := d.UnmarshalText([]byte(s))
	return d, err
}
