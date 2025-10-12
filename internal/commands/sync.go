package commands

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/synchroniser"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Sync struct {
	Month *types.Month `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *Sync) Run(config *config.Config) error {
	for _, synchroniser := range synchroniser.Synchronisers {
		go synchroniser.Sync(context.Background(), nil, nil, *s.Month)
	}
	return nil
}
