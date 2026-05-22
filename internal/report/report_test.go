package report

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/resolver"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

func createTestConfig(t *testing.T) *config.Config {
	t.Helper()
	tmpDir := t.TempDir()
	return &config.Config{
		Directory: types.Filename(tmpDir),
		Reporting: config.ReportingConfig{
			FullName:              "Test User",
			DayWorkedMinThreshold: 60,
			ProjectRates: map[string]float64{
				"project1": 100.0,
				"project2": 150.0,
			},
			CustomerRates: map[string]float64{
				"ACME Corp": 200.0,
			},
			CustomerAliases: map[string]string{
				"acme": "ACME Corporation",
			},
			CustomerAddresses: map[string]string{
				"ACME Corp": "123 Main St\nNew York, NY 10001",
			},
		},
	}
}

func mustParseTime(t *testing.T, s string) *types.Time {
	t.Helper()
	time, err := types.ParseTime(s)
	require.NoError(t, err, "Failed to parse time %q", s)
	return time
}

func createTestEntries(t *testing.T, cfg *config.Config, date types.Date, entries []struct {
	start, stop                                       string
	entryType, project, customer, taskID, description string
}) {
	t.Helper()
	el, err := timeentry.EntryListForDay(cfg, date)
	require.NoError(t, err, "Failed to create entry list")

	for _, e := range entries {
		el.Start(
			mustParseTime(t, e.start),
			e.description,
			timeentry.TimeEntryType(e.entryType),
			timeentry.Project(e.project),
			timeentry.Customer(e.customer),
			timeentry.TaskID(e.taskID),
		)
		if e.stop != "" {
			el.Stop(mustParseTime(t, e.stop))
		}
	}
}

func TestReportCreation(t *testing.T) {
	cfg := createTestConfig(t)
	date := types.Today()

	createTestEntries(t, cfg, date, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "10:00", "github", "project1", "ACME Corp", "123", "Task 1"},
		{"11:00", "12:30", "gitlab", "project2", "Client B", "456", "Task 2"},
		{"13:00", "15:00", "github", "project1", "ACME Corp", "789", "Task 3"},
	})

	el, err := timeentry.EntryListForDay(cfg, date)
	require.NoError(t, err)
	entries := el.EntriesForCustomer("")
	require.Len(t, entries, 3, "Should have 3 entries")

	report, err := New(context.Background(), cfg, date, date, "", "", SortAlphabetical, false, false, false)
	require.NoError(t, err)

	expectedCache := resolver.NewItemDetailCache(context.Background(), string(cfg.Directory), cfg, resolver.Resolvers)

	assert.EqualValues(t, &Report{
		ctx:       context.Background(),
		StartDate: date,
		EndDate:   date,
		Config:    cfg,
		AggregatedTime: map[CustomerProject]map[timeentry.TaskID]map[string][]*timeentry.TimeEntry{
			{Customer: "ACME Corp", Project: "project1"}: {
				"123": {"TASK 1": {entries[0]}},
				"789": {"TASK 3": {entries[2]}},
			},
			{Customer: "Client B", Project: "project2"}: {
				"456": {"TASK 2": {entries[1]}},
			},
		},
		DayStats: map[types.Date]time.Duration{
			date: 4*time.Hour + 30*time.Minute, // 1h + 1.5h + 2h
		},
		RateTotals: map[CustomerProject]RateTotal{
			{Customer: "ACME Corp", Project: "project1"}: {Rate: 100.0, Total: 300.0}, // 3 hours * $100/h
			{Customer: "Client B", Project: "project2"}:  {Rate: 150.0, Total: 225.0}, // 1.5 hours * $150/h
		},
		Customer:         "",
		Project:          "",
		Sort:             SortAlphabetical,
		Reverse:          false,
		TaskGrain:        false,
		DescriptionGrain: false,
		Cache:            expectedCache,
	}, report)
}

func TestReportFiltering(t *testing.T) {
	cfg := createTestConfig(t)
	date := types.Today()

	createTestEntries(t, cfg, date, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "10:00", "github", "project1", "ACME Corp", "123", "Task 1"},
		{"11:00", "12:00", "gitlab", "project2", "Client B", "456", "Task 2"},
		{"13:00", "14:00", "github", "project1", "ACME Corp", "789", "Task 3"},
	})

	t.Run("filter by customer", func(t *testing.T) {
		report, err := New(context.Background(), cfg, date, date, "ACME Corp", "", SortAlphabetical, false, false, false)
		require.NoError(t, err)

		assert.Len(t, report.AggregatedTime, 1)

		acmeCp := CustomerProject{Customer: "ACME Corp", Project: "project1"}
		assert.NotNil(t, report.AggregatedTime[acmeCp], "No entries for ACME Corp / project1")
	})

	t.Run("filter by project", func(t *testing.T) {
		report, err := New(context.Background(), cfg, date, date, "", "project1", SortAlphabetical, false, false, false)
		require.NoError(t, err)

		// Should only have project1 entries
		foundProject1 := false
		foundProject2 := false
		for cp := range report.AggregatedTime {
			if cp.Project == "project1" {
				foundProject1 = true
			}
			if cp.Project == "project2" {
				foundProject2 = true
			}
		}

		assert.True(t, foundProject1, "Expected to find project1 entries")
		assert.False(t, foundProject2, "Should not have project2 entries")
	})
}

func TestStatisticsCalculation(t *testing.T) {
	cfg := createTestConfig(t)

	// Create entries across multiple days
	monday := types.NewDate(2023, 10, 2)    // Monday
	tuesday := types.NewDate(2023, 10, 3)   // Tuesday
	wednesday := types.NewDate(2023, 10, 4) // Wednesday
	saturday := types.NewDate(2023, 10, 7)  // Saturday

	createTestEntries(t, cfg, monday, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "17:00", "github", "project1", "ACME Corp", "123", "Monday work"}, // 8 hours
	})

	createTestEntries(t, cfg, tuesday, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "17:00", "github", "project1", "ACME Corp", "123", "Tuesday work"}, // 8 hours
	})

	createTestEntries(t, cfg, wednesday, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "13:00", "github", "project1", "ACME Corp", "123", "Wednesday work"}, // 4 hours
	})

	createTestEntries(t, cfg, saturday, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"10:00", "12:00", "github", "project1", "ACME Corp", "123", "Saturday work"}, // 2 hours
	})

	report, err := New(context.Background(), cfg, monday, saturday, "", "", SortAlphabetical, false, false, false)
	require.NoError(t, err)

	stats := report.CalculateStatistics()

	// DaysWorked: Monday (8h), Tuesday (8h), Wednesday (4h), Saturday (2h) = 4 days
	// (threshold is 60 minutes, all days meet it)
	assert.Equal(t, 4, stats.DaysWorked)

	// WeekdaysWorked: Monday, Tuesday, Wednesday = 3 weekdays
	assert.Equal(t, 3, stats.WeekdaysWorked)

	// WeeksWorked: 3 weekdays / 5 = 0.6 weeks
	assert.Equal(t, 0.6, stats.WeeksWorked)

	// Total: 22 hours = 1320 minutes
	// AvgTimePerDay: 1320 / 4 = 330 minutes = 5:30
	assert.Equal(t, 5*time.Hour+30*time.Minute, stats.AvgTimePerDay)

	// AvgTimePerWeekday: 1320 / 3 = 440 minutes = 7:20
	assert.Equal(t, 7*time.Hour+20*time.Minute, stats.AvgTimePerWeekday)
}

func TestRateCalculation(t *testing.T) {
	cfg := createTestConfig(t)
	date := types.Today()

	createTestEntries(t, cfg, date, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "11:00", "github", "project1", "ACME Corp", "123", "Task 1"}, // 2 hours on project1 (rate: 100)
		{"11:00", "13:00", "github", "project2", "Client B", "456", "Task 2"},  // 2 hours on project2 (rate: 150)
	})

	report, err := New(context.Background(), cfg, date, date, "", "", SortAlphabetical, false, false, false)
	require.NoError(t, err)

	// Check project1 rate (customer rate should be overridden by project rate)
	cp1 := CustomerProject{Customer: "ACME Corp", Project: "project1"}
	rt, ok := report.RateTotals[cp1]
	require.True(t, ok, "No rate totals for project1")
	assert.Equal(t, 100.0, rt.Rate)
	// 2 hours * $100/hour = $200
	assert.Equal(t, 200.0, rt.Total)

	// Check project2 rate
	cp2 := CustomerProject{Customer: "Client B", Project: "project2"}
	rt, ok = report.RateTotals[cp2]
	require.True(t, ok, "No rate totals for project2")
	assert.Equal(t, 150.0, rt.Rate)
	// 2 hours * $150/hour = $300
	assert.Equal(t, 300.0, rt.Total)

	// Grand total should be $500
	assert.Equal(t, 500.0, report.grandTotal())
}

func TestSorting(t *testing.T) {
	cfg := createTestConfig(t)
	date := types.Today()

	createTestEntries(t, cfg, date, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "10:00", "github", "zebra", "Z Corp", "1", "Task Z"}, // 1 hour
		{"10:00", "13:00", "github", "alpha", "A Corp", "2", "Task A"}, // 3 hours
		{"13:00", "14:00", "github", "beta", "B Corp", "3", "Task B"},  // 1 hour
	})

	t.Run("sort alphabetically", func(t *testing.T) {
		report, err := New(context.Background(), cfg, date, date, "", "", SortAlphabetical, false, false, false)
		require.NoError(t, err)

		sorted := report.sortedCustomerProjects()
		require.Len(t, sorted, 3)

		// Should be alphabetical: A Corp, B Corp, Z Corp
		assert.Equal(t, timeentry.Customer("A Corp"), sorted[0].Customer)
		assert.Equal(t, timeentry.Customer("B Corp"), sorted[1].Customer)
		assert.Equal(t, timeentry.Customer("Z Corp"), sorted[2].Customer)
	})

	t.Run("sort by time spent", func(t *testing.T) {
		report, err := New(context.Background(), cfg, date, date, "", "", SortTimeSpent, false, false, false)
		require.NoError(t, err)

		sorted := report.sortedCustomerProjects()
		require.Len(t, sorted, 3)

		// Should be by time: A Corp (3h), B Corp (1h), Z Corp (1h)
		assert.Equal(t, timeentry.Customer("A Corp"), sorted[0].Customer, "First item should be A Corp (most time)")
	})

	t.Run("sort alphabetically reversed", func(t *testing.T) {
		report, err := New(context.Background(), cfg, date, date, "", "", SortAlphabetical, true, false, false)
		require.NoError(t, err)

		sorted := report.sortedCustomerProjects()

		// Should be reverse alphabetical: Z Corp, B Corp, A Corp
		assert.Equal(t, timeentry.Customer("Z Corp"), sorted[0].Customer)
		assert.Equal(t, timeentry.Customer("A Corp"), sorted[2].Customer)
	})
}

func TestHeaderText(t *testing.T) {
	cfg := createTestConfig(t)

	tests := []struct {
		name     string
		start    types.Date
		end      types.Date
		expected string
	}{
		{
			name:     "full year",
			start:    types.NewDate(2023, 1, 1),
			end:      types.NewDate(2023, 12, 31),
			expected: "Time Report: 2023",
		},
		{
			name:     "full month",
			start:    types.NewDate(2023, 10, 1),
			end:      types.NewDate(2023, 10, 31),
			expected: "Time Report: October 2023",
		},
		{
			name:     "single day",
			start:    types.NewDate(2023, 10, 15),
			end:      types.NewDate(2023, 10, 15),
			expected: "Time Report: 2023-10-15",
		},
		{
			name:     "date range",
			start:    types.NewDate(2023, 10, 1),
			end:      types.NewDate(2023, 10, 15),
			expected: "Time Report: 2023-10-01 - 2023-10-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &Report{
				StartDate: tt.start,
				EndDate:   tt.end,
				Config:    cfg,
			}

			assert.Equal(t, tt.expected, report.headerText())
		})
	}
}

func TestSynchroniserIntegration(t *testing.T) {
	cfg := createTestConfig(t)
	// Configure GitHub synchroniser
	cfg.Sync.Enable = true
	cfg.GitHub.Username = "testuser"
	cfg.GitHub.RootURI = "https://github.com"

	date := types.Today()

	// Create GitHub entries
	createTestEntries(t, cfg, date, []struct {
		start, stop                                       string
		entryType, project, customer, taskID, description string
	}{
		{"09:00", "10:00", "github", "testuser/myproject", "ACME Corp", "123", "Fix bug"},
		{"10:00", "11:00", "gh", "anotherorg/repo", "ACME Corp", "456", "Add feature"},
	})

	report, err := New(context.Background(), cfg, date, date, "", "", SortAlphabetical, false, false, false)
	require.NoError(t, err)

	t.Run("formatTaskName returns formatted GitHub task IDs", func(t *testing.T) {
		cp := CustomerProject{Customer: "ACME Corp", Project: "testuser/myproject"}
		taskName := report.formatTaskName(cp, "123")
		assert.Equal(t, "#123", taskName)
	})

	t.Run("getTaskLink returns GitHub URLs", func(t *testing.T) {
		cp := CustomerProject{Customer: "ACME Corp", Project: "testuser/myproject"}
		link := report.getTaskLink(cp, "123")
		assert.Equal(t, "https://github.com/testuser/myproject/issues/123", link)
	})

	t.Run("formatTaskName returns fallback for unhandled entry types", func(t *testing.T) {
		createTestEntries(t, cfg, date, []struct {
			start, stop                                       string
			entryType, project, customer, taskID, description string
		}{
			{"11:00", "12:00", "custom", "myproject", "Client B", "TASK-789", "Some work"},
		})

		report2, err := New(context.Background(), cfg, date, date, "", "", SortAlphabetical, false, false, false)
		require.NoError(t, err)

		cp := CustomerProject{Customer: "Client B", Project: "myproject"}
		taskName := report2.formatTaskName(cp, "TASK-789")
		// No resolver handles "custom" type — falls back to raw task ID
		assert.Equal(t, "TASK-789", taskName)
	})

	t.Run("Linear formatTaskName returns formatted task IDs", func(t *testing.T) {
		cfg2 := createTestConfig(t)
		cfg2.Linear.DefaultOrg = "myorg"
		date2 := types.Today()

		createTestEntries(t, cfg2, date2, []struct {
			start, stop                                       string
			entryType, project, customer, taskID, description string
		}{
			{"09:00", "10:00", "linear", "ENG", "Tech Corp", "123", "Linear task"},
		})

		report3, err := New(context.Background(), cfg2, date2, date2, "", "", SortAlphabetical, false, false, false)
		require.NoError(t, err)

		cp := CustomerProject{Customer: "Tech Corp", Project: "ENG"}
		taskName := report3.formatTaskName(cp, "123")
		assert.Equal(t, "ENG-123", taskName)
	})

	t.Run("Linear getTaskLink returns Linear URLs", func(t *testing.T) {
		cfg2 := createTestConfig(t)
		cfg2.Linear.DefaultOrg = "myorg"
		date2 := types.Today()

		createTestEntries(t, cfg2, date2, []struct {
			start, stop                                       string
			entryType, project, customer, taskID, description string
		}{
			{"09:00", "10:00", "linear", "PROD", "Tech Corp", "456", "Linear task"},
		})

		report3, err := New(context.Background(), cfg2, date2, date2, "", "", SortAlphabetical, false, false, false)
		require.NoError(t, err)

		cp := CustomerProject{Customer: "Tech Corp", Project: "PROD"}
		link := report3.getTaskLink(cp, "456")
		assert.Equal(t, "https://linear.app/myorg/issue/PROD-456", link)
	})
}

func TestCustomerProjectStr(t *testing.T) {
	cfg := createTestConfig(t)

	tests := []struct {
		name     string
		customer timeentry.Customer
		project  timeentry.Project
		cp       CustomerProject
		expected string
	}{
		{
			name:     "both customer and project",
			customer: "",
			project:  "",
			cp:       CustomerProject{Customer: "ACME Corp", Project: "myproject"},
			expected: "ACME Corp: myproject",
		},
		{
			name:     "only customer",
			customer: "",
			project:  "",
			cp:       CustomerProject{Customer: "ACME Corp", Project: ""},
			expected: "ACME Corp",
		},
		{
			name:     "only project",
			customer: "",
			project:  "",
			cp:       CustomerProject{Customer: "", Project: "myproject"},
			expected: "myproject",
		},
		{
			name:     "filtering by customer shows project",
			customer: "ACME Corp",
			project:  "",
			cp:       CustomerProject{Customer: "ACME Corp", Project: "myproject"},
			expected: "myproject",
		},
		{
			name:     "filtering by project shows customer",
			customer: "",
			project:  "myproject",
			cp:       CustomerProject{Customer: "ACME Corp", Project: "myproject"},
			expected: "ACME Corp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &Report{
				Config:   cfg,
				Customer: tt.customer,
				Project:  tt.project,
			}

			assert.Equal(t, tt.expected, report.customerProjectStr(tt.cp))
		})
	}
}
