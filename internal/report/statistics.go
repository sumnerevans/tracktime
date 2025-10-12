package report

import (
	"fmt"
	"time"
)

// Statistics holds calculated statistics for a report period
type Statistics struct {
	DaysWorked        int
	WeekdaysWorked    int
	WeeksWorked       float64
	AvgTimePerDay     time.Duration
	AvgTimePerWeekday time.Duration
	AvgTimePerWeek    time.Duration
}

// CalculateStatistics computes statistics from day stats
func (r *Report) CalculateStatistics() *Statistics {
	var stats Statistics

	// Count days where minutes >= threshold
	for day, duration := range r.DayStats {
		if int(duration.Minutes()) >= r.Config.Reporting.DayWorkedMinThreshold {
			stats.DaysWorked++

			// Detect whether it's a weekday
			if day.Weekday() >= time.Monday && day.Weekday() <= time.Friday {
				stats.WeekdaysWorked++
			}
		}
	}

	// Calculate total minutes across all days
	var totalMinutes float64
	for _, duration := range r.DayStats {
		totalMinutes += duration.Minutes()
	}

	// Calculate averages
	if stats.DaysWorked > 0 {
		stats.AvgTimePerDay = time.Duration(totalMinutes/float64(stats.DaysWorked)) * time.Minute
	}

	if stats.WeekdaysWorked > 0 {
		stats.AvgTimePerWeekday = time.Duration(totalMinutes/float64(stats.WeekdaysWorked)) * time.Minute
	}

	// Calculate weeks worked and average per week
	stats.WeeksWorked = float64(stats.WeekdaysWorked) / 5.0

	if stats.WeekdaysWorked == 0 {
		stats.AvgTimePerWeek = 0
	} else if stats.WeeksWorked < 1 {
		// Less than a week worked, use total
		stats.AvgTimePerWeek = time.Duration(totalMinutes) * time.Minute
	} else {
		stats.AvgTimePerWeek = time.Duration(totalMinutes/stats.WeeksWorked) * time.Minute
	}

	return &stats
}

// formatDuration formats duration as H:MM
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	return fmt.Sprintf("%d:%02d", minutes/60, minutes%60)
}
