package conversion

import (
	"testing"

	c "local/costscope/internal/core/focus/conversion/common"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Test_EnumCacheHits_Increments verifies that applying unified normalization
// triggers enum cache hit metrics for currency/region/unit without renaming series.
// Note: Region and unit caches warm twice (raw -> canonical), so the actual cache
// hit occurs on the third normalization pass; currency typically hits on the second.
func Test_EnumCacheHits_Increments(t *testing.T) {
	// Ensure metrics are registered
	defer func() { _ = recover() }()
	telemetry.Register()

	fr := &types.FocusRecord{
		BillingCurrency: "usd",
		UsageUnit:       "hrs",
		PricingUnit:     "gb",
	}
	region := "US East (N. Virginia)"
	fr.Region = &region

	beforeCurrency := testutil.ToFloat64(telemetry.EnumCacheHits.WithLabelValues("currency", "any"))
	beforeUnitAny := testutil.ToFloat64(telemetry.EnumCacheHits.WithLabelValues("unit", "any"))
	beforeRegion := testutil.ToFloat64(telemetry.EnumCacheHits.WithLabelValues("region", "gcp"))

	// First pass warms caches for raw keys (no hit expected).
	c.ApplyUnifiedNormalization("gcp", fr)
	// Second pass warms canonical keys for region/unit; currency may already hit here.
	c.ApplyUnifiedNormalization("gcp", fr)
	// Third pass should hit caches for all (currency/region/unit) and increment counters.
	c.ApplyUnifiedNormalization("gcp", fr)

	afterCurrency := testutil.ToFloat64(telemetry.EnumCacheHits.WithLabelValues("currency", "any"))
	afterUnitAny := testutil.ToFloat64(telemetry.EnumCacheHits.WithLabelValues("unit", "any"))
	afterRegion := testutil.ToFloat64(telemetry.EnumCacheHits.WithLabelValues("region", "gcp"))

	if afterCurrency <= beforeCurrency {
		t.Fatalf("expected currency enum cache_hits to increase, before=%f after=%f", beforeCurrency, afterCurrency)
	}
	if afterUnitAny <= beforeUnitAny {
		t.Fatalf("expected unit enum cache_hits to increase, before=%f after=%f", beforeUnitAny, afterUnitAny)
	}
	if afterRegion <= beforeRegion {
		t.Fatalf("expected region enum cache_hits to increase, before=%f after=%f", beforeRegion, afterRegion)
	}
}
