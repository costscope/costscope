package gcp_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gcpp "github.com/costscope/costscope/internal/core/focus/conversion/gcp"
	"github.com/costscope/costscope/internal/core/focus/types"
	"github.com/costscope/costscope/internal/testutil"
)

// Test_UnifiedMapper_Parity validates parity of selected fields between legacy and unified paths.
func Test_UnifiedMapper_Parity(t *testing.T) {
	// Minimal CSV row with key fields
	header := strings.Join([]string{
		"billing_account_id",
		"billing_account_name",
		"currency",
		"project.id",
		"project.name",
		"service.description",
		"sku.id",
		"usage.amount",
		"usage.unit",
		"usage_start_time",
		"usage_end_time",
		"cost",
	}, ",")
	row := strings.Join([]string{
		"BA-100",
		"Main",
		"USD",
		"proj-1",
		"My Project",
		"Compute Engine",
		"SKU-1",
		"5",
		"Hrs",
		"2024-01-01T00:00:00Z",
		"2024-01-01T01:00:00Z",
		"9.99",
	}, ",")
	csv := strings.Join([]string{header, row}, "\n")

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.csv")
	outFast := filepath.Join(tmp, "fast.ndjson")
	outUnified := filepath.Join(tmp, "unified.ndjson")
	if err := os.WriteFile(in, []byte(csv), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	conv := gcpp.NewGCPConverter()

	// Fast path
	cfgFast := &types.ConversionConfig{
		Provider:         "gcp",
		InputPath:        in,
		OutputPath:       outFast,
		Streaming:        true,
		ChunkSize:        1000,
		Workers:          1,
		ConversionId:     "parity-fast-gcp",
		UseUnifiedMapper: false,
	}
	if err := conv.ValidateInput(t.Context(), cfgFast); err != nil {
		t.Fatalf("validate (fast): %v", err)
	}
	if _, err := conv.ConvertStream(t.Context(), cfgFast, nil); err != nil {
		t.Fatalf("convert fast: %v", err)
	}

	// Unified path
	cfgUnified := &types.ConversionConfig{
		Provider:         "gcp",
		InputPath:        in,
		OutputPath:       outUnified,
		Streaming:        true,
		ChunkSize:        1000,
		Workers:          1,
		ConversionId:     "parity-unified-gcp",
		UseUnifiedMapper: true,
	}
	if err := conv.ValidateInput(t.Context(), cfgUnified); err != nil {
		t.Fatalf("validate (unified): %v", err)
	}
	if _, err := conv.ConvertStream(t.Context(), cfgUnified, nil); err != nil {
		t.Fatalf("convert unified: %v", err)
	}

	fastRec := readFirstJSONLineGCP(t, outFast)
	uniRec := readFirstJSONLineGCP(t, outUnified)

	// Compare canonical fields
	eqString(t, fastRec, uniRec, "billing_account_id")
	eqString(t, fastRec, uniRec, "billing_currency")
	eqFloat(t, fastRec, uniRec, "effective_cost")
	eqFloat(t, fastRec, uniRec, "usage_quantity")
	eqString(t, fastRec, uniRec, "provider_name")
	eqString(t, fastRec, uniRec, "publisher_name")

	// Classification parity (added for unified mapper classification coverage)
	eqString(t, fastRec, uniRec, "charge_category")
	eqString(t, fastRec, uniRec, "pricing_category")

	// Unit normalization parity (hrs/Hrs -> Hours etc.) Accept case-insensitive equality.
	if uFast, ok := fastRec["usage_unit"].(string); ok {
		if uUni, ok2 := uniRec["usage_unit"].(string); ok2 {
			fastNorm := canonicalUnit(uFast)
			uniNorm := canonicalUnit(uUni)
			if fastNorm != uniNorm {
				t.Fatalf("usage_unit normalization mismatch: fast=%q unified=%q", uFast, uUni)
			}
		}
	}
}

// Test_UnifiedMapper_ClassificationParity ensures charge/pricing/credit classifications & commitment fields
// match between legacy and unified paths across credit/spot/cud fixtures.
func Test_UnifiedMapper_ClassificationParity(t *testing.T) {
	repoRoot := testutil.FindRepoRoot(t)
	fixtures := []string{"credit_cud.csv", "credit_spot.csv", "credit_sustained_promo.csv", "usage_minimal.csv"}
	for _, f := range fixtures {
		in := filepath.Join(repoRoot, "tests", "fixtures", "gcp", f)
		tmp := t.TempDir()
		outFast := filepath.Join(tmp, f+".fast.ndjson")
		outUni := filepath.Join(tmp, f+".uni.ndjson")
		conv := gcpp.NewGCPConverter()

		fastCfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: outFast, Streaming: true, UseUnifiedMapper: false}
		if err := conv.ValidateInput(t.Context(), fastCfg); err != nil {
			t.Fatalf("validate fast %s: %v", f, err)
		}
		if _, err := conv.ConvertStream(t.Context(), fastCfg, nil); err != nil {
			t.Fatalf("convert fast %s: %v", f, err)
		}

		uniCfg := &types.ConversionConfig{Provider: "gcp", InputPath: in, OutputPath: outUni, Streaming: true, UseUnifiedMapper: true}
		if err := conv.ValidateInput(t.Context(), uniCfg); err != nil {
			t.Fatalf("validate unified %s: %v", f, err)
		}
		if _, err := conv.ConvertStream(t.Context(), uniCfg, nil); err != nil {
			t.Fatalf("convert unified %s: %v", f, err)
		}

		fastRec := readFirstJSONLineGCP(t, outFast)
		uniRec := readFirstJSONLineGCP(t, outUni)

		eqString(t, fastRec, uniRec, "charge_category")
		eqString(t, fastRec, uniRec, "pricing_category")
		// Commitment discount fields may not exist for all fixtures; assert parity when present in either.
		for _, k := range []string{"commitment_discount_id", "commitment_discount_type", "commitment_discount_name"} {
			_, fHas := fastRec[k]
			_, uHas := uniRec[k]
			if fHas || uHas {
				eqString(t, fastRec, uniRec, k)
			}
		}
	}
}

func readFirstJSONLineGCP(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	//nolint:gosec // test helper opens a controlled temp file
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

// credit classification and CommitmentDiscount* population in unified path
func Test_UnifiedMapper_CreditsClassification(t *testing.T) {
	tmp := t.TempDir()

	repoRoot := testutil.FindRepoRoot(t)
	must := func(rel string) string { return filepath.Join(repoRoot, rel) }

	// CUD credit fixture
	inCUD := must(filepath.Join("tests", "fixtures", "gcp", "credit_cud.csv"))
	outCUD := filepath.Join(tmp, "out_cud.ndjson")
	conv := gcpp.NewGCPConverter()
	cfg := &types.ConversionConfig{Provider: "gcp", InputPath: inCUD, OutputPath: outCUD, Streaming: true, UseUnifiedMapper: true}
	if err := conv.ValidateInput(t.Context(), cfg); err != nil {
		t.Fatalf("validate cud: %v", err)
	}
	if _, err := conv.ConvertStream(t.Context(), cfg, nil); err != nil {
		t.Fatalf("convert cud: %v", err)
	}
	rec := readFirstJSONLineGCP(t, outCUD)
	if rec["charge_category"] != types.ChargeCategories.Credit {
		t.Fatalf("expected Credit, got %v", rec["charge_category"])
	}
	// CommitmentDiscount* populated
	if _, ok := rec["commitment_discount_id"]; !ok {
		t.Fatalf("expected commitment_discount_id present")
	}
	if _, ok := rec["commitment_discount_type"]; !ok {
		t.Fatalf("expected commitment_discount_type present")
	}
	if _, ok := rec["commitment_discount_name"]; !ok {
		t.Fatalf("expected commitment_discount_name present")
	}

	// Spot credit fixture
	inSpot := must(filepath.Join("tests", "fixtures", "gcp", "credit_spot.csv"))
	outSpot := filepath.Join(tmp, "out_spot.ndjson")
	cfg2 := &types.ConversionConfig{Provider: "gcp", InputPath: inSpot, OutputPath: outSpot, Streaming: true, UseUnifiedMapper: true}
	if err := conv.ValidateInput(t.Context(), cfg2); err != nil {
		t.Fatalf("validate spot: %v", err)
	}
	if _, err := conv.ConvertStream(t.Context(), cfg2, nil); err != nil {
		t.Fatalf("convert spot: %v", err)
	}
	rec2 := readFirstJSONLineGCP(t, outSpot)
	if rec2["charge_category"] != types.ChargeCategories.Credit {
		t.Fatalf("expected Credit for spot, got %v", rec2["charge_category"])
	}
	if rec2["pricing_category"] != types.PricingCategories.Spot {
		t.Fatalf("expected pricing_category Spot, got %v", rec2["pricing_category"])
	}

	// Sustained + Promotional fixture
	inSP := must(filepath.Join("tests", "fixtures", "gcp", "credit_sustained_promo.csv"))
	outSP := filepath.Join(tmp, "out_sp.ndjson")
	cfg3 := &types.ConversionConfig{Provider: "gcp", InputPath: inSP, OutputPath: outSP, Streaming: true, UseUnifiedMapper: true}
	if err := conv.ValidateInput(t.Context(), cfg3); err != nil {
		t.Fatalf("validate sustained/promo: %v", err)
	}
	if _, err := conv.ConvertStream(t.Context(), cfg3, nil); err != nil {
		t.Fatalf("convert sustained/promo: %v", err)
	}
	rec3 := readFirstJSONLineGCP(t, outSP)
	if rec3["charge_category"] != types.ChargeCategories.Credit {
		t.Fatalf("expected Credit for sustained/promo, got %v", rec3["charge_category"])
	}
}

// repo root helper moved to internal/testutil

// helpers for equality assertions
func eqString(t *testing.T, a, b map[string]any, k string) {
	t.Helper()
	if a[k] != b[k] {
		t.Fatalf("%s mismatch: %v vs %v", k, a[k], b[k])
	}
}
func eqFloat(t *testing.T, a, b map[string]any, k string) {
	t.Helper()
	af, aok := a[k].(float64)
	bf, bok := b[k].(float64)
	if !aok || !bok || af != bf {
		t.Fatalf("%s mismatch: %v vs %v", k, a[k], b[k])
	}
}
func canonicalUnit(u string) string {
	switch strings.ToLower(u) {
	case "hrs", "hr", "hour", "hours":
		return "Hours"
	case "gib", "gibibyte", "gibibytes":
		return "GiB"
	default:
		return u
	}
}
