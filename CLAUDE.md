# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
Also read **README.md** for installation, configuration reference, usage examples, and architecture overview.

## Development Commands

**Configuration:**
Copy `examples/tracktimerc` to `~/.config/tracktime/tracktimerc`, or use the `--config` flag:
```bash
go run ./cmd/tt --config examples/tracktimerc
```

**Run:**
```bash
go run ./cmd/tt --help
go run ./cmd/tt report --thisweek
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

## Code Architecture

### Directory Structure

```
cmd/tt/                # CLI entry point (go-arg parsing, dispatch)
internal/
  ├── commands/        # Subcommand implementations (start, stop, resume, list, edit, sync, import, report)
  ├── config/          # Config parsing + auto-migration (go.mau.fi/util/configupgrade)
  ├── importer/        # Importer interface; concrete importers registered here
  │   └── tempo/       # Tempo (Jira time-tracking) importer
  ├── report/          # Report aggregation and rendering (stdout, markdown, html, typst, pdf)
  ├── resolver/        # Work item metadata cache (formatted ID, description, hyperlink)
  ├── timeentry/       # TimeEntry and EntryList types; CSV read/write
  └── types/           # Date, Time, Month, Filename types
```

### Key Types

**TimeEntry** (`internal/timeentry/timeentry.go`)
- Fields: `Start`, `Stop`, `Type`, `Project`, `TaskID`, `Customer`, `Description`

**EntryList** (`internal/timeentry/entrylist.go`)
- All time entries for a single day
- Handles insertion logic (auto-stops overlapping entries)
- `Save()` writes CSV atomically; `SaveAndSync()` also triggers `syncMonth`

**Config** (`internal/config/config.go`)
- Loaded from `~/.config/tracktime/tracktimerc` (YAML)
- Auto-migrated from older flat format on first read via `configupgrade`
- Secrets support pipe notation: a value ending with `|` is run as a shell command

**ItemDetailCache** (`internal/resolver/cache.go`)
- Caches formatted task IDs, descriptions, and hyperlinks in `item-cache.csv`
- Resolvers: GitHub, GitLab, Linear, Sourcehut, Jira (descriptions seeded from Tempo import only)

### Command Flow

1. `cmd/tt/` parses arguments using `go-arg`
2. Loads and auto-migrates config from `tracktimerc`
3. Dispatches to the appropriate command in `internal/commands/`
4. Mutation commands (`start`, `stop`, `resume`, `edit`) call `syncMonth` after saving

### Importers

Importers implement `importer.Importer` and self-register via `init()`. Currently: `tempo`.
The import command deduplicates entries by `{start, type, project, taskID}` and can update
the `customer` field on existing entries. `--dry-run` computes counts without writing.

### Type Shortcuts

- `gh` → `github`
- `gl` → `gitlab`
- Otherwise preserved as-is

## Coding Style

- **Go Visibility Rules**: export only identifiers used outside the package.
- **Avoid single-use variables**: inline calls unless the invocation hurts readability.
  - Bad: `header := r.headerText(); buf.WriteString(header)`
  - Good: `buf.WriteString(r.headerText())`

## Important Notes

- Unsupported edge cases: daylight saving time, multi-day entries, timezone switches within a day
- Time format is always `HH:MM` (24-hour)
- Default action (no subcommand): lists today's entries
- No push sync — the Go version pulls metadata only; it does not push to external services
