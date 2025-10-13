package azure_test

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	azure "github.com/costscope/costscope/internal/core/focus/conversion/azure"

	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/testutil"
)

// TestAzure_UnifiedMapper_Parity checks that core fields match between legacy and unified paths.
func TestAzure_UnifiedMapper_Parity(t *testing.T) {
	header := strings.Join([]string{
		"BillingAccountId",
		"BillingAccountName",
		"BillingCurrency",
		"SubscriptionId",
		"SubscriptionName",
		"ServiceName",
		"ServiceFamily",
		"ResourceId",
		"ResourceName",
		"ResourceType",
		"ResourceLocation",
		"Quantity",
		"UnitOfMeasure",
		"AmortizedCost",
		"RetailPrice",
		"UsageStart",
		"UsageEnd",
	}, ",")
	row := strings.Join([]string{
		"BA-1",
		"Main",
		"USD",
		"sub-123",
		"Dev",
		"Virtual Machines",
		"Compute",
		"/subs/sub-123/rg/rg1/vm/vm1",
		"vm1",
		"Microsoft.Compute/virtualMachines",
		"eastus",
		"10",
		"Hours",
		"12.34",
		"1.5",
		"2024-01-01T00:00:00Z",
		"2024-01-01T01:00:00Z",
	}, ",")
	csv := strings.Join([]string{header, row}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	outFast := filepath.Join(tmp, "fast.ndjson")
	outUnified := filepath.Join(tmp, "unified.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := azure.NewAzureConverter()

	// Fast path
	cfgFast := &types.ConversionConfig{
		Provider:         "azure",
		InputPath:        in,
		OutputPath:       outFast,
		Streaming:        true,
		ChunkSize:        1000,
		Workers:          1,
		ConversionId:     "parity-fast-azure",
		UseUnifiedMapper: false,
	}
	if err := conv.ValidateInput(context.Background(), cfgFast); err != nil {
		t.Fatalf("validate (fast): %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfgFast, nil); err != nil {
		t.Fatalf("convert fast: %v", err)
	}

	// Unified path
	cfgUnified := &types.ConversionConfig{
		Provider:         "azure",
		InputPath:        in,
		OutputPath:       outUnified,
		Streaming:        true,
		ChunkSize:        1000,
		Workers:          1,
		ConversionId:     "parity-unified-azure",
		UseUnifiedMapper: true,
	}
	if err := conv.ValidateInput(context.Background(), cfgUnified); err != nil {
		t.Fatalf("validate (unified): %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), cfgUnified, nil); err != nil {
		t.Fatalf("convert unified: %v", err)
	}

	fastRec := readFirstJSONLineAzure(t, outFast)
	uniRec := readFirstJSONLineAzure(t, outUnified)

	// Parity on key fields
	eqString(t, fastRec, uniRec, "billing_account_id")
	eqString(t, fastRec, uniRec, "billing_currency")
	eqFloat(t, fastRec, uniRec, "effective_cost")
	eqFloat(t, fastRec, uniRec, "usage_quantity")
	eqString(t, fastRec, uniRec, "provider_name")
	eqString(t, fastRec, uniRec, "publisher_name")
	// Classification parity
	eqString(t, fastRec, uniRec, "charge_category")
	eqString(t, fastRec, uniRec, "pricing_category")
}

// Covers benefits/credits/tax and normalization of Location/ResourceLocation and currency
func TestAzure_UnifiedMapper_Parity_Fixtures(t *testing.T) {
	tmp := t.TempDir()

	// Copy fixture helper: resolve from repo root using centralized helper
	copyFixture := func(name string) string {
		repoRoot := testutil.FindRepoRoot(t)
		src := filepath.Join(repoRoot, "tests", "fixtures", "azure", name)
		if _, err := os.Stat(src); err != nil {
			t.Fatalf("fixture not found: %s (%v)", src, err)
		}
		dst := filepath.Join(tmp, name)
		// #nosec G304 - reading project fixture
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read fixture %s: %v", src, err)
		}
		if err := os.WriteFile(dst, b, 0o600); err != nil {
			t.Fatalf("write temp fixture: %v", err)
		}
		return dst
	}

	cases := []string{"usage.csv", "reservation_credit.csv", "tax_refund.csv"}
	for _, file := range cases {
		in := copyFixture(file)
		outFast := filepath.Join(tmp, file+".fast.ndjson")
		outUnified := filepath.Join(tmp, file+".unified.ndjson")

		conv := azure.NewAzureConverter()

		cfgFast := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outFast, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "fast-" + file}
		if err := conv.ValidateInput(context.Background(), cfgFast); err != nil {
			t.Fatalf("validate fast: %v", err)
		}
		if _, err := conv.ConvertStream(context.Background(), cfgFast, nil); err != nil {
			t.Fatalf("convert fast: %v", err)
		}

		cfgUni := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outUnified, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "unified-" + file, UseUnifiedMapper: true}
		if err := conv.ValidateInput(context.Background(), cfgUni); err != nil {
			t.Fatalf("validate unified: %v", err)
		}
		if _, err := conv.ConvertStream(context.Background(), cfgUni, nil); err != nil {
			t.Fatalf("convert unified: %v", err)
		}

		fastRec := readFirstJSONLineAzure(t, outFast)
		uniRec := readFirstJSONLineAzure(t, outUnified)

		// Common parity fields
		eqString(t, fastRec, uniRec, "billing_account_id")
		eqString(t, fastRec, uniRec, "billing_currency")
		eqFloat(t, fastRec, uniRec, "effective_cost")
		eqFloat(t, fastRec, uniRec, "usage_quantity")
		eqString(t, fastRec, uniRec, "provider_name")
		eqString(t, fastRec, uniRec, "charge_category")
		eqString(t, fastRec, uniRec, "pricing_category")

		// Normalization checks
		if rFast, ok := fastRec["region"].(string); ok {
			if rUni, ok2 := uniRec["region"].(string); ok2 {
				lf := strings.ToLower(rFast)
				lfCompact := strings.ReplaceAll(strings.ReplaceAll(lf, " ", ""), "-", "")
				luCompact := strings.ReplaceAll(strings.ReplaceAll(rUni, " ", ""), "-", "")
				if lfCompact != luCompact {
					// Accept canonicalization of known azure patterns (east us -> eastus)
					t.Fatalf("region normalization mismatch: fast=%q unified=%q", rFast, rUni)
				}
			}
		}
		if c, ok := fastRec["billing_currency"].(string); ok {
			if strings.ToUpper(c) != c {
				t.Fatalf("fast path currency should be upper: %q", c)
			}
		}

		// Commitment discount fields parity when present in either output (legacy/unified)
		for _, k := range []string{"commitment_discount_type", "commitment_discount_id", "commitment_discount_name"} {
			_, fHas := fastRec[k]
			_, uHas := uniRec[k]
			if fHas || uHas {
				eqString(t, fastRec, uniRec, k)
			}
		}
	}
}

func readFirstJSONLineAzure(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	//nolint:gosec // test helper reads file from temp dir
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReader(f)
	line, _, err := r.ReadLine()
	if err != nil {
		t.Fatalf("read line: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(line, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

// Small helpers for map value equality in parity tests
func eqString(t *testing.T, a, b map[string]interface{}, key string) {
	t.Helper()
	sa, _ := a[key].(string)
	sb, _ := b[key].(string)
	if sa != sb {
		t.Fatalf("%s mismatch: %q vs %q", key, sa, sb)
	}
}

func eqFloat(t *testing.T, a, b map[string]interface{}, key string) {
	t.Helper()
	fa, _ := a[key].(float64)
	fb, _ := b[key].(float64)
	if (fa-fb) > 0.0000001 || (fb-fa) > 0.0000001 {
		t.Fatalf("%s mismatch: %v vs %v", key, fa, fb)
	}
}
