package report

import (
	"fmt"
	"html"
	"io"
)

// GenerateMarkdownReport generates a markdown-formatted report
func (r *Report) GenerateMarkdownReport(w io.Writer) error {
	// Header
	header := r.headerText()
	if _, err := fmt.Fprintf(w, "# %s\n\n", header); err != nil {
		return err
	}

	// User
	if _, err := fmt.Fprintf(w, "**User:** %s\n\n", r.Config.Reporting.FullName); err != nil {
		return err
	}

	// Customer address (if single customer report)
	if r.Customer != "" {
		if _, err := fmt.Fprint(w, "**Customer:**\n\n"); err != nil {
			return err
		}
		for _, line := range r.addressLines() {
			if _, err := fmt.Fprintf(w, "- %s\n", line); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
	}

	// Grand Total
	if _, err := fmt.Fprintf(w, "**Grand Total:** $%.2f\n\n", r.grandTotal()); err != nil {
		return err
	}

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		if _, err := fmt.Fprint(w, "## Statistics\n\n"); err != nil {
			return err
		}
		stats := r.CalculateStatistics()

		if _, err := fmt.Fprint(w, "| Metric | Value |\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "|--------|-------|\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "| Days worked | %d |\n", stats.DaysWorked); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "| Average time per day worked | %s |\n", formatDuration(stats.AvgTimePerDay)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "| Average time per weekday worked | %s |\n", formatDuration(stats.AvgTimePerWeekday)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "| Weeks* worked | %.2f |\n", stats.WeeksWorked); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "| Average time per week* worked | %s |\n", formatDuration(stats.AvgTimePerWeek)); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "\\* a week is any set of five weekdays (not necessarily within the same calendar week)\n\n"); err != nil {
			return err
		}
	}

	// Detailed Time Report
	if _, err := fmt.Fprint(w, "## Detailed Time Report\n\n"); err != nil {
		return err
	}

	// Create table header
	if _, err := fmt.Fprint(w, "|  | Hours | Rate ($/h) | Total ($) |\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "|--|------:|-----------:|----------:|\n"); err != nil {
		return err
	}

	// TOTAL row
	if _, err := fmt.Fprintf(w, "| **TOTAL** | **%.2f** | | **$%.2f** |\n",
		r.totalMinutes()/60.0,
		r.grandTotal()); err != nil {
		return err
	}

	// Customer/Project rows
	for _, cp := range r.sortedCustomerProjects() {
		rt := r.RateTotals[cp]

		rate := ""
		total := ""
		if rt.Rate != 0.0 {
			rate = fmt.Sprintf("%.2f", rt.Rate)
			total = fmt.Sprintf("%.2f", rt.Total)
		}

		if _, err := fmt.Fprintf(w, "| **%s** | %.2f | %s | %s |\n",
			html.EscapeString(r.customerProjectStr(cp, false)),
			r.totalMinutesForCustomerProject(cp)/60.0,
			rate,
			total); err != nil {
			return err
		}

		if !r.TaskGrain {
			continue
		}

		// Task level
		for _, taskID := range r.sortedTaskIDs(cp) {
			taskName := r.formatTaskName(cp, taskID)
			// Use &nbsp; for indentation in markdown tables
			if _, err := fmt.Fprintf(w, "| &nbsp;&nbsp;• %s | %.2f | | |\n",
				html.EscapeString(taskName),
				r.totalMinutesForTask(cp, taskID)/60.0); err != nil {
				return err
			}

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
				if _, err := fmt.Fprintf(w, "| &nbsp;&nbsp;&nbsp;&nbsp;◦ %s | %.2f | | |\n",
					html.EscapeString(displayDesc),
					r.totalMinutesForDescription(cp, taskID, desc)/60.0); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
