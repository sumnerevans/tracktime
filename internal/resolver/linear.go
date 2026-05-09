package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

type LinearResolver struct {
	Config config.LinearConfig
}

type LinearResult struct {
	Data struct {
		Issue struct {
			Title string `json:"title"`
		} `json:"issue"`
	} `json:"data"`
}

var _ ItemDetailResolver = (*LinearResolver)(nil)

func (l *LinearResolver) Init(cfg *config.Config) { l.Config = cfg.Linear }

func (l *LinearResolver) Handles(entry *timeentry.TimeEntry) bool {
	return entry.Type == "linear"
}

func (l *LinearResolver) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	if !l.Handles(entry) || entry.Project == "" || entry.TaskID == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", entry.Project, entry.TaskID)
}

func (l *LinearResolver) GetTaskLink(entry *timeentry.TimeEntry) string {
	if !l.Handles(entry) || entry.Project == "" || entry.TaskID == "" {
		return ""
	}
	if l.Config.DefaultOrg == "" {
		return ""
	}
	return fmt.Sprintf("https://linear.app/%s/issue/%s-%s", l.Config.DefaultOrg, entry.Project, entry.TaskID)
}

func (l *LinearResolver) FetchDescription(ctx context.Context, entry *timeentry.TimeEntry) (string, error) {
	if !l.Handles(entry) || entry.TaskID == "" {
		return "", nil
	}
	if l.Config.APIKey == "" || l.Config.DefaultOrg == "" {
		return "", nil
	}
	taskID := l.GetFormattedTaskID(entry)
	if taskID == "" {
		return "", nil
	}

	body, err := json.Marshal(map[string]string{
		"query": fmt.Sprintf(`{issue (id: "%s") { title }}`, taskID),
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linear.app/graphql", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", l.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result LinearResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.Issue.Title, err
}

func init() {
	Resolvers = append(Resolvers, &LinearResolver{})
}
