package report

import (
	"fmt"
	"io"
	"strings"

	"go.mau.fi/util/exerrors"
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
func (r *Report) GenerateTypstReport(w io.Writer) {
	// Header
	header := r.headerText()
	exerrors.Must(fmt.Fprintf(w, "= %s\n\n", escapeTypst(header)))

	// User
	exerrors.Must(fmt.Fprintf(w, "*User:* %s\n\n", escapeTypst(r.Config.Reporting.FullName)))

	// Customer address (if single customer report)
	if r.Customer != "" {
		exerrors.Must(fmt.Fprint(w, "*Customer:*\n\n"))
		for _, line := range r.addressLines() {
			exerrors.Must(fmt.Fprintf(w, "- %s\n", escapeTypst(line)))
		}
		exerrors.Must(fmt.Fprint(w, "\n"))
	}

	// Grand Total
	exerrors.Must(fmt.Fprintf(w, "*Grand Total:* \\$%s\n\n", formatFloat(r.grandTotal())))

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		exerrors.Must(fmt.Fprint(w, "== Statistics\n\n"))
		stats := r.CalculateStatistics()

		exerrors.Must(fmt.Fprint(w, "#table(\n"))
		exerrors.Must(fmt.Fprint(w, "  columns: 2,\n"))
		exerrors.Must(fmt.Fprint(w, "  stroke: 0.5pt,\n"))
		exerrors.Must(fmt.Fprint(w, "  align: (left, right),\n"))
		exerrors.Must(fmt.Fprint(w, "  table.header([*Metric*], [*Value*]),\n"))
		exerrors.Must(fmt.Fprintf(w, "  [Days worked], [%d],\n", stats.DaysWorked))
		exerrors.Must(fmt.Fprintf(w, "  [Average time per day worked], [%s],\n", formatDuration(stats.AvgTimePerDay)))
		exerrors.Must(fmt.Fprintf(w, "  [Average time per weekday worked], [%s],\n", formatDuration(stats.AvgTimePerWeekday)))
		exerrors.Must(fmt.Fprintf(w, "  [Weeks\\* worked], [%s],\n", formatFloat(stats.WeeksWorked)))
		exerrors.Must(fmt.Fprintf(w, "  [Average time per week\\* worked], [%s],\n", formatDuration(stats.AvgTimePerWeek)))
		exerrors.Must(fmt.Fprint(w, ")\n\n"))
		exerrors.Must(fmt.Fprint(w, "\\* a week is any set of five weekdays (not necessarily within the same calendar week)\n\n"))
	}

	// Detailed Time Report
	exerrors.Must(fmt.Fprint(w, "== Detailed Time Report\n\n"))

	// Build table rows
	var rows []string

	// TOTAL row
	rows = append(rows, fmt.Sprintf("  [*TOTAL*], [*%s*], [], [*\\$%s*],",
		formatFloat(r.totalMinutes()/60.0),
		formatFloat(r.grandTotal())))

	// Customer/Project rows
	for _, cp := range r.sortedCustomerProjects() {
		rt := r.RateTotals[cp]

		rate := ""
		total := ""
		if rt.Rate != 0.0 {
			rate = "\\$" + formatFloat(rt.Rate)
			total = "\\$" + formatFloat(rt.Total)
		}

		rows = append(rows, fmt.Sprintf("  [*%s*], [%s], [%s], [%s],",
			escapeTypst(r.customerProjectStr(cp)),
			formatFloat(r.totalMinutesForCustomerProject(cp)/60.0),
			rate,
			total))

		if !r.TaskGrain {
			continue
		}

		// Task level
		for _, taskID := range r.sortedTaskIDs(cp) {
			taskName := escapeTypst(r.formatTaskName(cp, taskID))
			if link := r.getTaskLink(cp, taskID); link != "" {
				taskName = fmt.Sprintf("#link(%q)[%s]", link, taskName)
			}
			rows = append(rows, fmt.Sprintf("  [#h(1em)• %s], [%s], [], [],",
				taskName,
				formatFloat(r.totalMinutesForTask(cp, taskID)/60.0)))

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
				rows = append(rows, fmt.Sprintf("  [#h(2em)◦ %s], [%s], [], [],",
					escapeTypst(displayDesc),
					formatFloat(r.totalMinutesForDescription(cp, taskID, desc)/60.0)))
			}
		}
	}

	// Create table
	exerrors.Must(fmt.Fprint(w, "#table(\n"))
	exerrors.Must(fmt.Fprint(w, "  columns: 4,\n"))
	exerrors.Must(fmt.Fprint(w, "  stroke: 0.5pt,\n"))
	exerrors.Must(fmt.Fprint(w, "  align: (left, right, right, right),\n"))
	exerrors.Must(fmt.Fprint(w, "  table.header([], [*Hours*], [*Rate (\\$\\/h)*], [*Total (\\$)*]),\n"))
	exerrors.Must(fmt.Fprint(w, strings.Join(rows, "\n")))
	exerrors.Must(fmt.Fprint(w, "\n)\n"))
}
