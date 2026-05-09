# Python to Go Migration Status

This document tracks the migration progress from the Python implementation of tracktime to Go.

**Branch:** `golang`
**Last Updated:** 2026-05-09
**Overall Completion:** ~90%

---

## Overview

The Go rewrite is located in the root directory alongside the legacy Python implementation (in `tracktime/`). Both implementations currently coexist, with the Go version being actively developed.

### Current Project Structure

```
cmd/tt/                     # Main entry point
internal/
  ├── commands/             # Command implementations
  ├── config/               # Configuration loading
  ├── exporter/             # Push time to external services (Exporter interface + .synced I/O)
  ├── importer/             # Pull time from external services (Importer interface)
  ├── report/               # Report generation
  ├── resolver/             # Work item metadata (ItemResolver interface + ItemCache)
  ├── synchroniser/         # Retained for package compatibility (empty)
  ├── timeentry/            # Time entry and entry list logic
  └── types/                # Core types (Date, Time, Month, Filename)
tracktime/                  # Legacy Python implementation
```

---

## Core Library Components

### ✅ Fully Implemented

| Component | Location | Description | Status |
|-----------|----------|-------------|--------|
| **Configuration** | `internal/config/config.go` | YAML config parser for `~/.config/tracktime/tracktimerc` | ✅ Complete |
| **Date Type** | `internal/types/date.go` | Date operations and parsing (full Python parity + weekdays) | ✅ Complete with tests |
| **Time Type** | `internal/types/time.go` | HH:MM time format handling | ✅ Complete with tests |
| **Month Type** | `internal/types/month.go` | Month operations and parsing | ✅ Complete with tests |
| **TimeEntry** | `internal/timeentry/entrylist.go` | Core time entry data structure | ✅ Complete |
| **EntryList** | `internal/timeentry/entrylist.go` | Day file management and CSV I/O | ✅ Complete |
| **AggregatedTime** | `internal/timeentry/aggregated.go` | Time aggregation types | ✅ Complete |
| **Filename** | `internal/types/filename.go` | Path expansion and file handling | ✅ Complete |

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
| **start** | `internal/commands/start.go` | Start new time entry with optional start time, type, project, customer, task ID | ✅ Complete |
| **stop** | `internal/commands/stop.go` | Stop current entry with optional stop time | ✅ Complete |
| **resume** | `internal/commands/resume.go` | Resume previous entry by index (defaults to last entry) | ✅ Complete |
| **list** | `internal/commands/list.go` | List entries for a date with customer filtering, formatted table output, total time | ✅ Complete |
| **edit** | `internal/commands/edit.go` | Open day file in editor (respects config, `$EDITOR`, `$VISUAL`) | ✅ Complete |
| **report** | `internal/commands/report.go` + `internal/report/*.go` | Full stdout + all export formats | ✅ Complete |

**Default behavior:** Running `tt` without subcommand lists today's entries ✅

### 🚧 Partially Implemented

| Command | File | What Works | What's Missing |
|---------|------|------------|----------------|
| **sync** | `internal/commands/sync.go` | Month aggregation, `.synced` read/write, exporter/importer dispatch skeleton | Concrete exporter/importer implementations |

#### Report Command

All output formats complete:
- ✅ Stdout (colors, ANSI-aware formatting)
- ✅ Markdown (`.md`)
- ✅ HTML (`.html` via goldmark)
- ✅ Typst (`.typ`)
- ✅ PDF (`.pdf` via go-typst — requires `typst` binary)

All options complete:
- ✅ Date range shortcuts (`--today`, `--yesterday`, `--thisweek`, `--lastweek`, `--thismonth`, `--lastmonth`, `--thisyear`, `--lastyear`)
- ✅ Month/year shorthand (`-m`, `-y`)
- ✅ Positional date range (start/end dates)
- ✅ Customer and project filtering (`-c`, `-p`)
- ✅ Sort by alphabetical or time-spent, ascending/descending
- ✅ Grain options (task/description level)
- ✅ Rate and total calculations
- ✅ Statistics (days worked, averages)

#### Sync Command

**Current state:** Aggregates the month's time entries, reads the `.synced` file, dispatches to registered `Importer` and `Exporter` instances, writes updated `.synced` file. Framework is complete; no concrete importers or exporters are registered yet so it is effectively a no-op.

---

## Resolvers and Sync Framework

The Go implementation uses a three-package architecture (replacing the old monolithic `Synchroniser` interface):

### `internal/resolver` — Work Item Metadata

**`ItemResolver` interface:**
```go
type ItemResolver interface {
    Init(cfg config.SyncConfig)
    Handles(entry *timeentry.TimeEntry) bool
    GetFormattedTaskID(entry *timeentry.TimeEntry) string
    GetTaskLink(entry *timeentry.TimeEntry) string
    FetchDescription(ctx context.Context, entry *timeentry.TimeEntry) (string, error)
}
```

**`ItemCache`** wraps all resolvers and provides a persistent CSV cache at `~/.tracktime/item-cache.csv` with soft TTL semantics: stale entries are refreshed on access but never deleted on failure, so descriptions survive loss of API access (e.g. leaving a company). TTL defaults to 30 days and is configurable via `item_cache_ttl_days` in `tracktimerc`.

**Concrete resolvers:**

| Service | File | Formatted ID | Task Link | Fetch Description |
|---------|------|-------------|-----------|-------------------|
| **GitHub** | `internal/resolver/github.go` | `#123` | ✅ | ❌ stub |
| **Linear** | `internal/resolver/linear.go` | `ENG-123` | ✅ | ✅ GraphQL API |
| **GitLab** | `internal/resolver/gitlab.go` | `#123` / `!42` | ✅ | ✅ REST API |
| **Sourcehut** | `internal/resolver/sourcehut.go` | `#123` | ✅ | ✅ REST API |

### `internal/exporter` — Push Time to External Services

**`Exporter` interface:**
```go
type Exporter interface {
    Name() string
    Init(cfg config.SyncConfig)
    Export(ctx context.Context, aggregatedTime, syncedTime timeentry.AggregatedTime, month types.Month) (timeentry.AggregatedTime, error)
}
```

`.synced` file I/O lives here (`ReadSyncedFile` / `WriteSyncedFile`).

**Concrete exporters:** None yet. GitLab is the obvious first candidate.

### `internal/importer` — Pull Time from External Services

**`Importer` interface:**
```go
type Importer interface {
    Name() string
    Init(cfg config.SyncConfig)
    Import(ctx context.Context, month types.Month) (timeentry.AggregatedTime, error)
}
```

**Concrete importers:** None yet. Tempo (Atlassian) is the planned first candidate.

---

## Python Implementation Features Not Yet in Go

### Sync Functionality
- ❌ GitLab exporter (push time to issues/MRs via `/add_spent_time`)
- ❌ Sourcehut exporter (update tracktime comment on tickets)
- ❌ Tempo importer (pull logged time from Jira/Tempo)
- ❌ GitHub `FetchDescription` (GraphQL fetch of issue/PR title)

---

## Testing Status

| Package | Test Coverage | Notes |
|---------|---------------|-------|
| `internal/types/` | ✅ ~61% | Time, Month, Date parsing and operations |
| `internal/config/` | ✅ 100% | Config loading fully tested |
| `internal/timeentry/` | ✅ ~85% | EntryList, AddEntry, Save, Stop, Resume, CSV I/O |
| `internal/commands/` | ✅ ~24% | Start, Stop, List, Resume, Edit, Sync + integration test |
| `internal/report/` | ✅ ~27% | Report creation, filtering, statistics, rates, sorting |
| `internal/resolver/` | ✅ Linear tested | Guard clauses, formatted ID, task link; no HTTP tests |
| `internal/exporter/` | ❌ 0% | No tests yet |
| `internal/importer/` | ❌ 0% | No tests yet |

All tests pass with all linters (go-imports, go-vet, go-staticcheck).

---

## Migration Priorities

1. **Low Priority** — Concrete exporters/importers:
   - GitLab exporter (push time to issues/MRs)
   - Sourcehut exporter (update tracktime comments on tickets)
   - Tempo importer (pull time from Jira/Tempo)
   - GitHub `FetchDescription` implementation

2. **Low Priority** — Test coverage:
   - Exporter `.synced` file I/O tests
   - Resolver HTTP fetch tests (with httptest)
   - More edge cases in commands/report

3. **Low Priority** — Documentation and polish

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
go build -o tt ./cmd/tt
```

**Run:**
```bash
go run ./cmd/tt --help
go run ./cmd/tt --config examples/tracktimerc.go-example report --thisweek
```

**Run tests:**
```bash
go test ./...
go test -v ./internal/resolver/... -run TestName
```

**Lint:**
```bash
pre-commit run --all-files
```

---

## Architecture Notes

**CSV Format:** `start,stop,type,project,taskid,customer,description`

**Directory Structure:** `~/.tracktime/YEAR/MONTH/DAY`

**Item Cache:** `~/.tracktime/item-cache.csv`
```csv
type,project,taskid,description,fetched_at
gitlab,acme-web,#123,Implement feature X,2026-01-15T10:30:00Z
linear,ENG,456,Fix the login bug,2026-03-01T08:00:00Z
```

**Synced File Format:** `~/.tracktime/YEAR/MONTH/.synced`
```csv
type,project,taskid,synced
gitlab,acme-web,#123,12600
```
(duration stored as integer seconds)

**Auto-stop Logic:** When starting a new entry, any currently running entry is automatically stopped at the new entry's start time.

**Type Shortcuts:** `gh` → `github`, `gl` → `gitlab`

---

## Version

Current version: **v0.11.0** (as declared in `cmd/tt/main.go`)

---

## Summary

The Go rewrite is **~90% complete** and production-ready for daily time tracking:

| Component | Completion |
|-----------|------------|
| Core library (types, config, entry list) | **100%** ✅ |
| Basic commands (start, stop, resume, list, edit) | **100%** ✅ |
| Report command (all formats) | **100%** ✅ |
| Resolver framework + item cache | **100%** ✅ |
| Item resolvers (GitHub, Linear, GitLab, Sourcehut) | **~85%** ✅ (GitHub FetchDescription stub) |
| Exporter/Importer framework | **100%** ✅ |
| Concrete exporters (GitLab, Sourcehut) | **0%** ❌ |
| Concrete importers (Tempo) | **0%** ❌ |
| Test coverage | **~50%** ✅ |

**Usage:**
```bash
go run ./cmd/tt report --thisweek
go run ./cmd/tt report --thisweek -o report.pdf   # requires typst binary
go run ./cmd/tt sync                               # no-op until exporters implemented
```
