package synchroniser

import (
	"context"
	"fmt"
	"strings"

	"github.com/sumnerevans/tracktime/lib"
)

type GitHubSynchroniser struct {
	Config lib.GitHubSyncConfig
}

func (gh *GitHubSynchroniser) Name() string { return "GitHub" }

func (gh *GitHubSynchroniser) Init(config lib.SyncConfig) {
	gh.Config = config.GitHub
}

func (gh *GitHubSynchroniser) Sync(ctx context.Context, aggregatedTime, syncedTime AggregatedTime, month lib.Month) (AggregatedTime, error) {
	return aggregatedTime, nil
}

func (gh *GitHubSynchroniser) cleanTaskID(taskID string) string {
	return strings.TrimPrefix(taskID, "#")
}

func (gh *GitHubSynchroniser) GetFormattedTaskID(entry *lib.TimeEntry) string {
	if (entry.Type != "github" && entry.Type != "gh") || entry.TaskID == "" {
		return ""
	}

	return fmt.Sprintf("#%s", gh.cleanTaskID(entry.TaskID))
}

func (gh *GitHubSynchroniser) GetTaskLink(entry *lib.TimeEntry) string {
	var owner, project string
	projectParts := strings.Split(entry.Project, "/")
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

func (gh *GitHubSynchroniser) GetTaskDescription(ctx context.Context, entry *lib.TimeEntry) string {
	return ""
}

func init() {
	Synchronisers = append(Synchronisers, &GitHubSynchroniser{})
}
