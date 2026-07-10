package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		content  string
		ext      string
		expected string
	}{
		{"mode: set\n", ".out", "go-cover"},
		{"<coverage line-rate=\"0.5\">", ".xml", "cobertura"},
		{"SF:/path/to/file.go\n", ".info", "lcov"},
		{"SF:/path/to/file.go\n", ".lcov", "lcov"},
		{`{"file.go":{"s":{}}}`, ".json", "istanbul"},
	}

	for _, tc := range tests {
		dir := t.TempDir()
		path := filepath.Join(dir, "coverage"+tc.ext)
		if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
			t.Fatal(err)
		}

		got := DetectFormat(path)
		if got != tc.expected {
			t.Errorf("DetectFormat(%s) = %q, want %q", tc.ext, got, tc.expected)
		}
	}
}

func TestParseGoCover(t *testing.T) {
	content := `mode: set
github.com/example/pkg/main.go:10.2,20.14 5 1
github.com/example/pkg/main.go:21.2,30.14 3 0
github.com/example/pkg/util.go:5.2,15.14 8 3
`

	dir := t.TempDir()
	path := filepath.Join(dir, "cover.out")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ParseGoCover(path)
	if err != nil {
		t.Fatalf("ParseGoCover failed: %v", err)
	}

	if report.Name != "cover.out" {
		t.Errorf("expected name 'cover.out', got %q", report.Name)
	}

	if len(report.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(report.Files))
	}

	// main.go: 11 lines (10-20) covered, 10 lines (21-30) uncovered
	// util.go: 11 lines (5-15) covered
	for _, f := range report.Files {
		switch f.FilePath {
		case "github.com/example/pkg/main.go":
			if f.TotalLines != 21 {
				t.Errorf("main.go: expected 21 total lines, got %d", f.TotalLines)
			}
			if f.CoveredLines != 11 {
				t.Errorf("main.go: expected 11 covered lines, got %d", f.CoveredLines)
			}
		case "github.com/example/pkg/util.go":
			if f.TotalLines != 11 {
				t.Errorf("util.go: expected 11 total lines, got %d", f.TotalLines)
			}
			if f.CoveredLines != 11 {
				t.Errorf("util.go: expected 11 covered lines, got %d", f.CoveredLines)
			}
		default:
			t.Errorf("unexpected file: %s", f.FilePath)
		}
	}
}

func TestParseCobertura(t *testing.T) {
	content := `<?xml version="1.0" ?>
<coverage line-rate="0.5" branch-rate="0.0">
  <packages>
    <package name="pkg" line-rate="0.5">
      <classes>
        <class name="main" filename="src/main.py">
          <lines>
            <line number="1" hits="1" branch="false"/>
            <line number="2" hits="1" branch="false"/>
            <line number="3" hits="0" branch="false"/>
            <line number="4" hits="0" branch="false"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.xml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ParseCobertura(path)
	if err != nil {
		t.Fatalf("ParseCobertura failed: %v", err)
	}

	if len(report.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(report.Files))
	}

	f := report.Files[0]
	if f.FilePath != "src/main.py" {
		t.Errorf("expected src/main.py, got %q", f.FilePath)
	}
	if f.TotalLines != 4 {
		t.Errorf("expected 4 lines, got %d", f.TotalLines)
	}
	if f.CoveredLines != 2 {
		t.Errorf("expected 2 covered lines, got %d", f.CoveredLines)
	}
}

func TestParseLCOV(t *testing.T) {
	content := `SF:src/main.go
DA:1,5
DA:2,3
DA:3,0
DA:4,7
DA:5,0
BRDA:1,0,1,2
BRDA:1,1,1,-
end_of_record
`

	dir := t.TempDir()
	path := filepath.Join(dir, "trace.info")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ParseLCOV(path)
	if err != nil {
		t.Fatalf("ParseLCOV failed: %v", err)
	}

	if len(report.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(report.Files))
	}

	f := report.Files[0]
	if f.TotalLines != 5 {
		t.Errorf("expected 5 lines, got %d", f.TotalLines)
	}
	if f.CoveredLines != 3 {
		t.Errorf("expected 3 covered lines, got %d", f.CoveredLines)
	}
	if f.TotalBranches != 2 {
		t.Errorf("expected 2 branches, got %d", f.TotalBranches)
	}
	if f.CoveredBranches != 1 {
		t.Errorf("expected 1 covered branch, got %d", f.CoveredBranches)
	}
}

func TestParseIstanbul(t *testing.T) {
	content := `{
  "src/main.js": {
    "path": "src/main.js",
    "statementMap": {
      "1": {"start": {"line": 1, "column": 0}, "end": {"line": 1, "column": 10}},
      "2": {"start": {"line": 2, "column": 0}, "end": {"line": 2, "column": 10}},
      "3": {"start": {"line": 3, "column": 0}, "end": {"line": 3, "column": 10}},
      "4": {"start": {"line": 5, "column": 0}, "end": {"line": 5, "column": 10}}
    },
    "s": {"1": 1, "2": 5, "3": 0, "4": 3},
    "branchMap": {},
    "b": {}
  }
}`

	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ParseIstanbul(path)
	if err != nil {
		t.Fatalf("ParseIstanbul failed: %v", err)
	}

	if len(report.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(report.Files))
	}

	f := report.Files[0]
	if f.FilePath != "src/main.js" {
		t.Errorf("expected src/main.js, got %q", f.FilePath)
	}
	if f.TotalLines != 4 {
		t.Errorf("expected 4 unique lines, got %d", f.TotalLines)
	}
	if f.CoveredLines != 3 {
		t.Errorf("expected 3 covered lines, got %d", f.CoveredLines)
	}
}

func TestParseAuto(t *testing.T) {
	// Test Go cover
	dir := t.TempDir()
	goPath := filepath.Join(dir, "cover.out")
	os.WriteFile(goPath, []byte("mode: set\npkg/main.go:1.2,5.14 3 1\n"), 0644)

	report, err := ParseAuto(goPath)
	if err != nil {
		t.Fatalf("ParseAuto(go) failed: %v", err)
	}
	if len(report.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(report.Files))
	}

	// Test unknown format
	unknownPath := filepath.Join(dir, "unknown.txt")
	os.WriteFile(unknownPath, []byte("some random data\n"), 0644)

	_, err = ParseAuto(unknownPath)
	if err == nil {
		t.Error("expected error for unknown format")
	}
}
