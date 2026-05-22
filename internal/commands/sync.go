package commands

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Sync struct {
	Month *types.Month `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *Sync) Run(ctx context.Context, cfg *config.Config) error {
	return syncMonth(ctx, cfg, *s.Month)
}

func syncMonth(_ context.Context, _ *config.Config, _ types.Month) error {
	// TODO: push metadata to external services
	return nil
}
