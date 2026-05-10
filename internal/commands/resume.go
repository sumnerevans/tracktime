package commands

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Resume struct {
	Entry       int         `arg:"-n,--entry" help:"entry number to resume (negative counts from end: -1=last, -2=second-to-last)" default:"-1"`
	Description *string     `arg:"positional" placeholder:"DESC" help:"the description for the new time entry (defaults to that of the entry being resumed)"`
	Start       *types.Time `arg:"-s,--start" help:"the start time of the resumed time entry" default:"now"`
}

func (s *Resume) Run(ctx context.Context, config *config.Config) error {
	log := zerolog.Ctx(ctx)
	today := types.Today()
	entryList, err := timeentry.EntryListForDay(config, today)
	if err != nil {
		return err
	}
	if err := entryList.Resume(s.Entry, s.Description, s.Start); err != nil {
		return err
	}
	log.Info().Int("entry", s.Entry).Stringer("start", s.Start).Msg("resumed time entry")
	return syncMonth(ctx, config, types.MonthOf(today))
}
