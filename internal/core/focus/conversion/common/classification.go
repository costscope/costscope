package common

import (
	"strings"

	"github.com/costscope/costscope/internal/core/focus/types"
)

// De-duplicated string constants
const (
	providerAWS   = "aws"
	providerAzure = "azure"
	providerGCP   = "gcp"
	spotStr       = "spot"
)

// Extended keyword constants for Azure commitment heuristics
const (
	kwReservation  = "reservation"
	kwReserved     = "reserved"
	kwSavingsPlan1 = "savingsplan"
	kwSavingsPlan2 = "savings plan"
	kwCommitment   = "commitment"
	kwAmortized    = "amortized"
)

// Local copies of commitment types (string equality across packages is preserved)
const (
	ctypeSavingsPlan      = "SavingsPlan"
	ctypeReservedInstance = "ReservedInstance"
)

// ApplyUnifiedClassification mutates a FocusRecord with any provider-specific
// adjustments not already applied in the unified path.
func ApplyUnifiedClassification(provider string, raw map[string]string, fr *types.FocusRecord) {
	switch strings.ToLower(provider) {
	case providerAWS:
		ClassifyAWSShared(raw, fr)
	case providerAzure:
		ClassifyAzureShared(raw, fr)
	case providerGCP:
		ClassifyGCPShared(raw, fr)
	}
}

func ClassifyAWSShared(raw map[string]string, fr *types.FocusRecord) {
	lineItemType := strings.TrimSpace(raw["lineItem/LineItemType"])
	usageType := strings.ToLower(raw["lineItem/UsageType"])
	operation := strings.ToLower(raw["lineItem/Operation"])
	spId := strings.TrimSpace(raw["savingsPlan/SavingsPlanId"])
	spArn := strings.TrimSpace(raw["savingsPlan/SavingsPlanArn"])
	riArn := strings.TrimSpace(raw["reservation/ReservationARN"])
	riSub := strings.TrimSpace(raw["reservation/SubscriptionId"])
	if fr.CommitmentDiscountType == nil && (spId != "" || spArn != "" || riArn != "" || riSub != "") {
		switch lineItemType {
		case "SavingsPlanCoveredUsage", "SavingsPlanRecurringFee", "SavingsPlanNegation":
			ctype := ctypeSavingsPlan
			fr.CommitmentDiscountType = &ctype
			if id := firstNonEmpty(spId, spArn); id != "" {
				fr.CommitmentDiscountId = &id
				fr.CommitmentDiscountName = &id
			}
		case "DiscountedUsage", "RIFee":
			ctype := ctypeReservedInstance
			fr.CommitmentDiscountType = &ctype
			if id := firstNonEmpty(riSub, riArn); id != "" {
				fr.CommitmentDiscountId = &id
			}
			if nm := firstNonEmpty(riArn, riSub); nm != "" {
				fr.CommitmentDiscountName = &nm
			}
		}
	}
	if fr.PricingCategory == "" && (strings.Contains(usageType, spotStr) || strings.Contains(operation, spotStr)) {
		fr.PricingCategory = types.PricingCategories.Spot
	}
}

func ClassifyAzureShared(raw map[string]string, fr *types.FocusRecord) {
	if fr.PricingCategory == "" || fr.PricingCategory == types.PricingCategories.Standard {
		pm := strings.ToLower(strings.TrimSpace(firstNonEmpty(raw["PricingModel"], raw["PricingModelName"], raw["Term"])))
		switch pm {
		case spotStr:
			fr.PricingCategory = types.PricingCategories.Spot
			if fr.ChargeClass == "" {
				fr.ChargeClass = types.ChargeClasses.OnDemand
			}
		case kwReservation, kwReserved, kwSavingsPlan1, kwSavingsPlan2, kwCommitment, kwAmortized:
			fr.PricingCategory = types.PricingCategories.Reserved
			if fr.ChargeClass == "" || fr.ChargeClass == types.ChargeClasses.OnDemand {
				fr.ChargeClass = types.ChargeClasses.Commitment
			}
		}
	}
	if fr.CommitmentDiscountType == nil {
		if b := strings.TrimSpace(firstNonEmpty(raw["BenefitId"], raw["ReservationId"], raw["SavingsPlanId"])); b != "" {
			btLower := strings.ToLower(firstNonEmpty(raw["BenefitType"], raw["Term"], raw["PricingModel"]))
			var ctype string
			switch {
			case strings.Contains(btLower, "saving"):
				ctype = ctypeSavingsPlan
			case strings.Contains(btLower, "reserv"):
				ctype = ctypeReservedInstance
			default:
				ctype = kwCommitment
			}
			fr.CommitmentDiscountType = &ctype
			fr.CommitmentDiscountId = &b
			if name := firstNonEmpty(raw["BenefitName"], raw["ReservationName"], raw["SavingsPlanName"], b); name != "" {
				fr.CommitmentDiscountName = &name
			}
		}
	}
}

func ClassifyGCPShared(raw map[string]string, fr *types.FocusRecord) {
	if fr.CommitmentDiscountType == nil {
		if cType := strings.TrimSpace(raw["credits.type"]); cType != "" {
			fr.CommitmentDiscountType = &cType
		}
		if cId := strings.TrimSpace(raw["credits.id"]); cId != "" {
			fr.CommitmentDiscountId = &cId
		}
		if cName := strings.TrimSpace(raw["credits.name"]); cName != "" {
			fr.CommitmentDiscountName = &cName
		}
	}
	if fr.PricingCategory == "" || fr.PricingCategory == types.PricingCategories.Standard {
		lc := strings.ToLower(firstNonEmpty(raw["credits.type"], raw["credits.name"]))
		if strings.Contains(lc, spotStr) || strings.Contains(lc, "preempt") {
			fr.PricingCategory = types.PricingCategories.Spot
		}
	}
}

// firstNonEmpty is deprecated; use FirstNonEmpty in this package.
func firstNonEmpty(ss ...string) string { return FirstNonEmpty(ss...) }
