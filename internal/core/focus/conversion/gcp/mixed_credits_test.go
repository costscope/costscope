package gcp

import (
	"testing"

	"local/costscope/internal/core/focus/types"
)

// Verifies that in a mixed credits array, commitment fields are taken from the first
// credit entry that contains metadata, and pricing_category is set to Spot if any
// entry indicates spot/preemptible, regardless of position.
func TestMixedCredits_FirstCommitment_And_SpotPrecedence(t *testing.T) {
	t.Run("commitmentFieldsFromFirst", func(t *testing.T) {
		fr := &types.FocusRecord{}
		obj := map[string]interface{}{
			"credits": []interface{}{
				map[string]interface{}{"id": "CUD-123", "type": "CommittedUseDiscount", "name": "CUD VM"},
				map[string]interface{}{"type": "SustainedUseDiscount", "name": "Sustained use"},
			},
		}
		ApplyUnifiedPostMapGCP(fr, obj)

		if fr.ChargeCategory != types.ChargeCategories.Credit {
			t.Fatalf("expected charge_category Credit, got %v", fr.ChargeCategory)
		}
		if fr.CommitmentDiscountId == nil || *fr.CommitmentDiscountId != "CUD-123" {
			t.Fatalf("expected commitment_discount_id from first entry, got %v", fr.CommitmentDiscountId)
		}
		if fr.CommitmentDiscountType == nil || *fr.CommitmentDiscountType != "CommittedUseDiscount" {
			t.Fatalf("expected commitment_discount_type from first entry, got %v", fr.CommitmentDiscountType)
		}
		if fr.CommitmentDiscountName == nil || *fr.CommitmentDiscountName != "CUD VM" {
			t.Fatalf("expected commitment_discount_name from first entry, got %v", fr.CommitmentDiscountName)
		}
		if fr.PricingCategory == types.PricingCategories.Spot {
			t.Fatalf("did not expect pricing_category Spot when no spot credit present")
		}
	})

	t.Run("spotPrecedenceEvenIfNotFirst", func(t *testing.T) {
		fr := &types.FocusRecord{}
		obj := map[string]interface{}{
			"credits": []interface{}{
				map[string]interface{}{"id": "CUD-xyz", "type": "CommittedUseDiscount", "name": "CUD"},
				map[string]interface{}{"type": "SustainedUseDiscount", "name": "Sustained"},
				// Any entry with a type/name containing "spot" or "preemptible" should set Spot
				map[string]interface{}{"type": "SpotInstance", "name": "Spot VM discount"},
			},
		}
		ApplyUnifiedPostMapGCP(fr, obj)

		if fr.ChargeCategory != types.ChargeCategories.Credit {
			t.Fatalf("expected charge_category Credit, got %v", fr.ChargeCategory)
		}
		if fr.CommitmentDiscountId == nil || *fr.CommitmentDiscountId != "CUD-xyz" {
			t.Fatalf("expected commitment_discount_id from first entry, got %v", fr.CommitmentDiscountId)
		}
		if fr.PricingCategory != types.PricingCategories.Spot {
			t.Fatalf("expected pricing_category Spot due to later spot credit, got %v", fr.PricingCategory)
		}
	})
}
