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
		avgMinutes := totalMinutes / float64(stats.DaysWorked)
		stats.AvgTimePerDay = time.Duration(avgMinutes) * time.Minute
	}

	if stats.WeekdaysWorked > 0 {
		avgMinutes := totalMinutes / float64(stats.WeekdaysWorked)
		stats.AvgTimePerWeekday = time.Duration(avgMinutes) * time.Minute
	}

	// Calculate weeks worked and average per week
	stats.WeeksWorked = float64(stats.WeekdaysWorked) / 5.0

	if stats.WeekdaysWorked == 0 {
		stats.AvgTimePerWeek = 0
	} else if stats.WeeksWorked < 1 {
		// Less than a week worked, use total
		stats.AvgTimePerWeek = time.Duration(totalMinutes) * time.Minute
	} else {
		avgMinutes := totalMinutes / stats.WeeksWorked
		stats.AvgTimePerWeek = time.Duration(avgMinutes) * time.Minute
	}

	return &stats
}

// formatDuration formats duration as H:MM
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%d:%02d", hours, mins)
}
