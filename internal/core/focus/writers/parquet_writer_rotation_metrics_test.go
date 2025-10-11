package writers

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"
)

// TestParquetWriter_RotationMetrics validates that ParquetRotationSize summary receives an observation
// per finalized rotated file.
func TestParquetWriter_RotationMetrics(t *testing.T) {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed)) //nolint:gosec
	tmp, err := os.MkdirTemp("", "rot-metrics-")
	if err != nil {
		t.Fatalf("[seed=%d] tmpdir: %v", seed, err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })

	base := filepath.Join(tmp, "focus.parquet")
	// Small threshold to force multiple rotations.
	rotateThreshold := int64(6 * 1024) // 6KB
	opts := &types.ParquetOptions{CompressionCodec: "snappy", RotateSizeBytes: rotateThreshold, RowGroupSizeBytes: 4 * 1024, PageSizeBytes: 4 * 1024, FilePrefix: "met"}
	ctx := WithParquetOptions(context.Background(), opts)
	wIface, _, err := NewWriter(ctx, base, "parquet", types.GetFocusV12Schema())
	if err != nil {
		t.Fatalf("[seed=%d] NewWriter: %v", seed, err)
	}
	w := wIface

	// Generate records (payload ~1KB) enough to trigger several rotations.
	payload := func() string {
		n := 900 + r.Intn(400)
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('a' + r.Intn(26))
		}
		return string(b)
	}
	total := 0
	for total < 500 { // ~500 rows
		recs := make([]types.FocusRecord, 0, 25)
		for i := 0; i < 25; i++ {
			now := time.UnixMilli(time.Now().UnixMilli()).UTC()
			recs = append(recs, types.FocusRecord{
				BillingAccountId: "A", BillingAccountName: "Acct", BillingCurrency: "USD",
				BillingPeriodStart: now, BillingPeriodEnd: now,
				ChargeCategory: types.ChargeCategories.Usage, ChargeClass: types.ChargeClasses.OnDemand,
				ChargeDescription: payload(), ChargeFrequency: "Hourly", ChargePeriodStart: now, ChargePeriodEnd: now,
				ChargeSubcategory: "General", EffectiveCost: 1.0, InvoiceIssuerName: "Issuer", ListCost: 1.0, ListUnitPrice: 1.0,
				PricingCategory: types.PricingCategories.Standard, PricingQuantity: 1.0, PricingUnit: "Hours",
				ProviderName: "test", PublisherName: "pub", ResourceId: "res-m-", ResourceName: "r", ResourceType: "rtype",
				ServiceCategory: "Compute", ServiceName: "svc", SkuId: "sku", SkuPriceId: "spid",
				SubAccountId: "sub", SubAccountName: "subname", UsageQuantity: 1, UsageUnit: "Hours",
				SourceProvider: "gen", SourceFileName: "gen.csv", ConversionTimestamp: now,
			})
		}
		if err := w.Write(context.Background(), recs); err != nil {
			t.Fatalf("[seed=%d] write: %v", seed, err)
		}
		total += len(recs)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("[seed=%d] close: %v", seed, err)
	}

	// Count rotated parquet files.
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("[seed=%d] readdir: %v", seed, err)
	}
	nameRe := regexp.MustCompile(`^met-\d{8}-\d{4}-\d{3}\.parquet$`)
	rotated := 0
	for _, e := range entries {
		if nameRe.MatchString(e.Name()) {
			rotated++
		}
	}
	if rotated < 2 {
		t.Fatalf("[seed=%d] expected >=2 rotated files, got %d", seed, rotated)
	}

	// Extract summary sample count.
	m := &dto.Metric{}
	if err := telemetry.ParquetRotationSize.Write(m); err != nil {
		t.Fatalf("[seed=%d] metric write: %v", seed, err)
	}
	sc := m.GetSummary().GetSampleCount()
	if sc < uint64(rotated) {
		t.Fatalf("[seed=%d] summary sample_count %d < rotated files %d", seed, sc, rotated)
	}

	// (Best-effort) ensure metric remains registered for other tests; if this registry type assertion fails we ignore.
	if reg, ok := prometheus.DefaultRegisterer.(*prometheus.Registry); ok {
		// Touch metric to avoid linter complaining about unused import; no-op logic.
		_ = reg
	}
}
