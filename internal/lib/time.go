package lib

import (
	"fmt"
	"time"
)

type Time struct {
	minutes int
}

func (d *Time) UnmarshalText(text []byte) (err error) {
	var t *Time
	t, err = ParseTime(string(text))
	d.minutes = t.minutes
	return
}

func TimeFromMinutes(minutes int) *Time {
	return &Time{minutes: minutes}
}

func TimeFrom(t time.Time) *Time {
	return &Time{minutes: t.Minute() + t.Hour()*60}
}

func CurrentTime() *Time {
	return TimeFrom(time.Now().Local())
}

func ParseTime(timeStr string) (*Time, error) {
	if timeStr == "" {
		return nil, nil
	} else if timeStr == "now" {
		return CurrentTime(), nil
	}
	for _, format := range []string{"1504", "15:04"} {
		parsed, err := time.Parse(format, timeStr)
		if err == nil {
			return TimeFrom(parsed), nil
		}
	}
	return &Time{}, fmt.Errorf("error parsing time '%s'", timeStr)
}

func (t *Time) Between(start *Time, end *Time) bool {
	return start.minutes <= t.minutes && t.minutes < end.minutes
}

func (t *Time) Before(other *Time) bool {
	return t.minutes < other.minutes
}

func (t *Time) String() string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", t.minutes/60, t.minutes%60)
}

func (t *Time) Sub(other *Time) time.Duration {
	return time.Duration(t.minutes-other.minutes) * time.Minute
}
