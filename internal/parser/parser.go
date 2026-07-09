// Package parser provides coverage report parsers for multiple formats.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

// DetectFormat detects the coverage format from file extension/content.
func DetectFormat(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".out", ".cov":
		return "go-cover"
	case ".xml":
		return "cobertura"
	case ".info", ".lcov":
		return "lcov"
	case ".json":
		return "istanbul"
	}

	// Try detecting by reading first line
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	firstLine := strings.SplitN(string(data), "\n", 2)[0]

	if strings.HasPrefix(firstLine, "mode:") {
		return "go-cover"
	}
	if strings.Contains(firstLine, "<coverage") {
		return "cobertura"
	}
	if strings.HasPrefix(firstLine, "SF:") {
		return "lcov"
	}

	return ""
}

// ParseAuto detects the format and parses the file.
func ParseAuto(path string) (*model.CoverageReport, error) {
	format := DetectFormat(path)
	if format == "" {
		return nil, fmt.Errorf("cannot detect coverage format for %s", path)
	}

	switch format {
	case "go-cover":
		return ParseGoCover(path)
	case "cobertura":
		return ParseCobertura(path)
	case "lcov":
		return ParseLCOV(path)
	case "istanbul":
		return ParseIstanbul(path)
	default:
		return nil, fmt.Errorf("unsupported coverage format: %s", format)
	}
}

// istanbulReport is the top-level Istanbul JSON structure.
type istanbulFileCoverage struct {
	Path         string                `json:"path"`
	StatementMap map[string]istanbulSpan `json:"statementMap"`
	S            map[string]int          `json:"s"`
	BranchMap    map[string]istanbulBranch `json:"branchMap"`
	B            map[string][]int         `json:"b"`
}

type istanbulSpan struct {
	Start struct {
		Line int `json:"line"`
	} `json:"start"`
	End struct {
		Line int `json:"line"`
	} `json:"end"`
}

type istanbulBranch struct {
	Line      int              `json:"line"`
	Type      string           `json:"type"`
	Locations []istanbulSpan   `json:"locations"`
}

// ParseIstanbul parses an Istanbul JSON coverage report.
func ParseIstanbul(path string) (*model.CoverageReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	// Istanbul JSON is keyed by file path
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("cannot parse JSON in %s: %w", path, err)
	}

	report := &model.CoverageReport{
		Name:      filepath.Base(path),
		SourceDir: filepath.Dir(path),
	}

	for filePath, raw := range rawMap {
		// Check if this entry looks like file coverage (has "s" field)
		var check struct {
			S map[string]int `json:"s"`
		}
		if err := json.Unmarshal(raw, &check); err != nil || check.S == nil {
			continue
		}

		var fcData istanbulFileCoverage
		if err := json.Unmarshal(raw, &fcData); err != nil {
			continue
		}

		fc := model.FileCoverage{
			FilePath: filePath,
		}

		// Count statement coverage
		coveredLines := make(map[int]bool)
		totalLines := make(map[int]bool)

		for stmtID, stmt := range fcData.StatementMap {
			for l := stmt.Start.Line; l <= stmt.End.Line; l++ {
				totalLines[l] = true
				if count, ok := fcData.S[stmtID]; ok && count > 0 {
					coveredLines[l] = true
				}
			}
		}

		fc.TotalLines = len(totalLines)
		fc.CoveredLines = len(coveredLines)

		// Count branch coverage
		if len(fcData.BranchMap) > 0 {
			report.HasBranch = true
			for branchID, branch := range fcData.BranchMap {
				if counts, ok := fcData.B[branchID]; ok {
					for _, c := range counts {
						fc.TotalBranches++
						if c > 0 {
							fc.CoveredBranches++
						}
					}
				}
				_ = branch // branch metadata not needed for counting
			}
		}

		if fc.TotalLines > 0 {
			report.Files = append(report.Files, fc)
		}
	}

	return report, nil
}

// LCOV parser helper functions
func parseLCOVLine(line string) (string, string, bool) {
	for i := 0; i < len(line); i++ {
		if line[i] == ':' {
			return line[:i], line[i+1:], true
		}
	}
	return "", "", false
}

// ParseLCOV parses an LCOV tracefile.
func ParseLCOV(path string) (*model.CoverageReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	report := &model.CoverageReport{
		Name:      filepath.Base(path),
		SourceDir: filepath.Dir(path),
	}

	type lcovLineDA struct {
		lineNo int
		count  int
	}
	type lcovBranchDA struct {
		lineNo int
		taken  string
	}

	files := make(map[string]struct {
		lines   []lcovLineDA
		branches []lcovBranchDA
	})
	var currentFile string

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		tag, value, ok := parseLCOVLine(line)
		if !ok {
			continue
		}

		switch tag {
		case "SF":
			currentFile = value
			if _, ok := files[currentFile]; !ok {
				files[currentFile] = struct {
					lines    []lcovLineDA
					branches []lcovBranchDA
				}{}
			}
		case "DA":
			// DA:lineNumber,count
			parts := strings.SplitN(value, ",", 2)
			if len(parts) != 2 {
				continue
			}
			lineNo := atoi(parts[0])
			count := atoi(parts[1])
			entry := files[currentFile]
			entry.lines = append(entry.lines, lcovLineDA{lineNo, count})
			files[currentFile] = entry
		case "BRDA":
			// BRDA:blockNo,branchNo,lineNo,taken
			parts := strings.Split(value, ",")
			if len(parts) < 4 {
				continue
			}
			lineNo := atoi(parts[2])
			entry := files[currentFile]
			entry.branches = append(entry.branches, lcovBranchDA{lineNo, parts[3]})
			files[currentFile] = entry
		}
	}

	for filePath, fd := range files {
		fc := model.FileCoverage{
			FilePath: filePath,
		}

		for _, ld := range fd.lines {
			fc.TotalLines++
			if ld.count > 0 {
				fc.CoveredLines++
			}
		}

		if len(fd.branches) > 0 {
			report.HasBranch = true
			for _, bd := range fd.branches {
				fc.TotalBranches++
				if bd.taken != "-" && bd.taken != "0" {
					fc.CoveredBranches++
				}
			}
		}

		if fc.TotalLines > 0 {
			report.Files = append(report.Files, fc)
		}
	}

	return report, nil
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}