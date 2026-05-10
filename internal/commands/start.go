package commands

import (
	"context"

	"github.com/rs/zerolog"

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

func (s *Start) Run(ctx context.Context, config *config.Config) error {
	log := zerolog.Ctx(ctx)
	today := types.Today()
	entryList, err := timeentry.EntryListForDay(config, today)
	if err != nil {
		return err
	}
	if err := entryList.Start(s.Start, s.Description, s.Type, s.Project, s.Customer, s.TaskID); err != nil {
		return err
	}
	log.Info().
		Str("description", s.Description).
		Str("type", string(s.Type)).
		Str("project", string(s.Project)).
		Str("customer", string(s.Customer)).
		Str("taskid", string(s.TaskID)).
		Stringer("start", s.Start).
		Msg("started time entry")
	return syncMonth(ctx, config, types.MonthOf(today))
}
