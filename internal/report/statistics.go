package report

import (
	"fmt"
	"time"

	"github.com/sumnerevans/tracktime/internal/lib"
)

// Statistics holds calculated statistics for a report period
type Statistics struct {
	DaysWorked            int
	WeekdaysWorked        int
	WeeksWorked           float64
	AvgTimePerDay         time.Duration
	AvgTimePerWeekday     time.Duration
	AvgTimePerWeek        time.Duration
	totalMinutesWorked    float64
	dayWorkedMinThreshold int
}

// CalculateStatistics computes statistics from day stats
func (r *Report) CalculateStatistics() *Statistics {
	stats := &Statistics{
		dayWorkedMinThreshold: r.Config.Reporting.DayWorkedMinThreshold,
	}

	// Count days where minutes >= threshold
	daysWorked := make(map[lib.Date]time.Duration)
	for day, duration := range r.DayStats {
		if int(duration.Minutes()) >= stats.dayWorkedMinThreshold {
			daysWorked[day] = duration
		}
	}

	// Calculate total minutes across all days
	for _, duration := range r.DayStats {
		stats.totalMinutesWorked += duration.Minutes()
	}

	// Count days and weekdays worked
	stats.DaysWorked = len(daysWorked)
	for day := range daysWorked {
		// Monday=1, ..., Friday=5, Saturday=6, Sunday=0
		if day.Weekday() >= time.Monday && day.Weekday() <= time.Friday {
			stats.WeekdaysWorked++
		}
	}

	// Calculate averages
	if stats.DaysWorked > 0 {
		avgMinutes := stats.totalMinutesWorked / float64(stats.DaysWorked)
		stats.AvgTimePerDay = time.Duration(avgMinutes) * time.Minute
	}

	if stats.WeekdaysWorked > 0 {
		avgMinutes := stats.totalMinutesWorked / float64(stats.WeekdaysWorked)
		stats.AvgTimePerWeekday = time.Duration(avgMinutes) * time.Minute
	}

	// Calculate weeks worked and average per week
	// Python logic (lines 76-83): weeks = weekdays / 5
	stats.WeeksWorked = float64(stats.WeekdaysWorked) / 5.0

	if stats.WeekdaysWorked == 0 {
		stats.AvgTimePerWeek = 0
	} else if stats.WeeksWorked < 1 {
		// Less than a week worked, use total
		stats.AvgTimePerWeek = time.Duration(stats.totalMinutesWorked) * time.Minute
	} else {
		avgMinutes := stats.totalMinutesWorked / stats.WeeksWorked
		stats.AvgTimePerWeek = time.Duration(avgMinutes) * time.Minute
	}

	return stats
}

// FormatDuration formats duration as H:MM
func FormatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%d:%02d", hours, mins)
}

// StatisticsMap returns statistics as a map for display
func (s *Statistics) StatisticsMap() map[string]string {
	return map[string]string{
		"Days worked":                     fmt.Sprintf("%d", s.DaysWorked),
		"Average time per day worked":     FormatDuration(s.AvgTimePerDay),
		"Average time per weekday worked": FormatDuration(s.AvgTimePerWeekday),
		"Weeks* worked":                   fmt.Sprintf("%.1f", s.WeeksWorked),
		"Average time per week* worked":   FormatDuration(s.AvgTimePerWeek),
	}
}
