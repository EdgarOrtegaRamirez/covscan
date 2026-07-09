# AGENTS.md — for AI Coding Agents

## Project Overview

**CovScan** is a multi-language code coverage report viewer & aggregator CLI written in Go. It parses coverage reports from Go (coverprofile), Python (Cobertura XML), C/C++/Rust (LCOV), and JavaScript/TypeScript (Istanbul JSON), displays per-file/per-directory coverage statistics, compares two runs, and supports CI/CD threshold enforcement.

## Key Files

| File | Purpose |
|------|---------|
| `cmd/covscan/main.go` | CLI entry point with Cobra commands |
| `internal/model/model.go` | Shared data structures (CoverageReport, FileCoverage) |
| `internal/parser/gocover.go` | Go coverprofile parser |
| `internal/parser/cobertura.go` | Cobertura XML parser |
| `internal/parser/parser.go` | LCOV & Istanbul parsers, format detection |
| `internal/reporter/reporter.go` | Text and JSON output formatters |
| `internal/comparer/comparer.go` | Coverage report comparison engine |
| `internal/model/model_test.go` | Model tests |
| `internal/parser/parser_test.go` | Parser tests |
| `internal/reporter/reporter_test.go` | Reporter tests |
| `internal/comparer/comparer_test.go` | Comparer tests |

## Architecture

```
CLI (Cobra)
  → Parser (auto-detect format, parse into CoverageReport)
    → Reporter (format output as text or JSON)
    → Comparer (diff two CoverageReports)
```

## Build & Test

```bash
go build ./cmd/covscan/
go test ./... -count=1
go vet ./...
```

## Adding a New Coverage Format

1. Add a new parser function in `internal/parser/` that returns `*model.CoverageReport`
2. Add detection in `DetectFormat()` in `parser.go`
3. Add the format to `ParseAuto()` switch in `parser.go`
4. Write tests for the new parser
5. Update README.md supported formats table

## Test Data

Test files in `internal/parser/parser_test.go` create temporary files to test each parser. Use `t.TempDir()` for test isolation.