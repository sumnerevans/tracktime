// Package resolver provides interfaces and implementations for resolving work
// item metadata.
package resolver

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

// ItemDetailResolver resolves work item metadata (formatted ID, link, description)
// for a specific external service.
type ItemDetailResolver interface {
	Init(cfg *config.Config)
	Handles(entry *timeentry.TimeEntry) bool
	GetFormattedTaskID(entry *timeentry.TimeEntry) string
	GetTaskLink(entry *timeentry.TimeEntry) string
	FetchDescription(ctx context.Context, entry *timeentry.TimeEntry) (string, error)
}

// Resolvers is the global list of registered item resolvers.
var Resolvers []ItemDetailResolver
