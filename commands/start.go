package commands

import (
	"github.com/sumnerevans/tracktime/lib"
)

type Start struct {
	Description string            `arg:"positional" placeholder:"DESC" help:"the descripiton of the time entry"`
	Start       *lib.Time         `arg:"-s,--start" help:"the start time of the time entry" default:"now"`
	Type        lib.TimeEntryType `arg:"-t,--type" help:"the type of the time entry"`
	Project     lib.Project       `arg:"-p,--project" help:"the project of the time entry"`
	Customer    lib.Customer      `arg:"-c,--customer" help:"the customer of the time entry"`
	TaskID      lib.TaskID        `arg:"-i,--taskid" help:"the task ID of the time entry"`
}

func (s *Start) Run(config *lib.Config) error {
	entryList, err := lib.EntryListForDay(config, lib.Today())
	if err != nil {
		return err
	}
	// TODO sync
	return entryList.Start(s.Start, s.Description, s.Type, s.Project, s.Customer, s.TaskID)
}
