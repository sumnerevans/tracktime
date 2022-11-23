package commands

import (
	"context"

	"github.com/sumnerevans/tracktime/lib"
	"github.com/sumnerevans/tracktime/synchroniser"
)

type Sync struct {
	Month *lib.Month `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *Sync) Run(config *lib.Config) error {
	for _, synchroniser := range synchroniser.Synchronisers {
		go synchroniser.Sync(context.Background(), nil, nil, *s.Month)
	}
	return nil
}
