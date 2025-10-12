package commands

import (
	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Stop struct {
	Stop *types.Time `arg:"-s,--stop" help:"the time at which to stop the current time entry" default:"now"`
}

func (s *Stop) Run(config *config.Config) error {
	entryList, err := timeentry.EntryListForDay(config, types.Today())
	if err != nil {
		return err
	}
	// TODO sync
	return entryList.Stop(s.Stop)
}
