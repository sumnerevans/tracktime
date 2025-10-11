package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/sumnerevans/tracktime/internal/lib"
	"github.com/sumnerevans/tracktime/internal/report"
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

func (s Sort) toReportSort() report.Sort {
	if s == SortTimeSpent {
		return report.SortTimeSpent
	}
	return report.SortAlphabetical
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
	DescriptionGrain   bool `arg:"--descriptiongrain" help:"report on the task grain"`
	NoDescriptionGrain bool `arg:"--no-descriptiongrain" help:"do not report on the task grain"`

	// Narrow the set of time entries to report on
	Customer lib.Customer `arg:"-c,--customer" help:"customer ID to generate a report for"`
	Project  lib.Project  `arg:"-p,--project" help:"project name to generate a report for"`

	// How to sort the report
	Sort Sort `arg:"-s,--sort" help:"the grain to sort the report by (alphabetical,alpha,a or time-spent,time,t)" default:"alphabetical"`
	Desc bool `arg:"--desc" help:"sort descending"`
	Asc  bool `arg:"--asc" help:"sort ascending"`

	// Output file
	OutputFile lib.Filename `arg:"-o,--outfile" help:"specify the filename to export the report to (supports PDF, HTML, and RST files, if set to '-' then the report is printed to stdout)" default:"-"`
}

func (r *Report) Run(config *lib.Config) error {
	var start, end lib.Date
	switch {
	case r.Today:
		start = lib.Today()
		end = lib.Today()
	case r.Yesterday:
		start = lib.Today().AddDays(-1)
		end = lib.Today().AddDays(-1)
	case r.ThisWeek:
		start = lib.Today().AddDays(-int(lib.Today().Weekday()))
		end = lib.Today().AddDays(6 - int(lib.Today().Weekday()))
	case r.LastWeek:
		start = lib.Today().AddDays(-int(lib.Today().Weekday()) - 7)
		end = lib.Today().AddDays(6 - int(lib.Today().Weekday()) - 7)
	case r.ThisMonth:
		start = lib.Today().AddDays(1 - int(lib.Today().Day()))
		end = lib.Today().AddMonths(1).AddDays(-int(lib.Today().AddMonths(1).Day()))
	case r.ThisYear:
		start = lib.NewDate(lib.Today().Year(), 1, 1)
		end = lib.NewDate(lib.Today().Year(), 12, 31)
	case r.LastYear:
		start = lib.NewDate(lib.Today().Year()-1, 1, 1)
		end = lib.NewDate(lib.Today().Year()-1, 12, 31)
	default: // Last month
		start = lib.Today().AddMonths(-1).AddDays(1 - int(lib.Today().AddMonths(-1).Day()))
		end = lib.Today().AddDays(-int(lib.Today().Day()))
	}
	if r.Start != nil {
		start = *r.Start
	}
	if r.End != nil {
		end = *r.End
	}

	if start.Equal(time.Time{}) {
		return fmt.Errorf("start date is required")
	}
	if end.Equal(time.Time{}) {
		return fmt.Errorf("end date is required")
	}
	if start.After(end.Time) {
		return fmt.Errorf("start date must be before end date")
	}

	// Determine sort direction (desc overrides asc)
	reverse := r.Desc

	// Determine grains
	taskGrain := r.TaskGrain || r.DescriptionGrain
	descriptionGrain := r.DescriptionGrain
	if r.NoTaskGrain {
		taskGrain = false
		descriptionGrain = false
	}
	if r.NoDescriptionGrain {
		descriptionGrain = false
	}

	// Create report
	rep, err := report.New(
		config,
		start,
		end,
		r.Customer,
		r.Project,
		r.Sort.toReportSort(),
		reverse,
		taskGrain,
		descriptionGrain,
	)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	// Generate output
	output := rep.GenerateTextReport()

	// Strip formatting for stdout (match Python behavior line 651)
	output = strings.ReplaceAll(output, "| ", "")
	output = strings.ReplaceAll(output, "**", "")

	// Print to stdout (or file, TODO)
	fmt.Println(output)

	return nil
}
