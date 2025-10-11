package azure_test

import (
	"context"
	"encoding/csv"
	azure "local/costscope/internal/core/focus/conversion/azure"
	"os"
	"path/filepath"
	"strings"
	"testing"

	types "local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// helper writes a minimal CSV with provided ChargeType and BillingType
func writeAzureCSV(t *testing.T, dir, chargeType, billingType, cost string) string {
	t.Helper()
	path := filepath.Join(dir, "in.csv")
	// #nosec G304 - path is under t.TempDir()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = f.Close() }()
	w := csv.NewWriter(f)
	headers := []string{"BillingAccountId", "CostInBillingCurrency", "Cost", "AmortizedCost", "Quantity", "UsageStart", "UsageEnd", "MeterCategory", "ServiceName", "MeterName", "ChargeType", "BillingType", "RetailPrice", "UnitOfMeasure", "SubscriptionId", "SubscriptionName"}
	_ = w.Write(headers)
	row := []string{"BA-1", cost, cost, cost, "1", "2025-01-01T00:00:00Z", "2025-01-01T01:00:00Z", "Compute", "VM", "Standard", chargeType, billingType, "0.5", "Hours", "sub-1", "SubOne"}
	_ = w.Write(row)
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	return path
}

func TestAzureDiscountClassificationVariants(t *testing.T) {
	conv := azure.NewAzureConverter()
	tmp := t.TempDir()
	cases := []struct{ ct, bt, cost, want string }{
		{"usage-discount", "reservation-discount", "1", azure.CategoryDiscount},
		{"discount", "", "1", azure.CategoryDiscount},
		{"UsAgE-DisCoUnT", "", "1", azure.CategoryDiscount},
		{"promo-discount", "", "1", azure.CategoryDiscount},
		{"", "reservation-discount", "1", azure.CategoryDiscount},
		{"promo-discount", "", "-0.5", azure.CategoryDiscount}, // normalization keeps Discount even if negative when discount token present
		{"usage", "", "1", types.ChargeCategories.Usage},
	}
	for _, tc := range cases {
		in := writeAzureCSV(t, tmp, tc.ct, tc.bt, tc.cost)
		out := filepath.Join(tmp, strings.ReplaceAll(tc.ct+tc.bt+tc.cost, "/", "_")+".ndjson")
		cfg := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: out, Streaming: true}
		if _, err := conv.ConvertStream(context.Background(), cfg, nil); err != nil {
			t.Fatalf("convert: %v", err)
		}
		recs := readAllFocusRecordsFromNDJSONLocal(t, out)
		if len(recs) != 1 {
			t.Fatalf("expected 1 record")
		}
		got := recs[0].ChargeCategory
		if got != tc.want {
			t.Fatalf("ct=%s bt=%s cost=%s got=%s want=%s", tc.ct, tc.bt, tc.cost, got, tc.want)
		}
	}
}

func TestAzureDiscountClassificationBypassFlag(t *testing.T) {
	conv := azure.NewAzureConverter()
	tmp := t.TempDir()
	// enable normalization (unset flag) then disable and compare
	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "")
	in1 := writeAzureCSV(t, tmp, "promo-discount", "", "1")
	out1 := filepath.Join(tmp, "out1.ndjson")
	if _, err := conv.ConvertStream(context.Background(), &types.ConversionConfig{Provider: "azure", InputPath: in1, OutputPath: out1, Streaming: true}, nil); err != nil {
		t.Fatal(err)
	}
	rec1 := readAllFocusRecordsFromNDJSONLocal(t, out1)[0]
	if rec1.ChargeCategory != azure.CategoryDiscount {
		t.Fatalf("want Discount with flag unset, got %s", rec1.ChargeCategory)
	}

	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "1")
	in2 := writeAzureCSV(t, tmp, "promo-discount", "", "1")
	out2 := filepath.Join(tmp, "out2.ndjson")
	if _, err := conv.ConvertStream(context.Background(), &types.ConversionConfig{Provider: "azure", InputPath: in2, OutputPath: out2, Streaming: true}, nil); err != nil {
		t.Fatal(err)
	}
	rec2 := readAllFocusRecordsFromNDJSONLocal(t, out2)[0]
	if rec2.ChargeCategory != types.ChargeCategories.Usage {
		t.Fatalf("want Usage with flag set, got %s", rec2.ChargeCategory)
	}
}

func TestAzureDiscountNormalizationSkipMetric(t *testing.T) {
	conv := azure.NewAzureConverter()
	tmp := t.TempDir()
	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "")
	before := testutil.ToFloat64(telemetry.AzureDiscountNormalizationSkips.WithLabelValues("azure"))
	// normalization active
	in1 := writeAzureCSV(t, tmp, "promo-discount", "", "1")
	if _, err := conv.ConvertStream(context.Background(), &types.ConversionConfig{Provider: "azure", InputPath: in1, OutputPath: filepath.Join(tmp, "o1.ndjson"), Streaming: true}, nil); err != nil {
		t.Fatal(err)
	}
	mid := testutil.ToFloat64(telemetry.AzureDiscountNormalizationSkips.WithLabelValues("azure"))
	if mid-before != 0 {
		t.Fatalf("expected no skip delta, got %f", mid-before)
	}

	// disable normalization
	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "1")
	in2 := writeAzureCSV(t, tmp, "promo-discount", "", "1")
	if _, err := conv.ConvertStream(context.Background(), &types.ConversionConfig{Provider: "azure", InputPath: in2, OutputPath: filepath.Join(tmp, "o2.ndjson"), Streaming: true}, nil); err != nil {
		t.Fatal(err)
	}
	after := testutil.ToFloat64(telemetry.AzureDiscountNormalizationSkips.WithLabelValues("azure"))
	if after-mid != 1 {
		t.Fatalf("expected skip delta 1, got %f", after-mid)
	}
}
