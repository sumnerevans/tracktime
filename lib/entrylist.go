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

type TimeEntry struct {
	Index       int
	Start       *Time
	Stop        *Time
	Type        TimeEntryType
	Project     string
	Customer    string
	TaskID      string
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
	te.Project = record[3]
	te.TaskID = record[4]
	te.Customer = record[5]
	te.Description = record[6]
	return
}

func (te *TimeEntry) Duration(allowUnended bool) (time.Duration, error) {
	stop := te.Stop
	if stop == nil {
		if allowUnended {
			stop = CurrentTime()
		} else {
			return time.Duration(0), fmt.Errorf("Unended time entries cannot have a duration.")
		}
	}
	return stop.Sub(te.Start), nil
}

type EntryList struct {
	Date    Date
	Config  *Config
	entries []*TimeEntry
}

func EntryListForDay(config *Config, date Date) (*EntryList, error) {
	el := EntryList{Date: date, Config: config}
	filename := filepath.Join(
		el.Config.Directory.Expand(),
		el.Date.Format("2006"),
		el.Date.Format("01"),
		el.Date.Format("02"),
	)

	file, err := os.OpenFile(filename, os.O_CREATE, 0x700)
	if err != nil {
		return nil, err
	}

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

func (el *EntryList) EntriesForCustomer(customer string) []*TimeEntry {
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

func (el *EntryList) TotalTimeForCustomer(customer string) time.Duration {
	minutes := time.Duration(0)
	for _, e := range el.entries {
		if len(customer) == 0 || e.Customer == customer {
			duration, _ := e.Duration(true)
			minutes += duration
		}
	}
	return minutes
}
