package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

type SourcehutResolver struct {
	Config config.SourceHutConfig
}

var _ ItemDetailResolver = (*SourcehutResolver)(nil)

var sourcehutAPIRootRe = regexp.MustCompile(`(.*)/api/?$`)

func (sh *SourcehutResolver) Init(cfg *config.Config) { sh.Config = cfg.SourceHut }

func (sh *SourcehutResolver) Handles(entry *timeentry.TimeEntry) bool {
	t := string(entry.Type)
	return t == "srht" || t == "sr.ht" || t == "sh" || t == "sourcehut"
}

func (sh *SourcehutResolver) extractUsernameAndTracker(project string) (username, tracker string) {
	if strings.Contains(project, "/") {
		parts := strings.SplitN(project, "/", 2)
		username, tracker = parts[0], parts[1]
	} else {
		username = sh.Config.Username
		tracker = project
	}
	if username != "" && !strings.HasPrefix(username, "~") {
		username = "~" + username
	}
	return
}

func (sh *SourcehutResolver) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	if !sh.Handles(entry) || entry.TaskID == "" {
		return ""
	}
	taskID := string(entry.TaskID)
	if strings.HasPrefix(taskID, "#") {
		return taskID
	}
	return "#" + taskID
}

func (sh *SourcehutResolver) GetTaskLink(entry *timeentry.TimeEntry) string {
	if !sh.Handles(entry) || entry.TaskID == "" || sh.Config.APIRoot == "" {
		return ""
	}
	match := sourcehutAPIRootRe.FindStringSubmatch(sh.Config.APIRoot)
	if match == nil {
		return ""
	}
	root := match[1]
	username, tracker := sh.extractUsernameAndTracker(string(entry.Project))
	taskNum := strings.TrimPrefix(string(entry.TaskID), "#")
	return fmt.Sprintf("%s/%s/%s/%s", root, username, tracker, taskNum)
}

func (sh *SourcehutResolver) FetchDescription(ctx context.Context, entry *timeentry.TimeEntry) (string, error) {
	if !sh.Handles(entry) || entry.TaskID == "" {
		return "", nil
	}
	if sh.Config.AccessToken == "" || sh.Config.APIRoot == "" {
		return "", nil
	}

	username, tracker := sh.extractUsernameAndTracker(string(entry.Project))
	if username == "" || tracker == "" {
		return "", nil
	}
	ticketID := strings.TrimPrefix(string(entry.TaskID), "#")

	apiRoot := strings.TrimSuffix(sh.Config.APIRoot, "/")
	endpoint := fmt.Sprintf("%s/user/%s/trackers/%s/tickets/%s", apiRoot, username, tracker, ticketID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+sh.Config.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sourcehut API returned %d", resp.StatusCode)
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
	Resolvers = append(Resolvers, &SourcehutResolver{})
}
