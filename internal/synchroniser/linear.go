package synchroniser

import (
	"context"
	"fmt"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type LinearSynchroniser struct {
	Config config.LinearSyncConfig
}

var _ Synchroniser = (*LinearSynchroniser)(nil)

func (l *LinearSynchroniser) Name() string { return "Linear" }

func (l *LinearSynchroniser) Init(cfg config.SyncConfig) {
	l.Config = cfg.Linear
}

func (l *LinearSynchroniser) Sync(ctx context.Context, aggregatedTime, syncedTime AggregatedTime, month types.Month) (AggregatedTime, error) {
	return aggregatedTime, nil
}

func (l *LinearSynchroniser) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	if entry.Type != "linear" || entry.Project == "" || entry.TaskID == "" {
		return ""
	}

	return fmt.Sprintf("%s-%s", entry.Project, entry.TaskID)
}

func (l *LinearSynchroniser) GetTaskLink(entry *timeentry.TimeEntry) string {
	if entry.Type != "linear" || entry.Project == "" || entry.TaskID == "" {
		return ""
	}

	if l.Config.DefaultOrg == "" {
		return ""
	}

	issueID := fmt.Sprintf("%s-%s", entry.Project, entry.TaskID)
	return fmt.Sprintf("https://linear.app/%s/issue/%s", l.Config.DefaultOrg, issueID)
}

func (l *LinearSynchroniser) GetTaskDescription(ctx context.Context, entry *timeentry.TimeEntry) string {
	// TODO: Implement GraphQL query to Linear API with caching
	return ""
}

func init() {
	Synchronisers = append(Synchronisers, &LinearSynchroniser{})
}
