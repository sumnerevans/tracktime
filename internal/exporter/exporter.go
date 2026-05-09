// Package exporter defines the Exporter interface for pushing time entries to external services.
package exporter

import (
	"context"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

// Exporter pushes aggregated time data to an external service.
type Exporter interface {
	Name() string
	Init(cfg config.SyncConfig)
	Export(ctx context.Context, aggregatedTime, syncedTime timeentry.AggregatedTime, month types.Month) (timeentry.AggregatedTime, error)
}

// Exporters is the global list of registered exporters.
var Exporters []Exporter
