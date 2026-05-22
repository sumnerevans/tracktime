package resolver

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

type JiraResolver struct{}

var _ ItemDetailResolver = (*JiraResolver)(nil)

func (r *JiraResolver) Init(_ *config.Config) {}

func (r *JiraResolver) Handles(entry *timeentry.TimeEntry) bool {
	return entry.Type == "jira"
}

// GetFormattedTaskID combines project and task number into the standard Jira
// issue key format, e.g. "CCDEV" + "276" → "CCDEV-276".
func (r *JiraResolver) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	return string(entry.Project) + "-" + string(entry.TaskID)
}

func (r *JiraResolver) GetTaskLink(_ *timeentry.TimeEntry) string {
	return ""
}

// FetchDescription is a no-op: Jira descriptions are seeded from the Tempo import.
func (r *JiraResolver) FetchDescription(_ context.Context, _ *timeentry.TimeEntry) (string, error) {
	return "", nil
}

func init() {
	Resolvers = append(Resolvers, &JiraResolver{})
}
