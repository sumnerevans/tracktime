# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

tracktime is a filesystem-backed time tracking solution that stores time tracking data in CSV files organized by date. The project is **transitioning from Python to Go** - both implementations coexist, with the Go version on the `golang` branch being actively developed.

### Key Architecture Principles

- Filesystem-based: Uses Git-friendly CSV files in `YEAR/MONTH/DAY` directory structure (e.g., `2023/01/15`)
- Offline-first: Must work without internet connectivity
- Manual editing support: CSV format allows direct file editing
- External sync: One-way push to GitLab, GitHub, Sourcehut (tracktime pushes, does not poll)

## Commands

### Go Development (current branch: golang)

**Build and run:**
```bash
go build -o tt tracktime.go
./tt --help
```

**Run tests:**
```bash
go test ./...                    # Run all tests
go test ./lib                    # Run tests in lib package
go test -v ./lib/... -run TestName  # Run specific test with verbose output
```

**Linting (via pre-commit):**
```bash
pre-commit run -av go-imports-repo    # Format imports
pre-commit run -av go-vet-repo-mod    # Run go vet
pre-commit run -av go-staticcheck-repo-mod  # Static analysis
```

### Python Development (legacy)

The Python version is still in `tracktime/` directory and uses Poetry for dependency management.

**Setup:**
```bash
poetry install                   # Install dependencies
poetry shell                     # Activate virtualenv
```

**Testing:**
```bash
poetry run pytest                # Run all tests (with coverage)
```

**Linting:**
```bash
poetry check                     # Validate pyproject.toml
poetry run flake8               # Lint
poetry run mypy tracktime       # Type checking
poetry run black --check .      # Format checking
./cicd/custom_style_check.py    # Custom style checks
```

## Code Architecture

### Directory Structure

```
tracktime.go           # Main entry point, CLI argument parsing
commands/              # Command implementations (start, stop, resume, list, edit, sync, report)
lib/                   # Core library code
  ├── config.go        # Configuration parsing (~/.config/tracktime/tracktimerc YAML)
  ├── entrylist.go     # Time entry list operations, CSV I/O
  ├── date.go          # Date type and operations
  ├── time.go          # Time type (HH:MM format)
  └── month.go         # Month type and operations
synchroniser/          # External service synchronizers
  ├── github.go
  └── syncroniser.go
tracktime/             # Legacy Python implementation
```

### Key Types and Concepts

**TimeEntry** (`lib/entrylist.go:30-39`)
- Core data structure representing a time tracking entry
- Fields: `Start`, `Stop`, `Project`, `Customer`, `TaskID`, `Type`, `Description`
- CSV header: `start,stop,type,project,taskid,customer,description`

**EntryList** (`lib/entrylist.go:71-75`)
- Manages all time entries for a single day
- Handles insertion logic (auto-stops overlapping entries)
- CSV I/O operations

**Config** (`lib/config.go:45-55`)
- Loaded from `~/.config/tracktime/tracktimerc` (YAML)
- Contains: reporting config, sync settings, editor preferences, data directory path

### Command Flow

1. `tracktime.go` parses arguments using `go-arg`
2. Loads config from `tracktimerc`
3. Dispatches to appropriate command in `commands/`
4. Commands use `lib.EntryListForDay()` to load/save day files
5. `EntryList.Save()` writes CSV atomically

### Data File Formats

**Day file** (e.g., `~/.tracktime/2023/01/15`):
```csv
start,stop,type,project,taskid,customer,description
09:00,12:30,gitlab,acme-web,123,ACME Corp,Implementing feature X
13:30,17:00,github,internal-tool,456,Internal,Bug fix
```

**Synced file** (e.g., `~/.tracktime/2023/01/.synced`):
```csv
type,project,taskid,synced
gitlab,acme-web,123,3.5h
```

### Time Entry Type Shortcuts

- `gh` → `github`
- `gl` → `gitlab`
- Otherwise preserved as-is

### Report Command Architecture

The report command (`commands/report.go`) aggregates time entries across date ranges with nested maps:
```
Customer → Project → TaskID → Description → []*TimeEntry
```

Supports multiple grains (task-level, description-level) and sorting options. Currently outputs debug format; full report generation is a TODO.

## Important Notes

- **Unsupported edge cases**: Daylight saving time, multi-day entries, timezone switches within a day
- Time format is always `HH:MM` in 24-hour format
- Default action (no subcommand): Lists today's entries
- Sync is one-way: tracktime → external services (never pulls)
- `.synced` files track what's been pushed to avoid duplicate syncing
