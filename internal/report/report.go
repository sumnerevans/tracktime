package report

import (
	"time"

	"github.com/sumnerevans/tracktime/internal/lib"
)

// CustomerProject represents a unique customer/project combination
type CustomerProject struct {
	Customer lib.Customer
	Project  lib.Project
}

// RateTotal contains the hourly rate and total amount for a customer/project
type RateTotal struct {
	Rate  float64
	Total float64
}

// Report holds aggregated time entry data and statistics
type Report struct {
	StartDate lib.Date
	EndDate   lib.Date
	Config    *lib.Config

	// Aggregated data: Customer/Project -> TaskID -> Description -> []*TimeEntry
	AggregatedTime map[CustomerProject]map[lib.TaskID]map[string][]*lib.TimeEntry
	DayStats       map[lib.Date]time.Duration
	RateTotals     map[CustomerProject]RateTotal

	// Filter options
	Customer lib.Customer
	Project  lib.Project

	// Display options
	Sort             Sort
	Reverse          bool
	TaskGrain        bool
	DescriptionGrain bool
}

// Sort represents how to sort report entries
type Sort int

const (
	SortAlphabetical Sort = iota
	SortTimeSpent
)

// New creates a new Report by aggregating time entries over the date range
func New(config *lib.Config, start, end lib.Date, customer lib.Customer, project lib.Project, sort Sort, reverse, taskGrain, descriptionGrain bool) (*Report, error) {
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
		AggregatedTime:   make(map[CustomerProject]map[lib.TaskID]map[string][]*lib.TimeEntry),
		DayStats:         make(map[lib.Date]time.Duration),
		RateTotals:       make(map[CustomerProject]RateTotal),
	}

	// Aggregate time entries across date range
	for day := start; day.Before(end.Time) || day.Equal(end.Time); day = day.AddDays(1) {
		entryList, err := lib.EntryListForDay(config, day)
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
				r.AggregatedTime[cp] = make(map[lib.TaskID]map[string][]*lib.TimeEntry)
			}
			if r.AggregatedTime[cp][entry.TaskID] == nil {
				r.AggregatedTime[cp][entry.TaskID] = make(map[string][]*lib.TimeEntry)
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
		totalMinutes := r.TotalMinutesForCustomerProject(cp)
		total := (totalMinutes / 60.0) * rate

		r.RateTotals[cp] = RateTotal{
			Rate:  rate,
			Total: total,
		}
	}
}

// TotalMinutesForCustomerProject returns total minutes for a customer/project combination
func (r *Report) TotalMinutesForCustomerProject(cp CustomerProject) float64 {
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

// GrandTotal returns the sum of all rate totals
func (r *Report) GrandTotal() float64 {
	var total float64
	for _, rt := range r.RateTotals {
		total += rt.Total
	}
	return total
}
