// Package comparer provides coverage report comparison between two runs.
package comparer

import (
	"fmt"
	"sort"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

// ComparisonResult holds the diff between two coverage reports.
type ComparisonResult struct {
	OldReport *model.CoverageReport
	NewReport *model.CoverageReport

	FileChanges []FileChange
}

// FileChange describes how a file's coverage changed between two runs.
type FileChange struct {
	FilePath         string
	OldLines         int
	NewLines         int
	OldCovered       int
	NewCovered       int
	OldRate          float64
	NewRate          float64
}

// Compare compares two coverage reports and returns the differences.
func Compare(old, new *model.CoverageReport) *ComparisonResult {
	result := &ComparisonResult{
		OldReport: old,
		NewReport: new,
	}

	// Build lookup maps
	oldByFile := make(map[string]model.FileCoverage)
	for _, f := range old.Files {
		oldByFile[f.FilePath] = f
	}
	newByFile := make(map[string]model.FileCoverage)
	for _, f := range new.Files {
		newByFile[f.FilePath] = f
	}

	// Find all unique file paths
	allFiles := make(map[string]bool)
	for _, f := range old.Files {
		allFiles[f.FilePath] = true
	}
	for _, f := range new.Files {
		allFiles[f.FilePath] = true
	}

	for filePath := range allFiles {
		oldF, oldExists := oldByFile[filePath]
		newF, newExists := newByFile[filePath]

		change := FileChange{FilePath: filePath}

		if oldExists {
			change.OldLines = oldF.TotalLines
			change.OldCovered = oldF.CoveredLines
			change.OldRate = oldF.LineRate()
		}
		if newExists {
			change.NewLines = newF.TotalLines
			change.NewCovered = newF.CoveredLines
			change.NewRate = newF.LineRate()
		}

		result.FileChanges = append(result.FileChanges, change)
	}

	sort.Slice(result.FileChanges, func(i, j int) bool {
		return result.FileChanges[i].FilePath < result.FileChanges[j].FilePath
	})

	return result
}

// Format returns a human-readable comparison report.
func (c *ComparisonResult) Format() string {
	oldRate := c.OldReport.LineCoverage()
	newRate := c.NewReport.LineCoverage()

	out := fmt.Sprintf("Coverage Comparison: %s → %s\n", c.OldReport.Name, c.NewReport.Name)
	out += fmt.Sprintf("  Old coverage: %s\n", formatPercent(oldRate))
	out += fmt.Sprintf("  New coverage: %s", formatPercent(newRate))

	diff := newRate - oldRate
	if diff > 0 {
		out += fmt.Sprintf("  (↑ %s)\n", formatPercent(diff))
	} else if diff < 0 {
		out += fmt.Sprintf("  (↓ %s)\n", formatPercent(-diff))
	} else {
		out += "  (no change)\n"
	}
	out += fmt.Sprintf("  Files: %d → %d\n", len(c.OldReport.Files), len(c.NewReport.Files))
	out += "\n"

	out += "  Changes:\n"
	for _, fc := range c.FileChanges {
		oldStr := fmt.Sprintf("%.1f%%", fc.OldRate*100)
		newStr := fmt.Sprintf("%.1f%%", fc.NewRate*100)

		arrow := "─"
		if fc.NewRate > fc.OldRate {
			arrow = "↑"
		} else if fc.NewRate < fc.OldRate {
			arrow = "↓"
		}

		out += fmt.Sprintf("    %s %s → %s %s %s\n",
			arrow, oldStr, newStr, formatPercentDiff(fc.NewRate-fc.OldRate), fc.FilePath)
	}

	return out
}

func formatPercent(rate float64) string {
	return fmt.Sprintf("%5.1f%%", rate*100)
}

func formatPercentDiff(diff float64) string {
	if diff >= 0 {
		return fmt.Sprintf("(+%.1f%%)", diff*100)
	}
	return fmt.Sprintf("(%.1f%%)", diff*100)
}