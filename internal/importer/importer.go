// Package importer defines the Importer interface for pulling time entries from files.
package importer

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

// ImportEntry is a single time entry to be imported, paired with its date.
type ImportEntry struct {
	Date  types.Date
	Entry *timeentry.TimeEntry
}

// ImportResult is returned by an Importer. The framework handles writing
// entries to disk (append-only, deduped by start time) and seeding the cache.
type ImportResult struct {
	Entries     []ImportEntry
	ItemDetails map[timeentry.AggregatedTimeKey]string
}

// Importer parses a file and returns the entries and item details it contains.
// It must not write any files itself.
type Importer interface {
	Name() string
	Import(ctx context.Context, cfg *config.Config, path string) (*ImportResult, error)
}

// Importers is the global list of registered importers.
var Importers []Importer
