package quality

import (
	focustypes "local/costscope/internal/core/focus/types"
	"math"
	"strconv"
	"testing"
)

// floatEquals checks approximate equality with fixed absolute epsilon (1e-9).
func floatEquals(a, b float64) bool { // test helper
	return math.Abs(a-b) <= 1e-9
}

// TestInvariantsAggregatorParity ensures streaming aggregator matches batch ComputeInvariants output.
func TestInvariantsAggregatorParity(t *testing.T) {
	records := []focustypes.FocusRecord{
		{EffectiveCost: 10.0, ListCost: 12.0, UsageQuantity: 100, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r1", ServiceName: "EC2"},
		{EffectiveCost: 5.5, ListCost: 8.0, UsageQuantity: 50, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2", ServiceName: "S3"},
		{EffectiveCost: 3.2, ListCost: 4.0, UsageQuantity: -10, ChargeCategory: focustypes.ChargeCategories.Adjustment, PricingCategory: focustypes.PricingCategories.Reserved, ProviderName: focustypes.ProviderNames.GCP, ResourceId: "g1", ServiceName: "BigQuery"},
		{EffectiveCost: 7.0, ListCost: 7.0, UsageQuantity: 25, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Spot, ProviderName: focustypes.ProviderNames.Azure, ResourceId: "a1", ServiceName: "VM"},
		{EffectiveCost: 0.75, ListCost: 0.9, UsageQuantity: 5, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.Azure, ResourceId: "a2", ServiceName: "VM"},
		// Negative usage violation (Usage category)
		{EffectiveCost: 1.0, ListCost: 1.1, UsageQuantity: -2, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "bad-usage", ServiceName: "EC2"},
	}

	batch := ComputeInvariants(records)

	agg := NewInvariantsAggregator()
	for _, r := range records {
		agg.Add(r)
	}
	streaming := agg.Produce()

	if batch.RowCount != streaming.RowCount {
		t.Fatalf("row count mismatch: batch=%d streaming=%d", batch.RowCount, streaming.RowCount)
	}
	if !floatEquals(batch.SumEffectiveCost, streaming.SumEffectiveCost) {
		t.Fatalf("effective cost mismatch: batch=%f streaming=%f", batch.SumEffectiveCost, streaming.SumEffectiveCost)
	}
	if !floatEquals(batch.SumListCost, streaming.SumListCost) {
		t.Fatalf("list cost mismatch: batch=%f streaming=%f", batch.SumListCost, streaming.SumListCost)
	}
	if !floatEquals(batch.SumUsageQuantity, streaming.SumUsageQuantity) {
		t.Fatalf("usage qty mismatch: batch=%f streaming=%f", batch.SumUsageQuantity, streaming.SumUsageQuantity)
	}
	if batch.NegativeUsageAllowedCount != streaming.NegativeUsageAllowedCount {
		t.Fatalf("negative usage allowed mismatch: batch=%d streaming=%d", batch.NegativeUsageAllowedCount, streaming.NegativeUsageAllowedCount)
	}
	if batch.NegativeUsageViolationCount != streaming.NegativeUsageViolationCount {
		t.Fatalf("negative usage violation mismatch: batch=%d streaming=%d", batch.NegativeUsageViolationCount, streaming.NegativeUsageViolationCount)
	}

	// Compare distributions (percentage values) with small epsilon (1e-9 adequate)
	compareDist := func(name string, a, b map[string]float64) {
		if len(a) != len(b) {
			t.Fatalf("%s distribution size mismatch: %d vs %d", name, len(a), len(b))
		}
		for k, va := range a {
			vb, ok := b[k]
			if !ok {
				t.Fatalf("%s distribution missing key %s in streaming", name, k)
			}
			if !floatEquals(va, vb) {
				t.Fatalf("%s distribution mismatch for %s: batch=%f streaming=%f", name, k, va, vb)
			}
		}
	}
	compareDist("charge_category", batch.ChargeCategoryDistribution, streaming.ChargeCategoryDistribution)
	compareDist("pricing_category", batch.PricingCategoryDistribution, streaming.PricingCategoryDistribution)
	compareDist("provider", batch.ProviderDistribution, streaming.ProviderDistribution)
}

// TestInvariantsAggregatorEmpty ensures empty aggregator returns zeroed metrics and no panic.
func TestInvariantsAggregatorEmpty(t *testing.T) {
	agg := NewInvariantsAggregator()
	m := agg.Produce()
	if m.RowCount != 0 || m.SumEffectiveCost != 0 || m.SumListCost != 0 || m.SumUsageQuantity != 0 {
		t.Fatalf("expected zero metrics, got %+v", m)
	}
	if len(m.ChargeCategoryDistribution) != 0 || len(m.PricingCategoryDistribution) != 0 || len(m.ProviderDistribution) != 0 {
		t.Fatalf("expected empty distributions, got %+v", m)
	}
}

// TestInvariantsAggregatorDistinctMetadata validates distinct resource/service counts encoded in metadata.
func TestInvariantsAggregatorDistinctMetadata(t *testing.T) {
	records := []focustypes.FocusRecord{
		{EffectiveCost: 1, ListCost: 1, UsageQuantity: 1, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r1", ServiceName: "EC2"},
		{EffectiveCost: 2, ListCost: 2, UsageQuantity: 2, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2", ServiceName: "S3"},
		{EffectiveCost: 3, ListCost: 3, UsageQuantity: 3, ChargeCategory: focustypes.ChargeCategories.Usage, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r2", ServiceName: "S3"},   // duplicate resource/service
		{EffectiveCost: 4, ListCost: 4, UsageQuantity: 4, ChargeCategory: focustypes.ChargeCategories.Credit, PricingCategory: focustypes.PricingCategories.Standard, ProviderName: focustypes.ProviderNames.AWS, ResourceId: "r3", ServiceName: "EC2"}, // service duplicate, new resource
	}
	agg := NewInvariantsAggregator()
	for _, r := range records {
		agg.Add(r)
	}
	m := agg.Produce()
	if m.Metadata == nil {
		t.Fatalf("metadata missing")
	}
	drs, ok := m.Metadata["distinct_resources"]
	if !ok {
		t.Fatalf("distinct_resources key missing")
	}
	dss, ok := m.Metadata["distinct_services"]
	if !ok {
		t.Fatalf("distinct_services key missing")
	}
	// Expect distinct resources: r1,r2,r3 => 3; services: EC2,S3 => 2
	rCount, _ := strconv.Atoi(drs)
	sCount, _ := strconv.Atoi(dss)
	if rCount != 3 || sCount != 2 {
		t.Fatalf("expected distinct counts (3 resources,2 services) got (%d,%d)", rCount, sCount)
	}
}
