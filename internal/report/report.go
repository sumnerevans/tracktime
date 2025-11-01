package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/synchroniser"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

// CustomerProject represents a unique customer/project combination
type CustomerProject struct {
	Customer timeentry.Customer
	Project  timeentry.Project
}

// RateTotal contains the hourly rate and total amount for a customer/project
type RateTotal struct {
	Rate  float64
	Total float64
}

// Report holds aggregated time entry data and statistics
type Report struct {
	StartDate types.Date
	EndDate   types.Date
	Config    *config.Config

	// Aggregated data: Customer/Project -> TaskID -> Description -> []*TimeEntry
	AggregatedTime map[CustomerProject]map[timeentry.TaskID]map[string][]*timeentry.TimeEntry
	DayStats       map[types.Date]time.Duration
	RateTotals     map[CustomerProject]RateTotal

	// Filter options
	Customer timeentry.Customer
	Project  timeentry.Project

	// Display options
	Sort             Sort
	Reverse          bool
	TaskGrain        bool
	DescriptionGrain bool

	// Synchronisers for formatting task IDs and getting links
	Synchronisers []synchroniser.Synchroniser
}

// Sort represents how to sort report entries
type Sort int

const (
	SortAlphabetical Sort = iota
	SortTimeSpent
)

// New creates a new Report by aggregating time entries over the date range
func New(config *config.Config, start, end types.Date, customer timeentry.Customer, project timeentry.Project, sort Sort, reverse, taskGrain, descriptionGrain bool) (*Report, error) {
	r := &Report{
		StartDate:        start,
		EndDate:          end,
		Config:           config,
		Customer:         customer,
		Project:          project,
		Sort:             sort,
		Reverse:          reverse,
		TaskGrain:        taskGrain || descriptionGrain, // Task grain implied by description grain
		DescriptionGrain: descriptionGrain,
		AggregatedTime:   make(map[CustomerProject]map[timeentry.TaskID]map[string][]*timeentry.TimeEntry),
		DayStats:         make(map[types.Date]time.Duration),
		RateTotals:       make(map[CustomerProject]RateTotal),
		Synchronisers:    make([]synchroniser.Synchroniser, len(synchroniser.Synchronisers)),
	}

	// Initialize synchronisers with config
	for i, sync := range synchroniser.Synchronisers {
		r.Synchronisers[i] = sync
		r.Synchronisers[i].Init(config.Sync)
	}

	// Aggregate time entries across date range
	for day := start; day.Before(end.Time) || day.Equal(end.Time); day = day.AddDays(1) {
		entryList, err := timeentry.EntryListForDay(config, day)
		if err != nil {
			return nil, err
		}

		for _, entry := range entryList.EntriesForCustomer(customer) {
			// Filter by project if specified
			if project != "" && entry.Project != project {
				continue
			}

			// Check for unended entries
			duration, err := entry.Duration(false)
			if err != nil {
				return nil, err
			}

			// Add to day stats
			r.DayStats[day] += duration

			// Add to aggregated time
			cp := CustomerProject{Customer: entry.Customer, Project: entry.Project}
			if r.AggregatedTime[cp] == nil {
				r.AggregatedTime[cp] = make(map[timeentry.TaskID]map[string][]*timeentry.TimeEntry)
			}
			if r.AggregatedTime[cp][entry.TaskID] == nil {
				r.AggregatedTime[cp][entry.TaskID] = make(map[string][]*timeentry.TimeEntry)
			}
			// Note: Python uppercases description (line 191 in report.py)
			// We'll keep it as-is for now and can uppercase in formatter if needed
			r.AggregatedTime[cp][entry.TaskID][entry.Description] = append(
				r.AggregatedTime[cp][entry.TaskID][entry.Description],
				entry,
			)
		}
	}

	// Calculate rates and totals
	r.calculateRates()

	return r, nil
}

// calculateRates computes hourly rates and totals for each customer/project
func (r *Report) calculateRates() {
	for cp := range r.AggregatedTime {
		rate := 0.0

		// Check customer rate
		if cp.Customer != "" && r.Config.Reporting.CustomerRates != nil {
			if customerRate, ok := r.Config.Reporting.CustomerRates[string(cp.Customer)]; ok {
				rate = customerRate
			}
		}

		// Project rate overrides customer rate (Python line 197-201)
		if cp.Project != "" && r.Config.Reporting.ProjectRates != nil {
			if projectRate, ok := r.Config.Reporting.ProjectRates[string(cp.Project)]; ok {
				rate = projectRate
			}
		}

		// Calculate total: minutes / 60 * rate
		totalMinutes := r.totalMinutesForCustomerProject(cp)
		total := (totalMinutes / 60.0) * rate

		r.RateTotals[cp] = RateTotal{
			Rate:  rate,
			Total: total,
		}
	}
}

// totalMinutesForCustomerProject returns total minutes for a customer/project combination
func (r *Report) totalMinutesForCustomerProject(cp CustomerProject) float64 {
	var total time.Duration
	for _, tasks := range r.AggregatedTime[cp] {
		for _, entries := range tasks {
			for _, entry := range entries {
				if duration, err := entry.Duration(false); err == nil {
					total += duration
				}
			}
		}
	}
	return total.Minutes()
}

// grandTotal returns the sum of all rate totals
func (r *Report) grandTotal() float64 {
	var total float64
	for _, rt := range r.RateTotals {
		total += rt.Total
	}
	return total
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
func (r *Report) customerProjectStr(cp CustomerProject) string {
	noProject := "<no project>"
	noCustomer := "<no customer>"
	noBoth := "<no project or customer>"

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

	// Try to get formatted task ID from synchronisers
	for _, sync := range r.Synchronisers {
		if formattedID := sync.GetFormattedTaskID(firstEntry); formattedID != "" {
			return formattedID
		}
	}

	// Fallback: if no synchroniser handles this entry, use raw task ID
	if taskID != "" {
		return string(taskID)
	}

	return "<NO TASK>"
}

// getTaskLink returns a link to the task from any applicable synchroniser
func (r *Report) getTaskLink(cp CustomerProject, taskID timeentry.TaskID) string {
	// Get first entry to check type
	var firstEntry *timeentry.TimeEntry
	for _, entries := range r.AggregatedTime[cp][taskID] {
		if len(entries) > 0 {
			firstEntry = entries[0]
			break
		}
	}

	if firstEntry == nil {
		return ""
	}

	// Try to get task link from synchronisers
	for _, sync := range r.Synchronisers {
		if link := sync.GetTaskLink(firstEntry); link != "" {
			return link
		}
	}

	return ""
}

// totalMinutes returns total minutes across all entries
func (r *Report) totalMinutes() float64 {
	var total float64
	for _, duration := range r.DayStats {
		total += duration.Minutes()
	}
	return total
}
