package report

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ansiRegex matches ANSI CSI sequences (colors) and OSC sequences (e.g. hyperlinks)
var ansiRegex = regexp.MustCompile(`\x1b(?:\[[0-9;]*m|\][^\x1b]*(?:\x1b\\|\x07))`)

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// visualWidth returns the visual width of a string (without ANSI codes)
func visualWidth(s string) int {
	return len([]rune(stripAnsi(s)))
}

// ellipsize truncates a string to maxLen characters, adding "..." if truncated
func ellipsize(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

type reportRow struct {
	label string
	hours string
	rate  string
	total string
}

// rightAlign prepends spaces to s so its visual width equals width.
// Empty strings are returned as-is (hidden cells).
func rightAlign(s string, width int) string {
	if s == "" {
		return ""
	}
	if pad := width - visualWidth(s); pad > 0 {
		return strings.Repeat(" ", pad) + s
	}
	return s
}

// GenerateTextReport generates a plain text report matching Python's output format
func (r *Report) GenerateTextReport() string {
	var buf strings.Builder

	// Force color output (important for when output is piped)
	color.NoColor = false

	// Define colors
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	boldYellow := color.New(color.Bold, color.FgYellow)
	greenUnderline := color.New(color.FgGreen, color.Underline)
	boldCyan := color.New(color.Bold, color.FgCyan)
	boldUnderline := color.New(color.Bold, color.Underline)

	// Header
	buf.WriteString(boldUnderline.Sprint(r.headerText()))
	buf.WriteString("\n\n")

	// User
	fmt.Fprintf(&buf, "%s %s\n", bold.Sprint("User:"), r.Config.Reporting.FullName)
	buf.WriteString("\n")

	// Customer address (if single customer report)
	if r.Customer != "" {
		buf.WriteString(bold.Sprint("Customer:"))
		buf.WriteString("\n\n")
		for _, line := range r.addressLines() {
			fmt.Fprintf(&buf, "    %s\n", line)
		}
		buf.WriteString("\n")
	}

	// Grand Total
	fmt.Fprintf(&buf, "%s %s\n", bold.Sprint("Grand Total:"), boldCyan.Sprintf("$%s", formatFloat(r.grandTotal())))
	buf.WriteString("\n")

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		buf.WriteString(bold.Sprint("Statistics:"))
		buf.WriteString("\n\n")
		stats := r.CalculateStatistics()

		table.New("", "").
			WithHeaderFormatter(func(_ string, _ ...any) string { return "" }).
			WithWidthFunc(visualWidth).
			AddRow("    Days worked:", cyan.Sprint(stats.DaysWorked)).
			AddRow("    Average time per day worked:", cyan.Sprint(formatDuration(stats.AvgTimePerDay))).
			AddRow("    Average time per weekday worked:", cyan.Sprint(formatDuration(stats.AvgTimePerWeekday))).
			AddRow("    Weeks* worked:", cyan.Sprint(formatFloat(stats.WeeksWorked))).
			AddRow("    Average time per week* worked:", cyan.Sprint(formatDuration(stats.AvgTimePerWeek))).
			WithWriter(&buf).
			Print()

		buf.WriteString("\n\n")
		buf.WriteString("* a week is any set of five weekdays (not necessarily within the same calendar week)\n")
		buf.WriteString("\n")
	}

	// Detailed Time Report
	buf.WriteString(bold.Sprint("Detailed Time Report:"))
	buf.WriteString("\n\n")

	// Collect all rows before adding to the table so we can right-align numeric columns.
	var rows []reportRow

	rows = append(rows, reportRow{
		label: boldYellow.Sprint(ellipsize("TOTAL", 40)),
		hours: bold.Sprint(formatFloat(r.totalMinutes() / 60.0)),
		total: bold.Sprint(cyan.Sprintf("$%s", formatFloat(r.grandTotal()))),
	})

	for _, cp := range r.sortedCustomerProjects() {
		rt := r.RateTotals[cp]
		rateStr := "$" + formatFloat(rt.Rate)
		totalStr := "$" + formatFloat(rt.Total)
		if rt.Rate == 0.0 {
			rateStr = ""
			totalStr = ""
		}
		rows = append(rows, reportRow{
			label: boldYellow.Sprint(ellipsize(r.customerProjectStr(cp), 40)),
			hours: formatFloat(r.totalMinutesForCustomerProject(cp) / 60.0),
			rate:  rateStr,
			total: totalStr,
		})

		if !r.TaskGrain {
			continue
		}

		for _, taskID := range r.sortedTaskIDs(cp) {
			name := ellipsize(r.formatTaskName(cp, taskID), 37)
			if link := r.getTaskLink(cp, taskID); link != "" {
				name = fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", link, name)
			}
			rows = append(rows, reportRow{
				label: " * " + name,
				hours: formatFloat(r.totalMinutesForTask(cp, taskID) / 60.0),
			})

			if !r.DescriptionGrain {
				continue
			}

			descriptions := r.AggregatedTime[cp][taskID]
			if len(descriptions) == 1 {
				if _, hasEmpty := descriptions[""]; hasEmpty {
					continue
				}
			}

			for _, desc := range r.sortedDescriptions(cp, taskID) {
				if desc == "" {
					desc = "<NO DESCRIPTION>"
				}
				rows = append(rows, reportRow{
					label: ellipsize("    * "+desc, 40),
					hours: formatFloat(r.totalMinutesForDescription(cp, taskID, desc) / 60.0),
				})
			}
		}
	}

	// Compute max visual widths for numeric columns (seed with header widths).
	maxHours := len("Hours")
	maxRate := len("Rate ($/h)")
	maxTotal := len("Total ($)")
	for _, row := range rows {
		if w := visualWidth(row.hours); w > maxHours {
			maxHours = w
		}
		if w := visualWidth(row.rate); w > maxRate {
			maxRate = w
		}
		if w := visualWidth(row.total); w > maxTotal {
			maxTotal = w
		}
	}

	reportTable := table.New("", "Hours", "Rate ($/h)", "Total ($)").
		WithWidthFunc(visualWidth).
		WithPadding(3).
		WithHeaderFormatter(greenUnderline.SprintfFunc())

	for _, row := range rows {
		reportTable.AddRow(
			row.label,
			rightAlign(row.hours, maxHours),
			rightAlign(row.rate, maxRate),
			rightAlign(row.total, maxTotal),
		)
	}

	reportTable.WithWriter(&buf).Print()
	buf.WriteString("\n")
	return buf.String()
}
