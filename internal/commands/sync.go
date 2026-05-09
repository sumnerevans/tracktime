package commands

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/importer"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Sync struct {
	Month *types.Month `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *Sync) Run(ctx context.Context, cfg *config.Config) error {
	return syncMonth(ctx, cfg, *s.Month)
}

func syncMonth(ctx context.Context, cfg *config.Config, month types.Month) error {
	log := zerolog.Ctx(ctx)
	for _, imp := range importer.Importers {
		imp.Init(cfg.Sync)
		if _, err := imp.Import(ctx, month); err != nil {
			log.Error().Err(err).Str("importer", imp.Name()).Msg("import failed")
		}
	}
	return nil
}
