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

func (d *Date) UnmarshalText(text []byte) error {
	now := time.Now().Local()
	switch strings.ToLower(string(text)) {
	case "today":
		d.Time = now
	case "yesterday":
		d.Time = now.AddDate(0, 0, -1)
	case "monday":
		if now.Weekday() == time.Monday {
			d.Time = now
		} else {
			d.Time = now.AddDate(0, 0, int(time.Monday-now.Weekday()))
		}
	case "tuesday":
		if now.Weekday() == time.Tuesday {
			d.Time = now
		} else {
			d.Time = now.AddDate(0, 0, int(time.Tuesday-now.Weekday()))
		}
	default:
		// Try each format in priority order
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
	return nil
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
