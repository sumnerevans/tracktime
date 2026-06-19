package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/importer"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{Directory: types.Filename(t.TempDir())}
}

func mustTime(t *testing.T, s string) *types.Time {
	t.Helper()
	tm, err := types.ParseTime(s)
	require.NoError(t, err)
	return tm
}

func importEntry(date types.Date, start, stop, entType, project, taskID string) importer.ImportEntry {
	return importer.ImportEntry{
		Date: date,
		Entry: &timeentry.TimeEntry{
			Start:   mustParseTimeInTest(start),
			Stop:    mustParseTimeInTest(stop),
			Type:    timeentry.TimeEntryType(entType),
			Project: timeentry.Project(project),
			TaskID:  timeentry.TaskID(taskID),
		},
	}
}

// mustParseTimeInTest parses without a *testing.T for use in composite literals.
func mustParseTimeInTest(s string) *types.Time {
	tm, err := types.ParseTime(s)
	if err != nil {
		panic(err)
	}
	return tm
}

func loadEntries(t *testing.T, cfg *config.Config, date types.Date) []*timeentry.TimeEntry {
	t.Helper()
	el, err := timeentry.EntryListForDay(cfg, date)
	require.NoError(t, err)
	return el.Entries
}

func TestApplyImport_AddsNewEntries(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	result := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
			importEntry(date, "09:30", "10:00", "jira", "COSO", "2"),
		},
	}

	added, skipped, _, err := applyImport(context.Background(), cfg, result, false)
	require.NoError(t, err)
	assert.Equal(t, 2, added)
	assert.Equal(t, 0, skipped)

	entries := loadEntries(t, cfg, date)
	require.Len(t, entries, 2)
	assert.Equal(t, "09:00", entries[0].Start.String())
	assert.Equal(t, timeentry.Project("IMP"), entries[0].Project)
	assert.Equal(t, "09:30", entries[1].Start.String())
	assert.Equal(t, timeentry.Project("COSO"), entries[1].Project)
}

func TestApplyImport_SkipsDuplicateOnReimport(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	result := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
		},
	}

	// First import
	added, skipped, _, err := applyImport(context.Background(), cfg, result, false)
	require.NoError(t, err)
	assert.Equal(t, 1, added)
	assert.Equal(t, 0, skipped)

	// Second import — same entry, should be skipped
	added, skipped, _, err = applyImport(context.Background(), cfg, result, false)
	require.NoError(t, err)
	assert.Equal(t, 0, added)
	assert.Equal(t, 1, skipped)

	// Still only one entry on disk
	assert.Len(t, loadEntries(t, cfg, date), 1)
}

func TestApplyImport_SameStartDifferentTaskBothKept(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	// Seed an existing github entry at 09:00
	existing := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "github", "myrepo", "42"),
		},
	}
	_, _, _, err := applyImport(context.Background(), cfg, existing, false)
	require.NoError(t, err)

	// Import a jira entry at the same start time
	incoming := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
		},
	}
	added, skipped, _, err := applyImport(context.Background(), cfg, incoming, false)
	require.NoError(t, err)
	assert.Equal(t, 1, added)
	assert.Equal(t, 0, skipped)

	// Both entries present
	entries := loadEntries(t, cfg, date)
	assert.Len(t, entries, 2)
}

func TestApplyImport_SortsByStartTime(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	// Seed a later entry first
	existing := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "11:00", "12:00", "jira", "IMP", "11"),
		},
	}
	_, _, _, err := applyImport(context.Background(), cfg, existing, false)
	require.NoError(t, err)

	// Import an earlier entry
	incoming := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
		},
	}
	_, _, _, err = applyImport(context.Background(), cfg, incoming, false)
	require.NoError(t, err)

	entries := loadEntries(t, cfg, date)
	require.Len(t, entries, 2)
	assert.Equal(t, "09:00", entries[0].Start.String())
	assert.Equal(t, "11:00", entries[1].Start.String())
}

func TestApplyImport_MultipleDates(t *testing.T) {
	cfg := testConfig(t)
	date1 := types.NewDate(2025, 9, 23)
	date2 := types.NewDate(2025, 9, 24)

	result := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date1, "09:00", "10:00", "jira", "TOP", "2879"),
			importEntry(date2, "09:00", "09:30", "jira", "IMP", "13"),
			importEntry(date2, "09:30", "10:00", "jira", "COSO", "2"),
		},
	}

	added, skipped, _, err := applyImport(context.Background(), cfg, result, false)
	require.NoError(t, err)
	assert.Equal(t, 3, added)
	assert.Equal(t, 0, skipped)

	assert.Len(t, loadEntries(t, cfg, date1), 1)
	assert.Len(t, loadEntries(t, cfg, date2), 2)
}

func TestApplyImport_PreservesNonMatchingExistingEntries(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	// Seed a github entry manually
	existing := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "08:30", "09:00", "github", "myrepo", "42"),
		},
	}
	_, _, _, err := applyImport(context.Background(), cfg, existing, false)
	require.NoError(t, err)

	// Import jira entries
	incoming := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
		},
	}
	added, skipped, _, err := applyImport(context.Background(), cfg, incoming, false)
	require.NoError(t, err)
	assert.Equal(t, 1, added)
	assert.Equal(t, 0, skipped)

	entries := loadEntries(t, cfg, date)
	require.Len(t, entries, 2)
	assert.Equal(t, timeentry.TimeEntryType("github"), entries[0].Type)
	assert.Equal(t, timeentry.TimeEntryType("jira"), entries[1].Type)
}

func TestApplyImport_EmptyResult(t *testing.T) {
	cfg := testConfig(t)

	added, skipped, _, err := applyImport(context.Background(), cfg, &importer.ImportResult{}, false)
	require.NoError(t, err)
	assert.Equal(t, 0, added)
	assert.Equal(t, 0, skipped)
}

func TestApplyImport_UpdatesCustomerOnReimport(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	// First import: no customer set
	first := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
		},
	}
	added, skipped, updated, err := applyImport(context.Background(), cfg, first, false)
	require.NoError(t, err)
	assert.Equal(t, 1, added)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 0, updated)

	// Re-import with customer — same key, no customer on existing entry
	withCustomer := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			{
				Date: date,
				Entry: &timeentry.TimeEntry{
					Start:    mustParseTimeInTest("09:00"),
					Stop:     mustParseTimeInTest("09:30"),
					Type:     timeentry.TimeEntryType("jira"),
					Project:  timeentry.Project("IMP"),
					TaskID:   timeentry.TaskID("13"),
					Customer: timeentry.Customer("ACME"),
				},
			},
		},
	}
	added, skipped, updated, err = applyImport(context.Background(), cfg, withCustomer, false)
	require.NoError(t, err)
	assert.Equal(t, 0, added)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 1, updated)

	entries := loadEntries(t, cfg, date)
	require.Len(t, entries, 1)
	assert.Equal(t, timeentry.Customer("ACME"), entries[0].Customer)
}

func TestApplyImport_DoesNotOverwriteExistingCustomer(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	// First import with customer already set
	first := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			{
				Date: date,
				Entry: &timeentry.TimeEntry{
					Start:    mustParseTimeInTest("09:00"),
					Stop:     mustParseTimeInTest("09:30"),
					Type:     timeentry.TimeEntryType("jira"),
					Project:  timeentry.Project("IMP"),
					TaskID:   timeentry.TaskID("13"),
					Customer: timeentry.Customer("ACME"),
				},
			},
		},
	}
	_, _, _, err := applyImport(context.Background(), cfg, first, false)
	require.NoError(t, err)

	// Re-import with a different customer — should be skipped, not overwritten
	second := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			{
				Date: date,
				Entry: &timeentry.TimeEntry{
					Start:    mustParseTimeInTest("09:00"),
					Stop:     mustParseTimeInTest("09:30"),
					Type:     timeentry.TimeEntryType("jira"),
					Project:  timeentry.Project("IMP"),
					TaskID:   timeentry.TaskID("13"),
					Customer: timeentry.Customer("OTHER"),
				},
			},
		},
	}
	added, skipped, updated, err := applyImport(context.Background(), cfg, second, false)
	require.NoError(t, err)
	assert.Equal(t, 0, added)
	assert.Equal(t, 1, skipped)
	assert.Equal(t, 0, updated)

	entries := loadEntries(t, cfg, date)
	require.Len(t, entries, 1)
	assert.Equal(t, timeentry.Customer("ACME"), entries[0].Customer)
}

func TestApplyImport_DryRun_DoesNotWriteToDisk(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	result := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
			importEntry(date, "09:30", "10:00", "jira", "COSO", "2"),
		},
	}

	added, skipped, _, err := applyImport(context.Background(), cfg, result, true)
	require.NoError(t, err)
	assert.Equal(t, 2, added)
	assert.Equal(t, 0, skipped)

	// Nothing written to disk
	assert.Empty(t, loadEntries(t, cfg, date))
}

func TestApplyImport_DryRun_SkipCountsCorrect(t *testing.T) {
	cfg := testConfig(t)
	date := types.NewDate(2025, 9, 24)

	result := &importer.ImportResult{
		Entries: []importer.ImportEntry{
			importEntry(date, "09:00", "09:30", "jira", "IMP", "13"),
		},
	}

	// Real import first
	_, _, _, err := applyImport(context.Background(), cfg, result, false)
	require.NoError(t, err)

	// Dry run re-import — should count the skip without touching disk
	added, skipped, _, err := applyImport(context.Background(), cfg, result, true)
	require.NoError(t, err)
	assert.Equal(t, 0, added)
	assert.Equal(t, 1, skipped)

	// Still only one entry on disk from the real import
	assert.Len(t, loadEntries(t, cfg, date), 1)
}

func TestApplyImport_MustTime(t *testing.T) {
	// Ensure mustTime helper works correctly (used transitively in test helpers).
	tm := mustTime(t, "09:30")
	assert.Equal(t, "09:30", tm.String())
}
