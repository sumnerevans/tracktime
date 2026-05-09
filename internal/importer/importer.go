// Package importer defines the Importer interface for pulling time entries from external services.
package importer

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

// Importer pulls time data from an external service into tracktime.
type Importer interface {
	Name() string
	Init(cfg config.SyncConfig)
	Import(ctx context.Context, month types.Month) (timeentry.AggregatedTime, error)
}

// Importers is the global list of registered importers.
var Importers []Importer
