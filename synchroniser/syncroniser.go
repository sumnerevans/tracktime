package synchroniser

import (
	"context"
	"time"

	"github.com/sumnerevans/tracktime/lib"
)

type AggregatedTimeKey struct {
	Type    lib.TimeEntryType
	Project string
	TaskID  string
}

type AggregatedTime map[AggregatedTimeKey]time.Duration

type Synchroniser interface {
	Init(config lib.SyncConfig)
	Name() string
	Sync(ctx context.Context, aggregatedTime, syncedTime AggregatedTime, month lib.Month) (AggregatedTime, error)
	GetFormattedTaskID(entry *lib.TimeEntry) string
	GetTaskLink(entry *lib.TimeEntry) string
	GetTaskDescription(ctx context.Context, entry *lib.TimeEntry) string
}

var Synchronisers []Synchroniser
