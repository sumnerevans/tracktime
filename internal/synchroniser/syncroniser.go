package synchroniser

import (
	"context"
	"time"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type AggregatedTimeKey struct {
	Type    timeentry.TimeEntryType
	Project string
	TaskID  string
}

type AggregatedTime map[AggregatedTimeKey]time.Duration

type Synchroniser interface {
	Init(config config.SyncConfig)
	Name() string
	Sync(ctx context.Context, aggregatedTime, syncedTime AggregatedTime, month types.Month) (AggregatedTime, error)
	GetFormattedTaskID(entry *timeentry.TimeEntry) string
	GetTaskLink(entry *timeentry.TimeEntry) string
	GetTaskDescription(ctx context.Context, entry *timeentry.TimeEntry) string
}

var Synchronisers []Synchroniser
