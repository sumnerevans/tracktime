package timeentry_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

func TestTimeEntryDuration(t *testing.T) {
	start := mustParseTime(t, "09:00")
	stop := mustParseTime(t, "12:30")

	t.Run("with stop time", func(t *testing.T) {
		entry := &timeentry.TimeEntry{
			Start: start,
			Stop:  stop,
		}
		duration, err := entry.Duration(false)
		assert.NoError(t, err)
		assert.Equal(t, 3*time.Hour+30*time.Minute, duration)
	})

	t.Run("without stop time, not allowed", func(t *testing.T) {
		entry := &timeentry.TimeEntry{
			Start: start,
			Stop:  nil,
		}
		_, err := entry.Duration(false)
		assert.Error(t, err)
	})

	t.Run("without stop time, allowed", func(t *testing.T) {
		entry := &timeentry.TimeEntry{
			Start: start,
			Stop:  nil,
		}
		duration, err := entry.Duration(true)
		assert.NoError(t, err)
		assert.Positive(t, duration)
	})
}

func TestEntryListForDay(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Directory: types.Filename(tmpDir),
	}
	date := types.NewDate(2023, 1, 15)

	t.Run("empty file", func(t *testing.T) {
		el, err := timeentry.EntryListForDay(cfg, date)
		require.NoError(t, err)
		assert.Empty(t, el.Entries)
	})

	t.Run("file with entries", func(t *testing.T) {
		// Create a day file with test data
		dayFile := timeentry.DayFilename(cfg, date)
		os.MkdirAll(filepath.Dir(dayFile), 0755)
		csvContent := `start,stop,type,project,taskid,customer,description
09:00,12:30,github,myproject,123,ACME Corp,Fixing bug
13:30,17:00,gitlab,otherproject,456,Client B,Feature work
`
		require.NoError(t, os.WriteFile(dayFile, []byte(csvContent), 0644))

		el, err := timeentry.EntryListForDay(cfg, date)
		require.NoError(t, err)

		expected := []*timeentry.TimeEntry{
			{
				Index:       1,
				Start:       mustParseTime(t, "09:00"),
				Stop:        mustParseTime(t, "12:30"),
				Type:        "github",
				Project:     "myproject",
				TaskID:      "123",
				Customer:    "ACME Corp",
				Description: "Fixing bug",
			},
			{
				Index:       2,
				Start:       mustParseTime(t, "13:30"),
				Stop:        mustParseTime(t, "17:00"),
				Type:        "gitlab",
				Project:     "otherproject",
				TaskID:      "456",
				Customer:    "Client B",
				Description: "Feature work",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})
}

func TestEntriesForCustomer(t *testing.T) {
	el := &timeentry.EntryList{
		Entries: []*timeentry.TimeEntry{
			{Customer: "ACME Corp", Description: "Task 1"},
			{Customer: "Client B", Description: "Task 2"},
			{Customer: "ACME Corp", Description: "Task 3"},
		},
	}

	t.Run("filter by customer", func(t *testing.T) {
		filtered := el.EntriesForCustomer("ACME Corp")
		assert.EqualValues(t, []*timeentry.TimeEntry{
			{Customer: "ACME Corp", Description: "Task 1"},
			{Customer: "ACME Corp", Description: "Task 3"},
		}, filtered)
	})

	t.Run("empty customer returns all", func(t *testing.T) {
		filtered := el.EntriesForCustomer("")
		assert.EqualValues(t, el.Entries, filtered)
	})
}

func TestTotalTimeForCustomer(t *testing.T) {
	el := &timeentry.EntryList{
		Entries: []*timeentry.TimeEntry{
			{Customer: "ACME Corp", Start: mustParseTime(t, "09:00"), Stop: mustParseTime(t, "10:00")}, // 1 hour
			{Customer: "Client B", Start: mustParseTime(t, "11:00"), Stop: mustParseTime(t, "12:30")},  // 1.5 hours
			{Customer: "ACME Corp", Start: mustParseTime(t, "13:00"), Stop: mustParseTime(t, "14:00")}, // 1 hour
		},
	}

	t.Run("specific customer", func(t *testing.T) {
		total := el.TotalTimeForCustomer("ACME Corp")
		assert.Equal(t, 2*time.Hour, total)
	})

	t.Run("all customers", func(t *testing.T) {
		total := el.TotalTimeForCustomer("")
		assert.Equal(t, 3*time.Hour+30*time.Minute, total)
	})
}

func TestAddEntry(t *testing.T) {
	t.Run("add to empty list", func(t *testing.T) {
		el := &timeentry.EntryList{Entries: []*timeentry.TimeEntry{}}
		newEntry := &timeentry.TimeEntry{
			Start:       mustParseTime(t, "09:00"),
			Description: "New task",
		}
		el.AddEntry(newEntry)

		expected := []*timeentry.TimeEntry{
			{
				Start:       mustParseTime(t, "09:00"),
				Stop:        nil,
				Description: "New task",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})

	t.Run("add after existing entry", func(t *testing.T) {
		el := &timeentry.EntryList{
			Entries: []*timeentry.TimeEntry{
				{Start: mustParseTime(t, "09:00"), Stop: mustParseTime(t, "10:00")},
			},
		}
		newEntry := &timeentry.TimeEntry{
			Start:       mustParseTime(t, "11:00"),
			Description: "New task",
		}
		el.AddEntry(newEntry)

		expected := []*timeentry.TimeEntry{
			{
				Start: mustParseTime(t, "09:00"),
				Stop:  mustParseTime(t, "10:00"),
			},
			{
				Start:       mustParseTime(t, "11:00"),
				Stop:        nil,
				Description: "New task",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})

	t.Run("auto-stop unended entry", func(t *testing.T) {
		el := &timeentry.EntryList{
			Entries: []*timeentry.TimeEntry{
				{Start: mustParseTime(t, "09:00"), Stop: nil},
			},
		}
		newEntry := &timeentry.TimeEntry{
			Start:       mustParseTime(t, "11:00"),
			Description: "New task",
		}
		el.AddEntry(newEntry)

		expected := []*timeentry.TimeEntry{
			{
				Start: mustParseTime(t, "09:00"),
				Stop:  mustParseTime(t, "11:00"),
			},
			{
				Start:       mustParseTime(t, "11:00"),
				Stop:        nil,
				Description: "New task",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})

	t.Run("insert in middle splits existing entry", func(t *testing.T) {
		el := &timeentry.EntryList{
			Entries: []*timeentry.TimeEntry{
				{Start: mustParseTime(t, "09:00"), Stop: mustParseTime(t, "12:00")},
			},
		}
		newEntry := &timeentry.TimeEntry{
			Start:       mustParseTime(t, "10:00"),
			Description: "Inserted task",
		}
		el.AddEntry(newEntry)

		expected := []*timeentry.TimeEntry{
			{
				Start: mustParseTime(t, "09:00"),
				Stop:  mustParseTime(t, "10:00"),
			},
			{
				Start:       mustParseTime(t, "10:00"),
				Stop:        mustParseTime(t, "12:00"),
				Description: "Inserted task",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})

	t.Run("insert before existing entry", func(t *testing.T) {
		el := &timeentry.EntryList{
			Entries: []*timeentry.TimeEntry{
				{Start: mustParseTime(t, "11:00"), Stop: mustParseTime(t, "12:00")},
			},
		}
		newEntry := &timeentry.TimeEntry{
			Start:       mustParseTime(t, "09:00"),
			Description: "Earlier task",
		}
		el.AddEntry(newEntry)

		expected := []*timeentry.TimeEntry{
			{
				Start:       mustParseTime(t, "09:00"),
				Stop:        mustParseTime(t, "11:00"),
				Description: "Earlier task",
			},
			{
				Start: mustParseTime(t, "11:00"),
				Stop:  mustParseTime(t, "12:00"),
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Directory: types.Filename(tmpDir),
	}
	date := types.NewDate(2023, 1, 15)

	el := &timeentry.EntryList{
		Date:   date,
		Config: cfg,
		Entries: []*timeentry.TimeEntry{
			{
				Start:       mustParseTime(t, "09:00"),
				Stop:        mustParseTime(t, "12:30"),
				Type:        "github",
				Project:     "myproject",
				TaskID:      "123",
				Customer:    "ACME Corp",
				Description: "Fixing bug",
			},
			{
				Start:       mustParseTime(t, "13:30"),
				Stop:        nil,
				Type:        "gitlab",
				Project:     "otherproject",
				TaskID:      "456",
				Customer:    "Client B",
				Description: "Feature work",
			},
		},
	}

	err := el.Save()
	require.NoError(t, err)

	// Read back the file and verify
	el2, err := timeentry.EntryListForDay(cfg, date)
	require.NoError(t, err)
	assert.Len(t, el2.Entries, 2)

	// Verify first entry
	assert.Equal(t, "09:00", el2.Entries[0].Start.String())
	assert.Equal(t, "12:30", el2.Entries[0].Stop.String())
	assert.Equal(t, "Fixing bug", el2.Entries[0].Description)

	// Verify second entry (with nil stop)
	assert.Nil(t, el2.Entries[1].Stop)
}

func TestStop(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Directory: types.Filename(tmpDir),
	}
	date := types.NewDate(2023, 1, 15)

	t.Run("stop running entry", func(t *testing.T) {
		el := &timeentry.EntryList{
			Date:   date,
			Config: cfg,
			Entries: []*timeentry.TimeEntry{
				{Start: mustParseTime(t, "09:00"), Stop: nil},
			},
		}

		err := el.Stop(mustParseTime(t, "12:00"))
		assert.NoError(t, err)
		assert.NotNil(t, el.Entries[0].Stop)
		assert.Equal(t, "12:00", el.Entries[0].Stop.String())
	})

	t.Run("no entry to stop", func(t *testing.T) {
		el := &timeentry.EntryList{
			Date:    date,
			Config:  cfg,
			Entries: []*timeentry.TimeEntry{},
		}

		err := el.Stop(mustParseTime(t, "12:00"))
		assert.Error(t, err)
	})

	t.Run("last entry already stopped", func(t *testing.T) {
		el := &timeentry.EntryList{
			Date:   date,
			Config: cfg,
			Entries: []*timeentry.TimeEntry{
				{Start: mustParseTime(t, "09:00"), Stop: mustParseTime(t, "10:00")},
			},
		}

		err := el.Stop(mustParseTime(t, "12:00"))
		assert.Error(t, err)
	})
}

func TestResume(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Directory: types.Filename(tmpDir),
	}
	date := types.NewDate(2023, 1, 15)

	t.Run("resume last entry from same day", func(t *testing.T) {
		el := &timeentry.EntryList{
			Date:   date,
			Config: cfg,
			Entries: []*timeentry.TimeEntry{
				{
					Start:       mustParseTime(t, "09:00"),
					Stop:        mustParseTime(t, "10:00"),
					Type:        "github",
					Project:     "myproject",
					Customer:    "ACME",
					TaskID:      "123",
					Description: "Original task",
				},
			},
		}

		err := el.Resume(-1, nil, mustParseTime(t, "11:00"))
		assert.NoError(t, err)

		expected := []*timeentry.TimeEntry{
			{
				Start:       mustParseTime(t, "09:00"),
				Stop:        mustParseTime(t, "10:00"),
				Type:        "github",
				Project:     "myproject",
				Customer:    "ACME",
				TaskID:      "123",
				Description: "Original task",
			},
			{
				Start:       mustParseTime(t, "11:00"),
				Stop:        nil,
				Type:        "github",
				Project:     "myproject",
				Customer:    "ACME",
				TaskID:      "123",
				Description: "Original task",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})

	t.Run("resume specific entry by index", func(t *testing.T) {
		el := &timeentry.EntryList{
			Date:   date,
			Config: cfg,
			Entries: []*timeentry.TimeEntry{
				{
					Start:       mustParseTime(t, "09:00"),
					Stop:        mustParseTime(t, "10:00"),
					Type:        "gitlab",
					Project:     "project1",
					Customer:    "Client A",
					TaskID:      "111",
					Description: "First task",
				},
				{
					Start:       mustParseTime(t, "11:00"),
					Stop:        mustParseTime(t, "12:00"),
					Type:        "github",
					Project:     "project2",
					Customer:    "Client B",
					TaskID:      "222",
					Description: "Second task",
				},
			},
		}

		// Resume the first entry (index 1, since indices are 1-based in the UI)
		err := el.Resume(1, nil, mustParseTime(t, "13:00"))
		assert.NoError(t, err)

		expected := []*timeentry.TimeEntry{
			{
				Start:       mustParseTime(t, "09:00"),
				Stop:        mustParseTime(t, "10:00"),
				Type:        "gitlab",
				Project:     "project1",
				Customer:    "Client A",
				TaskID:      "111",
				Description: "First task",
			},
			{
				Start:       mustParseTime(t, "11:00"),
				Stop:        mustParseTime(t, "12:00"),
				Type:        "github",
				Project:     "project2",
				Customer:    "Client B",
				TaskID:      "222",
				Description: "Second task",
			},
			{
				Start:       mustParseTime(t, "13:00"),
				Stop:        nil,
				Type:        "gitlab",
				Project:     "project1",
				Customer:    "Client A",
				TaskID:      "111",
				Description: "First task",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})

	t.Run("resume with custom description", func(t *testing.T) {
		el := &timeentry.EntryList{
			Date:   date,
			Config: cfg,
			Entries: []*timeentry.TimeEntry{
				{
					Start:       mustParseTime(t, "09:00"),
					Stop:        mustParseTime(t, "10:00"),
					Type:        "github",
					Project:     "myproject",
					Customer:    "ACME",
					TaskID:      "123",
					Description: "Original task",
				},
			},
		}

		customDesc := "Custom description"
		err := el.Resume(-1, &customDesc, mustParseTime(t, "11:00"))
		assert.NoError(t, err)

		expected := []*timeentry.TimeEntry{
			{
				Start:       mustParseTime(t, "09:00"),
				Stop:        mustParseTime(t, "10:00"),
				Type:        "github",
				Project:     "myproject",
				Customer:    "ACME",
				TaskID:      "123",
				Description: "Original task",
			},
			{
				Start:       mustParseTime(t, "11:00"),
				Stop:        nil,
				Type:        "github",
				Project:     "myproject",
				Customer:    "ACME",
				TaskID:      "123",
				Description: "Custom description",
			},
		}
		assert.EqualValues(t, expected, el.Entries)
	})
}

func mustParseTime(t *testing.T, s string) *types.Time {
	t.Helper()
	time, err := types.ParseTime(s)
	require.NoError(t, err, "Failed to parse time %q", s)
	return time
}
