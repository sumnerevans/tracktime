package commands

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Stop struct {
	Stop *types.Time `arg:"-s,--stop" help:"the time at which to stop the current time entry" default:"now"`
}

func (s *Stop) Run(ctx context.Context, config *config.Config) error {
	log := zerolog.Ctx(ctx)
	today := types.Today()
	entryList, err := timeentry.EntryListForDay(config, today)
	if err != nil {
		return err
	}
	if err := entryList.Stop(s.Stop); err != nil {
		return err
	}
	log.Info().Stringer("stop", s.Stop).Msg("stopped time entry")
	return syncMonth(ctx, config, types.MonthOf(today))
}
