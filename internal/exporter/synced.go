package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

func syncedFilePath(dir string, month types.Month) string {
	return filepath.Join(dir, fmt.Sprintf("%04d", month.Year()), fmt.Sprintf("%02d", int(month.Month())), ".synced")
}

// ReadSyncedFile reads the .synced CSV for month from dir.
// Returns an empty AggregatedTime if the file does not exist.
func ReadSyncedFile(dir string, month types.Month) (timeentry.AggregatedTime, error) {
	synced := make(timeentry.AggregatedTime)
	path := syncedFilePath(dir, month)

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return synced, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		if i == 0 || len(record) < 4 {
			continue
		}
		seconds, err := strconv.ParseInt(record[3], 10, 64)
		if err != nil {
			continue
		}
		key := timeentry.AggregatedTimeKey{
			Type:    timeentry.TimeEntryType(record[0]),
			Project: timeentry.Project(record[1]),
			TaskID:  timeentry.TaskID(record[2]),
		}
		synced[key] = time.Duration(seconds) * time.Second
	}

	return synced, nil
}

// WriteSyncedFile writes synced to the .synced CSV for month under dir.
func WriteSyncedFile(dir string, month types.Month, synced timeentry.AggregatedTime) error {
	path := syncedFilePath(dir, month)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"type", "project", "taskid", "synced"}); err != nil {
		return err
	}
	for key, duration := range synced {
		if err := w.Write([]string{
			string(key.Type),
			string(key.Project),
			string(key.TaskID),
			strconv.FormatInt(int64(duration.Seconds()), 10),
		}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
