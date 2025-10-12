package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/rodaine/table"

	"github.com/sumnerevans/tracktime/internal/timeentry"
)

// GenerateTextReport generates a plain text report matching Python's output format
func (r *Report) GenerateTextReport() string {
	var buf strings.Builder

	// Header
	header := r.headerText()
	buf.WriteString(header)
	buf.WriteString("\n")
	buf.WriteString(strings.Repeat("=", len(header)))
	buf.WriteString("\n\n")

	// User
	fmt.Fprintf(&buf, "**User:** %s\n", r.Config.Reporting.FullName)
	buf.WriteString("\n")

	// Customer address (if single customer report)
	if r.Customer != "" {
		buf.WriteString("**Customer:**\n")
		buf.WriteString("\n")
		addressLines := r.addressLines()
		for _, line := range addressLines {
			fmt.Fprintf(&buf, "    | %s\n", line)
		}
		buf.WriteString("\n")
	}

	// Grand Total
	fmt.Fprintf(&buf, "**Grand Total:** $%.2f\n", r.GrandTotal())
	buf.WriteString("\n")

	// Statistics (if enabled)
	if r.Config.Reporting.ReportStatistics {
		buf.WriteString("**Statistics:**\n")
		buf.WriteString("\n")
		stats := r.CalculateStatistics()

		table.New("", "").
			WithHeaderFormatter(func(_ string, _ ...any) string { return "" }).
			AddRow("Days worked:", stats.DaysWorked).
			AddRow("Average time per day worked:", formatDuration(stats.AvgTimePerDay)).
			AddRow("Average time per weekday worked:", formatDuration(stats.AvgTimePerWeekday)).
			AddRow("Weeks* worked:", fmt.Sprintf("%.2f", stats.WeeksWorked)).
			AddRow("Average time per week* worked:", formatDuration(stats.AvgTimePerWeek)).
			WithWriter(&buf).
			Print()

		buf.WriteString("\n\n")
		buf.WriteString("* a week is any set of five weekdays (not necessarily within the same calendar week)\n")
		buf.WriteString("\n")
	}

	// Detailed Time Report
	buf.WriteString("**Detailed Time Report:**\n")
	buf.WriteString("\n")

	// TOTAL row
	reportTable := table.New("", "Hours", "Rate ($/h)", "Total ($)").
		AddRow("TOTAL", fmt.Sprintf("%.2f", r.totalMinutes()/60.0), "", fmt.Sprintf("%.2f", r.GrandTotal()))

	// Customer/Project rows
	cps := r.SortedCustomerProjects()
	for _, cp := range cps {
		// Customer/project summary row
		cpMinutes := r.TotalMinutesForCustomerProject(cp)
		rt := r.RateTotals[cp]

		reportTable.AddRow(r.customerProjectStr(cp, false), fmt.Sprintf("%.2f", cpMinutes/60.0), fmt.Sprintf("%.2f", rt.Rate), fmt.Sprintf("%.2f", rt.Total))

		if !r.TaskGrain {
			continue
		}

		// Task level
		taskIDs := r.SortedTaskIDs(cp)
		for _, taskID := range taskIDs {
			taskMinutes := r.TotalMinutesForTask(cp, taskID)
			reportTable.AddRow(" * "+r.formatTaskName(cp, taskID), fmt.Sprintf("%.2f", taskMinutes/60.0), "", "")

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
			descs := r.SortedDescriptions(cp, taskID)
			for _, desc := range descs {
				if desc == "" {
					desc = "<NO DESCRIPTION>"
				}
				descMinutes := r.TotalMinutesForDescription(cp, taskID, desc)
				reportTable.AddRow("    * "+desc, fmt.Sprintf("%.2f", descMinutes/60.0), "", "")
			}
		}
	}

	reportTable.WithWriter(&buf).Print()
	buf.WriteString("\n")
	return buf.String()
}

// headerText returns the report header with smart date formatting
func (r *Report) headerText() string {
	// Check for whole year
	if r.StartDate.Year() == r.EndDate.Year() &&
		r.StartDate.Month() == time.January && r.StartDate.Day() == 1 &&
		r.EndDate.Month() == time.December && r.EndDate.Day() == 31 {
		return fmt.Sprintf("Time Report: %d", r.StartDate.Year())
	}

	// Check for single month
	if r.StartDate.Year() == r.EndDate.Year() && r.StartDate.Month() == r.EndDate.Month() {
		daysInMonth := time.Date(r.StartDate.Year(), r.StartDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if r.StartDate.Day() == 1 && r.EndDate.Day() == daysInMonth {
			return fmt.Sprintf("Time Report: %s", r.StartDate.Format("January 2006"))
		}

		// Check for single day
		if r.StartDate.Day() == r.EndDate.Day() {
			return fmt.Sprintf("Time Report: %s", r.StartDate.Format("2006-01-02"))
		}
	}

	// Default: show range
	return fmt.Sprintf("Time Report: %s - %s", r.StartDate.Format("2006-01-02"), r.EndDate.Format("2006-01-02"))
}

// customerProjectStr formats customer/project string
func (r *Report) customerProjectStr(cp CustomerProject, html bool) string {
	noProject := "<no project>"
	noCustomer := "<no customer>"
	noBoth := "<no project or customer>"

	if html {
		noProject = "<i>no project</i>"
		noCustomer = "<i>no customer</i>"
		noBoth = "<i>no project or customer</i>"
	}

	if r.Customer != "" {
		// Filtering by customer, show project
		if cp.Project != "" {
			return string(cp.Project)
		}
		return noProject
	}

	if r.Project != "" {
		// Filtering by project, show customer
		if cp.Customer != "" {
			return string(cp.Customer)
		}
		return noCustomer
	}

	// Show both
	if cp.Customer == "" && cp.Project == "" {
		return noBoth
	}
	if cp.Customer != "" && cp.Project != "" {
		return fmt.Sprintf("%s: %s", cp.Customer, cp.Project)
	}
	if cp.Customer != "" {
		return string(cp.Customer)
	}
	return string(cp.Project)
}

// addressLines returns customer address lines
func (r *Report) addressLines() []string {
	var lines []string

	// Add alias
	alias := string(r.Customer)
	if r.Config.Reporting.CustomerAliases != nil {
		if a, ok := r.Config.Reporting.CustomerAliases[string(r.Customer)]; ok {
			alias = a
		}
	}
	lines = append(lines, alias)

	// Add address
	if r.Config.Reporting.CustomerAddresses != nil {
		if addr, ok := r.Config.Reporting.CustomerAddresses[string(r.Customer)]; ok {
			addrLines := strings.Split(strings.TrimSpace(addr), "\n")
			lines = append(lines, addrLines...)
		}
	}

	return lines
}

// formatTaskName formats a task name with ID and description
func (r *Report) formatTaskName(cp CustomerProject, taskID timeentry.TaskID) string {
	// Get first entry to check type
	var firstEntry *timeentry.TimeEntry
	for _, entries := range r.AggregatedTime[cp][taskID] {
		if len(entries) > 0 {
			firstEntry = entries[0]
			break
		}
	}

	if firstEntry == nil {
		return "<NO TASK>"
	}

	taskName := "<NO TASK>"
	if taskID != "" {
		// Task IDs already include prefixes in the CSV data (#123, !456, etc.)
		// so just use them directly
		taskName = string(taskID)
	}

	// TODO: Add synchronizer task description lookup
	// For now, just return the formatted task ID
	return taskName
}

// totalMinutes returns total minutes across all entries
func (r *Report) totalMinutes() float64 {
	var total float64
	for _, duration := range r.DayStats {
		total += duration.Minutes()
	}
	return total
}
