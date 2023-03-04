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
	switch strings.ToLower(string(text)) {
	case "this month", "thismonth":
		d.year = now.Year()
		d.month = now.Month()
	default:
		formats := []string{"01", "January", "Jan", "1"}
		for _, format := range formats {
			parsed, err := time.Parse(format, string(text))
			if err == nil {
				d.year = now.Year()
				d.month = parsed.Month()
				return nil
			}
		}
		return fmt.Errorf("invalid month '%s'", string(text))
	}
	return nil
}
