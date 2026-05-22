// Package timeentry defines the TimeEntry struct and related functions for
// parsing and handling time entries.
package timeentry

import (
	"fmt"
	"strings"
	"time"

	"github.com/sumnerevans/tracktime/internal/types"
)

type TimeEntryType string

func (t TimeEntryType) String() string { return string(t) }

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

func (c Customer) String() string { return string(c) }
func (p Project) String() string  { return string(p) }
func (t TaskID) String() string   { return string(t) }

type TimeEntry struct {
	Index       int
	Start       *types.Time
	Stop        *types.Time
	Type        TimeEntryType
	Project     Project
	Customer    Customer
	TaskID      TaskID
	Description string
}

func NewEntryFromRecord(idx int, record []string) (te *TimeEntry, err error) {
	te = &TimeEntry{Index: idx}
	te.Start, err = types.ParseTime(record[0])
	if err != nil {
		return
	}
	te.Stop, err = types.ParseTime(record[1])
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
			stop = types.CurrentTime()
		} else {
			return time.Duration(0), fmt.Errorf("unended time entries cannot have a duration")
		}
	}
	return stop.Sub(te.Start), nil
}
