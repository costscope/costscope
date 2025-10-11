package common

import (
	"testing"

	"local/costscope/internal/core/focus/types"
)

func TestAzureEnsureDiscount_Table(t *testing.T) {
	// Ensure env toggle is off
	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "")

	cases := []struct {
		name        string
		chargeType  string
		billingType string
		inCat       string
		wantCat     string
		changed     bool
	}{
		{"usage+chargeType token", "Promo-Discount", "", types.ChargeCategories.Usage, CategoryDiscount, true},
		{"usage+billingType token", "", "DISCOUNT credit", types.ChargeCategories.Usage, CategoryDiscount, true},
		{"already Discount variant", "", "", "usage-discount", CategoryDiscount, true},
		{"non-usage other category", "", "", types.ChargeCategories.Tax, types.ChargeCategories.Tax, false},
		{"no tokens stays usage", "", "", types.ChargeCategories.Usage, types.ChargeCategories.Usage, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fr := &types.FocusRecord{ChargeCategory: tc.inCat, SourceFileName: "azure_legacy.csv"}
			changed := AzureEnsureDiscount(tc.chargeType, tc.billingType, fr)
			if changed != tc.changed || fr.ChargeCategory != tc.wantCat {
				t.Fatalf("got changed=%v cat=%q; want changed=%v cat=%q", changed, fr.ChargeCategory, tc.changed, tc.wantCat)
			}
		})
	}

	// Env toggle disables normalization
	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "true")
	fr := &types.FocusRecord{ChargeCategory: types.ChargeCategories.Usage}
	changed := AzureEnsureDiscount("discount", "", fr)
	if changed || fr.ChargeCategory != types.ChargeCategories.Usage {
		t.Fatalf("env disabled: changed=%v cat=%q", changed, fr.ChargeCategory)
	}

	// Ensure non-nil safety
	if AzureEnsureDiscount("", "", nil) {
		t.Fatal("nil record should not change")
	}

	// Heuristic unified path label shouldn't panic
	t.Setenv("COSTSCOPE_DISABLE_AZURE_DISCOUNT_NORMALIZATION", "")
	fr = &types.FocusRecord{ChargeCategory: types.ChargeCategories.Usage, SourceFileName: "azure_unified.json"}
	_ = AzureEnsureDiscount("discount", "", fr)
}
