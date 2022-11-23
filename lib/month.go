package lib

import (
	"fmt"
	"time"
)

type Month struct {
	year  int
	month time.Month
}

func ThisMonth() Month {
	now := time.Now().Local()
	return Month{year: now.Year(), month: now.Month()}
}

func (d *Month) UnmarshalText(text []byte) error {
	now := time.Now().Local()
	switch string(text) {
	case "this month", "thismonth":
		d.year = now.Year()
		d.month = now.Month()
	default:
		// TODO
		return fmt.Errorf("Invalid date '%s'.", string(text))
	}
	return nil
}
