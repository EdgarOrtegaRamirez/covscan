package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EdgarOrtegaRamirez/covscan/internal/model"
)

// coberturaXml is the root XML element for Cobertura format.
type coberturaXml struct {
	XMLName  xml.Name           `xml:"coverage"`
	Packages []coberturaPackage `xml:"packages>package"`
}

type coberturaPackage struct {
	Name    string           `xml:"name,attr"`
	Classes []coberturaClass `xml:"classes>class"`
}

type coberturaClass struct {
	Name     string          `xml:"name,attr"`
	Filename string          `xml:"filename,attr"`
	Lines    []coberturaLine `xml:"lines>line"`
}

type coberturaLine struct {
	Number int    `xml:"number,attr"`
	Hits   int    `xml:"hits,attr"`
	Branch string `xml:"branch,attr"`
}

// ParseCobertura parses a Python coverage.py XML report (Cobertura format).
func ParseCobertura(path string) (*model.CoverageReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	var c coberturaXml
	if err := xml.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("cannot parse XML in %s: %w", path, err)
	}

	report := &model.CoverageReport{
		Name:      filepath.Base(path),
		SourceDir: filepath.Dir(path),
	}

	for _, pkg := range c.Packages {
		for _, cls := range pkg.Classes {
			filename := cls.Filename
			if filename == "" {
				filename = cls.Name
			}

			fc := model.FileCoverage{
				FilePath: filename,
			}

			for _, line := range cls.Lines {
				fc.TotalLines++
				if line.Hits > 0 {
					fc.CoveredLines++
				}
				if line.Branch == "true" {
					fc.TotalBranches++
					if line.Hits > 0 {
						fc.CoveredBranches++
					}
				}
			}

			if fc.TotalLines > 0 {
				if fc.TotalBranches > 0 {
					report.HasBranch = true
				}
				report.Files = append(report.Files, fc)
			}
		}
	}

	return report, nil
}
