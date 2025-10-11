# Python to Go Migration Status

This document tracks the migration progress from the Python implementation of tracktime to Go.

**Branch:** `golang`
**Last Updated:** 2025-10-11

---

## Overview

The Go rewrite is located in the root directory alongside the legacy Python implementation (in `tracktime/`). Both implementations currently coexist, with the Go version being actively developed.

---

## Core Library Components

### ✅ Fully Implemented

| Component | Location | Description | Status |
|-----------|----------|-------------|--------|
| **Configuration** | `lib/config.go:45-55` | YAML config parser for `~/.config/tracktime/tracktimerc` | ✅ Complete |
| **Date Type** | `lib/date.go` | Date operations and parsing (full Python parity) | ✅ Complete with tests |
| **Time Type** | `lib/time.go` | HH:MM time format handling | ✅ Complete with tests |
| **Month Type** | `lib/month.go` | Month operations and parsing | ✅ Complete with tests |
| **TimeEntry** | `lib/entrylist.go:30-39` | Core time entry data structure | ✅ Complete |
| **EntryList** | `lib/entrylist.go:71-75` | Day file management and CSV I/O | ✅ Complete |
| **Filename** | `lib/filename.go` | Path expansion and file handling | ✅ Complete |

**Key Features:**
- CSV reading/writing with atomic saves
- Auto-stop logic for overlapping entries
- Day file path generation (`YEAR/MONTH/DAY` structure)
- Time entry type shortcuts (`gh` → `github`, `gl` → `gitlab`)
- Flexible date parsing matching Python implementation:
  - Full dates: `YYYY-MM-DD`, `YYYY/MM/DD`, `YY-MM-DD`, `YY/MM/DD`
  - Partial dates: `MM-DD`, `M-D` (defaults to current year)
  - Day only: `DD`, `D` (defaults to current year and month)
  - Keywords: `today`, `yesterday`

---

## Commands

### ✅ Fully Implemented

| Command | File | Features | Status |
|---------|------|----------|--------|
| **start** | `commands/start.go` | Start new time entry with optional start time, type, project, customer, task ID | ✅ Complete |
| **stop** | `commands/stop.go` | Stop current entry with optional stop time | ✅ Complete |
| **resume** | `commands/resume.go` | Resume previous entry by index (defaults to last entry) | ✅ Complete |
| **list** | `commands/list.go` | List entries for a date with customer filtering, formatted table output, total time | ✅ Complete |
| **edit** | `commands/edit.go` | Open day file in editor (respects config, `$EDITOR`, `$VISUAL`) | ✅ Complete |

**Default behavior:** Running `tt` without subcommand lists today's entries ✅

### 🚧 Partially Implemented

| Command | File | What Works | What's Missing |
|---------|------|------------|----------------|
| **sync** | `commands/sync.go` | Command structure, argument parsing | Full implementation (currently just spawns goroutines) |
| **report** | `commands/report.go` | Date range calculation, aggregation logic, data collection | Output formatting (currently debug prints only) |

#### Sync Command Details

**Current State:**
- Accepts month argument with default "this month"
- Spawns synchroniser goroutines
- Has no blocking/completion logic

**Missing:**
- Actual API calls to external services
- Reading/writing `.synced` files
- Error handling and reporting
- Synchroniser coordination

#### Report Command Details

**Current State:**
- ✅ All date range shortcuts work (`--today`, `--yesterday`, `--thisweek`, `--lastweek`, `--thismonth`, `--lastmonth`, `--thisyear`, `--lastyear`)
- ✅ Month/year shorthand (`-m`, `-y`)
- ✅ Positional date range (start/end dates)
- ✅ Customer and project filtering (`-c`, `-p`)
- ✅ Data aggregation into nested maps: `Customer → Project → TaskID → Description → []*TimeEntry`
- ✅ Day-level statistics collection
- ✅ Sort options parsed (alphabetical/time-spent, asc/desc)
- ✅ Grain options parsed (task/description level)

**Missing:**
- Actual report output formatting (PDF, HTML, RST, or stdout)
- Respect for grain and sort options
- Output file writing
- Human-readable report format

**Current Output:**
```
ENTRY  {...}
REPORT  {...}
```

---

## Synchronisers

### 🚧 Partially Implemented

| Service | File | What Works | What's Missing |
|---------|------|------------|----------------|
| **GitHub** | `synchroniser/github.go` | Task ID formatting, task link generation | API calls, actual sync logic, task description fetching |

**Interface Definition:** `synchroniser/syncroniser.go:18-25`

**Implemented Methods:**
- `Init()` - Config loading ✅
- `Name()` - Returns "GitHub" ✅
- `GetFormattedTaskID()` - Formats as `#123` ✅
- `GetTaskLink()` - Generates GitHub issue/PR URL ✅
- `GetTaskDescription()` - Stub (returns empty string) ❌
- `Sync()` - Stub (no-op) ❌

**Missing Synchronisers:**
- GitLab (Python has `tracktime/synchronisers/gitlab.py`)
- Sourcehut (Python has `tracktime/synchronisers/sourcehut.py`)
- Linear (Python has `tracktime/synchronisers/linear.py`)

---

## Python Implementation Features Not Yet in Go

### Commands
- None! All Python commands have Go equivalents (though sync/report need completion)

### Synchronisers
- ❌ GitLab synchroniser
- ❌ Sourcehut synchroniser
- ❌ Linear synchroniser

### Report Functionality
- ❌ PDF export (Python uses ReportLab)
- ❌ HTML export
- ❌ RST export
- ❌ Formatted stdout output

### Sync Functionality
- ❌ `.synced` file reading/writing
- ❌ API integration with external services
- ❌ Time aggregation and submission logic
- ❌ Deduplication (checking what's already synced)

---

## Testing Status

| Package | Test Coverage | Notes |
|---------|---------------|-------|
| `lib/time.go` | ✅ Has tests (`lib/time_test.go`) | Time parsing and formatting |
| `lib/month.go` | ✅ Has tests (`lib/month_test.go`) | Month parsing |
| `lib/date.go` | ✅ Has tests (`lib/date_test.go`) | Date parsing (all formats), AddDays, AddMonths |
| Other packages | ❌ No tests yet | Need test coverage |

---

## Migration Priorities

Based on current state, suggested priorities:

1. **High Priority** - Complete sync command:
   - Implement `.synced` file I/O
   - Add aggregation logic (similar to report)
   - Complete GitHub synchroniser API calls
   - Add GitLab, Sourcehut, Linear synchronisers

2. **High Priority** - Complete report command:
   - Implement output formatters (stdout, HTML, RST, PDF)
   - Apply grain and sort options
   - Match Python output format

3. **Medium Priority** - Testing:
   - Add unit tests for commands
   - Add integration tests
   - Test edge cases (overlapping entries, invalid times, etc.)

4. **Low Priority** - Feature parity:
   - Ensure all Python features are covered
   - Documentation

---

## Known Limitations (Same as Python)

- ❌ Daylight saving time handling
- ❌ Multi-day time entries
- ❌ Timezone switches within a day
- ❌ Time entry validation (e.g., start > stop)

---

## Build and Test

**Build:**
```bash
go build -o tt tracktime.go
```

**Run tests:**
```bash
go test ./...                           # All tests
go test -v ./lib/... -run TestName     # Specific test
```

**Lint:**
```bash
pre-commit run -av go-imports-repo
pre-commit run -av go-vet-repo-mod
pre-commit run -av go-staticcheck-repo-mod
```

---

## Architecture Notes

**CSV Format:** `start,stop,type,project,taskid,customer,description`

**Directory Structure:** `~/.tracktime/YEAR/MONTH/DAY`

**Synced File Format:** `~/.tracktime/YEAR/MONTH/.synced`
```csv
type,project,taskid,synced
gitlab,acme-web,123,3.5h
```

**Auto-stop Logic:** When starting a new entry, any currently running entry is automatically stopped at the new entry's start time.

**Type Shortcuts:** `gh` → `github`, `gl` → `gitlab`

---

## Version

Current version: **v0.11.0** (as declared in `tracktime.go:27`)
