# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

tracktime is a filesystem-backed time tracking solution. The project is **transitioning from Python to Go** — both implementations coexist, with the Go version on the `golang` branch being actively developed. The README describes the Python version; this file focuses on the Go implementation.

### Key Architecture Principles

- Filesystem-based: Git-friendly CSV files in `YEAR/MONTH/DAY` directory structure
- Offline-first: must work without internet connectivity
- Manual editing support: CSV format allows direct file editing
- External sync: importers only (pull from external services); push sync is intentionally not ported

## Commands

### Go Development (current branch: golang)

**Configuration:**
Use the example config at `examples/tracktimerc.go-example`. Either:
- Copy it to `~/.config/tracktime/tracktimerc`, or
- Use the `--config` flag: `go run ./cmd/tt --config examples/tracktimerc.go-example`

**Run:**
```bash
go run ./cmd/tt --help
go run ./cmd/tt --config examples/tracktimerc.go-example report --thisweek
```

**Run tests:**
```bash
go test ./...
go test -v ./internal/... -run TestName
```

**Linting (via pre-commit):**
```bash
pre-commit run -av go-imports-repo
pre-commit run -av go-vet-repo-mod
pre-commit run -av go-staticcheck-repo-mod
```

## Go Code Architecture

### Directory Structure

```
cmd/tt/                # CLI entry point
internal/
  ├── commands/        # Subcommand implementations (start, stop, resume, list, edit, sync, report)
  ├── config/          # Config parsing + auto-migration (go.mau.fi/util/configupgrade)
  ├── importer/        # Importer interface; concrete importers registered here
  ├── report/          # Report aggregation and rendering (stdout, markdown, html, typst, pdf)
  ├── resolver/        # Work item metadata cache (formatted ID, description, hyperlink)
  └── types/           # Date, Time, Month, Filename types
```

### Key Types and Concepts

**TimeEntry** (`internal/timeentry/entrylist.go`)
- Fields: `Start`, `Stop`, `Type`, `Project`, `TaskID`, `Customer`, `Description`
- CSV header: `start,stop,type,project,taskid,customer,description`

**EntryList** (`internal/timeentry/entrylist.go`)
- All time entries for a single day
- Handles insertion logic (auto-stops overlapping entries)
- `Save()` writes CSV atomically; `SaveAndSync()` also triggers `syncMonth`

**Config** (`internal/config/config.go`)
- Loaded from `~/.config/tracktime/tracktimerc` (YAML)
- Auto-migrated from Python flat format on first read via `configupgrade`
- Secrets support pipe notation: `cat /path/to/secret|` runs the command and uses stdout

**Resolver / ItemDetailCache** (`internal/resolver/`)
- Fetches formatted task IDs, descriptions, and hyperlinks from GitHub/GitLab/Linear/Sourcehut
- Caches results in `item-cache.csv` with configurable TTL (`item_cache_ttl_days`, default 30 days)

### Command Flow

1. `cmd/tt/` parses arguments using `go-arg`
2. Loads and auto-migrates config from `tracktimerc`
3. Dispatches to appropriate command in `internal/commands/`
4. Mutation commands (`start`, `stop`, `resume`, `edit`) call `syncMonth` after saving

### Time Entry Type Shortcuts

- `gh` → `github`
- `gl` → `gitlab`
- Otherwise preserved as-is

## Coding Style

- **Go Visibility Rules**: export only identifiers that need to be used outside the package.

- **Avoid single-use variables**: inline function calls unless the invocation is complex enough to hurt readability.
  - Bad: `header := r.headerText(); buf.WriteString(header)`
  - Good: `buf.WriteString(r.headerText())`

## Important Notes

- **Unsupported edge cases**: daylight saving time, multi-day entries, timezone switches within a day
- Time format is always `HH:MM` in 24-hour format
- Default action (no subcommand): lists today's entries
- No `.synced` files — the Go version does not push to external services
