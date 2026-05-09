package report

import (
	"fmt"
	"html"
	"io"

	"go.mau.fi/util/exerrors"
)

// GenerateMarkdownReport generates a markdown-formatted report
func (r *Report) GenerateMarkdownReport(w io.Writer) {
	// Header
	header := r.headerText()
	exerrors.Must(fmt.Fprintf(w, "# %s\n\n", header))

	// User
	exerrors.Must(fmt.Fprintf(w, "**User:** %s\n\n", r.Config.Reporting.FullName))

	// Customer address (if single customer report)
	if r.Customer != "" {
		exerrors.Must(fmt.Fprint(w, "**Customer:**\n\n"))
		for _, line := range r.addressLines() {
			exerrors.Must(fmt.Fprintf(w, "- %s\n", line))
		}
		exerrors.Must(fmt.Fprint(w, "\n"))
	}

	// Grand Total
	exerrors.Must(fmt.Fprintf(w, "**Grand Total:** $%.2f\n\n", r.grandTotal()))

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		exerrors.Must(fmt.Fprint(w, "## Statistics\n\n"))
		stats := r.CalculateStatistics()

		exerrors.Must(fmt.Fprint(w, "| Metric | Value |\n"))
		exerrors.Must(fmt.Fprint(w, "|--------|-------|\n"))
		exerrors.Must(fmt.Fprintf(w, "| Days worked | %d |\n", stats.DaysWorked))
		exerrors.Must(fmt.Fprintf(w, "| Average time per day worked | %s |\n", formatDuration(stats.AvgTimePerDay)))
		exerrors.Must(fmt.Fprintf(w, "| Average time per weekday worked | %s |\n", formatDuration(stats.AvgTimePerWeekday)))
		exerrors.Must(fmt.Fprintf(w, "| Weeks* worked | %.2f |\n", stats.WeeksWorked))
		exerrors.Must(fmt.Fprintf(w, "| Average time per week* worked | %s |\n", formatDuration(stats.AvgTimePerWeek)))
		exerrors.Must(fmt.Fprint(w, "\n"))
		exerrors.Must(fmt.Fprint(w, "\\* a week is any set of five weekdays (not necessarily within the same calendar week)\n\n"))
	}

	// Detailed Time Report
	exerrors.Must(fmt.Fprint(w, "## Detailed Time Report\n\n"))

	// Create table header
	exerrors.Must(fmt.Fprint(w, "|  | Hours | Rate ($/h) | Total ($) |\n"))
	exerrors.Must(fmt.Fprint(w, "|--|------:|-----------:|----------:|\n"))

	// TOTAL row
	exerrors.Must(fmt.Fprintf(w, "| **TOTAL** | **%.2f** | | **$%.2f** |\n",
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

		exerrors.Must(fmt.Fprintf(w, "| **%s** | %.2f | %s | %s |\n",
			html.EscapeString(r.customerProjectStr(cp)),
			r.totalMinutesForCustomerProject(cp)/60.0,
			rate,
			total))

		if !r.TaskGrain {
			continue
		}

		// Task level
		for _, taskID := range r.sortedTaskIDs(cp) {
			taskName := r.formatTaskName(cp, taskID)
			if link := r.getTaskLink(cp, taskID); link != "" {
				taskName = fmt.Sprintf("[%s](%s)", html.EscapeString(taskName), link)
			} else {
				taskName = html.EscapeString(taskName)
			}
			// Use &nbsp; for indentation in markdown tables
			exerrors.Must(fmt.Fprintf(w, "| &nbsp;&nbsp;• %s | %.2f | | |\n",
				taskName,
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
				// Use more &nbsp; for deeper indentation
				exerrors.Must(fmt.Fprintf(w, "| &nbsp;&nbsp;&nbsp;&nbsp;◦ %s | %.2f | | |\n",
					html.EscapeString(displayDesc),
					r.totalMinutesForDescription(cp, taskID, desc)/60.0))
			}
		}
	}
}
