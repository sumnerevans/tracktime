package commands

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/sumnerevans/tracktime/lib"
)

type Edit struct {
	Date lib.Date `arg:"-d,--date" help:"the date to list time entries for" default:"today"`
}

func (s *Edit) Run(config *lib.Config) error {
	// TODO make sure the header exists

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
	args = append(args, lib.DayFilename(config, s.Date))

	cmd := exec.Command(editor, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msg("Couldn't open editor")
		return err
	}
	// TODO Sync
	return nil
}
