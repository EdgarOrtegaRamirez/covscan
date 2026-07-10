// Package model defines the shared data structures for coverage reports.
package model

import (
	"fmt"
	"strings"
)

// CoverageType indicates the type of coverage data.
type CoverageType string

const (
	CoverageLine   CoverageType = "line"
	CoverageBranch CoverageType = "branch"
)

// FileCoverage holds coverage data for a single file.
type FileCoverage struct {
	FilePath        string
	TotalLines      int
	CoveredLines    int
	TotalBranches   int
	CoveredBranches int
}

// LineRate returns the line coverage ratio (0.0 to 1.0).
func (f FileCoverage) LineRate() float64 {
	if f.TotalLines == 0 {
		return 0
	}
	return float64(f.CoveredLines) / float64(f.TotalLines)
}

// BranchRate returns the branch coverage ratio (0.0 to 1.0).
func (f FileCoverage) BranchRate() float64 {
	if f.TotalBranches == 0 {
		return 0
	}
	return float64(f.CoveredBranches) / float64(f.TotalBranches)
}

// OverallRate returns the combined coverage rate.
func (f FileCoverage) OverallRate() float64 {
	total := f.TotalLines + f.TotalBranches
	if total == 0 {
		return 0
	}
	return float64(f.CoveredLines+f.CoveredBranches) / float64(total)
}

// CoverageReport holds the complete coverage report.
type CoverageReport struct {
	Name      string
	Files     []FileCoverage
	SourceDir string
	HasBranch bool
}

// TotalLines sums all file lines.
func (r CoverageReport) TotalLines() int {
	t := 0
	for _, f := range r.Files {
		t += f.TotalLines
	}
	return t
}

// TotalCoveredLines sums all covered lines.
func (r CoverageReport) TotalCoveredLines() int {
	t := 0
	for _, f := range r.Files {
		t += f.CoveredLines
	}
	return t
}

// LineCoverage returns the overall line coverage ratio.
func (r CoverageReport) LineCoverage() float64 {
	total := r.TotalLines()
	if total == 0 {
		return 0
	}
	return float64(r.TotalCoveredLines()) / float64(total)
}

// DirSummary aggregates coverage by directory.
type DirSummary struct {
	DirPath      string
	TotalLines   int
	CoveredLines int
	FileCount    int
}

// GroupByDir groups file coverage by directory.
func (r CoverageReport) GroupByDir() []DirSummary {
	dirs := make(map[string]*DirSummary)
	var keys []string
	for _, f := range r.Files {
		dir := dirFromPath(f.FilePath)
		if _, ok := dirs[dir]; !ok {
			dirs[dir] = &DirSummary{DirPath: dir}
			keys = append(keys, dir)
		}
		ds := dirs[dir]
		ds.TotalLines += f.TotalLines
		ds.CoveredLines += f.CoveredLines
		ds.FileCount++
	}
	result := make([]DirSummary, len(keys))
	for i, k := range keys {
		result[i] = *dirs[k]
	}
	return result
}

func dirFromPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return "."
	}
	if idx == 0 {
		return "/"
	}
	return path[:idx]
}

// CoverageBar returns a colored coverage bar string.
func CoverageBar(rate float64, width int) string {
	filled := int(rate * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	empty := width - filled

	var color string
	switch {
	case rate >= 0.8:
		color = "\033[32m" // green
	case rate >= 0.5:
		color = "\033[33m" // yellow
	default:
		color = "\033[31m" // red
	}
	reset := "\033[0m"

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("%s%s%s", color, bar, reset)
}

// FormatPercent formats a rate as a percentage string.
func FormatPercent(rate float64) string {
	return fmt.Sprintf("%5.1f%%", rate*100)
}
