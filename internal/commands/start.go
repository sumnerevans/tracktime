package commands

import (
	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Start struct {
	Description string                  `arg:"positional" placeholder:"DESC" help:"the descripiton of the time entry"`
	Start       *types.Time             `arg:"-s,--start" help:"the start time of the time entry" default:"now"`
	Type        timeentry.TimeEntryType `arg:"-t,--type" help:"the type of the time entry"`
	Project     timeentry.Project       `arg:"-p,--project" help:"the project of the time entry"`
	Customer    timeentry.Customer      `arg:"-c,--customer" help:"the customer of the time entry"`
	TaskID      timeentry.TaskID        `arg:"-i,--taskid" help:"the task ID of the time entry"`
}

func (s *Start) Run(config *config.Config) error {
	entryList, err := timeentry.EntryListForDay(config, types.Today())
	if err != nil {
		return err
	}
	// TODO sync
	return entryList.Start(s.Start, s.Description, s.Type, s.Project, s.Customer, s.TaskID)
}
