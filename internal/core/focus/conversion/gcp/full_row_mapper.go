package gcp

import (
	"strconv"
	"strings"
	"time"

	c "github.com/costscope/costscope/internal/core/focus/conversion/common"
	"github.com/costscope/costscope/internal/core/focus/types"
)

// FullRowMapper maps a CSV row to a FocusRecord using header name lookup.
// It performs pure field extraction/enrichment and credit parsing but leaves
// final charge classification to the caller to avoid import cycles.
type FullRowMapper struct{ idx map[string]int }

func NewFullRowMapper(headers []string) *FullRowMapper {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.TrimSpace(h)] = i
	}
	return &FullRowMapper{idx: m}
}

func (m *FullRowMapper) col(row []string, name string) string {
	if i, ok := m.idx[name]; ok && i < len(row) {
		return row[i]
	}
	return ""
}

// Map converts a row to FocusRecord and returns flags (hasCredits, spotCredit).
// Classification is intentionally not done here.
func (m *FullRowMapper) Map(row []string) (types.FocusRecord, bool, bool, error) { //nolint:cyclop
	// numeric
	cost, _ := strconv.ParseFloat(m.col(row, "cost"), 64)
	usageQty, _ := strconv.ParseFloat(m.col(row, "usage.amount"), 64)
	// time
	start := c.ParseTimeAny(m.col(row, "usage_start_time"))
	end := c.ParseTimeAny(m.col(row, "usage_end_time"))

	// labels -> tags
	tags := map[string]string{}
	for _, labelCol := range []string{"labels", "system_labels", "resource.labels"} {
		raw := strings.TrimSpace(m.col(row, labelCol))
		if raw == "" {
			continue
		}
		for k, v := range c.ParseLabelsJSON(raw) { // tolerant
			tags[k] = v
		}
	}

	fr := types.FocusRecord{
		BillingAccountId:   firstNonEmpty(m.col(row, "billing_account_id"), m.col(row, "billing.account_id")),
		BillingAccountName: m.col(row, "billing_account_name"),
		BillingCurrency:    firstNonEmpty(m.col(row, "currency"), m.col(row, "billing_currency")),
		BillingPeriodStart: start,
		BillingPeriodEnd:   end,
		ChargeCategory:     types.ChargeCategories.Usage,
		ChargeClass:        types.ChargeClasses.OnDemand,
		ChargeDescription:  firstNonEmpty(m.col(row, "sku.description"), m.col(row, "sku.description_truncated")),
		ChargeFrequency:    "Daily",
		ChargePeriodStart:  start,
		ChargePeriodEnd:    end,
		ChargeSubcategory:  m.col(row, "service.description"),
		EffectiveCost:      cost,
		InvoiceIssuerName:  "Google",
		ListCost:           cost,
		ListUnitPrice:      unitPrice(cost, usageQty),
		PricingCategory:    types.PricingCategories.Standard,
		PricingQuantity:    usageQty,
		PricingUnit:        firstNonEmpty(m.col(row, "usage.unit"), m.col(row, "usage.pricing_unit")),
		ProviderName:       types.ProviderNames.GCP,
		PublisherName:      "Google",
		ResourceId:         firstNonEmpty(m.col(row, "resource.name"), m.col(row, "resource.global_name")),
		ResourceName:       firstNonEmpty(m.col(row, "resource.name"), m.col(row, "resource.global_name")),
		ResourceType:       m.col(row, "resource.name"),
		ServiceCategory:    m.col(row, "service.category"),
		ServiceName:        firstNonEmpty(m.col(row, "service.description"), m.col(row, "service.id")),
		SkuId:              firstNonEmpty(m.col(row, "sku.id"), m.col(row, "sku")),
		SkuPriceId:         firstNonEmpty(m.col(row, "pricing.price_id"), m.col(row, "sku.id")),
		SubAccountId:       firstNonEmpty(m.col(row, "project.id"), m.col(row, "project.number")),
		SubAccountName:     m.col(row, "project.name"),
		UsageQuantity:      usageQty,
		UsageUnit:          firstNonEmpty(m.col(row, "usage.unit"), m.col(row, "usage.pricing_unit")),
		Tags:               tags,

		SourceProvider:      "gcp",
		SourceFileName:      "",
		ConversionTimestamp: time.Now(),
	}

	if v := m.col(row, "location.region"); v != "" {
		fr.Region = &v
	}
	if v := m.col(row, "location.zone"); v != "" {
		fr.AvailabilityZone = &v
	}

	// credits capture and spot detection (flags only)
	rawCredits := strings.TrimSpace(m.col(row, "credits"))
	hasCredits := false
	spotCredit := false
	if rawCredits != "" {
		if cID, cType, cName, ok := c.ParseCredits(rawCredits); ok {
			if cID != "" {
				fr.CommitmentDiscountId = &cID
			}
			if cType != "" {
				fr.CommitmentDiscountType = &cType
			}
			if cName != "" {
				fr.CommitmentDiscountName = &cName
			}
			hasCredits = true
			lcType, lcName := strings.ToLower(cType), strings.ToLower(cName)
			if strings.Contains(lcType, "spot") || strings.Contains(lcName, "spot") || strings.Contains(lcType, "preempt") || strings.Contains(lcName, "preempt") {
				spotCredit = true
			}
		}
		if !spotCredit {
			lcRaw := strings.ToLower(rawCredits)
			if strings.Contains(lcRaw, "\"type\":\"spot\"") || strings.Contains(lcRaw, "spot vm") || strings.Contains(lcRaw, "preempt") {
				spotCredit = true
			}
		}
	}

	if spotCredit {
		fr.PricingCategory = types.PricingCategories.Spot
	}

	// Provider-local classification to remove root wrapper dependency:
	// - Negative effective cost => Credit (GCP)
	// - Presence of credits => Credit
	if fr.EffectiveCost < 0 {
		fr.ChargeCategory = types.ChargeCategories.Credit
	}
	if hasCredits {
		fr.ChargeCategory = types.ChargeCategories.Credit
	}

	// Apply shared unified classification tweaks (no-op if not applicable)
	c.ApplyUnifiedClassification("gcp", map[string]string{}, &fr)

	return fr, hasCredits, spotCredit, nil
}
