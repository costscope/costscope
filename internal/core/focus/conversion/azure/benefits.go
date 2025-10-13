package azure

import (
	"strings"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// local constants to avoid repeating string literals
const (
	bTypeReservation = "Reservation"
	bTypeSavingsPlan = "SavingsPlan"
)

// ApplyBenefitsRow inspects benefit-related columns and populates CommitmentDiscount*.
func ApplyBenefitsRow(idx *HeaderIndex, r []string, fr *types.FocusRecord) {
	bt := strings.ToLower(strings.TrimSpace(Get(r, idx.BenefitType)))
	var bType, bId, bName string
	if bt != "" {
		if strings.Contains(bt, "reservation") {
			bType = bTypeReservation
		} else if strings.Contains(bt, "saving") || strings.Contains(bt, "savings") || strings.Contains(bt, "savingsplan") {
			bType = bTypeSavingsPlan
		}
		bId = FirstNonEmpty(r, idx.BenefitId, idx.ReservationId, idx.SavingsPlanId)
		bName = FirstNonEmpty(r, idx.BenefitName, idx.ReservationName, idx.SavingsPlanName)
	} else {
		if s := FirstNonEmpty(r, idx.ReservationId, idx.ReservationName); s != "" {
			bType = bTypeReservation
			bId = Get(r, idx.ReservationId)
			bName = Get(r, idx.ReservationName)
		} else if s := FirstNonEmpty(r, idx.SavingsPlanId, idx.SavingsPlanName); s != "" {
			bType = bTypeSavingsPlan
			bId = Get(r, idx.SavingsPlanId)
			bName = Get(r, idx.SavingsPlanName)
		}
	}
	if bType != "" {
		fr.CommitmentDiscountType = stringPtr(bType)
		if strings.TrimSpace(bId) != "" {
			fr.CommitmentDiscountId = stringPtr(bId)
		}
		if strings.TrimSpace(bName) != "" {
			fr.CommitmentDiscountName = stringPtr(bName)
		}
		if fr.PricingCategory == types.PricingCategories.Standard {
			fr.PricingCategory = types.PricingCategories.Reserved
		}
		if fr.ChargeClass == types.ChargeClasses.OnDemand {
			fr.ChargeClass = types.ChargeClasses.Commitment
		}
	}
}

// ApplyBenefitsMap performs the same as ApplyBenefitsRow but for a map record.
func ApplyBenefitsMap(rec map[string]string, fr *types.FocusRecord) {
	bt := strings.ToLower(strings.TrimSpace(FirstNonEmptyField(rec, "BenefitType")))
	var bType, bId, bName string
	if bt != "" {
		if strings.Contains(bt, "reservation") {
			bType = bTypeReservation
		} else if strings.Contains(bt, "saving") || strings.Contains(bt, "savings") || strings.Contains(bt, "savingsplan") {
			bType = bTypeSavingsPlan
		}
		bId = FirstNonEmptyField(rec, "BenefitId", "ReservationId", "SavingsPlanId")
		bName = FirstNonEmptyField(rec, "BenefitName", "ReservationName", "SavingsPlanName")
	} else {
		if s := FirstNonEmptyField(rec, "ReservationId", "ReservationName"); s != "" {
			bType = bTypeReservation
			bId = FirstNonEmptyField(rec, "ReservationId")
			bName = FirstNonEmptyField(rec, "ReservationName")
		} else if s := FirstNonEmptyField(rec, "SavingsPlanId", "SavingsPlanName"); s != "" {
			bType = bTypeSavingsPlan
			bId = FirstNonEmptyField(rec, "SavingsPlanId")
			bName = FirstNonEmptyField(rec, "SavingsPlanName")
		}
	}
	if bType != "" {
		fr.CommitmentDiscountType = stringPtr(bType)
		if strings.TrimSpace(bId) != "" {
			fr.CommitmentDiscountId = stringPtr(bId)
		}
		if strings.TrimSpace(bName) != "" {
			fr.CommitmentDiscountName = stringPtr(bName)
		}
		if fr.PricingCategory == types.PricingCategories.Standard {
			fr.PricingCategory = types.PricingCategories.Reserved
		}
		if fr.ChargeClass == types.ChargeClasses.OnDemand {
			fr.ChargeClass = types.ChargeClasses.Commitment
		}
	}
}

// local helper consistent with root package utility
// stringPtr is defined in azure/row_mapper.go; reuse that implementation.
