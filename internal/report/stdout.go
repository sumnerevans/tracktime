package report

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ansiRegex matches ANSI escape codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

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
	header := r.headerText()
	buf.WriteString(boldUnderline.Sprint(header))
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
	fmt.Fprintf(&buf, "%s %s\n", bold.Sprint("Grand Total:"), boldCyan.Sprintf("$%.2f", r.grandTotal()))
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
			AddRow("    Weeks* worked:", cyan.Sprintf("%.2f", stats.WeeksWorked)).
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

	// Create table
	reportTable := table.New("", "Hours", "Rate ($/h)", "Total ($)").
		WithWidthFunc(visualWidth).
		WithPadding(3).
		WithHeaderFormatter(greenUnderline.SprintfFunc())

	// TOTAL row
	reportTable.AddRow(
		boldYellow.Sprint(ellipsize("TOTAL", 40)),
		bold.Sprintf("%.2f", r.totalMinutes()/60.0),
		"",
		bold.Sprint(cyan.Sprintf("$%.2f", r.grandTotal())),
	)

	// Customer/Project rows
	for _, cp := range r.SortedCustomerProjects() {
		// Customer/project summary row
		rt := r.RateTotals[cp]

		rate := fmt.Sprintf("%.2f", rt.Rate)
		total := fmt.Sprintf("%.2f", rt.Total)

		// Hide zero rates
		if rt.Rate == 0.0 {
			rate = ""
			total = ""
		}

		reportTable.AddRow(
			boldYellow.Sprint(ellipsize(r.customerProjectStr(cp, false), 40)),
			fmt.Sprintf("%.2f", r.totalMinutesForCustomerProject(cp)/60.0),
			rate,
			total,
		)

		if !r.TaskGrain {
			continue
		}

		// Task level
		for _, taskID := range r.SortedTaskIDs(cp) {
			taskName := " * " + r.formatTaskName(cp, taskID)
			reportTable.AddRow(ellipsize(taskName, 40), fmt.Sprintf("%.2f", r.TotalMinutesForTask(cp, taskID)/60.0), "", "")

			if !r.DescriptionGrain {
				continue
			}

			// Skip if only one empty description
			descriptions := r.AggregatedTime[cp][taskID]
			if len(descriptions) == 1 {
				if _, hasEmpty := descriptions[""]; hasEmpty {
					continue
				}
			}

			// Description level
			for _, desc := range r.SortedDescriptions(cp, taskID) {
				if desc == "" {
					desc = "<NO DESCRIPTION>"
				}
				descName := "    * " + desc
				reportTable.AddRow(ellipsize(descName, 40), fmt.Sprintf("%.2f", r.TotalMinutesForDescription(cp, taskID, desc)/60.0), "", "")
			}
		}
	}

	reportTable.WithWriter(&buf).Print()
	buf.WriteString("\n")
	return buf.String()
}
