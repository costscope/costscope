package quality

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	focustypes "github.com/costscope/costscope/internal/core/focus/types"
)

// InvariantMetrics captures aggregate invariants for a FOCUS dataset.
type InvariantMetrics struct {
	RowCount                    int                `json:"row_count"`
	SumEffectiveCost            float64            `json:"sum_effective_cost"`
	SumListCost                 float64            `json:"sum_list_cost"`
	SumUsageQuantity            float64            `json:"sum_usage_quantity"`
	ChargeCategoryDistribution  map[string]float64 `json:"charge_category_distribution"`
	PricingCategoryDistribution map[string]float64 `json:"pricing_category_distribution"`
	ProviderDistribution        map[string]float64 `json:"provider_distribution"`
	NegativeUsageAllowedCount   int                `json:"negative_usage_allowed_count"`
	NegativeUsageViolationCount int                `json:"negative_usage_violation_count"`
	GeneratedAt                 time.Time          `json:"generated_at"`
	Violations                  []string           `json:"violations"`
	Metadata                    map[string]string  `json:"metadata,omitempty"`
}

// ComputeInvariants computes invariants from a slice of FocusRecord.
func ComputeInvariants(records []focustypes.FocusRecord) InvariantMetrics {
	m := InvariantMetrics{
		ChargeCategoryDistribution:  make(map[string]float64),
		PricingCategoryDistribution: make(map[string]float64),
		ProviderDistribution:        make(map[string]float64),
		GeneratedAt:                 time.Now().UTC(),
	}
	if len(records) == 0 {
		m.Violations = append(m.Violations, "no_records")
		return m
	}
	for _, r := range records {
		m.RowCount++
		m.SumEffectiveCost += r.EffectiveCost
		m.SumListCost += r.ListCost
		m.SumUsageQuantity += r.UsageQuantity
		m.ChargeCategoryDistribution[r.ChargeCategory]++
		m.PricingCategoryDistribution[r.PricingCategory]++
		m.ProviderDistribution[r.ProviderName]++
		if r.UsageQuantity < 0 {
			if r.ChargeCategory == focustypes.ChargeCategories.Credit || r.ChargeCategory == focustypes.ChargeCategories.Adjustment {
				m.NegativeUsageAllowedCount++
			} else {
				m.NegativeUsageViolationCount++
				m.Violations = append(m.Violations, fmt.Sprintf("negative_usage_quantity:%s", r.ResourceId))
			}
		}
	}
	// Convert counts to percentages
	toPct := func(dist map[string]float64) {
		for k, v := range dist {
			dist[k] = (v / float64(m.RowCount)) * 100.0
		}
	}
	toPct(m.ChargeCategoryDistribution)
	toPct(m.PricingCategoryDistribution)
	toPct(m.ProviderDistribution)
	return m
}

// CompareInvariants compares current metrics with baseline applying a relative tolerance (e.g. 0.01 for ±1%).
// Violations appended to current.Violations.
func CompareInvariants(current *InvariantMetrics, baseline InvariantMetrics, tolerance float64) {
	// Allow explicit placeholder baselines to bypass enforcement without failing the run.
	if baseline.Metadata != nil {
		if marker, ok := baseline.Metadata["baseline_placeholder"]; ok && marker == "true" {
			return
		}
	}
	if baseline.RowCount == 0 {
		current.Violations = append(current.Violations, "baseline_rowcount_zero")
		return
	}
	relDiff := func(a, b float64) float64 {
		if b == 0 { // avoid div by zero; treat as absolute diff
			return math.Abs(a - b)
		}
		return math.Abs(a-b) / math.Abs(b)
	}
	// Aggregates
	if relDiff(current.SumEffectiveCost, baseline.SumEffectiveCost) > tolerance {
		current.Violations = append(current.Violations, fmt.Sprintf("sum_effective_cost_drift:cur=%.6f base=%.6f", current.SumEffectiveCost, baseline.SumEffectiveCost))
	}
	if relDiff(current.SumListCost, baseline.SumListCost) > tolerance {
		current.Violations = append(current.Violations, fmt.Sprintf("sum_list_cost_drift:cur=%.6f base=%.6f", current.SumListCost, baseline.SumListCost))
	}
	if relDiff(float64(current.RowCount), float64(baseline.RowCount)) > tolerance {
		current.Violations = append(current.Violations, fmt.Sprintf("row_count_drift:cur=%d base=%d", current.RowCount, baseline.RowCount))
	}
	// Distribution comparisons by key set union
	compareDist := func(name string, cur, base map[string]float64) {
		keys := map[string]struct{}{}
		for k := range cur {
			keys[k] = struct{}{}
		}
		for k := range base {
			keys[k] = struct{}{}
		}
		sorted := make([]string, 0, len(keys))
		for k := range keys {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)
		for _, k := range sorted {
			cVal := cur[k]
			bVal := base[k]
			// For ±1% absolute drift we check absolute difference with tolerance*100 (if tolerance small)
			if tolerance <= 0.05 { // treat tolerance as relative (0.01) but percentages are 0-100 scale
				if math.Abs(cVal-bVal) > tolerance*100.0 {
					current.Violations = append(current.Violations, fmt.Sprintf("%s_dist_drift:%s:cur=%.4f base=%.4f", name, k, cVal, bVal))
				}
			} else { // fallback to relative on percentages
				if bVal == 0 {
					if cVal > tolerance*100.0 {
						current.Violations = append(current.Violations, fmt.Sprintf("%s_dist_new:%s:cur=%.4f base=%.4f", name, k, cVal, bVal))
					}
				} else if relDiff(cVal, bVal) > tolerance {
					current.Violations = append(current.Violations, fmt.Sprintf("%s_dist_rel_drift:%s:cur=%.4f base=%.4f", name, k, cVal, bVal))
				}
			}
		}
	}
	compareDist("charge_category", current.ChargeCategoryDistribution, baseline.ChargeCategoryDistribution)
	compareDist("pricing_category", current.PricingCategoryDistribution, baseline.PricingCategoryDistribution)
	compareDist("provider", current.ProviderDistribution, baseline.ProviderDistribution)
	if current.NegativeUsageViolationCount > 0 {
		current.Violations = append(current.Violations, fmt.Sprintf("negative_usage_violations=%d", current.NegativeUsageViolationCount))
	}
}

// LoadBaseline loads baseline metrics from a JSON file.
func LoadBaseline(path string) (InvariantMetrics, error) { //nolint:gosec // G304: path provided by trusted CLI/ API caller; validated existence & schema
	var m InvariantMetrics
	// Security: baseline path is passed from CLI / config, documented as local file.
	// Not fetched over network; used only for JSON unmarshal & validation.
	//nolint:gosec // G304: intentional file read from user-specified baseline
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	if err := ValidateBaseline(&m); err != nil { // schema / logical validation
		return m, fmt.Errorf("baseline schema invalid: %w", err)
	}
	return m, nil
}

// ValidateBaseline performs lightweight schema & logical checks on a baseline.
// It ensures required numeric fields are non-negative, distributions (if present) sum to ~100% (±0.5),
// and violation counters are consistent. Placeholders (metadata.baseline_placeholder=true) skip checks.
func ValidateBaseline(m *InvariantMetrics) error {
	if m.Metadata != nil {
		if ph, ok := m.Metadata["baseline_placeholder"]; ok && ph == "true" {
			return nil
		}
	}
	if m.RowCount < 0 {
		return errors.New("row_count negative")
	}
	if m.NegativeUsageViolationCount < 0 || m.NegativeUsageAllowedCount < 0 {
		return errors.New("negative usage counters invalid")
	}
	// helper to check sum ~100
	checkDist := func(name string, dist map[string]float64) error {
		if dist == nil {
			return fmt.Errorf("%s distribution missing", name)
		}
		var sum float64
		for _, v := range dist {
			sum += v
		}
		if sum == 0 { // allow empty distributions for edge baselines
			return nil
		}
		if math.Abs(sum-100.0) > 0.5 { // allow small rounding drift
			return fmt.Errorf("%s distribution sum %.4f not ~100", name, sum)
		}
		return nil
	}
	if err := checkDist("charge_category", m.ChargeCategoryDistribution); err != nil {
		return err
	}
	if err := checkDist("pricing_category", m.PricingCategoryDistribution); err != nil {
		return err
	}
	if err := checkDist("provider", m.ProviderDistribution); err != nil {
		return err
	}
	return nil
}

// SaveReport writes current metrics to JSON file (creating parent dirs).
func SaveReport(path string, metrics InvariantMetrics) error {
	if path == "" {
		return errors.New("empty path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil { // restrict dir permissions
		return err
	}
	b, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
