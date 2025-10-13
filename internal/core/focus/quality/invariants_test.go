package quality

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	focustypes "github.com/costscope/costscope/internal/core/focus/types"
)

// helper to load baseline
func loadBaselineT(t *testing.T, name string) InvariantMetrics {
	t.Helper()
	p := filepath.Join("../../../../tests/fixtures/quality", name)
	b, err := os.ReadFile(p) //nolint:gosec // test fixture path controlled
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	var m InvariantMetrics
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal baseline: %v", err)
	}
	return m
}

// TestInvariantsGolden ensures aggregates & distributions stay within ±1% drift and usage_quantity rule holds.
func TestInvariantsGolden(t *testing.T) {
	baseline := loadBaselineT(t, "baseline_invariants.json")
	// Construct synthetic current records matching baseline
	records := []focustypes.FocusRecord{
		{EffectiveCost: 10, ListCost: 10, UsageQuantity: 100, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r1"},
		{EffectiveCost: 15, ListCost: 15, UsageQuantity: 200, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2"},
		{EffectiveCost: 5, ListCost: 5, UsageQuantity: 0, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r3"},
	}
	cur := ComputeInvariants(records)
	CompareInvariants(&cur, baseline, 0.01) // 1% tolerance

	// Save JSON report (artifact for CI)
	reportPath := filepath.Join(os.TempDir(), "invariants_report.json")
	if err := SaveReport(reportPath, cur); err != nil {
		t.Fatalf("save report: %v", err)
	}

	if len(cur.Violations) > 0 {
		t.Fatalf("invariant violations: %v", cur.Violations)
	}
}

// TestAzureInvariantsGolden validates Azure baseline invariants using synthetic records.
func TestAzureInvariantsGolden(t *testing.T) {
	baseline := loadBaselineT(t, "baseline_azure_invariants.json")
	// 3 records to match distribution: Credit, Tax, Usage (each ~33.33%)
	// Pricing: 2 Standard, 1 Reserved (66.66 / 33.33)
	records := []focustypes.FocusRecord{
		{EffectiveCost: 3.00, ListCost: 8.00, UsageQuantity: 5, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.Azure, ResourceId: "az-r1"},
		{EffectiveCost: -0.50, ListCost: 5.60, UsageQuantity: 3, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.Azure, ResourceId: "az-r2"},
		{EffectiveCost: 5.54, ListCost: 9.00, UsageQuantity: 8, ChargeCategory: "Tax", PricingCategory: focustypes.PricingCategories.Reserved, ProviderName: focustypes.ProviderNames.Azure, ResourceId: "az-r3"},
	}
	cur := ComputeInvariants(records)
	CompareInvariants(&cur, baseline, 0.01)
	if len(cur.Violations) > 0 {
		t.Fatalf("azure invariant violations: %v", cur.Violations)
	}
}

// TestGCPInvariantsGolden validates GCP baseline invariants using synthetic records.
func TestGCPInvariantsGolden(t *testing.T) {
	baseline := loadBaselineT(t, "baseline_gcp_invariants.json")
	// 4 records: 3 Credit (75%), 1 Usage (25%)
	// Pricing: 3 Standard (75%), 1 Spot (25%)
	records := []focustypes.FocusRecord{
		{EffectiveCost: -2.00, ListCost: -2.00, UsageQuantity: 30, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.GCP, ResourceId: "gcp-r1"},
		{EffectiveCost: -1.50, ListCost: -1.50, UsageQuantity: 25, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.GCP, ResourceId: "gcp-r2"},
		{EffectiveCost: -1.50, ListCost: -1.50, UsageQuantity: 35, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.GCP, ResourceId: "gcp-r3"},
		{EffectiveCost: 0.24, ListCost: 0.24, UsageQuantity: 25, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Spot, ProviderName: focustypes.ProviderNames.GCP, ResourceId: "gcp-r4"},
	}
	cur := ComputeInvariants(records)
	CompareInvariants(&cur, baseline, 0.01)
	if len(cur.Violations) > 0 {
		t.Fatalf("gcp invariant violations: %v", cur.Violations)
	}
}

// TestDriftDetection ensures that a drift in distribution or aggregates produces violations.
func TestDriftDetection(t *testing.T) {
	baseline := loadBaselineT(t, "baseline_invariants.json")
	// Introduce drift: change one record to Credit (was all Usage) and alter costs >1%.
	records := []focustypes.FocusRecord{
		{EffectiveCost: 10, ListCost: 10, UsageQuantity: 100, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r1"},
		{EffectiveCost: 14, ListCost: 14, UsageQuantity: 200, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2"},
		{EffectiveCost: 5, ListCost: 5, UsageQuantity: 0, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r3"},
	}
	cur := ComputeInvariants(records)
	CompareInvariants(&cur, baseline, 0.01)
	if len(cur.Violations) == 0 {
		t.Fatalf("expected drift violations but none found")
	}
}

// TestBaselineTamperTriggersViolation simulates a modified baseline file (tampered) and ensures
// that the previously matching synthetic records now produce drift violations (>1%).
func TestBaselineTamperTriggersViolation(t *testing.T) {
	// Load original baseline used by golden test
	original := loadBaselineT(t, "baseline_invariants.json")
	// Create tampered baseline copy in temp dir
	tampered := original
	tampered.SumEffectiveCost = original.SumEffectiveCost * 1.10 // +10% drift
	tampered.RowCount = original.RowCount + 1                    // add drift in count
	// Write tampered file
	tmpPath := filepath.Join(t.TempDir(), "tampered_baseline.json")
	b, _ := json.Marshal(tampered)
	if err := os.WriteFile(tmpPath, b, 0o600); err != nil {
		t.Fatalf("write tampered baseline: %v", err)
	}
	// Reload via loader to simulate normal path
	loaded, err := LoadBaseline(tmpPath)
	if err != nil {
		t.Fatalf("load tampered baseline: %v", err)
	}
	// Use the same synthetic records from golden test (redeclared here for clarity)
	records := []focustypes.FocusRecord{
		{EffectiveCost: 10, ListCost: 10, UsageQuantity: 100, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r1"},
		{EffectiveCost: 15, ListCost: 15, UsageQuantity: 200, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2"},
		{EffectiveCost: 5, ListCost: 5, UsageQuantity: 0, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r3"},
	}
	cur := ComputeInvariants(records)
	CompareInvariants(&cur, loaded, 0.01)
	if len(cur.Violations) == 0 {
		t.Fatalf("expected violations against tampered baseline but found none")
	}
}

// TestNegativeUsageViolation ensures negative usage outside allowed categories fails.
func TestNegativeUsageViolation(t *testing.T) {
	baseline := loadBaselineT(t, "baseline_invariants.json")
	// Copy baseline like records but introduce invalid negative usage in Usage category
	records := []focustypes.FocusRecord{
		{EffectiveCost: 10, ListCost: 11, UsageQuantity: -5, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "bad"},
		{EffectiveCost: 15, ListCost: 16, UsageQuantity: 200, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2"},
		{EffectiveCost: 5, ListCost: 6, UsageQuantity: 105, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r3"},
	}
	cur := ComputeInvariants(records)
	CompareInvariants(&cur, baseline, 0.01)
	if cur.NegativeUsageViolationCount == 0 {
		t.Fatalf("expected negative usage violation to be detected")
	}
	found := false
	for _, v := range cur.Violations {
		if v == "negative_usage_violations=1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected violation token not found: %v", cur.Violations)
	}
}
