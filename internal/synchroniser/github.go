package synchroniser

import (
	"context"
	"fmt"
	"strings"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type GitHubSynchroniser struct {
	Config config.GitHubSyncConfig
}

var _ Synchroniser = (*GitHubSynchroniser)(nil)

func (gh *GitHubSynchroniser) Name() string { return "GitHub" }

func (gh *GitHubSynchroniser) Init(cfg config.SyncConfig) {
	gh.Config = cfg.GitHub
}

func (gh *GitHubSynchroniser) Sync(ctx context.Context, aggregatedTime, syncedTime AggregatedTime, month types.Month) (AggregatedTime, error) {
	return aggregatedTime, nil
}

func (gh *GitHubSynchroniser) cleanTaskID(taskID timeentry.TaskID) string {
	return strings.TrimPrefix(string(taskID), "#")
}

func (gh *GitHubSynchroniser) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	if (entry.Type != "github" && entry.Type != "gh") || entry.TaskID == "" {
		return ""
	}

	return fmt.Sprintf("#%s", gh.cleanTaskID(entry.TaskID))
}

func (gh *GitHubSynchroniser) GetTaskLink(entry *timeentry.TimeEntry) string {
	var owner, project string
	projectParts := strings.Split(string(entry.Project), "/")
	if len(projectParts) == 1 {
		if gh.Config.Username == "" {
			return ""
		}
		owner = gh.Config.Username
		project = projectParts[0]
	} else if len(projectParts) == 2 {
		owner = projectParts[0]
		project = projectParts[1]
	} else {
		return ""
	}
	// Always link to /issues/ because it will redirect to /pull/ if necessary.
	return fmt.Sprintf("%s/%s/%s/issues/%s", gh.Config.RootURI, owner, project, gh.cleanTaskID(entry.TaskID))
}

func (gh *GitHubSynchroniser) GetTaskDescription(ctx context.Context, entry *timeentry.TimeEntry) string {
	return ""
}

func init() {
	Synchronisers = append(Synchronisers, &GitHubSynchroniser{})
}
