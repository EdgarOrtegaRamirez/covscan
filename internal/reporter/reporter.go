// Package reporter formats and outputs coverage reports.
package reporter

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

// TextReporter outputs coverage data in a human-readable text format.
type TextReporter struct {
	ShowAllFiles bool
	Width        int
	Threshold    float64
}

// NewTextReporter creates a new TextReporter.
func NewTextReporter() *TextReporter {
	return &TextReporter{
		ShowAllFiles: false,
		Width:        30,
		Threshold:    0.0,
	}
}

// Report writes the coverage report to stdout.
func (r *TextReporter) Report(report *model.CoverageReport) int {
	lines := r.Format(report)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Check threshold (threshold is in percentage, e.g. 80 = 80%)
	coveragePct := report.LineCoverage() * 100
	if r.Threshold > 0 && coveragePct < r.Threshold {
		fmt.Fprintf(os.Stderr, "\n❌ Coverage %.1f%% is below threshold %.1f%%\n",
			coveragePct, r.Threshold)
		return 1
	}
	return 0
}

// Format returns the coverage report as formatted strings.
func (r *TextReporter) Format(report *model.CoverageReport) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Coverage Report: %s", report.Name))
	lines = append(lines, strings.Repeat("─", 60))
	lines = append(lines, "")

	// Sort files
	sorted := make([]model.FileCoverage, len(report.Files))
	copy(sorted, report.Files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FilePath < sorted[j].FilePath
	})

	// Summary line
	lineRate := report.LineCoverage()
	lines = append(lines, fmt.Sprintf("  Overall: %s %s",
		model.CoverageBar(lineRate, r.Width),
		model.FormatPercent(lineRate)))
	lines = append(lines, fmt.Sprintf("  Files:   %d", len(report.Files)))
	lines = append(lines, fmt.Sprintf("  Lines:   %d / %d covered",
		report.TotalCoveredLines(), report.TotalLines()))

	if report.HasBranch {
		totalBranches := 0
		coveredBranches := 0
		for _, f := range report.Files {
			totalBranches += f.TotalBranches
			coveredBranches += f.CoveredBranches
		}
		branchRate := float64(0)
		if totalBranches > 0 {
			branchRate = float64(coveredBranches) / float64(totalBranches)
		}
		lines = append(lines, fmt.Sprintf("  Branches: %s %s",
			model.CoverageBar(branchRate, r.Width),
			model.FormatPercent(branchRate)))
	}

	lines = append(lines, "")

	// Per-directory summary
	dirs := report.GroupByDir()
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].DirPath < dirs[j].DirPath
	})
	lines = append(lines, "  Per-Directory Summary:")
	lines = append(lines, "")
	for _, d := range dirs {
		rate := float64(0)
		if d.TotalLines > 0 {
			rate = float64(d.CoveredLines) / float64(d.TotalLines)
		}
		lines = append(lines, fmt.Sprintf("    %s %s  %s (%d files)",
			model.CoverageBar(rate, r.Width),
			model.FormatPercent(rate),
			d.DirPath,
			d.FileCount))
	}

	lines = append(lines, "")

	// Per-file details
	if r.ShowAllFiles {
		lines = append(lines, "  Per-File Breakdown:")
		lines = append(lines, "")
		for _, f := range sorted {
			rate := f.LineRate()
			lines = append(lines, fmt.Sprintf("    %s %s  %s",
				model.CoverageBar(rate, r.Width),
				model.FormatPercent(rate),
				f.FilePath))
		}
		lines = append(lines, "")
	}

	return lines
}

// JSONReporter outputs coverage data as JSON.
type JSONReporter struct{}

// Report writes the coverage report as JSON to stdout.
func (j *JSONReporter) Report(report *model.CoverageReport) int {
	lines := j.Format(report)
	for _, line := range lines {
		fmt.Println(line)
	}
	return 0
}

// Format returns the coverage report as JSON strings.
func (j *JSONReporter) Format(report *model.CoverageReport) []string {
	type fileJSON struct {
		File            string  `json:"file"`
		TotalLines      int     `json:"total_lines"`
		CoveredLines    int     `json:"covered_lines"`
		LineCoverage    float64 `json:"line_coverage"`
		TotalBranches   int     `json:"total_branches,omitempty"`
		CoveredBranches int     `json:"covered_branches,omitempty"`
	}

	type reportJSON struct {
		Name         string     `json:"name"`
		TotalFiles   int        `json:"total_files"`
		TotalLines   int        `json:"total_lines"`
		CoveredLines int        `json:"covered_lines"`
		LineCoverage float64    `json:"line_coverage"`
		Files        []fileJSON `json:"files"`
	}

	rj := reportJSON{
		Name:         report.Name,
		TotalFiles:   len(report.Files),
		TotalLines:   report.TotalLines(),
		CoveredLines: report.TotalCoveredLines(),
		LineCoverage: report.LineCoverage(),
	}

	for _, f := range report.Files {
		rj.Files = append(rj.Files, fileJSON{
			File:            f.FilePath,
			TotalLines:      f.TotalLines,
			CoveredLines:    f.CoveredLines,
			LineCoverage:    f.LineRate(),
			TotalBranches:   f.TotalBranches,
			CoveredBranches: f.CoveredBranches,
		})
	}

	// Simple JSON output without encoding/json dependency
	lines := []string{fmt.Sprintf(
		`{"name":%q,"total_files":%d,"total_lines":%d,"covered_lines":%d,"line_coverage":%.4f,"files":[`,
		rj.Name, rj.TotalFiles, rj.TotalLines, rj.CoveredLines, rj.LineCoverage,
	)}

	for i, f := range rj.Files {
		if i > 0 {
			lines[0] += ","
		}
		lines[0] += fmt.Sprintf(
			`{"file":%q,"total_lines":%d,"covered_lines":%d,"line_coverage":%.4f}`,
			f.File, f.TotalLines, f.CoveredLines, f.LineCoverage,
		)
	}
	lines[0] += "]}"

	return lines
}
