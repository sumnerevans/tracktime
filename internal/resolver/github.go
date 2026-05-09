package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

type GitHubResolver struct {
	Config config.GitHubConfig
}

var _ ItemDetailResolver = (*GitHubResolver)(nil)

func (gh *GitHubResolver) Init(cfg *config.Config) { gh.Config = cfg.GitHub }

func (gh *GitHubResolver) Handles(entry *timeentry.TimeEntry) bool {
	return entry.Type == "github" || entry.Type == "gh"
}

func (gh *GitHubResolver) cleanTaskID(taskID timeentry.TaskID) string {
	return strings.TrimPrefix(string(taskID), "#")
}

func (gh *GitHubResolver) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	if !gh.Handles(entry) || entry.TaskID == "" {
		return ""
	}
	return fmt.Sprintf("#%s", gh.cleanTaskID(entry.TaskID))
}

func (gh *GitHubResolver) GetTaskLink(entry *timeentry.TimeEntry) string {
	if !gh.Handles(entry) || entry.TaskID == "" {
		return ""
	}
	var owner, project string
	projectParts := strings.Split(string(entry.Project), "/")
	switch len(projectParts) {
	case 1:
		if gh.Config.Username == "" {
			return ""
		}
		owner = gh.Config.Username
		project = projectParts[0]
	case 2:
		owner = projectParts[0]
		project = projectParts[1]
	default:
		return ""
	}
	// Link to /issues/ — GitHub redirects to /pull/ for PRs.
	return fmt.Sprintf("%s/%s/%s/issues/%s", gh.Config.RootURI, owner, project, gh.cleanTaskID(entry.TaskID))
}

func (gh *GitHubResolver) FetchDescription(_ context.Context, _ *timeentry.TimeEntry) (string, error) {
	// GitHub description fetch not yet implemented.
	return "", nil
}

func init() {
	Resolvers = append(Resolvers, &GitHubResolver{})
}
