package lib

import (
	"fmt"
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
	switch string(text) {
	case "today":
		d.Time = time.Now().Local()
	case "yesterday":
		d.Time = time.Now().Local().AddDate(0, 0, -1)
	default:
		// formatsWithoutYear := []string{"01", "January", "Jan", "1"}
		// for _, format := range formats {
		// 	parsed, err := time.Parse(format, string(text))
		// 	if err == nil {
		// 		d.year = now.Year()
		// 		d.month = parsed.Month()
		// 		return nil
		// 	}
		// }
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
