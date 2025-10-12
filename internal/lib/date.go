package lib

import (
	"fmt"
	"strings"
	"time"
)

type Date struct {
	time.Time
}

func NewDate(year, month, day int) Date {
	return Date{Time: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)}
}

func NewDateFromTime(t time.Time) Date {
	return Date{Time: t}
}

func Today() Date {
	return NewDateFromTime(time.Now())
}

// mostRecentWeekday returns the most recent occurrence of the target weekday,
// including today if today matches the target weekday.
func mostRecentWeekday(now time.Time, target time.Weekday) time.Time {
	if now.Weekday() == target {
		return now
	}
	return now.AddDate(0, 0, int(target-now.Weekday()))
}

// UnmarshalText parses a date from text input, supporting multiple formats:
//
// Special keywords:
//   - "today": Current date
//   - "yesterday": Previous day
//
// Weekday names (case-insensitive):
//   - "monday", "tuesday", etc.: Most recent occurrence of that weekday
//     (includes today if today is that weekday, otherwise goes backwards)
//
// Date formats:
//   - YYYY-MM-DD, YYYY/MM/DD: Full dates
//   - YY-MM-DD, YY/MM/DD: Two-digit year
//   - MM-DD, M-D: Defaults to current year
//   - DD, D: Defaults to current year and month
func (d *Date) UnmarshalText(text []byte) error {
	now := time.Now().Local()
	input := strings.ToLower(string(text))

	// Handle special keywords
	switch input {
	case "today":
		d.Time = now
		return nil
	case "yesterday":
		d.Time = now.AddDate(0, 0, -1)
		return nil
	}

	// Handle weekday names
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	if weekday, ok := weekdays[input]; ok {
		d.Time = mostRecentWeekday(now, weekday)
		return nil
	}

	// Handle date formats - try each format in priority order
	formats := []struct {
		layout       string
		defaultYear  bool
		defaultMonth bool
	}{
		{"2006-01-02", false, false}, // YYYY-MM-DD
		{"2006/01/02", false, false}, // YYYY/MM/DD
		{"06-01-02", false, false},   // YY-MM-DD
		{"06/01/02", false, false},   // YY/MM/DD
		{"01-02", true, false},       // MM-DD (default year)
		{"1-2", true, false},         // M-D (default year)
		{"02", true, true},           // DD (default year+month)
		{"2", true, true},            // D (default year+month)
	}

	for _, fmt := range formats {
		parsed, err := time.Parse(fmt.layout, string(text))
		if err == nil {
			year, month, day := parsed.Date()
			if fmt.defaultYear {
				year = now.Year()
			}
			if fmt.defaultMonth {
				month = now.Month()
			}
			d.Time = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
			return nil
		}
	}
	return fmt.Errorf("invalid date '%s'", string(text))
}

func (d Date) AddDays(numDays int) Date {
	return Date{Time: d.Time.AddDate(0, 0, numDays)}
}

func (d Date) AddMonths(numMonths int) Date {
	return Date{Time: d.Time.AddDate(0, numMonths, 0)}
}

func (d Date) DaysInMonth() int {
	year, month, _ := d.Date()
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
