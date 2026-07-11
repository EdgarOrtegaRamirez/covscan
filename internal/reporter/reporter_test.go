package reporter

import (
	"strings"
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

func TestMarkdownReporterFormat(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "src/main.go", TotalLines: 100, CoveredLines: 80},
			{FilePath: "src/util.go", TotalLines: 50, CoveredLines: 25},
		},
	}

	r := NewMarkdownReporter()
	lines := r.Format(report)

	if len(lines) == 0 {
		t.Fatal("expected non-empty output")
	}

	// Check heading
	found := false
	for _, l := range lines {
		if l == "## Coverage Report: test.out" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected markdown heading with report name")
	}

	// Check table header
	found = false
	for _, l := range lines {
		if l == "| Metric | Value |" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected markdown table header")
	}

	// Check coverage percentage in table
	found = false
	for _, l := range lines {
		if l == "| **Overall Coverage** | 70.0% |" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected overall coverage in markdown output")
	}

	// Check badge
	found = false
	for _, l := range lines {
		if strings.HasPrefix(l, "![Coverage]") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected coverage badge in markdown output")
	}

	// Check directory summary table
	found = false
	for _, l := range lines {
		if l == "| Directory | Coverage | Files |" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected directory summary table header in markdown output")
	}
}

func TestMarkdownReporterShowAllFiles(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 10, CoveredLines: 8},
			{FilePath: "util.go", TotalLines: 20, CoveredLines: 15},
		},
	}

	r := NewMarkdownReporter()
	r.ShowAllFiles = true
	lines := r.Format(report)

	// Check per-file breakdown table
	found := false
	for _, l := range lines {
		if l == "### Per-File Breakdown" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected per-file breakdown section when ShowAllFiles is true")
	}

	// Check file listing in table
	found = false
	for _, l := range lines {
		if l == "| `main.go` | 80.0% |" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected file listing in per-file breakdown")
	}
}

func TestMarkdownReporterWithBranches(t *testing.T) {
	report := &model.CoverageReport{
		Name:      "test.xml",
		HasBranch: true,
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 80, TotalBranches: 10, CoveredBranches: 7},
		},
	}

	r := NewMarkdownReporter()
	lines := r.Format(report)

	found := false
	for _, l := range lines {
		if len(l) > 12 && l[:5] == "| **B" && l[5:12] == "ranch C" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected branch coverage row in markdown output")
	}
}

func TestMarkdownReporterThreshold(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 10, CoveredLines: 3},
		},
	}

	r := NewMarkdownReporter()
	r.Threshold = 80

	// Coverage is 30%, below 80% threshold
	code := r.Report(report)
	if code != 1 {
		t.Errorf("expected exit code 1 (below threshold), got %d", code)
	}

	// Test passing threshold
	r2 := NewMarkdownReporter()
	r2.Threshold = 20
	code2 := r2.Report(report)
	if code2 != 0 {
		t.Errorf("expected exit code 0 (above threshold), got %d", code2)
	}
}

func TestMarkdownReporterEmptyReport(t *testing.T) {
	report := &model.CoverageReport{
		Name: "empty.out",
	}

	r := NewMarkdownReporter()
	lines := r.Format(report)

	if len(lines) == 0 {
		t.Fatal("expected non-empty output for empty report")
	}

	// Should still have heading
	if lines[0] != "## Coverage Report: empty.out" {
		t.Errorf("expected heading, got %q", lines[0])
	}
}

func TestCoverageBadge(t *testing.T) {
	tests := []struct {
		rate     float64
		expected string
	}{
		{0.90, "brightgreen"},
		{0.80, "brightgreen"},
		{0.70, "yellow"},
		{0.50, "yellow"},
		{0.40, "red"},
		{0.00, "red"},
	}

	for _, tt := range tests {
		result := coverageBadge(tt.rate)
		if len(result) == 0 {
			t.Errorf("expected non-empty badge for rate %.2f", tt.rate)
		}
		// Check it starts with the image tag
		if !strings.HasPrefix(result, "![Coverage]") {
			t.Errorf("expected badge to start with ![Coverage], got %q", result[:15])
		}
	}
}

func TestMarkdownReporterNoShowAllFiles(t *testing.T) {
	report := &model.CoverageReport{
		Name: "test.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 10, CoveredLines: 8},
			{FilePath: "util.go", TotalLines: 20, CoveredLines: 15},
		},
	}

	r := NewMarkdownReporter()
	r.ShowAllFiles = false
	lines := r.Format(report)

	// Should NOT have per-file breakdown
	for _, l := range lines {
		if l == "### Per-File Breakdown" {
			t.Error("should not have per-file breakdown when ShowAllFiles is false")
			break
		}
	}
}
