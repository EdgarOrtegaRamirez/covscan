# CovScan — Multi-Language Code Coverage Report Viewer & Aggregator

**CovScan** is a fast, cross-platform CLI tool for parsing, aggregating, viewing, and comparing code coverage reports from multiple formats. It helps developers and CI/CD pipelines quickly assess code coverage quality across any language.

## Features

- **Multi-Format Parsing** — Parse coverage reports from:
  - **Go**: `go test -coverprofile=cover.out` format
  - **Python**: Cobertura XML (`coverage.xml`)
  - **C/C++/Rust**: LCOV tracefiles (`.info`, `.lcov`)
  - **JavaScript/TypeScript**: Istanbul JSON (`coverage/coverage-final.json`)
- **Auto-Detection** — Automatically detects format from file extension or first-line content
- **Per-Directory Aggregation** — Groups files by directory with aggregate coverage
- **Per-File Breakdown** — Optional `--all` flag shows every file's coverage
- **Color-Coded Bars** — Visual coverage bars (green ≥80%, yellow ≥50%, red <50%)
- **Branch Coverage** — Displays branch coverage when available (LCOV, Cobertura)
- **Coverage Comparison** — Compare two coverage reports with per-file diffs
- **CI/CD Mode** — `--threshold 80` exits with code 1 if coverage is below the minimum
- **Multiple Output Formats** — Text (default) and JSON (`--json`)
- **Compact Summary** — One-line per-file summary with `covscan summary`

## Installation

### From source (Go 1.25+)

```bash
go install github.com/EdgarOrtegaRamirez/covscan/cmd/covscan@latest
```

### From binary

Download the latest release from [GitHub Releases](https://github.com/EdgarOrtegaRamirez/covscan/releases).

## Quick Start

```bash
# Generate a Go coverage profile
go test ./... -coverprofile=cover.out

# View coverage
covscan cover cover.out

# View with per-file breakdown
covscan cover cover.out --all

# View as JSON
covscan cover cover.out --json

# CI/CD: fail if below 80%
covscan cover cover.out --threshold 80

# Compare two coverage runs
covscan compare old.out new.out

# Compact summary
covscan summary cover.out
```

## Usage

### `covscan cover [files...]`

Parse one or more coverage report files and display coverage statistics.

**Flags:**
| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--all` | `-a` | `false` | Show per-file breakdown |
| `--json` | `-j` | `false` | Output as JSON |
| `--threshold` | `-t` | `0` | Minimum coverage percentage (exit 1 if below) |
| `--quiet` | `-q` | `false` | Suppress output, only check threshold |

### `covscan summary [files...]`

Display a one-line summary per coverage report file.

### `covscan compare <old> <new>`

Parse and compare two coverage reports, showing per-file differences.

## Supported Coverage Formats

| Format | File Extensions | First-Line Detection | Branch Coverage |
|--------|----------------|---------------------|-----------------|
| Go coverprofile | `.out`, `.cov` | `mode:` | No |
| Cobertura XML | `.xml` | `<coverage` | Yes (partial) |
| LCOV tracefile | `.info`, `.lcov` | `SF:` | Yes |
| Istanbul JSON | `.json` | `{"file":` | Yes |

## Examples

### Python project

```bash
# Generate coverage
coverage run -m pytest
coverage xml -o coverage.xml

# View
covscan cover coverage.xml --all
```

### Rust project

```bash
# Generate LCOV coverage
cargo tarpaulin --out lcov --output-dir coverage

# View
covscan cover coverage/lcov.info
```

### JavaScript/TypeScript project

```bash
# Generate coverage (using c8, nyc, or jest)
npx c8 report --reporter=json

# View
covscan cover coverage/coverage-final.json
```

### CI/CD pipeline

```bash
covscan cover coverage.out --threshold 80
if [ $? -ne 0 ]; then
  echo "❌ Coverage below threshold!"
  exit 1
fi
```

## Architecture

```
covscan/
├── cmd/covscan/        # CLI entry point (Cobra)
├── internal/
│   ├── model/          # Shared data structures (CoverageReport, FileCoverage)
│   ├── parser/         # Coverage format parsers
│   │   ├── gocover.go  # Go coverprofile parser
│   │   ├── cobertura.go # Cobertura XML parser
│   │   └── parser.go   # LCOV, Istanbul parsers, auto-detection
│   ├── reporter/       # Output formatters (text, JSON)
│   └── comparer/       # Coverage comparison engine
└── go.mod
```

## License

MIT — see [LICENSE](LICENSE).