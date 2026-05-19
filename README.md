# tracktime

[![Build](https://github.com/sumnerevans/tracktime/actions/workflows/go.yml/badge.svg)](https://github.com/sumnerevans/tracktime/actions/workflows/go.yml)
[![LiberaPay Donation Status](https://img.shields.io/liberapay/receives/sumner.svg?logo=liberapay)](https://liberapay.com/sumner/donate)

tracktime is a filesystem-backed time tracking solution. It stores time tracking
data in Git-friendly CSV files organized by date, and can generate rich reports
with time totals, rates, and billing amounts.

## Features

- Start, stop, and resume time entries
- List and edit time entries for any day
- Generate reports for arbitrary date ranges (stdout, Markdown, HTML, Typst, PDF)
- Filter reports by customer or project
- Sync task metadata (titles, links) from GitHub, GitLab, Linear, and Sourcehut
- Color terminal output with clickable task hyperlinks (OSC 8)
- Config auto-migration from older formats

## Installation

### Using `go install`

```
go install github.com/sumnerevans/tracktime/cmd/tt@latest
```

### Using Nix

Run tracktime ad-hoc:

```
nix run github:sumnerevans/tracktime
```

Using flakes:

```nix
{
  inputs.tracktime = {
    url = "github:sumnerevans/tracktime";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { tracktime, ... }: {
    // use the package: tracktime.packages."x86_64-linux".tracktime
  };
}
```

### From source

```
git clone https://github.com/sumnerevans/tracktime
cd tracktime
go build -o tt ./cmd/tt
```

## Configuration

Copy `examples/tracktimerc` to `~/.config/tracktime/tracktimerc` and edit it.
The file is YAML. Key options:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `directory` | string | `$HOME/.tracktime` | Where CSV files are stored |
| `editor` | string | `$EDITOR` / `vim` | Editor opened by `tt edit` |
| `editor_args` | list | `[]` | Extra arguments passed to the editor |
| `typst_path` | string | `typst` | Path to the Typst compiler (for PDF export) |
| `reporting.fullname` | string | | Your name, used in report headers |
| `reporting.project_rates` | map | | Hourly rates per project (for billing totals) |
| `reporting.customer_rates` | map | | Hourly rates per customer |
| `reporting.customer_aliases` | map | | Short name → full name mappings |
| `reporting.customer_addresses` | map | | Customer addresses for report headers |
| `reporting.day_worked_min_threshold` | int | 120 | Minimum minutes to count a day as worked |
| `reporting.report_statistics` | bool | true | Include statistics section in reports |
| `sync.enable` | bool | false | Enable sync on `tt sync` |
| `github.username` | string | | GitHub username |
| `github.access_token` | string | | GitHub access token (supports pipe notation) |
| `github.root_uri` | string | `https://github.com` | Override for GitHub Enterprise |
| `gitlab.api_root` | string | `https://gitlab.com/api/v4/` | GitLab API root |
| `gitlab.api_key` | string | | GitLab API key (supports pipe notation) |
| `sourcehut.api_root` | string | | Sourcehut API root |
| `sourcehut.access_token` | string | | Sourcehut access token (supports pipe notation) |
| `sourcehut.username` | string | | Sourcehut username |
| `linear.default_org` | string | | Linear organization slug |
| `linear.api_key` | string | | Linear API key (supports pipe notation) |
| `item_cache_ttl_days` | int | 30 | Days to cache task metadata |

**Pipe notation:** any secret value ending with `|` is treated as a shell
command whose stdout is the secret. Example:
```yaml
github:
  access_token: cat /run/secrets/github-token|
```

## Usage

```
tt                          # list today's entries (default)
tt start "writing tests"    # start a new entry
tt stop                     # stop the current entry
tt resume                   # resume the last entry
tt resume -2                # resume the second-to-last entry
tt list                     # list today's entries
tt list -d monday           # list entries for last Monday
tt edit                     # open today's file in $EDITOR
tt edit -d 2024-03-15       # open a specific date
tt sync                     # pull task metadata from external services
tt report --thisweek        # report for the current week
tt report --lastmonth -f html -o report.html
```

### Date formats

Most commands accept `-d DATE`. Supported formats:

- `today`, `yesterday`
- `monday` / `mon`, `tuesday` / `tue`, … (most recent occurrence)
- `YYYY-MM-DD`, `YYYY/MM/DD`
- `MM-DD`, `DD`

### Report formats

| Flag | Output |
|------|--------|
| *(default)* | Colour terminal table |
| `-f md` | Markdown |
| `-f html` | HTML |
| `-f typst` | Typst source |
| `-f pdf` | PDF (requires `typst` in PATH) |

## Architecture

Time entries are stored as CSV files under `directory`:

```
$HOME/.tracktime/
  2024/
    01/
      15        ← one file per day
      16
    02/
      ...
```

Each day file has the header:
```
start,stop,type,project,taskid,customer,description
```

`type` is the service the task lives on (`github`, `gitlab`, `sourcehut`,
`linear`, or arbitrary text). `taskid` is the issue/PR/MR number. `tt` uses
these to fetch task titles and hyperlinks.

## Guiding Principles

- Filesystem-based — use Git to version your time entries
- Easy to edit manually — plain CSV, no binary format
- Offline-first — works without internet; metadata is cached locally

## Contributing

Contributions welcome. Please open issues or submit PRs on GitHub.

```
go test ./...
pre-commit run -av
```
