package commands

import "github.com/sumnerevans/tracktime/lib"

type Stop struct {
	Stop *lib.Time `arg:"-s,--stop" help:"the time at which to stop the current time entry" default:"now"`
}

func (s *Stop) Run(config *lib.Config) error {
	entryList, err := lib.EntryListForDay(config, lib.Today())
	if err != nil {
		return err
	}
	return entryList.Stop(s.Stop)
}
