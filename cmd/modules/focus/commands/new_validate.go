package commands

import (
	"encoding/json"
	"fmt"
	"local/costscope/internal/core/focus/quality"
	"local/costscope/internal/core/focus/validation"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// formatter string literals centralized to satisfy goconst
	formatterJSON               = "json"
	formatterHTML               = "html"
	formatterCSV                = "csv"
	validateSpec                string
	validateCompliance          string
	validateComplianceEnabled   bool
	validateFormat              string
	validateOutput              string
	validateQuiet               bool
	validateVerbose             bool
	validateSchema              bool
	validateQuality             bool
	validatePerformance         bool
	validateAnomalies           bool
	validateInputDir            string
	validateOutputDir           string
	validatePattern             string
	validateReportHTML          bool
	validateReportCSV           bool
	validateMinScore            float64
	validateFailFast            bool
	validateInvariants          bool
	validateInvariantsBaseline  string
	validateInvariantsReport    string
	validateInvariantsTolerance float64
)

func BuildValidateCommand() *cobra.Command {
	c := &cobra.Command{Use: "validate [file]", Short: "Validate FOCUS files", Args: cobra.MaximumNArgs(1), RunE: runValidate}
	c.Flags().StringVar(&validateSpec, "spec", "v1.2", "FOCUS spec version (v1.2, v1.1, v1.0)")
	c.Flags().StringVar(&validateCompliance, "compliance", "focus", "Compliance framework")
	c.Flags().BoolVar(&validateComplianceEnabled, "compliance-enable", true, "Enable compliance validation")
	c.Flags().StringVar(&validateFormat, "format", "auto", "File format hint")
	c.Flags().StringVarP(&validateOutput, "output", "o", "", "Output file (infer format by extension)")
	c.Flags().BoolVar(&validateQuiet, "quiet", false, "Quiet mode")
	c.Flags().BoolVar(&validateVerbose, "verbose", false, "Verbose output")
	c.Flags().BoolVar(&validateSchema, "schema", false, "Schema only")
	c.Flags().BoolVar(&validateQuality, "quality", false, "Quality only")
	c.Flags().BoolVar(&validatePerformance, "performance", false, "Performance only")
	c.Flags().BoolVar(&validateAnomalies, "anomalies", false, "Anomalies only")
	c.Flags().Bool("all", false, "Enable all domains")
	c.Flags().StringVar(&validateInputDir, "input-dir", "", "Input directory (batch)")
	c.Flags().StringVar(&validateOutputDir, "output-dir", "", "Output directory (batch)")
	c.Flags().StringVar(&validatePattern, "pattern", "*.parquet", "File pattern (batch)")
	c.Flags().BoolVar(&validateReportHTML, "report-html", false, "(deprecated) HTML output")
	c.Flags().BoolVar(&validateReportCSV, "report-csv", false, "(deprecated) CSV output")
	c.Flags().Float64Var(&validateMinScore, "min-score", 80.0, "Minimum score")
	c.Flags().BoolVar(&validateFailFast, "fail-fast", false, "Fail fast")
	c.Flags().BoolVar(&validateInvariants, "invariants", false, "Compute invariants")
	c.Flags().StringVar(&validateInvariantsBaseline, "invariants-baseline", "", "Baseline invariants JSON")
	c.Flags().StringVar(&validateInvariantsReport, "invariants-report", "", "Write invariants report JSON")
	c.Flags().Float64Var(&validateInvariantsTolerance, "invariants-tolerance", 0.01, "Relative/absolute drift tolerance")
	// Discovery helper: list available schemas and exit
	c.Flags().Bool("list-schemas", false, "List available FOCUS schemas and exit")
	c.Flags().Bool("list-schemas-json", false, "List available FOCUS schemas as JSON and exit (non-breaking experimental)")
	c.AddCommand(buildBatchValidateCommand())
	c.AddCommand(buildInvariantsDiffCommand())
	return c
}

func runValidate(cmd *cobra.Command, args []string) error {
	// If --list-schemas or --list-schemas-json is set, print schemas and exit (no input required)
	if list, _ := cmd.Flags().GetBool("list-schemas"); list {
		eng := validation.NewEngine()
		specs := eng.GetSupportedSpecs()
		if len(specs) == 0 {
			fmt.Println("No schemas available")
			return nil
		}
		fmt.Println("Available FOCUS schemas:")
		for _, s := range specs {
			fmt.Printf("- %s\n", s)
		}
		return nil
	}
	if listJSON, _ := cmd.Flags().GetBool("list-schemas-json"); listJSON {
		eng := validation.NewEngine()
		specs := eng.GetSupportedSpecs()
		out := struct {
			Schemas []validation.ValidationSpec `json:"schemas"`
			Latest  string                      `json:"latest"`
			Count   int                         `json:"count"`
		}{Schemas: specs, Latest: func() string {
			if len(specs) > 0 {
				return string(specs[0])
			}
			return ""
		}(), Count: len(specs)}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("input file required")
	}
	input := args[0]
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", input)
	}
	// Enable all domains when --all is passed
	if all, _ := cmd.Flags().GetBool("all"); all {
		validateSchema, validateQuality, validatePerformance, validateAnomalies = true, true, true, true
	}
	// Back-compat: if user didn’t set any domain flags, default to all domains
	if !validateSchema && !validateQuality && !validatePerformance && !validateAnomalies {
		validateSchema, validateQuality, validatePerformance, validateAnomalies = true, true, true, true
	}
	opts := validation.ValidationOpts{InputPath: input, Spec: validateSpec, ComplianceFramework: validateCompliance, EnableCompliance: validateComplianceEnabled, FormatHint: validateFormat, RunSchema: validateSchema, RunQuality: validateQuality, RunPerformance: validatePerformance, RunAnomalies: validateAnomalies, FailFast: validateFailFast, MinScore: validateMinScore, Quiet: validateQuiet, Verbose: validateVerbose, OutputPath: validateOutput, InvariantsEnabled: validateInvariants, InvariantsBaseline: validateInvariantsBaseline, InvariantsReportPath: validateInvariantsReport, InvariantsTolerance: validateInvariantsTolerance, E2EMode: os.Getenv("COSTSCOPE_E2E_MODE") == "1"}
	full, err := validation.RunValidation(opts)
	// Determine explicit formatter preference. Prefer explicit --format, then legacy report flags.
	explicitFmt := ""
	if strings.ToLower(validateFormat) == formatterJSON {
		explicitFmt = formatterJSON
	} else if validateReportHTML {
		explicitFmt = formatterHTML
	} else if validateReportCSV {
		explicitFmt = formatterCSV
	}
	// Emit deprecation warnings for legacy report flags (unless quiet).
	if !validateQuiet {
		if validateReportHTML {
			fmt.Fprintln(os.Stderr, "[deprecation] --report-html is deprecated; prefer --output <path>.html")
		}
		if validateReportCSV {
			fmt.Fprintln(os.Stderr, "[deprecation] --report-csv is deprecated; prefer --output <path>.csv")
		}
	}
	formatter := SelectFormatter(explicitFmt, validateOutput)
	out, ferr := formatter.Format(full)
	if ferr != nil {
		return ferr
	}
	if validateOutput != "" {
		if werr := os.WriteFile(validateOutput, out, 0600); werr != nil {
			return werr
		}
		if !validateQuiet {
			fmt.Printf(" Report saved to: %s (%s)\n", validateOutput, formatter.Name())
		}
	} else {
		fmt.Print(string(out))
	}
	if err != nil {
		return err
	}
	if !full.Core.IsValid && full.Core.OverallScore < validateMinScore {
		os.Exit(1)
	}
	return nil
}

func buildBatchValidateCommand() *cobra.Command {
	return &cobra.Command{Use: "batch [directory]", Short: "Batch validate multiple files", Args: cobra.ExactArgs(1), RunE: runBatchValidate}
}

func runBatchValidate(_ *cobra.Command, args []string) error {
	dir := args[0]
	pattern := filepath.Join(dir, validatePattern)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("find files: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files matched %s", pattern)
	}
	start := time.Now()
	engine := validation.NewEngine()
	cfg := validation.ValidationConfig{Level: validation.ValidationLevelStandard, Spec: validation.SpecFOCUS12, Format: validateFormat, EnableCompliance: true, EnableQuality: true, EnablePerformance: validatePerformance, EnableAnomalyDetection: validateAnomalies, Quiet: validateQuiet, Verbose: validateVerbose}
	switch validateSpec {
	case "v1.1":
		cfg.Spec = validation.SpecFOCUS11
	case "v1.0":
		cfg.Spec = validation.SpecFOCUS10
	}
	results, err := engine.ValidateBatch(dir, cfg)
	if err != nil {
		return err
	}
	if validateOutputDir != "" {
		_ = os.MkdirAll(validateOutputDir, 0750)
		ts := time.Now().Format("20060102_150405")
		path := filepath.Join(validateOutputDir, fmt.Sprintf("batch_validation_%s.json", ts))
		data, _ := json.MarshalIndent(results, "", "  ")
		_ = os.WriteFile(path, data, 0600)
		if !validateQuiet {
			fmt.Printf(" Batch report: %s\n", path)
		}
	}
	if !validateQuiet {
		fmt.Printf(" Batch validated %d files in %v\n", len(results), time.Since(start))
	}
	return nil
}

func buildInvariantsDiffCommand() *cobra.Command {
	var baseline, report, artifact string
	var tolerance float64
	cmd := &cobra.Command{Use: "invariants-diff <current>", Short: "Compare invariants against baseline", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		curPath := args[0]
		if baseline == "" {
			return fmt.Errorf("--baseline required")
		}
		cur, err := validation.ComputeInvariantsFromFile(curPath)
		if err != nil {
			return err
		}
		base, err := quality.LoadBaseline(baseline)
		if err != nil {
			return fmt.Errorf("load baseline: %w", err)
		}
		quality.CompareInvariants(&cur, base, tolerance)
		validation.ExportInvariantMetrics(cur, base)
		if report != "" {
			if err := quality.SaveReport(report, cur); err != nil {
				return err
			}
			if artifact != "" {
				_ = os.MkdirAll(filepath.Dir(artifact), 0750)
				_ = os.WriteFile(artifact, []byte(report), 0600)
			}
		}
		b, _ := json.MarshalIndent(cur, "", "  ")
		fmt.Println(string(b))
		if len(cur.Violations) > 0 {
			return fmt.Errorf("invariants violations: %v", cur.Violations)
		}
		return nil
	}}
	cmd.Flags().StringVar(&baseline, "baseline", "", "Baseline invariants JSON")
	cmd.Flags().StringVar(&report, "report", "", "Write invariants report JSON")
	cmd.Flags().StringVar(&artifact, "artifact-path", "", "(CI) write path of report file to artifact file")
	cmd.Flags().Float64Var(&tolerance, "tolerance", 0.01, "Relative/absolute drift tolerance")
	return cmd
}
