# Contributing

Contributions are welcome! Please open issues on the repository, or submit a PR.

## Issue Reporting

Please report any bugs and suggest features by creating an issue.

**When reporting a bug**, please be as specific as possible, and include steps
to reproduce.

## Code

If you want to propose a code change, please submit a PR.

### Setup

You need Go 1.25+ and [pre-commit](https://pre-commit.com/):

```
git clone https://github.com/sumnerevans/tracktime
cd tracktime
pre-commit install
```

A Nix devShell is available if you use Nix + direnv:

```
direnv allow
```

### Running tests

```
go test ./...
```

### Linting

```
pre-commit run -av
```

The CI runs `goimports`, `go vet`, and `staticcheck` via pre-commit.

### Code style

Standard Go conventions enforced by `goimports` and `staticcheck`. Export only
identifiers that need to be used outside the package.
