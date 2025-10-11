package azure_test

import (
	"bufio"
	"context"
	"encoding/csv"
	jsonpkg "encoding/json"
	"os"
	"path/filepath"
	"testing"

	azure "local/costscope/internal/core/focus/conversion/azure"
	"local/costscope/internal/core/focus/types"
)

// writeCSV writes a minimal Azure CSV with headers and rows and returns the file path.
func writeCSV(t *testing.T, dir string, headers []string, rows [][]string) string {
	t.Helper()
	p := filepath.Join(dir, "in.csv")
	f, err := os.Create(p) // #nosec G304 – temp test path
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = f.Close() }()
	w := csv.NewWriter(bufio.NewWriter(f))
	_ = w.Write(headers)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatalf("csv write: %v", err)
	}
	return p
}

// TestAzure_Parity_SpotAndCommitment covers Spot and commitment (Reservation/SavingsPlan) classification parity
// and ensures CommitmentDiscount* fields are populated consistently in unified path.
func TestAzure_Parity_SpotAndCommitment(t *testing.T) {
	tmp := t.TempDir()

	headers := []string{
		"BillingAccountId", "BillingAccountName", "BillingCurrency",
		"SubscriptionId", "SubscriptionName", "ServiceName", "ServiceFamily",
		"ResourceId", "ResourceName", "ResourceType", "ResourceLocation",
		"Quantity", "UnitOfMeasure", "AmortizedCost", "RetailPrice",
		"UsageStart", "UsageEnd",
		// classification-related
		"PricingModel", "PricingModelName", "BenefitId", "BenefitName", "ReservationId", "ReservationName", "SavingsPlanId", "SavingsPlanName",
		// category hints
		"ChargeType", "BillingType",
	}

	cases := []struct {
		name        string
		pricing     string
		chargeType  string
		benefitId   string
		benefitNm   string
		resId       string
		resNm       string
		spId        string
		spNm        string
		wantPricing string
		wantClass   string
		wantCType   string // commitment_discount_type expected when present
	}{
		{name: "spot-usage", pricing: "Spot", chargeType: "Usage", wantPricing: types.PricingCategories.Spot, wantClass: types.ChargeClasses.OnDemand},
		// Azure uses "Reservation" (not "ReservedInstance") as the commitment_discount_type token
		{name: "reservation-benefit", pricing: "Reservation", chargeType: "Usage", resId: "res-123", resNm: "RI-Name", wantPricing: types.PricingCategories.Reserved, wantClass: types.ChargeClasses.Commitment, wantCType: "Reservation"},
		{name: "savingsplan-benefit", pricing: "SavingsPlan", chargeType: "Usage", spId: "sp-abc", spNm: "SP-Name", wantPricing: types.PricingCategories.Reserved, wantClass: types.ChargeClasses.Commitment, wantCType: "SavingsPlan"},
	}

	// Common base row values
	base := []string{
		"BA-9", "Main", "USD",
		"sub-1", "Dev", "Virtual Machines", "Compute",
		"/subs/sub-1/rg/rg1/vm/vm1", "vm1", "Microsoft.Compute/virtualMachines", "eastus",
		"1", "Hours", "0.10", "0.10",
		"2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z",
	}

	// Build rows per case
	rows := make([][]string, 0, len(cases))
	for _, tc := range cases {
		r := append([]string{}, base...)
		// pricing model/name
		r = append(r, tc.pricing, tc.pricing)
		// benefits
		r = append(r, tc.benefitId, tc.benefitNm, tc.resId, tc.resNm, tc.spId, tc.spNm)
		// category hints
		r = append(r, tc.chargeType, "")
		rows = append(rows, r)
	}

	in := writeCSV(t, tmp, headers, rows)
	outFast := filepath.Join(tmp, "fast.ndjson")
	outUni := filepath.Join(tmp, "unified.ndjson")

	conv := azure.NewAzureConverter()

	// Legacy
	fastCfg := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outFast, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-spot-commit-fast"}
	if err := conv.ValidateInput(context.Background(), fastCfg); err != nil {
		t.Fatalf("validate fast: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), fastCfg, nil); err != nil {
		t.Fatalf("convert fast: %v", err)
	}

	// Unified
	uniCfg := &types.ConversionConfig{Provider: "azure", InputPath: in, OutputPath: outUni, Streaming: true, ChunkSize: 1000, Workers: 1, ConversionId: "azure-spot-commit-uni", UseUnifiedMapper: true}
	if err := conv.ValidateInput(context.Background(), uniCfg); err != nil {
		t.Fatalf("validate unified: %v", err)
	}
	if _, err := conv.ConvertStream(context.Background(), uniCfg, nil); err != nil {
		t.Fatalf("convert unified: %v", err)
	}

	// Read back first three records and compare parity for relevant fields
	fastF, err := os.Open(outFast) // #nosec G304 — test opens a temp file path constructed within test scope
	if err != nil {
		t.Fatalf("open fast: %v", err)
	}
	defer func() { _ = fastF.Close() }()
	uniF, err := os.Open(outUni) // #nosec G304 — test opens a temp file path constructed within test scope
	if err != nil {
		t.Fatalf("open unified: %v", err)
	}
	defer func() { _ = uniF.Close() }()
	fr := bufio.NewReader(fastF)
	ur := bufio.NewReader(uniF)

	for i, tc := range cases {
		fLine, _, err := fr.ReadLine()
		if err != nil {
			t.Fatalf("read fast line %d: %v", i, err)
		}
		uLine, _, err := ur.ReadLine()
		if err != nil {
			t.Fatalf("read uni line %d: %v", i, err)
		}
		var fRec map[string]any
		var uRec map[string]any
		if err := jsonpkg.Unmarshal(fLine, &fRec); err != nil {
			t.Fatalf("json fast %d: %v", i, err)
		}
		if err := jsonpkg.Unmarshal(uLine, &uRec); err != nil {
			t.Fatalf("json uni %d: %v", i, err)
		}

		// Parity on classification-related fields
		if fRec["charge_category"] != uRec["charge_category"] {
			t.Fatalf("row %d charge_category mismatch: fast=%v uni=%v", i, fRec["charge_category"], uRec["charge_category"])
		}
		if fRec["pricing_category"] != uRec["pricing_category"] {
			t.Fatalf("row %d pricing_category mismatch: fast=%v uni=%v", i, fRec["pricing_category"], uRec["pricing_category"])
		}
		if fRec["charge_class"] != uRec["charge_class"] {
			t.Fatalf("row %d charge_class mismatch: fast=%v uni=%v", i, fRec["charge_class"], uRec["charge_class"])
		}

		// Expected semantics checks
		if got := uRec["pricing_category"]; got != tc.wantPricing {
			t.Fatalf("row %d expected pricing_category %s, got %v", i, tc.wantPricing, got)
		}
		if got := uRec["charge_class"]; got != tc.wantClass {
			t.Fatalf("row %d expected charge_class %s, got %v", i, tc.wantClass, got)
		}

		// Commitment discount fields when applicable
		if tc.wantCType != "" {
			if uRec["commitment_discount_type"] != tc.wantCType {
				t.Fatalf("row %d expected commitment_discount_type %s, got %v", i, tc.wantCType, uRec["commitment_discount_type"])
			}
			if _, ok := uRec["commitment_discount_id"]; !ok {
				t.Fatalf("row %d expected commitment_discount_id present", i)
			}
			if _, ok := uRec["commitment_discount_name"]; !ok {
				t.Fatalf("row %d expected commitment_discount_name present", i)
			}
		}
	}
}
