package main

import (
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"

	"github.com/sumnerevans/tracktime/commands"
	"github.com/sumnerevans/tracktime/lib"
)

type stop struct {
	Stop string `arg:"-s,--stop" help:"the time at which to stop the current time entry" default:"now"`
}

func (s *stop) Run(config *lib.Config) error {
	return nil
}

type resume struct {
	Entry       int    `arg:"positional" help:"the entry number to resume" default:"-1"`
	Description string `arg:"positional" placeholder:"DESC" help:"the description for the new time entry (defaults to that of the entry being resumed)"`
	Start       string `arg:"-s,--start" help:"the start time of the resumed time entry" default:"now"`
}

func (s *resume) Run(config *lib.Config) error {
	return nil
}

type sync struct {
	Month string `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *sync) Run(config *lib.Config) error {
	return nil
}

type args struct {
	Start      *commands.Start  `arg:"subcommand" help:"start a new time entry for today"`
	Stop       *stop            `arg:"subcommand" help:"stop the current time entry"`
	Resume     *resume          `arg:"subcommand" help:"resume a time entry from today"`
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
		args.Start.Run(config)
	case args.Stop != nil:
		args.Stop.Run(config)
	case args.Resume != nil:
		args.Resume.Run(config)
	case args.List != nil:
		args.List.Run(config)
	case args.Edit != nil:
		args.Edit.Run(config)
	case args.Sync != nil:
		args.Sync.Run(config)
	case args.Report != nil:
		args.Report.Run(config)
	default:
		args.List = &commands.List{Date: lib.Date{Time: time.Now()}}
		args.List.Run(config)
	}
}
