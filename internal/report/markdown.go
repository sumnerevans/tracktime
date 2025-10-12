package report

import (
	"fmt"
	"html"
	"strings"
)

// GenerateMarkdownReport generates a markdown-formatted report
func (r *Report) GenerateMarkdownReport() string {
	var buf strings.Builder

	// Header
	header := r.headerText()
	buf.WriteString(fmt.Sprintf("# %s\n\n", header))

	// User
	buf.WriteString(fmt.Sprintf("**User:** %s\n\n", r.Config.Reporting.FullName))

	// Customer address (if single customer report)
	if r.Customer != "" {
		buf.WriteString("**Customer:**\n\n")
		for _, line := range r.addressLines() {
			buf.WriteString(fmt.Sprintf("- %s\n", line))
		}
		buf.WriteString("\n")
	}

	// Grand Total
	buf.WriteString(fmt.Sprintf("**Grand Total:** $%.2f\n\n", r.grandTotal()))

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		buf.WriteString("## Statistics\n\n")
		stats := r.CalculateStatistics()

		buf.WriteString("| Metric | Value |\n")
		buf.WriteString("|--------|-------|\n")
		buf.WriteString(fmt.Sprintf("| Days worked | %d |\n", stats.DaysWorked))
		buf.WriteString(fmt.Sprintf("| Average time per day worked | %s |\n", formatDuration(stats.AvgTimePerDay)))
		buf.WriteString(fmt.Sprintf("| Average time per weekday worked | %s |\n", formatDuration(stats.AvgTimePerWeekday)))
		buf.WriteString(fmt.Sprintf("| Weeks* worked | %.2f |\n", stats.WeeksWorked))
		buf.WriteString(fmt.Sprintf("| Average time per week* worked | %s |\n", formatDuration(stats.AvgTimePerWeek)))
		buf.WriteString("\n")
		buf.WriteString("\\* a week is any set of five weekdays (not necessarily within the same calendar week)\n\n")
	}

	// Detailed Time Report
	buf.WriteString("## Detailed Time Report\n\n")

	// Create table header
	buf.WriteString("|  | Hours | Rate ($/h) | Total ($) |\n")
	buf.WriteString("|--|------:|-----------:|----------:|\n")

	// TOTAL row
	buf.WriteString(fmt.Sprintf("| **TOTAL** | **%.2f** | | **$%.2f** |\n",
		r.totalMinutes()/60.0,
		r.grandTotal()))

	// Customer/Project rows
	for _, cp := range r.SortedCustomerProjects() {
		rt := r.RateTotals[cp]

		rate := ""
		total := ""
		if rt.Rate != 0.0 {
			rate = fmt.Sprintf("%.2f", rt.Rate)
			total = fmt.Sprintf("%.2f", rt.Total)
		}

		buf.WriteString(fmt.Sprintf("| **%s** | %.2f | %s | %s |\n",
			html.EscapeString(r.customerProjectStr(cp, false)),
			r.totalMinutesForCustomerProject(cp)/60.0,
			rate,
			total))

		if !r.TaskGrain {
			continue
		}

		// Task level
		for _, taskID := range r.SortedTaskIDs(cp) {
			taskName := r.formatTaskName(cp, taskID)
			// Use &nbsp; for indentation in markdown tables
			buf.WriteString(fmt.Sprintf("| &nbsp;&nbsp;• %s | %.2f | | |\n",
				html.EscapeString(taskName),
				r.TotalMinutesForTask(cp, taskID)/60.0))

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
				displayDesc := desc
				if displayDesc == "" {
					displayDesc = "(no description)"
				}
				// Use more &nbsp; for deeper indentation
				buf.WriteString(fmt.Sprintf("| &nbsp;&nbsp;&nbsp;&nbsp;◦ %s | %.2f | | |\n",
					html.EscapeString(displayDesc),
					r.TotalMinutesForDescription(cp, taskID, desc)/60.0))
			}
		}
	}

	return buf.String()
}
