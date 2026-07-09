package comparer

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

func TestCompare(t *testing.T) {
	old := &model.CoverageReport{
		Name: "old.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 80},
			{FilePath: "util.go", TotalLines: 50, CoveredLines: 40},
		},
	}

	newR := &model.CoverageReport{
		Name: "new.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 85},
			{FilePath: "util.go", TotalLines: 50, CoveredLines: 45},
			{FilePath: "extra.go", TotalLines: 30, CoveredLines: 20},
		},
	}

	result := Compare(old, newR)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.FileChanges) != 3 {
		t.Errorf("expected 3 file changes, got %d", len(result.FileChanges))
	}

	// Check main.go improvement
	for _, fc := range result.FileChanges {
		if fc.FilePath == "main.go" {
			if fc.OldRate != 0.8 {
				t.Errorf("main.go old rate: expected 0.8, got %f", fc.OldRate)
			}
			if fc.NewRate != 0.85 {
				t.Errorf("main.go new rate: expected 0.85, got %f", fc.NewRate)
			}
		}
	}
}

func TestCompareEmpty(t *testing.T) {
	old := &model.CoverageReport{Name: "old"}
	newR := &model.CoverageReport{Name: "new"}

	result := Compare(old, newR)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.FileChanges) != 0 {
		t.Errorf("expected 0 changes, got %d", len(result.FileChanges))
	}
}

func TestCompareFormat(t *testing.T) {
	old := &model.CoverageReport{
		Name: "old.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 80},
		},
	}
	newR := &model.CoverageReport{
		Name: "new.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 90},
		},
	}

	result := Compare(old, newR)
	output := result.Format()

	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
	if len(output) < 20 {
		t.Error("expected substantial output")
	}
}

func TestCompareRegression(t *testing.T) {
	old := &model.CoverageReport{
		Name: "old.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 90},
		},
	}
	newR := &model.CoverageReport{
		Name: "new.out",
		Files: []model.FileCoverage{
			{FilePath: "main.go", TotalLines: 100, CoveredLines: 70},
		},
	}

	result := Compare(old, newR)

	for _, fc := range result.FileChanges {
		if fc.FilePath == "main.go" {
			if fc.NewRate >= fc.OldRate {
				t.Error("expected regression (new rate < old rate)")
			}
		}
	}
}