package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/covscan/internal/comparer"
	"github.com/EdgarOrtegaRamirez/covscan/internal/parser"
	"github.com/EdgarOrtegaRamirez/covscan/internal/reporter"
)

var (
	showAllFiles bool
	jsonOutput   bool
	format       string
	threshold    float64
	quiet        bool
)

var rootCmd = &cobra.Command{
	Use:   "covscan",
	Short: "Multi-language code coverage report viewer & aggregator",
	Long: `CovScan parses, aggregates, views, and compares code coverage reports
from multiple formats (Go coverprofile, Cobertura XML, LCOV, Istanbul JSON).

Examples:
  covscan cover coverage.out
  covscan cover coverage.xml --json
  covscan cover lcov.info --threshold 80
  covscan compare old.out new.out`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return nil
	},
}

var coverCmd = &cobra.Command{
	Use:   "cover [files...]",
	Short: "Parse and display coverage reports",
	Long: `Parse one or more coverage report files and display coverage statistics.
Supports: Go coverprofile (.out, .cov), Cobertura XML (.xml),
LCOV tracefiles (.info, .lcov), Istanbul JSON (.json).`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		exitCode := 0
		for _, arg := range args {
			report, err := parser.ParseAuto(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", arg, err)
				exitCode = 1
				continue
			}

			// Determine format: --json overrides --format, default "text"
			useFormat := format
			if jsonOutput {
				useFormat = "json"
			}

			switch useFormat {
			case "json":
				r := &reporter.JSONReporter{}
				code := r.Report(report)
				if code != 0 {
					exitCode = code
				}
			case "markdown":
				r := reporter.NewMarkdownReporter()
				r.ShowAllFiles = showAllFiles
				r.Threshold = threshold
				code := r.Report(report)
				if code != 0 {
					exitCode = code
				}
			default:
				r := reporter.NewTextReporter()
				r.ShowAllFiles = showAllFiles
				r.Threshold = threshold
				code := r.Report(report)
				if code != 0 {
					exitCode = code
				}
			}
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary [files...]",
	Short: "Show a compact summary of coverage reports",
	Long:  `Display a one-line summary per coverage report file.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			report, err := parser.ParseAuto(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", arg, err)
				continue
			}
			fmt.Printf("%s: %.1f%% (%d/%d lines, %d files)\n",
				report.Name,
				report.LineCoverage()*100,
				report.TotalCoveredLines(),
				report.TotalLines(),
				len(report.Files))
		}
		return nil
	},
}

var compareCmd = &cobra.Command{
	Use:   "compare <old> <new>",
	Short: "Compare two coverage reports",
	Long:  `Parse and compare two coverage reports, showing differences per file.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		old, err := parser.ParseAuto(args[0])
		if err != nil {
			return fmt.Errorf("error parsing old report %s: %w", args[0], err)
		}

		newReport, err := parser.ParseAuto(args[1])
		if err != nil {
			return fmt.Errorf("error parsing new report %s: %w", args[1], err)
		}

		result := comparer.Compare(old, newReport)
		fmt.Print(result.Format())
		return nil
	},
}

func init() {
	coverCmd.Flags().BoolVarP(&showAllFiles, "all", "a", false, "Show per-file breakdown")
	coverCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON (deprecated: use --format json)")
	coverCmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, markdown")
	coverCmd.Flags().Float64VarP(&threshold, "threshold", "t", 0, "Minimum coverage threshold (exit 1 if below)")
	coverCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress output, only check threshold")

	rootCmd.AddCommand(coverCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(compareCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
