package lib

import (
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func Today() Date {
	return Date{Time: time.Now()}
}

func (d *Date) UnmarshalText(text []byte) error {
	switch string(text) {
	case "today":
		d.Time = time.Now()
	case "yesterday":
		d.Time = time.Now().AddDate(0, 0, -1)
	default:
		// TODO
		return fmt.Errorf("Invalid date '%s'.", string(text))
	}
	return nil
}

func (d *Date) AddDays(numDays int) Date {
	return Date{Time: d.Time.AddDate(0, 0, numDays)}
}
