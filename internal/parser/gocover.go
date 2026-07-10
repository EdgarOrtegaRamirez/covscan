package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

// ParseGoCover parses a Go coverprofile file.
// Format: "mode: set" / "path/to/file.go:startLine.startCol,endLine.endCol numStmt count"
func ParseGoCover(path string) (*model.CoverageReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	report := &model.CoverageReport{
		Name:      filepath.Base(path),
		SourceDir: filepath.Dir(path),
	}

	fileLines := make(map[string]map[int]bool) // file -> covered line numbers
	fileTotal := make(map[string]int)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 3 {
			continue
		}

		locPart := parts[0]
		countStr := parts[len(parts)-1]

		colonIdx := strings.LastIndex(locPart, ":")
		if colonIdx < 0 {
			continue
		}
		filePath := locPart[:colonIdx]
		rangePart := locPart[colonIdx+1:]

		commaIdx := strings.Index(rangePart, ",")
		if commaIdx < 0 {
			continue
		}
		startPart := rangePart[:commaIdx]
		endPart := rangePart[commaIdx+1:]

		dotIdx := strings.Index(startPart, ".")
		if dotIdx < 0 {
			continue
		}
		startLine, err := strconv.Atoi(startPart[:dotIdx])
		if err != nil {
			continue
		}

		endDot := strings.Index(endPart, ".")
		if endDot < 0 {
			continue
		}
		endLine, err := strconv.Atoi(endPart[:endDot])
		if err != nil {
			continue
		}

		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}

		if _, ok := fileTotal[filePath]; !ok {
			fileTotal[filePath] = 0
			fileLines[filePath] = make(map[int]bool)
		}

		blockLines := endLine - startLine + 1
		fileTotal[filePath] += blockLines

		if count > 0 {
			for l := startLine; l <= endLine; l++ {
				fileLines[filePath][l] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	for filePath, totalLines := range fileTotal {
		report.Files = append(report.Files, model.FileCoverage{
			FilePath:     filePath,
			TotalLines:   totalLines,
			CoveredLines: len(fileLines[filePath]),
		})
	}

	return report, nil
}
