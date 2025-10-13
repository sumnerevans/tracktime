package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/report"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
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
	Start *types.Date `arg:"positional" help:"specify the start of the reporting range (defaults to the beginning of last month)"`
	End   *types.Date `arg:"positional" help:"specify the end of the reporting range (defaults to the end of last month)"`

	// Specify the range using shorthand
	Month     types.Month `arg:"-m,--month" help:"shorthand for reporting over an entire month (can be combined with --year, accepted formats: 01, 1, Jan, January, 2019-01)"`
	Year      int         `arg:"-y,--year" help:"shorthand for reporting over an entire year (can be combined with --month)"`
	Today     bool        `arg:"--today" help:"shorthand for reporting on today"`
	Yesterday bool        `arg:"--yesterday" help:"shorthand for reporting on yesterday"`
	ThisWeek  bool        `arg:"--thisweek" help:"shorthand for reporting on the current week (Sunday-today)"`
	LastWeek  bool        `arg:"--lastweek" help:"shorthand for reporting on last week (Sunday-Saturday)"`
	ThisMonth bool        `arg:"--thismonth" help:"shorthand for reporting on the current month"`
	LastMonth bool        `arg:"--lastmonth" help:"shorthand for reporting on last month"`
	ThisYear  bool        `arg:"--thisyear" help:"shorthand for reporting on the current year"`
	LastYear  bool        `arg:"--lastyear" help:"shorthand for reporting on last year"`

	// Specify the grains to show
	TaskGrain          bool `arg:"--taskgrain" help:"report on the task grain"`
	NoTaskGrain        bool `arg:"--no-taskgrain" help:"do not report on the task grain"`
	DescriptionGrain   bool `arg:"--descriptiongrain" help:"report on the task grain"`
	NoDescriptionGrain bool `arg:"--no-descriptiongrain" help:"do not report on the task grain"`

	// Narrow the set of time entries to report on
	Customer timeentry.Customer `arg:"-c,--customer" help:"customer ID to generate a report for"`
	Project  timeentry.Project  `arg:"-p,--project" help:"project name to generate a report for"`

	// How to sort the report
	Sort Sort `arg:"-s,--sort" help:"the grain to sort the report by (alphabetical,alpha,a or time-spent,time,t)" default:"time-spent"`
	Desc bool `arg:"--desc" help:"sort descending"`
	Asc  bool `arg:"--asc" help:"sort ascending"`

	// Output file
	OutputFile types.Filename `arg:"-o,--outfile" help:"specify the filename to export the report to (supports .md, .html, .typ, .pdf files, if set to '-' then the report is printed to stdout)" default:"-"`
}

func (r *Report) Run(config *config.Config) error {
	var start, end types.Date
	today := types.Today()

	// Handle positional range arguments first
	if r.Start != nil && r.End != nil {
		start = *r.Start
		end = *r.End
	} else if r.Year != 0 || r.ThisYear || r.LastYear {
		// Yearly range
		year := today.Year()
		if r.Year != 0 {
			year = r.Year
		} else if r.LastYear {
			year--
		}
		start = types.NewDate(year, 1, 1)
		end = types.NewDate(year, 12, 31)

		// If month is also specified, narrow to that month
		if !r.Month.IsZero() {
			monthWithYear := r.Month
			if r.Month.Year() == 0 {
				// Month was specified without year, use the year from above
				monthWithYear = types.NewMonth(year, r.Month.Month())
			}
			if monthWithYear.Year() != year {
				return fmt.Errorf("when specifying a year, the month must be in the same year")
			}
			start = types.NewDate(monthWithYear.Year(), int(monthWithYear.Month()), 1)
			end = types.NewDate(monthWithYear.Year(), int(monthWithYear.Month()), monthWithYear.DaysInMonth())
		}
	} else if r.Today {
		start = today
		end = today
	} else if r.Yesterday {
		start = today.AddDays(-1)
		end = today.AddDays(-1)
	} else if r.ThisWeek {
		start = today.AddDays(-int(today.Weekday()))
		end = today.AddDays(6 - int(today.Weekday()))
	} else if r.LastWeek {
		start = today.AddDays(-int(today.Weekday()) - 7)
		end = today.AddDays(6 - int(today.Weekday()) - 7)
	} else {
		// Monthly (default)
		// Default to last month
		lastMonth := today.AddMonths(-1)
		start = types.NewDate(lastMonth.Year(), int(lastMonth.Month()), 1)

		if !r.Month.IsZero() {
			monthWithYear := r.Month
			if r.Month.Year() == 0 {
				// Month was specified without year, use current year
				monthWithYear = types.NewMonth(today.Year(), r.Month.Month())
			}
			start = types.NewDate(monthWithYear.Year(), int(monthWithYear.Month()), 1)
		} else if r.ThisMonth {
			start = types.NewDate(today.Year(), int(today.Month()), 1)
		}

		end = types.NewDate(start.Year(), int(start.Month()), start.DaysInMonth())
	}

	// Allow positional arguments to override
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

	// Determine sort direction
	// Reverse=false gives: alphabetical A-Z, time-spent largest-first (both are defaults)
	// User can override with --desc (Reverse=true) or --asc (Reverse=false)
	reverse := r.Desc

	// Determine grains with defaults based on date range (match Python logic)
	dateDiff := int(end.Sub(start.Time).Hours() / 24)
	taskGrain := dateDiff <= 31       // Default: enabled for ranges <= 31 days
	descriptionGrain := dateDiff <= 7 // Default: enabled for ranges <= 7 days

	// Override with explicit flags
	if r.TaskGrain {
		taskGrain = true
	}
	if r.DescriptionGrain {
		descriptionGrain = true
		taskGrain = true // Description grain requires task grain
	}
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

	// Determine output format based on file extension
	outputPath := string(r.OutputFile)

	if outputPath == "-" {
		fmt.Println(rep.GenerateTextReport())
		return nil
	}

	// Helper to write reports that use io.Writer
	writeReport := func(generator func(io.Writer) error, formatName string) error {
		expandedPath := r.OutputFile.Expand()
		file, err := os.Create(expandedPath)
		if err != nil {
			return fmt.Errorf("failed to create %s report: %w", formatName, err)
		}
		defer file.Close()

		if err := generator(file); err != nil {
			return fmt.Errorf("failed to write %s report: %w", formatName, err)
		}
		fmt.Printf("%s report exported to %s\n", formatName, expandedPath)
		return nil
	}

	lowerPath := strings.ToLower(outputPath)
	switch {
	case strings.HasSuffix(lowerPath, ".md"):
		return writeReport(rep.GenerateMarkdownReport, "Markdown")
	case strings.HasSuffix(lowerPath, ".html"):
		return writeReport(rep.GenerateHTMLReport, "HTML")
	case strings.HasSuffix(lowerPath, ".typ"):
		return writeReport(rep.GenerateTypstReport, "Typst")
	case strings.HasSuffix(lowerPath, ".pdf"):
		expandedPath := r.OutputFile.Expand()
		if err := rep.GeneratePDFReport(expandedPath); err != nil {
			return fmt.Errorf("failed to generate PDF report: %w", err)
		}
		fmt.Printf("PDF report exported to %s\n", expandedPath)
		return nil
	default:
		return fmt.Errorf("unsupported output format for file %s (supported: .md, .html, .typ, .pdf, or use '-' for stdout)", outputPath)
	}
}
