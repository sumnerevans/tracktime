package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	root := gh.Config.RootURI
	if root == "" {
		root = "https://github.com"
	}
	// Link to /issues/ — GitHub redirects to /pull/ for PRs.
	return fmt.Sprintf("%s/%s/%s/issues/%s", root, owner, project, gh.cleanTaskID(entry.TaskID))
}

func (gh *GitHubResolver) FetchDescription(ctx context.Context, entry *timeentry.TimeEntry) (string, error) {
	if gh.Config.AccessToken == "" {
		return "", nil
	}

	var owner, repo string
	parts := strings.Split(string(entry.Project), "/")
	switch len(parts) {
	case 1:
		if gh.Config.Username == "" {
			return "", nil
		}
		owner = gh.Config.Username
		repo = parts[0]
	case 2:
		owner = parts[0]
		repo = parts[1]
	default:
		return "", nil
	}
	id := gh.cleanTaskID(entry.TaskID)

	gqlQuery := fmt.Sprintf(
		`query{repository(owner:"%s",name:"%s"){issue(number:%s){title}pullRequest(number:%s){title}discussion(number:%s){title}}}`,
		owner, repo, id, id, id,
	)
	body, _ := json.Marshal(map[string]string{"query": gqlQuery})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+gh.Config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Repository struct {
				Issue       *struct{ Title string } `json:"issue"`
				PullRequest *struct{ Title string } `json:"pullRequest"`
				Discussion  *struct{ Title string } `json:"discussion"`
			} `json:"repository"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	repo2 := result.Data.Repository
	if repo2.Issue != nil {
		return repo2.Issue.Title, nil
	}
	if repo2.PullRequest != nil {
		return repo2.PullRequest.Title, nil
	}
	if repo2.Discussion != nil {
		return repo2.Discussion.Title, nil
	}
	return "", nil
}

func init() {
	Resolvers = append(Resolvers, &GitHubResolver{})
}
