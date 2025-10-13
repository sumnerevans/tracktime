package report

import (
	"fmt"
	"strings"
)

// escapeTypst escapes special characters for Typst
func escapeTypst(s string) string {
	// Typst special characters that need escaping
	replacer := strings.NewReplacer(
		"#", "\\#",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"@", "\\@",
		"`", "\\`",
		"<", "\\<",
		">", "\\>",
		"$", "\\$",
		"/", "\\/",
	)
	return replacer.Replace(s)
}

// GenerateTypstReport generates a Typst-formatted report
func (r *Report) GenerateTypstReport() string {
	var buf strings.Builder

	// Header
	header := r.headerText()
	buf.WriteString(fmt.Sprintf("= %s\n\n", escapeTypst(header)))

	// User
	buf.WriteString(fmt.Sprintf("*User:* %s\n\n", escapeTypst(r.Config.Reporting.FullName)))

	// Customer address (if single customer report)
	if r.Customer != "" {
		buf.WriteString("*Customer:*\n\n")
		for _, line := range r.addressLines() {
			buf.WriteString(fmt.Sprintf("- %s\n", escapeTypst(line)))
		}
		buf.WriteString("\n")
	}

	// Grand Total
	buf.WriteString(fmt.Sprintf("*Grand Total:* \\$%.2f\n\n", r.grandTotal()))

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		buf.WriteString("== Statistics\n\n")
		stats := r.CalculateStatistics()

		buf.WriteString("#table(\n")
		buf.WriteString("  columns: 2,\n")
		buf.WriteString("  stroke: 0.5pt,\n")
		buf.WriteString("  align: (left, right),\n")
		buf.WriteString("  table.header([*Metric*], [*Value*]),\n")
		buf.WriteString(fmt.Sprintf("  [Days worked], [%d],\n", stats.DaysWorked))
		buf.WriteString(fmt.Sprintf("  [Average time per day worked], [%s],\n", formatDuration(stats.AvgTimePerDay)))
		buf.WriteString(fmt.Sprintf("  [Average time per weekday worked], [%s],\n", formatDuration(stats.AvgTimePerWeekday)))
		buf.WriteString(fmt.Sprintf("  [Weeks\\* worked], [%.2f],\n", stats.WeeksWorked))
		buf.WriteString(fmt.Sprintf("  [Average time per week\\* worked], [%s],\n", formatDuration(stats.AvgTimePerWeek)))
		buf.WriteString(")\n\n")
		buf.WriteString("\\* a week is any set of five weekdays (not necessarily within the same calendar week)\n\n")
	}

	// Detailed Time Report
	buf.WriteString("== Detailed Time Report\n\n")

	// Build table rows
	var rows []string

	// TOTAL row
	rows = append(rows, fmt.Sprintf("  [*TOTAL*], [*%.2f*], [], [*\\$%.2f*],",
		r.totalMinutes()/60.0,
		r.grandTotal()))

	// Customer/Project rows
	for _, cp := range r.sortedCustomerProjects() {
		rt := r.RateTotals[cp]

		rate := ""
		total := ""
		if rt.Rate != 0.0 {
			rate = fmt.Sprintf("%.2f", rt.Rate)
			total = fmt.Sprintf("%.2f", rt.Total)
		}

		rows = append(rows, fmt.Sprintf("  [*%s*], [%.2f], [%s], [%s],",
			escapeTypst(r.customerProjectStr(cp, false)),
			r.totalMinutesForCustomerProject(cp)/60.0,
			rate,
			total))

		if !r.TaskGrain {
			continue
		}

		// Task level
		for _, taskID := range r.sortedTaskIDs(cp) {
			taskName := r.formatTaskName(cp, taskID)
			rows = append(rows, fmt.Sprintf("  [#h(1em)• %s], [%.2f], [], [],",
				escapeTypst(taskName),
				r.totalMinutesForTask(cp, taskID)/60.0))

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
			for _, desc := range r.sortedDescriptions(cp, taskID) {
				displayDesc := desc
				if displayDesc == "" {
					displayDesc = "(no description)"
				}
				rows = append(rows, fmt.Sprintf("  [#h(2em)◦ %s], [%.2f], [], [],",
					escapeTypst(displayDesc),
					r.totalMinutesForDescription(cp, taskID, desc)/60.0))
			}
		}
	}

	// Create table
	buf.WriteString("#table(\n")
	buf.WriteString("  columns: 4,\n")
	buf.WriteString("  stroke: 0.5pt,\n")
	buf.WriteString("  align: (left, right, right, right),\n")
	buf.WriteString("  table.header([], [*Hours*], [*Rate (\\$\\/h)*], [*Total (\\$)*]),\n")
	buf.WriteString(strings.Join(rows, "\n"))
	buf.WriteString("\n)\n")

	return buf.String()
}
