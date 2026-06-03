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

// formatFloat formats f with 2 decimal places and thousands separators.
func formatFloat(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	dotIdx := len(s) - 3 // always ends in .XX
	intPart := s[:dotIdx]
	fracPart := s[dotIdx:]
	negative := len(intPart) > 0 && intPart[0] == '-'
	if negative {
		intPart = intPart[1:]
	}
	n := len(intPart)
	var b []byte
	for i := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, intPart[i])
	}
	if negative {
		return "-" + string(b) + fracPart
	}
	return string(b) + fracPart
}
