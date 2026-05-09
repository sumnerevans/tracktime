# Python → Go Gap Report

Analysis of feature parity between the Python implementation (`tracktime/`) and the Go
implementation (`internal/`, `cmd/tt/`, branch `golang`), as of 2026-05-09.

---

## 1. Intentionally Not Ported

These Python features were deliberately excluded from the Go rewrite.

### 1.1 Sync push to external services

Python pushes spent time to **GitLab** (via `POST .../add_spent_time`) and **Sourcehut** (via
ticket comment create/edit). Go deliberately omits the push direction: the `sync` command runs
importers only. GitHub and Linear were read-only in Python too.

The `.synced` file tracking and `Exporter` interface have been removed. Custom push integrations
should be implemented as standalone tools that read the CSV data directly.

### 1.2 External synchroniser plugin system

Python supports `external_synchroniser_files`, allowing user-supplied Python files to implement
`ExternalSynchroniser` (e.g. `examples/jira.py`). Go's static type system and compilation model
make a dynamic plugin system impractical without significant complexity (e.g. Go plugins via
`plugin.Open` are fragile and platform-limited). Custom integrations should instead be implemented
as standalone tools that read the CSV data directly.

### 1.3 Internet connectivity check before sync

Python pings `8.8.8.8` before syncing and silently skips on failure. In Go this is replaced by
`context.Context` timeouts and cancellation: sync operations respect the request context and will
fail fast with a clear error rather than hanging or silently skipping.

### 1.4 RST report format

Python exports `.rst` via `RSTExporter` (a thin wrapper around the tabulate text output). In Go,
Markdown (`.md`) and Typst (`.typ`) cover the structured-document use cases better. RST is not
planned.

### 1.5 `tableformat` config / tabulate-style output

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
| Auto-sync after mutations | ✓ | ✓ | Done |
| Sync push to GitLab/Sourcehut | ✓ | n/a | Intentionally not ported |
| Task hyperlinks in reports | ✓ | ✓ | Done |
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
