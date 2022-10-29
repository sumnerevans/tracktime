package lib

import (
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalText(text []byte) error {
	switch string(text) {
	case "today":
		d.Time = time.Now()
	case "yesterday":
		d.Time = time.Now().AddDate(0, 0, -1)
	default:
		return fmt.Errorf("Invalid date '%s'.", string(text))
	}
	return nil
}
