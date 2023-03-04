package lib

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TimeEntryType string

func ParseTimeEntryType(typeStr string) TimeEntryType {
	switch strings.ToLower(typeStr) {
	case "gh":
		return "github"
	case "gl":
		return "gitlab"
	default:
		return TimeEntryType(typeStr)
	}
}

type Customer string
type Project string
type TaskID string

type TimeEntry struct {
	Index       int
	Start       *Time
	Stop        *Time
	Type        TimeEntryType
	Project     Project
	Customer    Customer
	TaskID      TaskID
	Description string
}

func NewEntryFromRecord(idx int, record []string) (te *TimeEntry, err error) {
	te = &TimeEntry{Index: idx}
	te.Start, err = ParseTime(record[0])
	if err != nil {
		return
	}
	te.Stop, err = ParseTime(record[1])
	if err != nil {
		return
	}
	te.Type = ParseTimeEntryType(record[2])
	te.Project = Project(record[3])
	te.TaskID = TaskID(record[4])
	te.Customer = Customer(record[5])
	te.Description = record[6]
	return
}

func (te *TimeEntry) Duration(allowUnended bool) (time.Duration, error) {
	stop := te.Stop
	if stop == nil {
		if allowUnended {
			stop = CurrentTime()
		} else {
			return time.Duration(0), fmt.Errorf("unended time entries cannot have a duration")
		}
	}
	return stop.Sub(te.Start), nil
}

type EntryList struct {
	Date    Date
	Config  *Config
	entries []*TimeEntry
}

func DayFilename(config *Config, date Date) string {
	return filepath.Join(
		config.Directory.Expand(),
		date.Format("2006"),
		date.Format("01"),
		date.Format("02"),
	)
}

func EntryListForDay(config *Config, date Date) (*EntryList, error) {
	file, err := os.OpenFile(DayFilename(config, date), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

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
		el.entries = append(el.entries, entry)
		i++
	}

	return &el, nil
}

func (el *EntryList) EntriesForCustomer(customer Customer) []*TimeEntry {
	if len(customer) == 0 {
		return el.entries
	}

	entries := []*TimeEntry{}
	for _, e := range el.entries {
		if e.Customer == customer {
			entries = append(entries, e)
		}
	}
	return entries
}

func (el *EntryList) TotalTimeForCustomer(customer Customer) time.Duration {
	minutes := time.Duration(0)
	for _, e := range el.entries {
		if len(customer) == 0 || e.Customer == customer {
			duration, _ := e.Duration(true)
			minutes += duration
		}
	}
	return minutes
}

func (el *EntryList) AddEntry(entry *TimeEntry) {
	insertIdx := len(el.entries) + 1
	for i, e := range el.entries {
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

	newEntries := make([]*TimeEntry, len(el.entries)+1)
	for i := 0; i < len(newEntries); i++ {
		if i < insertIdx {
			newEntries[i] = el.entries[i]
		} else if i == insertIdx {
			newEntries[i] = entry
		} else {
			newEntries[i] = el.entries[i-1]
		}
	}
	el.entries = newEntries
}

func (el *EntryList) Save() error {
	file, err := os.OpenFile(DayFilename(el.Config, el.Date), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	err = writer.Write([]string{"start", "stop", "type", "project", "taskid", "customer", "description"})
	if err != nil {
		return err
	}

	for _, e := range el.entries {
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

func (el *EntryList) Start(start *Time, description string, taskType TimeEntryType, project Project, customer Customer, taskID TaskID) error {
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

func (el *EntryList) Stop(stop *Time) error {
	if len(el.entries) == 0 || el.entries[len(el.entries)-1].Stop != nil {
		return fmt.Errorf("no time entry to stop")
	}
	el.entries[len(el.entries)-1].Stop = stop
	return el.SaveAndSync()
}

func (el *EntryList) Resume(resumeIndex int, description *string, start *Time) error {
	var oldEntry *TimeEntry
	if resumeIndex == -1 {
		if len(el.entries) > 0 {
			oldEntry = el.entries[len(el.entries)-1]
		} else {
			yesterdayEntries, err := EntryListForDay(el.Config, el.Date.AddDays(-1))
			if err != nil {
				return err
			}
			if len(yesterdayEntries.entries) == 0 {
				return fmt.Errorf("no time entry to resume")
			}
			oldEntry = yesterdayEntries.entries[len(yesterdayEntries.entries)-1]
		}
	} else {
		oldEntry = el.entries[resumeIndex-1]
	}

	if description == nil {
		description = &oldEntry.Description
	}
	return el.Start(start, *description, oldEntry.Type, oldEntry.Project, oldEntry.Customer, oldEntry.TaskID)
}
