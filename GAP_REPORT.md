# Python → Go Gap Report

Analysis of feature parity between the Python implementation (`tracktime/`) and the Go
implementation (`internal/`, `cmd/tt/`, branch `golang`), as of 2026-05-09.

---

## 1. Critical Gaps — Python features absent in Go

### 1.1 Sync push not implemented

The Go `sync` command has the **infrastructure** (interfaces in `internal/exporter/` and
`internal/importer/`) but **no concrete implementations are registered**. `exporter.Exporters`
and `importer.Importers` are always empty slices, so the sync command:

- Reads aggregated time ✓
- Reads the `.synced` file ✓
- Loops over zero exporters — does nothing
- Writes the `.synced` file back unchanged ✓

Python pushes spent time to **GitLab** (via `POST .../add_spent_time`) and **Sourcehut** (via
ticket comment create/edit). Neither exists in Go. GitHub and Linear have no push sync in Python
either (they are read-only resolvers in both implementations).

**Files affected:** `internal/exporter/exporter.go` (interface only, no registrations)

### 1.2 Auto-sync after mutations is TODO

Python calls `EntryList.sync()` after every `start`, `stop`, `resume`, and (via `edit`) after the
editor closes. All Go command handlers carry a `// TODO sync` comment with no implementation.

**Files affected:** `internal/commands/start.go:26`, `stop.go:21`, `resume.go:23`, `edit.go:55`

### 1.3 Task hyperlinks missing from all report formats

Python's HTML report wraps task names in `<a href>` anchors using `get_task_link()`. Go's
`GetTaskLink()` exists in the cache layer (`internal/resolver/cache.go`) but **none of the
renderers call it** — not stdout, not Markdown, not HTML. The HTML renderer generates via
Markdown conversion so it cannot insert per-cell anchors anyway.

**Files affected:** `internal/report/stdout.go`, `markdown.go`, `html.go`

---

## 2. Intentionally Not Ported

These Python features were deliberately excluded from the Go rewrite.

### 2.1 External synchroniser plugin system

Python supports `external_synchroniser_files`, allowing user-supplied Python files to implement
`ExternalSynchroniser` (e.g. `examples/jira.py`). Go's static type system and compilation model
make a dynamic plugin system impractical without significant complexity (e.g. Go plugins via
`plugin.Open` are fragile and platform-limited). Custom integrations should instead be implemented
as standalone tools that read the CSV data directly.

### 2.2 Internet connectivity check before sync

Python pings `8.8.8.8` before syncing and silently skips on failure. In Go this is replaced by
`context.Context` timeouts and cancellation: sync operations respect the request context and will
fail fast with a clear error rather than hanging or silently skipping.

### 2.3 RST report format

Python exports `.rst` via `RSTExporter` (a thin wrapper around the tabulate text output). In Go,
Markdown (`.md`) and Typst (`.typ`) cover the structured-document use cases better. RST is not
planned.

### 2.4 `tableformat` config / tabulate-style output

Python's stdout report is rendered via the `tabulate` library, with the format controlled by
`tableformat` in config (e.g. `fancy_grid`, `rst`). Go uses `rodaine/table` with ANSI-color
output, which is a deliberate improvement and not configurable by format name. The config key
`reporting.table_format` was parsed in Go but is unused and should be removed.

---

## 3. New in Go (not in Python)

These features exist in Go but have no Python equivalent.

| Feature | Details |
|---------|---------|
| **Markdown export** | `.md` output via `GenerateMarkdownReport` |
| **Typst export** | `.typ` output; PDF compiled via Typst instead of pdfkit/wkhtmltopdf |
| **Color terminal output** | `list` and `report --stdout` use ANSI colors via `fatih/color` |
| **Structured logging** | Configurable log file, format, level via `logging` config key |
| **Configurable cache TTL** | `item_cache_ttl_days` (default 30); Python caches forever |
| **Shared CSV item cache** | Single `item-cache.csv` replaces per-service pickle files; survives API failures (stale-while-revalidate) |
| **GitHub Enterprise support** | `github.root_uri` overrides `https://github.com` |
| **`--lastmonth` flag** | Explicit flag for last month; Python only had this as the default |
| **`--lastweek` starts on Sunday** | Both start the week on Sunday (consistent) |
| **`context.Context` propagation** | All commands pass a context for cancellation and logging |

---

## 4. Summary Table

| Area | Python | Go | Status |
|------|--------|----|--------|
| `start` / `stop` / `resume` / `list` / `edit` | ✓ | ✓ | — |
| Auto-sync after mutations | ✓ | ✗ (TODO) | **Gap — high** |
| Sync push to GitLab | ✓ | ✗ | **Gap — high** |
| Sync push to Sourcehut | ✓ | ✗ | **Gap — high** |
| Task hyperlinks in reports | ✓ | ✗ | **Gap — medium** |
| Config auto-migration (configupgrade) | — | ✓ | Done |
| Description case-folding in report | ✓ | ✓ | Done |
| Abbreviated weekday date parsing | ✓ | ✓ | Done |
| External synchroniser plugins | ✓ | n/a | Intentionally not ported |
| Internet check before sync | ✓ | n/a | Replaced by context timeouts |
| RST export | ✓ | n/a | Intentionally not ported |
| `tableformat` config | ✓ | n/a | Replaced by color output |
| Markdown export | ✗ | ✓ | New in Go |
| Typst / PDF via Typst | ✗ | ✓ | New in Go |
| Color terminal output | ✗ | ✓ | New in Go |
| Structured logging | ✗ | ✓ | New in Go |
| Configurable cache TTL | ✗ | ✓ | New in Go |
| GitHub Enterprise URI | ✗ | ✓ | New in Go |
