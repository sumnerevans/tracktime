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
		formatsWithoutYear := []string{"02", "2"}
		for _, format := range formatsWithoutYear {
			parsed, err := time.Parse(format, string(text))
			if err == nil {
				fmt.Printf("Parsed %v\n", parsed)
				d.Time = time.Date(now.Year(), now.Month(), parsed.Day(), 0, 0, 0, 0, time.Local)
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
