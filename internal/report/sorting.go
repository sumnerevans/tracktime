package report

import (
	"sort"
	"strings"
	"time"

	"github.com/sumnerevans/tracktime/internal/timeentry"
)

// SortedCustomerProjects returns CustomerProject keys sorted according to report settings
func (r *Report) SortedCustomerProjects() []CustomerProject {
	cps := make([]CustomerProject, 0, len(r.AggregatedTime))
	for cp := range r.AggregatedTime {
		cps = append(cps, cp)
	}

	sort.Slice(cps, func(i, j int) bool {
		if r.Sort == SortAlphabetical {
			// Alphabetical: sort by customer then project (case-insensitive)
			si := strings.ToLower(string(cps[i].Customer) + string(cps[i].Project))
			sj := strings.ToLower(string(cps[j].Customer) + string(cps[j].Project))
			if r.Reverse {
				return si > sj
			}
			return si < sj
		}
		// Time spent: sort by total minutes
		minutesI := r.TotalMinutesForCustomerProject(cps[i])
		minutesJ := r.TotalMinutesForCustomerProject(cps[j])
		if r.Reverse {
			return minutesI < minutesJ
		}
		return minutesI > minutesJ
	})

	return cps
}

// SortedTaskIDs returns TaskID keys for a customer/project sorted according to report settings
func (r *Report) SortedTaskIDs(cp CustomerProject) []timeentry.TaskID {
	tasks := r.AggregatedTime[cp]
	taskIDs := make([]timeentry.TaskID, 0, len(tasks))
	for taskID := range tasks {
		taskIDs = append(taskIDs, taskID)
	}

	sort.Slice(taskIDs, func(i, j int) bool {
		if r.Sort == SortAlphabetical {
			// Alphabetical: sort by task ID (case-insensitive)
			si := strings.ToLower(string(taskIDs[i]))
			sj := strings.ToLower(string(taskIDs[j]))
			if r.Reverse {
				return si > sj
			}
			return si < sj
		}
		// Time spent: sort by total minutes for this task
		minutesI := r.TotalMinutesForTask(cp, taskIDs[i])
		minutesJ := r.TotalMinutesForTask(cp, taskIDs[j])
		if r.Reverse {
			return minutesI < minutesJ
		}
		return minutesI > minutesJ
	})

	return taskIDs
}

// SortedDescriptions returns description keys for a task sorted according to report settings
func (r *Report) SortedDescriptions(cp CustomerProject, taskID timeentry.TaskID) []string {
	descriptions := r.AggregatedTime[cp][taskID]
	descs := make([]string, 0, len(descriptions))
	for desc := range descriptions {
		descs = append(descs, desc)
	}

	sort.Slice(descs, func(i, j int) bool {
		if r.Sort == SortAlphabetical {
			// Alphabetical: sort by description (case-insensitive)
			si := strings.ToLower(descs[i])
			sj := strings.ToLower(descs[j])
			if r.Reverse {
				return si > sj
			}
			return si < sj
		}
		// Time spent: sort by total minutes for this description
		minutesI := r.TotalMinutesForDescription(cp, taskID, descs[i])
		minutesJ := r.TotalMinutesForDescription(cp, taskID, descs[j])
		if r.Reverse {
			return minutesI < minutesJ
		}
		return minutesI > minutesJ
	})

	return descs
}

// TotalMinutesForTask returns total minutes for a specific task
func (r *Report) TotalMinutesForTask(cp CustomerProject, taskID timeentry.TaskID) float64 {
	var total time.Duration
	for _, entries := range r.AggregatedTime[cp][taskID] {
		for _, entry := range entries {
			if duration, err := entry.Duration(false); err == nil {
				total += duration
			}
		}
	}
	return total.Minutes()
}

// TotalMinutesForDescription returns total minutes for a specific description
func (r *Report) TotalMinutesForDescription(cp CustomerProject, taskID timeentry.TaskID, description string) float64 {
	var total time.Duration
	for _, entry := range r.AggregatedTime[cp][taskID][description] {
		if duration, err := entry.Duration(false); err == nil {
			total += duration
		}
	}
	return total.Minutes()
}
