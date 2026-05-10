// Package commands implements the CLI subcommands for the tt tool.
package commands

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type Edit struct {
	Date types.Date `arg:"-d,--date" help:"the date to list time entries for" default:"today"`
}

func (s *Edit) Run(ctx context.Context, config *config.Config) error {
	// Make sure the header exists
	if entryList, err := timeentry.EntryListForDay(config, s.Date); err != nil {
		return err
	} else if err := entryList.Save(); err != nil {
		return err
	}

	editor := config.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		switch runtime.GOOS {
		case "windows":
			editor = "notepad"
		default:
			editor = "vim"
		}
	}

	args := config.EditorArgs
	args = append(args, timeentry.DayFilename(config, s.Date))

	cmd := exec.Command(editor, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Couldn't open editor")
		return err
	}
	zerolog.Ctx(ctx).Info().Stringer("date", s.Date).Msg("edit complete")
	return syncMonth(ctx, config, types.MonthOf(s.Date))
}
