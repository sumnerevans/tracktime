package report

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/rodaine/table"

	"github.com/sumnerevans/tracktime/internal/lib"
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
		statsMap := stats.StatisticsMap()

		// Find max length for alignment
		maxLen := 0
		for key := range statsMap {
			if len(key) > maxLen {
				maxLen = len(key)
			}
		}

		// Output in specific order to match Python
		order := []string{
			"Days worked",
			"Average time per day worked",
			"Average time per weekday worked",
			"Weeks* worked",
			"Average time per week* worked",
		}
		for _, key := range order {
			value := statsMap[key]
			fmt.Fprintf(&buf, "    | %s %s\n", padRight(key+":", maxLen+2), value)
		}
		buf.WriteString("\n")
		buf.WriteString("* a week is any set of five weekdays (not necessarily within the same calendar week)\n")
		buf.WriteString("\n")
	}

	// Detailed Time Report
	buf.WriteString("**Detailed Time Report:**\n")
	buf.WriteString("\n")

	// TOTAL row
	reportTable := table.New("", "Hours", "Rate ($/h)", "Total ($)")
	reportTable.AddRow("TOTAL", fmt.Sprintf("%.2f", r.totalMinutes()/60.0), "", fmt.Sprintf("%.2f", r.GrandTotal()))

	// Customer/Project rows
	cps := r.SortedCustomerProjects()
	for _, cp := range cps {
		// Customer/project summary row
		cpName := r.customerProjectStr(cp, false)
		cpMinutes := r.TotalMinutesForCustomerProject(cp)
		rt := r.RateTotals[cp]

		reportTable.AddRow(cpName, fmt.Sprintf("%.2f", cpMinutes/60.0), fmt.Sprintf("%.2f", rt.Rate), fmt.Sprintf("%.2f", rt.Total))

		if !r.TaskGrain {
			continue
		}

		// Task level
		taskIDs := r.SortedTaskIDs(cp)
		for _, taskID := range taskIDs {
			taskName := r.formatTaskName(cp, taskID)
			taskMinutes := r.TotalMinutesForTask(cp, taskID)
			reportTable.AddRow(" * "+taskName, fmt.Sprintf("%.2f", taskMinutes/60.0), "", "")

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
				displayDesc := desc
				if displayDesc == "" {
					displayDesc = "<NO DESCRIPTION>"
				}
				descMinutes := r.TotalMinutesForDescription(cp, taskID, desc)
				reportTable.AddRow("    * "+displayDesc, fmt.Sprintf("%.2f", descMinutes/60.0), "", "")
			}
		}
	}

	buf.WriteString(r.tableToString(reportTable))
	buf.WriteString("\n")

	return buf.String()
}

// headerText returns the report header with smart date formatting
func (r *Report) headerText() string {
	start := r.StartDate
	end := r.EndDate

	// Check for whole year
	if start.Year() == end.Year() &&
		start.Month() == time.January && start.Day() == 1 &&
		end.Month() == time.December && end.Day() == 31 {
		return fmt.Sprintf("Time Report: %d", start.Year())
	}

	// Check for single month
	if start.Year() == end.Year() && start.Month() == end.Month() {
		daysInMonth := time.Date(start.Year(), start.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if start.Day() == 1 && end.Day() == daysInMonth {
			return fmt.Sprintf("Time Report: %s", start.Format("January 2006"))
		}
		// Check for single day
		if start.Day() == end.Day() {
			return fmt.Sprintf("Time Report: %s", start.Format("2006-01-02"))
		}
	}

	// Default: show range
	return fmt.Sprintf("Time Report: %s - %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
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
func (r *Report) formatTaskName(cp CustomerProject, taskID lib.TaskID) string {
	// Get first entry to check type
	var firstEntry *lib.TimeEntry
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

// tableToString converts a table to its string representation
func (r *Report) tableToString(tbl table.Table) string {
	var buf bytes.Buffer
	tbl.WithWriter(&buf).Print()
	// Trim the leading/trailing newlines that Print() adds
	return strings.TrimSpace(buf.String())
}

// padRight pads a string to the right with spaces
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// totalMinutes returns total minutes across all entries
func (r *Report) totalMinutes() float64 {
	var total float64
	for _, duration := range r.DayStats {
		total += duration.Minutes()
	}
	return total
}
