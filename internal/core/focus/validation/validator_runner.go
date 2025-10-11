package validation

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"local/costscope/internal/core/focus/quality"

	"github.com/prometheus/client_golang/prometheus"
)

// ValidationOpts holds execution parameters for the validation runner.
type ValidationOpts struct {
	InputPath            string
	Spec                 string
	ComplianceFramework  string
	EnableCompliance     bool
	FormatHint           string
	RunSchema            bool
	RunQuality           bool
	RunPerformance       bool
	RunAnomalies         bool
	FailFast             bool
	MinScore             float64
	Quiet                bool
	Verbose              bool
	OutputPath           string
	InvariantsEnabled    bool
	InvariantsBaseline   string
	InvariantsReportPath string
	InvariantsTolerance  float64
	E2EMode              bool
}

// ValidationFullResult aggregates core result + invariants + timing.
type ValidationFullResult struct {
	Core               *ValidationResult         `json:"core"`
	Invariants         *quality.InvariantMetrics `json:"invariants,omitempty"`
	Duration           time.Duration             `json:"duration"`
	InvariantsBaseline string                    `json:"invariants_baseline,omitempty"`
}

var (
	invariantsDriftGauge    = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: "costscope", Subsystem: "invariants", Name: "drift_ratio", Help: "Relative drift ratios for aggregate invariants (sum_effective_cost, sum_list_cost, row_count)"}, []string{"metric"})
	distributionDriftGauge  = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: "costscope", Subsystem: "invariants", Name: "distribution_drift_pp", Help: "Absolute percentage point drift for distributions (charge_category, pricing_category, provider)"}, []string{"distribution", "key"})
	runnerMetricsRegistered bool
)

func ensureInvariantMetricsRegistered() {
	if runnerMetricsRegistered {
		return
	}
	prometheus.MustRegister(invariantsDriftGauge, distributionDriftGauge)
	runnerMetricsRegistered = true
}

// RunValidation performs core validation plus optional invariants.
func RunValidation(opts ValidationOpts) (*ValidationFullResult, error) {
	if opts.InputPath == "" {
		return nil, errors.New("input path required")
	}
	if _, err := os.Stat(opts.InputPath); err != nil {
		return nil, fmt.Errorf("stat input: %w", err)
	}
	if !opts.RunSchema && !opts.RunQuality && !opts.RunPerformance && !opts.RunAnomalies {
		opts.RunSchema, opts.RunQuality, opts.RunPerformance, opts.RunAnomalies = true, true, true, true
	}
	cfg := ValidationConfig{Level: ValidationLevelStandard, Spec: SpecFOCUS12, Format: opts.FormatHint, EnableCompliance: opts.EnableCompliance, EnableQuality: opts.RunQuality, EnablePerformance: opts.RunPerformance, EnableAnomalyDetection: opts.RunAnomalies, OutputPath: opts.OutputPath, Quiet: opts.Quiet, Verbose: opts.Verbose}
	switch strings.ToLower(opts.Spec) {
	case "v1.1", "focus-1.1":
		cfg.Spec = SpecFOCUS11
	case "v1.0", "focus-1.0":
		cfg.Spec = SpecFOCUS10
	}
	if opts.E2EMode {
		cfg.EnableAnomalyDetection, cfg.EnableCompliance, cfg.EnablePerformance = false, false, false
	}
	eng := NewEngine()
	start := time.Now()
	core, err := eng.Validate(opts.InputPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	full := &ValidationFullResult{Core: core, Duration: time.Since(start)}
	if opts.FailFast && !core.IsValid {
		return full, fmt.Errorf("validation failed (score %.1f)", core.OverallScore)
	}
	if opts.InvariantsEnabled {
		ensureInvariantMetricsRegistered()
		inv, ierr := ComputeInvariantsFromFile(opts.InputPath)
		if ierr != nil {
			return full, fmt.Errorf("invariants: %w", ierr)
		}
		if opts.InvariantsBaseline != "" {
			base, berr := quality.LoadBaseline(opts.InvariantsBaseline)
			if berr != nil {
				return full, fmt.Errorf("baseline: %w", berr)
			}
			quality.CompareInvariants(&inv, base, opts.InvariantsTolerance)
			ExportInvariantMetrics(inv, base)
			full.InvariantsBaseline = opts.InvariantsBaseline
			if len(inv.Violations) > 0 {
				full.Invariants = &inv
				if opts.InvariantsReportPath != "" {
					_ = quality.SaveReport(opts.InvariantsReportPath, inv)
				}
				return full, fmt.Errorf("invariants violations (%d)", len(inv.Violations))
			}
		}
		if opts.InvariantsReportPath != "" {
			_ = quality.SaveReport(opts.InvariantsReportPath, inv)
		}
		full.Invariants = &inv
	}
	return full, nil
}

// ExportInvariantMetrics publishes invariants drift metrics.
func ExportInvariantMetrics(cur quality.InvariantMetrics, base quality.InvariantMetrics) {
	ensureInvariantMetricsRegistered()
	if base.RowCount > 0 {
		invariantsDriftGauge.WithLabelValues("row_count").Set(relDiffFloat(float64(cur.RowCount), float64(base.RowCount)))
	}
	if base.SumEffectiveCost != 0 {
		invariantsDriftGauge.WithLabelValues("sum_effective_cost").Set(relDiffFloat(cur.SumEffectiveCost, base.SumEffectiveCost))
	}
	if base.SumListCost != 0 {
		invariantsDriftGauge.WithLabelValues("sum_list_cost").Set(relDiffFloat(cur.SumListCost, base.SumListCost))
	}
	push := func(n string, cm, bm map[string]float64) {
		keys := map[string]struct{}{}
		for k := range cm {
			keys[k] = struct{}{}
		}
		for k := range bm {
			keys[k] = struct{}{}
		}
		for k := range keys {
			distributionDriftGauge.WithLabelValues(n, k).Set(mathAbs(cm[k] - bm[k]))
		}
	}
	push("charge_category", cur.ChargeCategoryDistribution, base.ChargeCategoryDistribution)
	push("pricing_category", cur.PricingCategoryDistribution, base.PricingCategoryDistribution)
	push("provider", cur.ProviderDistribution, base.ProviderDistribution)
}

func relDiffFloat(a, b float64) float64 {
	if b == 0 {
		if a == 0 {
			return 0
		}
		if a < 0 {
			return -1
		}
		return 1
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	return d / mathAbs(b)
}
func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
