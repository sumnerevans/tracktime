package main

import (
	"fmt"
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

type edit struct {
	Date string `arg:"-d,--date" help:"the date to list time entries for" default:"today"`
}

func (s *edit) Run(config *lib.Config) error {
	return nil
}

type sync struct {
	Month string `arg:"positional" help:"the month to synchronize time entries for (accepted formats: 01, 1, Jan, January, 2019-01)" default:"this month"`
}

func (s *sync) Run(config *lib.Config) error {
	return nil
}

type Sort int

const (
	SortAlphabetical Sort = iota
	SortTimeSpent
)

func (s *Sort) UnmarshalText(b []byte) error {
	switch string(b) {
	case "alphabetical", "alpha", "a":
		*s = SortAlphabetical
	case "time-spent", "time", "t":
		*s = SortTimeSpent
	default:
		return fmt.Errorf("invalid sort value '%s'", string(b))
	}
	return nil
}

type report struct {
	// Specify the range positionally
	Start string `arg:"positional" help:"specify the start of the reporting range (defaults to the beginning of last month)"`
	End   string `arg:"positional" help:"specify the end of the reporting range (defaults to the end of last month)"`

	// Specify the range using shorthand
	Month     string `arg:"-m,--month" help:"shorthand for reporting over an entire month (can be combined with --year, accepted formats: 01, 1, Jan, January, 2019-01)"`
	Year      int    `arg:"-y,--year" help:"shorthand for reporting over an entire year (can be combined with --month)"`
	Today     bool   `arg:"--today" help:"shorthand for reporting on today"`
	Yesterday bool   `arg:"--yesterday" help:"shorthand for reporting on yesterday"`
	ThisWeek  bool   `arg:"--thisweek" help:"shorthand for reporting on the current week (Sunday-today)"`
	LastWeek  bool   `arg:"--lastweek" help:"shorthand for reporting on last week (Sunday-Saturday)"`
	ThisMonth bool   `arg:"--thismonth" help:"shorthand for reporting on the current month"`
	LastMonth bool   `arg:"--lastmonth" help:"shorthand for reporting on last month"`
	ThisYear  bool   `arg:"--thisyear" help:"shorthand for reporting on the current year"`
	LastYear  bool   `arg:"--lastyear" help:"shorthand for reporting on last year"`

	// Specify the grains to show
	TaskGrain          bool `arg:"--taskgrain" help:"report on the task grain"`
	NoTaskGrain        bool `arg:"--no-taskgrain" help:"do not report on the task grain"`
	DescriptionGrain   bool `arg:"--no-descriptiongrain" help:"report on the task grain"`
	NoDescriptionGrain bool `arg:"--no-descriptiongrain" help:"do not report on the task grain"`

	// Narrow the set of time entries to report on
	Customer string `arg:"-c,--customer" help:"customer ID to generate a report for"`
	Project  string `arg:"-p,--project" help:"project name to generate a report for"`

	// How to sort the report
	Sort Sort `arg:"-s,--sort" help:"the grain to sort the report by (alphabetical,alpha,a or time-spent,time,t)" default:"alphabetical"`
	Desc bool `arg:"--desc" help:"sort descending"`
	Asc  bool `arg:"--asc" help:"sort ascending"`

	// Output file
	OutputFile string `arg:"-o,--outfile" help:"specify the filename to export the report to (supports PDF, HTML, and RST files, if set to '-' then the report is printed to stdout)" default:"-"`
}

func (s *report) Run(config *lib.Config) error {
	return nil
}

type args struct {
	Start      *commands.Start `arg:"subcommand" help:"start a new time entry for today"`
	Stop       *stop           `arg:"subcommand" help:"stop the current time entry"`
	Resume     *resume         `arg:"subcommand" help:"resume a time entry from today"`
	List       *commands.List  `arg:"subcommand" help:"list the time entries for a date"`
	Edit       *edit           `arg:"subcommand" help:"edit time entries for a date"`
	Sync       *sync           `arg:"subcommand" help:"synchronize time spent on tasks for a month to external services"`
	Report     *report         `arg:"subcommand" help:"output a report about time spent in a time range"`
	ConfigFile lib.Filename    `arg:"--config" help:"the configuration file to use" default:"$HOME/.config/tracktime/tracktimerc"`
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
