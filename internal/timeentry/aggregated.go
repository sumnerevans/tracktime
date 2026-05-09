package timeentry

import "time"

type AggregatedTimeKey struct {
	Type    TimeEntryType
	Project Project
	TaskID  TaskID
}

type AggregatedTime map[AggregatedTimeKey]time.Duration
