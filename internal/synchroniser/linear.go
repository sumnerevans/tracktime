package synchroniser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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
	if entry.Type != "linear" || entry.TaskID == "" {
		return ""
	}
	if l.Config.APIKey == "" || l.Config.DefaultOrg == "" {
		return ""
	}

	taskID := l.GetFormattedTaskID(entry)
	if taskID == "" {
		return ""
	}

	// Load cache
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "tracktime")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return ""
	}
	cacheFile := filepath.Join(cacheDir, "linear.json")

	cache := map[string]string{}
	if data, err := os.ReadFile(cacheFile); err == nil {
		_ = json.Unmarshal(data, &cache)
	}

	if desc := cache[taskID]; desc != "" {
		return desc
	}

	// Query Linear GraphQL API
	body, err := json.Marshal(map[string]string{
		"query": fmt.Sprintf(`{issue (id: "%s") { title }}`, taskID),
	})
	if err != nil {
		return ""
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linear.app/graphql", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", l.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Issue struct {
				Title string `json:"title"`
			} `json:"issue"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	title := result.Data.Issue.Title
	if title != "" {
		cache[taskID] = title
		if data, err := json.Marshal(cache); err == nil {
			_ = os.WriteFile(cacheFile, data, 0o644)
		}
	}

	return title
}

func init() {
	Synchronisers = append(Synchronisers, &LinearSynchroniser{})
}
