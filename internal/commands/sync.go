package commands

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/exporter"
	"github.com/sumnerevans/tracktime/internal/importer"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Sync struct {
	Month *types.Month `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *Sync) Run(ctx context.Context, cfg *config.Config) error {
	log := zerolog.Ctx(ctx)
	month := *s.Month

	aggregatedTime, err := aggregateMonth(cfg, month)
	if err != nil {
		return err
	}

	syncedTime, err := exporter.ReadSyncedFile(string(cfg.Directory), month)
	if err != nil {
		return err
	}

	for _, imp := range importer.Importers {
		imp.Init(cfg.Sync)
		if _, err := imp.Import(ctx, month); err != nil {
			log.Error().Err(err).Str("importer", imp.Name()).Msg("import failed")
		}
	}

	newSynced := syncedTime
	for _, exp := range exporter.Exporters {
		exp.Init(cfg.Sync)
		result, err := exp.Export(ctx, aggregatedTime, syncedTime, month)
		if err != nil {
			log.Error().Err(err).Str("exporter", exp.Name()).Msg("export failed")
			continue
		}
		for k, v := range result {
			newSynced[k] = v
		}
	}

	return exporter.WriteSyncedFile(string(cfg.Directory), month, newSynced)
}

// aggregateMonth sums all time entries for every day in month.
func aggregateMonth(cfg *config.Config, month types.Month) (timeentry.AggregatedTime, error) {
	result := make(timeentry.AggregatedTime)
	for day := 1; day <= month.DaysInMonth(); day++ {
		date := types.NewDate(month.Year(), int(month.Month()), day)
		el, err := timeentry.EntryListForDay(cfg, date)
		if err != nil {
			return nil, err
		}
		for _, entry := range el.EntriesForCustomer("") {
			duration, err := entry.Duration(false)
			if err != nil {
				continue // skip unended entries
			}
			key := timeentry.AggregatedTimeKey{
				Type:    entry.Type,
				Project: entry.Project,
				TaskID:  entry.TaskID,
			}
			result[key] += duration
		}
	}
	return result, nil
}
