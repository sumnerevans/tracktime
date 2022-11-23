package main

import (
	"fmt"
	"os"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"

	"github.com/sumnerevans/tracktime/commands"
	"github.com/sumnerevans/tracktime/lib"
)

type sync struct {
	Month string `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *sync) Run(config *lib.Config) error {
	return nil
}

type args struct {
	Start      *commands.Start  `arg:"subcommand" help:"start a new time entry for today"`
	Stop       *commands.Stop   `arg:"subcommand" help:"stop the current time entry"`
	Resume     *commands.Resume `arg:"subcommand" help:"resume a time entry from today"`
	List       *commands.List   `arg:"subcommand" help:"list the time entries for a date"`
	Edit       *commands.Edit   `arg:"subcommand" help:"edit time entries for a date"`
	Sync       *sync            `arg:"subcommand" help:"synchronize time spent on tasks for a month to external services"`
	Report     *commands.Report `arg:"subcommand" help:"output a report about time spent in a time range"`
	ConfigFile lib.Filename     `arg:"--config" help:"the configuration file to use" default:"$HOME/.config/tracktime/tracktimerc"`
}

func (args) Version() string {
	return "tracktime v0.11.0"
}

func (args) Epilogue() string {
	return "For more information visit https://github.com/sumnerevans/tracktime"
}

func main() {
	var args args
	arg.MustParse(&args)

	config, err := lib.ReadConfig(args.ConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't read config file")
	}

	switch {
	case args.Start != nil:
		err = args.Start.Run(config)
	case args.Stop != nil:
		err = args.Stop.Run(config)
	case args.Resume != nil:
		err = args.Resume.Run(config)
	case args.List != nil:
		err = args.List.Run(config)
	case args.Edit != nil:
		err = args.Edit.Run(config)
	case args.Sync != nil:
		err = args.Sync.Run(config)
	case args.Report != nil:
		err = args.Report.Run(config)
	default:
		args.List = &commands.List{Date: lib.Date{Time: time.Now()}}
		err = args.List.Run(config)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}
