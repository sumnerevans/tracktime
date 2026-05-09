package timeentry

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/types"
)

type EntryList struct {
	Date    types.Date
	Config  *config.Config
	Entries []*TimeEntry
}

func DayFilename(config *config.Config, date types.Date) string {
	return filepath.Join(
		config.Directory.Expand(),
		date.Format("2006"),
		date.Format("01"),
		date.Format("02"),
	)
}

func EntryListForDay(config *config.Config, date types.Date) (*EntryList, error) {
	dayFile := DayFilename(config, date)
	if err := os.MkdirAll(filepath.Dir(dayFile), 0755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(dayFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	el := EntryList{Date: date, Config: config}
	reader := csv.NewReader(file)
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if i == 0 {
			i++
			continue
		}

		entry, err := NewEntryFromRecord(i, record)
		if err != nil {
			return nil, err
		}
		el.Entries = append(el.Entries, entry)
		i++
	}

	return &el, nil
}

func (el *EntryList) EntriesForCustomer(customer Customer) []*TimeEntry {
	if len(customer) == 0 {
		return el.Entries
	}

	entries := []*TimeEntry{}
	for _, e := range el.Entries {
		if e.Customer == customer {
			entries = append(entries, e)
		}
	}
	return entries
}

func (el *EntryList) TotalTimeForCustomer(customer Customer) time.Duration {
	minutes := time.Duration(0)
	for _, e := range el.Entries {
		if len(customer) == 0 || e.Customer == customer {
			duration, _ := e.Duration(true)
			minutes += duration
		}
	}
	return minutes
}

func (el *EntryList) AddEntry(entry *TimeEntry) {
	insertIdx := len(el.Entries)
	for i, e := range el.Entries {
		if e.Stop != nil && entry.Start.Between(e.Start, e.Stop) {
			// The entry is being started in the middle of this one
			entry.Stop = e.Stop
			e.Stop = entry.Start
			insertIdx = i + 1
			break
		}

		if entry.Start.Before(e.Start) {
			// The entry is being started before this.
			entry.Stop = e.Start
			insertIdx = i
			break
		}

		if e.Start.Before(entry.Start) && e.Stop == nil {
			// There is an unended time entry. Stop it, and start the new one.
			e.Stop = entry.Start
			insertIdx = i + 1
			break
		}
	}

	newEntries := make([]*TimeEntry, len(el.Entries)+1)
	for i := range newEntries {
		if i < insertIdx {
			newEntries[i] = el.Entries[i]
		} else if i == insertIdx {
			newEntries[i] = entry
		} else {
			newEntries[i] = el.Entries[i-1]
		}
	}
	el.Entries = newEntries
}

func (el *EntryList) Save() error {
	dayFile := DayFilename(el.Config, el.Date)
	if err := os.MkdirAll(filepath.Dir(dayFile), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(dayFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.Write([]string{"start", "stop", "type", "project", "taskid", "customer", "description"})
	if err != nil {
		return err
	}

	for _, e := range el.Entries {
		err = writer.Write([]string{
			e.Start.String(),
			e.Stop.String(),
			string(e.Type),
			string(e.Project),
			string(e.TaskID),
			string(e.Customer),
			e.Description,
		})
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func (el *EntryList) Sync() error {
	// TODO
	return nil
}

func (el *EntryList) SaveAndSync() error {
	err := el.Save()
	if err != nil {
		return err
	}
	return el.Sync()
}

func (el *EntryList) Start(start *types.Time, description string, taskType TimeEntryType, project Project, customer Customer, taskID TaskID) error {
	newEntry := &TimeEntry{
		Start:       start,
		Type:        taskType,
		Project:     project,
		Customer:    customer,
		TaskID:      taskID,
		Description: description,
	}
	el.AddEntry(newEntry)
	return el.SaveAndSync()
}

func (el *EntryList) Stop(stop *types.Time) error {
	if len(el.Entries) == 0 || el.Entries[len(el.Entries)-1].Stop != nil {
		return fmt.Errorf("no time entry to stop")
	}
	el.Entries[len(el.Entries)-1].Stop = stop
	return el.SaveAndSync()
}

func (el *EntryList) Resume(resumeIndex int, description *string, start *types.Time) error {
	var oldEntry *TimeEntry
	if resumeIndex == -1 {
		if len(el.Entries) > 0 {
			oldEntry = el.Entries[len(el.Entries)-1]
		} else {
			yesterdayEntries, err := EntryListForDay(el.Config, el.Date.AddDays(-1))
			if err != nil {
				return err
			}
			if len(yesterdayEntries.Entries) == 0 {
				return fmt.Errorf("no time entry to resume")
			}
			oldEntry = yesterdayEntries.Entries[len(yesterdayEntries.Entries)-1]
		}
	} else {
		oldEntry = el.Entries[resumeIndex-1]
	}

	if description == nil {
		description = &oldEntry.Description
	}
	return el.Start(start, *description, oldEntry.Type, oldEntry.Project, oldEntry.Customer, oldEntry.TaskID)
}
