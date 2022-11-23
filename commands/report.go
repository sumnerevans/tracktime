package commands

import (
	"fmt"

	"github.com/sumnerevans/tracktime/lib"
)

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

type Report struct {
	// Specify the range positionally
	Start *lib.Date `arg:"positional" help:"specify the start of the reporting range (defaults to the beginning of last month)"`
	End   *lib.Date `arg:"positional" help:"specify the end of the reporting range (defaults to the end of last month)"`

	// Specify the range using shorthand
	Month     lib.Month `arg:"-m,--month" help:"shorthand for reporting over an entire month (can be combined with --year, accepted formats: 01, 1, Jan, January, 2019-01)"`
	Year      int       `arg:"-y,--year" help:"shorthand for reporting over an entire year (can be combined with --month)"`
	Today     bool      `arg:"--today" help:"shorthand for reporting on today"`
	Yesterday bool      `arg:"--yesterday" help:"shorthand for reporting on yesterday"`
	ThisWeek  bool      `arg:"--thisweek" help:"shorthand for reporting on the current week (Sunday-today)"`
	LastWeek  bool      `arg:"--lastweek" help:"shorthand for reporting on last week (Sunday-Saturday)"`
	ThisMonth bool      `arg:"--thismonth" help:"shorthand for reporting on the current month"`
	LastMonth bool      `arg:"--lastmonth" help:"shorthand for reporting on last month"`
	ThisYear  bool      `arg:"--thisyear" help:"shorthand for reporting on the current year"`
	LastYear  bool      `arg:"--lastyear" help:"shorthand for reporting on last year"`

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
	OutputFile lib.Filename `arg:"-o,--outfile" help:"specify the filename to export the report to (supports PDF, HTML, and RST files, if set to '-' then the report is printed to stdout)" default:"-"`
}

func (s *Report) Run(config *lib.Config) error {
	return nil
}
