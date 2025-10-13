//go:build duckdb
// +build duckdb

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/focus/quality"
	pipeline "github.com/costscope/costscope/tests/e2e/pipeline"
)

type Config struct {
	Providers           []string `json:"providers"`
	Tolerance           float64  `json:"tolerance"`
	BaselineDir         string   `json:"baseline_dir"`
	SkipMissingFixtures bool     `json:"skip_missing_fixtures"`
	Deterministic       bool     `json:"deterministic"`
}

type ProviderResult struct {
	Provider      string                   `json:"provider"`
	Passed        bool                     `json:"passed"`
	Notes         []string                 `json:"notes,omitempty"`
	Invariants    quality.InvariantMetrics `json:"invariants"`
	Baseline      quality.InvariantMetrics `json:"baseline"`
	RelativeDrift map[string]float64       `json:"relative_drift"`
	Report        *pipeline.Report         `json:"report,omitempty"`
}

type Summary struct {
	StartedAt time.Time        `json:"started_at"`
	EndedAt   time.Time        `json:"ended_at"`
	Config    Config           `json:"config"`
	Results   []ProviderResult `json:"results"`
	Passed    bool             `json:"passed"`
}

func defaultConfig() Config {
	return Config{
		Providers:           []string{"aws", "azure", "gcp"},
		Tolerance:           0.01,
		BaselineDir:         filepath.Join("tests", "fixtures", "quality"),
		SkipMissingFixtures: true,
		Deterministic:       true,
	}
}

func inputsForProvider(provider string) ([]string, error) {
	fixturesDir := filepath.Join("tests", "fixtures", provider)
	switch provider {
	case "azure":
		return []string{
			filepath.Join(fixturesDir, "usage.csv"),
			filepath.Join(fixturesDir, "tax_refund.csv"),
			filepath.Join(fixturesDir, "reservation_credit.csv"),
		}, nil
	case "gcp":
		return []string{
			filepath.Join(fixturesDir, "usage_minimal.csv"),
			filepath.Join(fixturesDir, "credit_cud.csv"),
			filepath.Join(fixturesDir, "credit_spot.csv"),
			filepath.Join(fixturesDir, "credit_sustained_promo.csv"),
		}, nil
	case "aws":
		return []string{filepath.Join(fixturesDir, "cur_baseline_sample.csv")}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func anyMissing(paths []string) (bool, []string) {
	var missing []string
	for _, p := range paths {
		if !fileExists(p) {
			missing = append(missing, p)
		}
	}
	return len(missing) > 0, missing
}

func runOne(ctx context.Context, cfg Config, provider string) (ProviderResult, error) {
	pr := ProviderResult{Provider: provider, RelativeDrift: map[string]float64{}}
	inputs, err := inputsForProvider(provider)
	if err != nil {
		return pr, err
	}
	// Determine baseline path with provider-aware fallbacks (aws uses baseline_invariants.json)
	baselinePath := resolveBaselinePath(cfg.BaselineDir, provider)
	if miss, list := anyMissing(inputs); miss {
		msg := fmt.Sprintf("missing fixtures for %s: %s", provider, strings.Join(list, ", "))
		if cfg.SkipMissingFixtures {
			pr.Passed = true // soft-skip
			pr.Notes = append(pr.Notes, "skip_missing_fixtures")
			pr.Notes = append(pr.Notes, msg)
			// Load baseline if present just to include it
			if baselinePath != "" {
				_ = loadBaselineInto(&pr.Baseline, baselinePath)
			}
			return pr, nil
		}
		return pr, errors.New(msg)
	}

	rep, err := pipeline.Run(ctx, pipeline.RunConfig{
		Provider:           provider,
		InputFiles:         inputs,
		DriftTolerance:     0.001,
		ValidateOutput:     true,
		BaselinePath:       baselinePath,
		InvariantTolerance: cfg.Tolerance,
	})
	if err != nil {
		pr.Passed = false
		pr.Notes = append(pr.Notes, fmt.Sprintf("pipeline_error:%v", err))
		return pr, nil
	}
	pr.Invariants = rep.Invariants
	pr.Baseline = rep.BaselineInvariants
	pr.RelativeDrift = rep.RelativeDrift
	pr.Passed = rep.Passed
	pr.Report = rep
	return pr, nil
}

func loadBaselineInto(dst *quality.InvariantMetrics, path string) error {
	base, err := quality.LoadBaseline(path)
	if err != nil {
		return err
	}
	*dst = base
	return nil
}

// resolveBaselinePath returns the best-effort baseline file path for a provider.
// Preferred pattern is baseline_<provider>_invariants.json; for AWS also accept
// baseline_invariants.json (historical filename). Returns empty string if none exist.
func resolveBaselinePath(dir, provider string) string {
	// Preferred provider-specific file
	p1 := filepath.Join(dir, fmt.Sprintf("baseline_%s_invariants.json", provider))
	if fileExists(p1) {
		return p1
	}
	// AWS historical name
	if provider == "aws" {
		p2 := filepath.Join(dir, "baseline_invariants.json")
		if fileExists(p2) {
			return p2
		}
	}
	return ""
}

func main() {
	var (
		providersCSV string
		tolerance    float64
		baselineDir  string
		skipMissing  bool
		configPath   string
	)
	def := defaultConfig()
	flag.StringVar(&providersCSV, "providers", strings.Join(def.Providers, ","), "comma-separated providers: aws,azure,gcp")
	flag.Float64Var(&tolerance, "tolerance", def.Tolerance, "relative tolerance for invariants (e.g., 0.01 = 1%)")
	flag.StringVar(&baselineDir, "baseline-dir", def.BaselineDir, "directory containing baseline_*.json")
	flag.BoolVar(&skipMissing, "skip-missing-fixtures", def.SkipMissingFixtures, "skip providers with missing fixtures instead of failing")
	flag.StringVar(&configPath, "config", "", "optional JSON config file with {providers, tolerance, baseline_dir, skip_missing_fixtures}")
	flag.Parse()

	cfg := defaultConfig()
	// Load from config file if provided
	if configPath != "" {
		if b, err := os.ReadFile(configPath); err == nil {
			_ = json.Unmarshal(b, &cfg)
		}
	}
	// CLI flags override
	if providersCSV != "" {
		cfg.Providers = strings.Split(strings.TrimSpace(providersCSV), ",")
	}
	if tolerance > 0 {
		cfg.Tolerance = tolerance
	}
	if baselineDir != "" {
		cfg.BaselineDir = baselineDir
	}
	cfg.SkipMissingFixtures = skipMissing
	cfg.Deterministic = true

	sum := Summary{StartedAt: time.Now().UTC(), Config: cfg}
	ctx := context.Background()
	passedAll := true
	for _, p := range cfg.Providers {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		res, err := runOne(ctx, cfg, p)
		if err != nil {
			passedAll = false
			res.Passed = false
			res.Notes = append(res.Notes, err.Error())
		}
		if !res.Passed {
			passedAll = false
		}
		sum.Results = append(sum.Results, res)
	}
	sum.EndedAt = time.Now().UTC()
	sum.Passed = passedAll

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(sum)
	if !passedAll {
		os.Exit(1)
	}
}
