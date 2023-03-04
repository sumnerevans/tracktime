package commands

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

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

	aggregatedTime := map[lib.Customer]map[lib.Project]map[lib.TaskID]map[string][]*lib.TimeEntry{}
	dayStats := map[lib.Date]time.Duration{}

	for day := start; day.Before(end.Time) || day.Equal(end.Time); day = day.AddDays(1) {
		entryList, err := lib.EntryListForDay(config, day)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read entry list")
		}

		for _, entry := range entryList.EntriesForCustomer(r.Customer) {
			if r.Project != "" && entry.Project != r.Project {
				continue
			}

			if _, ok := dayStats[day]; !ok {
				dayStats[day] = time.Duration(0)
			}
			duration, err := entry.Duration(false)
			if err != nil {
				return fmt.Errorf("unended time entry on %s", day.Time.Format("2006-01-02"))
			}
			dayStats[day] = dayStats[day] + duration

			if _, ok := aggregatedTime[entry.Customer]; !ok {
				aggregatedTime[entry.Customer] = map[lib.Project]map[lib.TaskID]map[string][]*lib.TimeEntry{}
			}
			if _, ok := aggregatedTime[entry.Customer][entry.Project]; !ok {
				aggregatedTime[entry.Customer][entry.Project] = map[lib.TaskID]map[string][]*lib.TimeEntry{}
			}
			if _, ok := aggregatedTime[entry.Customer][entry.Project][entry.TaskID]; !ok {
				aggregatedTime[entry.Customer][entry.Project][entry.TaskID] = map[string][]*lib.TimeEntry{}
			}
			if _, ok := aggregatedTime[entry.Customer][entry.Project][entry.TaskID][entry.Description]; !ok {
				aggregatedTime[entry.Customer][entry.Project][entry.TaskID][entry.Description] = append(aggregatedTime[entry.Customer][entry.Project][entry.TaskID][entry.Description], entry)
			}

			fmt.Printf(" ENTRY  %v\n", entry)
		}
	}

	fmt.Printf("REPORT  %v\n", aggregatedTime)

	return nil
}
