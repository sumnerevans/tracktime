package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

type GitLabResolver struct {
	Config config.GitLabConfig
}

var _ ItemDetailResolver = (*GitLabResolver)(nil)

var gitlabAPIRootRe = regexp.MustCompile(`(.*)/api/v4/?$`)

func (gl *GitLabResolver) Init(cfg *config.Config) { gl.Config = cfg.GitLab }

func (gl *GitLabResolver) Handles(entry *timeentry.TimeEntry) bool {
	return entry.Type == "gitlab" || entry.Type == "gl"
}

// taskTypeAndNumber splits a GitLab task ID like "!42" or "#7" or "7" into
// ("merge_requests","42") or ("issues","7").
func (gl *GitLabResolver) taskTypeAndNumber(taskID string) (taskType, number string) {
	if strings.HasPrefix(taskID, "!") {
		return "merge_requests", taskID[1:]
	}
	return "issues", strings.TrimPrefix(taskID, "#")
}

func (gl *GitLabResolver) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	if !gl.Handles(entry) || entry.TaskID == "" {
		return ""
	}
	taskID := string(entry.TaskID)
	if strings.HasPrefix(taskID, "!") || strings.HasPrefix(taskID, "#") {
		return taskID
	}
	return "#" + taskID
}

func (gl *GitLabResolver) GetTaskLink(entry *timeentry.TimeEntry) string {
	if !gl.Handles(entry) || entry.TaskID == "" || gl.Config.APIRoot == "" {
		return ""
	}
	match := gitlabAPIRootRe.FindStringSubmatch(gl.Config.APIRoot)
	if match == nil {
		return ""
	}
	root := match[1]
	taskType, number := gl.taskTypeAndNumber(string(entry.TaskID))
	return fmt.Sprintf("%s/%s/%s/%s", root, entry.Project, taskType, number)
}

func (gl *GitLabResolver) FetchDescription(ctx context.Context, entry *timeentry.TimeEntry) (string, error) {
	if !gl.Handles(entry) || entry.TaskID == "" {
		return "", nil
	}
	if gl.Config.APIKey == "" || gl.Config.APIRoot == "" {
		return "", nil
	}

	taskType, number := gl.taskTypeAndNumber(string(entry.TaskID))
	escapedProject := url.PathEscape(string(entry.Project))
	escapedProject = strings.ReplaceAll(escapedProject, "/", "%2F")

	apiRoot := strings.TrimSuffix(gl.Config.APIRoot, "/")
	endpoint := fmt.Sprintf("%s/projects/%s/%s/%s", apiRoot, escapedProject, taskType, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", gl.Config.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gitlab API returned %d", resp.StatusCode)
	}

	var result struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Title, nil
}

func init() {
	Resolvers = append(Resolvers, &GitLabResolver{})
}
