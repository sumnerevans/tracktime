// Package main is the entry point for the tt command-line tool.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/commands"
	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/types"
)

type args struct {
	Start      *commands.Start  `arg:"subcommand" help:"start a new time entry for today"`
	Stop       *commands.Stop   `arg:"subcommand" help:"stop the current time entry"`
	Resume     *commands.Resume `arg:"subcommand" help:"resume a time entry from today"`
	List       *commands.List   `arg:"subcommand" help:"list the time entries for a date"`
	Edit       *commands.Edit   `arg:"subcommand" help:"edit time entries for a date"`
	Sync       *commands.Sync   `arg:"subcommand" help:"synchronize time spent on tasks for a month to external services"`
	Report     *commands.Report `arg:"subcommand" help:"output a report about time spent in a time range"`
	ConfigFile types.Filename   `arg:"--config" help:"the configuration file to use" default:"$HOME/.config/tracktime/tracktimerc"`
}

var _ arg.Versioned = (*args)(nil)
var _ arg.Epilogued = (*args)(nil)
var _ arg.Described = (*args)(nil)

func (args) Version() string {
	return "tracktime v0.11.0"
}

func (a *args) Description() string {
	return "tracktime -- a filesystem-backed time tracking solution"
}

func (args) Epilogue() string {
	return "For more information visit https://github.com/sumnerevans/tracktime"
}

func main() {
	// Bootstrap logger used only until the config is loaded.
	bootstrap := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx := bootstrap.WithContext(context.Background())

	var args args
	arg.MustParse(&args)

	cfg, err := config.ReadConfig(args.ConfigFile)
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Msg("Couldn't read config file")
	}

	// Replace bootstrap logger with the one from config.
	logger, err := cfg.Logging.Compile()
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Msg("Couldn't compile logging config")
	}
	ctx = logger.WithContext(context.Background())

	logger.Info().Any("args", args).Msg("Starting tracktime")

	switch {
	case args.Start != nil:
		err = args.Start.Run(ctx, cfg)
	case args.Stop != nil:
		err = args.Stop.Run(ctx, cfg)
	case args.Resume != nil:
		err = args.Resume.Run(ctx, cfg)
	case args.Edit != nil:
		err = args.Edit.Run(ctx, cfg)
	case args.Sync != nil:
		err = args.Sync.Run(ctx, cfg)
	case args.Report != nil:
		err = args.Report.Run(ctx, cfg)
	default:
		if args.List == nil {
			args.List = &commands.List{Date: types.Date{Time: time.Now()}}
		}
		err = args.List.Run(ctx, cfg)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}
