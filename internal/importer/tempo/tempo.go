// Package tempo imports time entries from a Tempo (Jira time tracking) CSV export.
package tempo

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/importer"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

func init() {
	importer.Importers = append(importer.Importers, &Importer{})
}

type Importer struct{}

func (i *Importer) Name() string { return "tempo" }

func (i *Importer) Import(ctx context.Context, _ *config.Config, path string) (*importer.ImportResult, error) {
	log := zerolog.Ctx(ctx)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &importer.ImportResult{}, nil
	}

	// Build column-name to index map; strip UTF-8 BOM from first field.
	header := records[0]
	header[0] = strings.TrimPrefix(header[0], "\xef\xbb\xbf")
	col := make(map[string]int, len(header))
	for idx, name := range header {
		col[name] = idx
	}

	result := &importer.ImportResult{
		ItemDetails: make(map[timeentry.AggregatedTimeKey]string),
	}

	for _, rec := range records[1:] {
		workDate, err := time.ParseInLocation("2006-01-02 15:04", rec[col["Work date"]], time.Local)
		if err != nil {
			log.Warn().Str("work_date", rec[col["Work date"]]).Err(err).Msg("skipping row")
			continue
		}

		loggedSecs, err := strconv.Atoi(rec[col["Logged Seconds"]])
		if err != nil {
			log.Warn().Str("logged_seconds", rec[col["Logged Seconds"]]).Err(err).Msg("skipping row with invalid logged seconds")
			continue
		}

		issueKey := rec[col["Issue Key"]]
		projectKey := rec[col["Project Key"]]
		issueSummary := rec[col["Issue summary"]]

		// Store just the numeric part of the issue key (e.g. "276" not "CCDEV-276").
		// The Jira resolver reconstructs the full key for display.
		taskID := strings.TrimPrefix(issueKey, projectKey+"-")

		if issueSummary != "" {
			key := timeentry.AggregatedTimeKey{Type: "jira", Project: timeentry.Project(projectKey), TaskID: timeentry.TaskID(taskID)}
			result.ItemDetails[key] = issueSummary
		}

		customer := timeentry.Customer(rec[col["Account Customer"]])
		if customer == "" {
			customer = timeentry.Customer(rec[col["Account Name"]])
		}

		start := types.TimeFrom(workDate)
		stopMins := min(start.Minutes()+loggedSecs/60, 23*60+59)
		stop := types.TimeFromMinutes(stopMins)
		date := types.NewDate(workDate.Year(), int(workDate.Month()), workDate.Day())

		result.Entries = append(result.Entries, importer.ImportEntry{
			Date: date,
			Entry: &timeentry.TimeEntry{
				Start:       start,
				Stop:        stop,
				Type:        "jira",
				Project:     timeentry.Project(projectKey),
				TaskID:      timeentry.TaskID(taskID),
				Customer:    customer,
				Description: rec[col["Work Description"]],
			},
		})
	}

	return result, nil
}
