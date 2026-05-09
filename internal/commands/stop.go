package commands

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Stop struct {
	Stop *types.Time `arg:"-s,--stop" help:"the time at which to stop the current time entry" default:"now"`
}

func (s *Stop) Run(ctx context.Context, config *config.Config) error {
	today := types.Today()
	entryList, err := timeentry.EntryListForDay(config, today)
	if err != nil {
		return err
	}
	if err := entryList.Stop(s.Stop); err != nil {
		return err
	}
	return syncMonth(ctx, config, types.MonthOf(today))
}
