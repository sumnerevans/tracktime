package lib

import (
	"fmt"
	"strings"
	"time"
)

type Month struct {
	year  int
	month time.Month
}

func NewMonth(year int, month time.Month) Month {
	return Month{year: year, month: month}
}

func ThisMonth() Month {
	now := time.Now().Local()
	return Month{year: now.Year(), month: now.Month()}
}

func (d *Month) UnmarshalText(text []byte) error {
	now := time.Now().Local()
	s := string(text)

	// Check for YYYY-MM format first
	if strings.Contains(s, "-") {
		parsed, err := time.Parse("2006-01", s)
		if err == nil {
			d.year = parsed.Year()
			d.month = parsed.Month()
			return nil
		}
	}

	switch strings.ToLower(s) {
	case "this month", "thismonth":
		d.year = now.Year()
		d.month = now.Month()
	default:
		formats := []string{"01", "January", "Jan", "1"}
		for _, format := range formats {
			parsed, err := time.Parse(format, s)
			if err == nil {
				d.year = now.Year()
				d.month = parsed.Month()
				return nil
			}
		}
		return fmt.Errorf("invalid month '%s'", s)
	}
	return nil
}

// IsZero returns whether this Month is the zero value
func (m Month) IsZero() bool {
	return m.year == 0 && m.month == 0
}

// WithDefaultYear returns a time.Time with the given default year if not already set
func (m Month) WithDefaultYear(defaultYear int) time.Time {
	year := m.year
	if year == 0 {
		year = defaultYear
	}
	return time.Date(year, m.month, 1, 0, 0, 0, 0, time.Local)
}

// Year returns the year component
func (m Month) Year() int {
	return m.year
}

// Month returns the month component
func (m Month) Month() time.Month {
	return m.month
}

// DaysInMonth returns the number of days in this month
func (m Month) DaysInMonth() int {
	return time.Date(m.year, m.month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
