package report

import (
	"fmt"
	"io"
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
func (r *Report) GenerateTypstReport(w io.Writer) error {
	// Header
	header := r.headerText()
	if _, err := fmt.Fprintf(w, "= %s\n\n", escapeTypst(header)); err != nil {
		return err
	}

	// User
	if _, err := fmt.Fprintf(w, "*User:* %s\n\n", escapeTypst(r.Config.Reporting.FullName)); err != nil {
		return err
	}

	// Customer address (if single customer report)
	if r.Customer != "" {
		if _, err := fmt.Fprint(w, "*Customer:*\n\n"); err != nil {
			return err
		}
		for _, line := range r.addressLines() {
			if _, err := fmt.Fprintf(w, "- %s\n", escapeTypst(line)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
	}

	// Grand Total
	if _, err := fmt.Fprintf(w, "*Grand Total:* \\$%.2f\n\n", r.grandTotal()); err != nil {
		return err
	}

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		if _, err := fmt.Fprint(w, "== Statistics\n\n"); err != nil {
			return err
		}
		stats := r.CalculateStatistics()

		if _, err := fmt.Fprint(w, "#table(\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "  columns: 2,\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "  stroke: 0.5pt,\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "  align: (left, right),\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "  table.header([*Metric*], [*Value*]),\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  [Days worked], [%d],\n", stats.DaysWorked); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  [Average time per day worked], [%s],\n", formatDuration(stats.AvgTimePerDay)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  [Average time per weekday worked], [%s],\n", formatDuration(stats.AvgTimePerWeekday)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  [Weeks\\* worked], [%.2f],\n", stats.WeeksWorked); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  [Average time per week\\* worked], [%s],\n", formatDuration(stats.AvgTimePerWeek)); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, ")\n\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "\\* a week is any set of five weekdays (not necessarily within the same calendar week)\n\n"); err != nil {
			return err
		}
	}

	// Detailed Time Report
	if _, err := fmt.Fprint(w, "== Detailed Time Report\n\n"); err != nil {
		return err
	}

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
	if _, err := fmt.Fprint(w, "#table(\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "  columns: 4,\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "  stroke: 0.5pt,\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "  align: (left, right, right, right),\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "  table.header([], [*Hours*], [*Rate (\\$\\/h)*], [*Total (\\$)*]),\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, strings.Join(rows, "\n")); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "\n)\n"); err != nil {
		return err
	}

	return nil
}
