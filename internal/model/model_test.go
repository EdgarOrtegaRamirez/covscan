package model

import (
	"testing"
)

func TestFileCoverageLineRate(t *testing.T) {
	f := FileCoverage{TotalLines: 100, CoveredLines: 75}
	if f.LineRate() != 0.75 {
		t.Errorf("expected 0.75, got %f", f.LineRate())
	}

	f2 := FileCoverage{TotalLines: 0}
	if f2.LineRate() != 0 {
		t.Errorf("expected 0, got %f", f2.LineRate())
	}
}

func TestFileCoverageBranchRate(t *testing.T) {
	f := FileCoverage{TotalBranches: 10, CoveredBranches: 8}
	if f.BranchRate() != 0.8 {
		t.Errorf("expected 0.8, got %f", f.BranchRate())
	}

	f2 := FileCoverage{}
	if f2.BranchRate() != 0 {
		t.Errorf("expected 0, got %f", f2.BranchRate())
	}
}

func TestReportTotals(t *testing.T) {
	r := CoverageReport{
		Files: []FileCoverage{
			{TotalLines: 100, CoveredLines: 80},
			{TotalLines: 50, CoveredLines: 40},
		},
	}

	if r.TotalLines() != 150 {
		t.Errorf("expected 150, got %d", r.TotalLines())
	}
	if r.TotalCoveredLines() != 120 {
		t.Errorf("expected 120, got %d", r.TotalCoveredLines())
	}
	if r.LineCoverage() != 0.8 {
		t.Errorf("expected 0.8, got %f", r.LineCoverage())
	}
}

func TestGroupByDir(t *testing.T) {
	r := CoverageReport{
		Files: []FileCoverage{
			{FilePath: "src/main.go", TotalLines: 100, CoveredLines: 80},
			{FilePath: "src/util.go", TotalLines: 50, CoveredLines: 30},
			{FilePath: "README.md", TotalLines: 10, CoveredLines: 5},
		},
	}

	dirs := r.GroupByDir()
	if len(dirs) != 2 {
		t.Errorf("expected 2 dirs, got %d", len(dirs))
	}

	for _, d := range dirs {
		switch d.DirPath {
		case "src":
			if d.FileCount != 2 {
				t.Errorf("expected 2 files in src, got %d", d.FileCount)
			}
			if d.TotalLines != 150 {
				t.Errorf("expected 150 lines in src, got %d", d.TotalLines)
			}
		case ".":
			if d.FileCount != 1 {
				t.Errorf("expected 1 file in ., got %d", d.FileCount)
			}
		default:
			t.Errorf("unexpected dir: %s", d.DirPath)
		}
	}
}

func TestCoverageBar(t *testing.T) {
	bar := CoverageBar(0.9, 10)
	if len(bar) == 0 {
		t.Error("expected non-empty bar")
	}
	// Should contain ANSI color codes
	if bar[0] != '\033' {
		t.Error("expected ANSI escape code")
	}
}

func TestFormatPercent(t *testing.T) {
	s := FormatPercent(0.856)
	if len(s) == 0 {
		t.Error("expected non-empty string")
	}
}

func TestDirFromPath(t *testing.T) {
	tests := []struct {
		path string
		dir  string
	}{
		{"src/main.go", "src"},
		{"a/b/c/d.go", "a/b/c"},
		{"root.go", "."},
	}
	for _, tc := range tests {
		got := dirFromPath(tc.path)
		if got != tc.dir {
			t.Errorf("dirFromPath(%q) = %q, want %q", tc.path, got, tc.dir)
		}
	}
}
