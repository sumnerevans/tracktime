package lib

import (
	"fmt"
	"time"
)

type Time struct {
	minutes int
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
	}
	for _, format := range []string{"1504", "15:04"} {
		parsed, err := time.Parse(format, timeStr)
		if err == nil {
			return TimeFrom(parsed), nil
		}
	}
	return &Time{}, fmt.Errorf("Error parsing time '%s'", timeStr)
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
