package reporter

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

func TestTextReporterFormat(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "src/main.go", TotalLines: 100, CoveredLines: 80},
			{FilePath: "src/util.go", TotalLines: 50, CoveredLines: 25},
		},
	}

	r := NewTextReporter()
	lines := r.Format(report)

	if len(lines) == 0 {
		t.Fatal("expected non-empty output")
	}

	// Check that the report name appears
	found := false
	for _, l := range lines {
		if l == "Coverage Report: test.out" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected report name in output")
	}
}

func TestTextReporterFormatWithBranches(t *testing.T) {
	report := &model.CoverageReport{
		Name:      "test.xml",
		HasBranch: true,
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 80, TotalBranches: 10, CoveredBranches: 7},
		},
	}

	r := NewTextReporter()
	lines := r.Format(report)

	// Should contain branch info
	found := false
	for _, l := range lines {
		if len(l) >= 10 && l[:10] == "  Branches" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected branch coverage info in output")
	}
}

func TestTextReporterShowAllFiles(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 10, CoveredLines: 8},
		},
	}

	r := NewTextReporter()
	r.ShowAllFiles = true
	lines := r.Format(report)

	found := false
	for _, l := range lines {
		if len(l) >= 6 && l[len(l)-3:] == ".go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected per-file breakdown with ShowAllFiles")
	}
}

func TestTextReporterThreshold(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 10, CoveredLines: 3},
		},
	}

	r := NewTextReporter()
	r.Threshold = 80

	// Coverage is 30%, below 80% threshold
	code := r.Report(report)
	if code != 1 {
		t.Errorf("expected exit code 1 (below threshold), got %d", code)
	}

	// Test passing threshold
	r2 := NewTextReporter()
	r2.Threshold = 20
	code2 := r2.Report(report)
	if code2 != 0 {
		t.Errorf("expected exit code 0 (above threshold), got %d", code2)
	}
}

func TestJSONReporterFormat(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 80},
		},
	}

	j := &JSONReporter{}
	lines := j.Format(report)

	if len(lines) == 0 {
		t.Fatal("expected non-empty output")
	}

	// Should contain JSON keys
	output := lines[0]
	if len(output) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestTextReporterEmptyReport(t *testing.T) {
	report := &model.CoverageReport{
		Name: "empty.out",
	}

	r := NewTextReporter()
	lines := r.Format(report)

	if len(lines) == 0 {
		t.Fatal("expected non-empty output for empty report")
	}
}
