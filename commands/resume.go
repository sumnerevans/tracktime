package commands

import "github.com/sumnerevans/tracktime/lib"

type Resume struct {
	Entry       int       `arg:"positional" help:"the entry number to resume" default:"-1"`
	Description *string   `arg:"positional" placeholder:"DESC" help:"the description for the new time entry (defaults to that of the entry being resumed)"`
	Start       *lib.Time `arg:"-s,--start" help:"the start time of the resumed time entry" default:"now"`
}

func (s *Resume) Run(config *lib.Config) error {
	entryList, err := lib.EntryListForDay(config, lib.Today())
	if err != nil {
		return err
	}
	return entryList.Resume(s.Entry, s.Description, s.Start)
}
