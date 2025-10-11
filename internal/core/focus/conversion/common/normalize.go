package common

import (
	"strings"

	"local/costscope/internal/core/cache"
	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/focus/validation"
	"local/costscope/internal/core/monitoring/telemetry"
	"local/costscope/internal/core/normalization"
)

var (
	currencyCache = cache.NewLRU(256)
	unitCache     = cache.NewLRU(1024)
	regionCache   = cache.NewLRU(4096)
)

func NormalizeCurrency(raw string) string {
	if raw == "" {
		return raw
	}
	key := strings.ToUpper(strings.TrimSpace(raw))
	if v, ok := currencyCache.Get(key); ok {
		telemetry.EnumCacheHits.WithLabelValues("currency", "any").Inc()
		return v
	}
	normalized, _ := validation.NormalizeCurrency(key)
	currencyCache.Add(key, normalized)
	if _, _, ev, sz := currencyCache.Stats(); sz >= 0 {
		telemetry.CacheSize.WithLabelValues("currency_enum").Set(float64(sz))
		telemetry.CacheEvictions.WithLabelValues("currency_enum").Set(float64(ev))
	}
	return normalized
}

func NormalizeRegion(provider string, ptr *string) *string {
	if ptr == nil || *ptr == "" {
		return ptr
	}
	raw := strings.TrimSpace(*ptr)
	lower := strings.ToLower(raw)
	cacheKey := provider + "|" + lower
	if v, ok := regionCache.Get(cacheKey); ok {
		telemetry.EnumCacheHits.WithLabelValues("region", provider).Inc()
		return &v
	}
	canon, ok := normalization.NormalizeRegion(provider, raw)
	if !ok {
		canon = lower
	}
	regionCache.Add(cacheKey, canon)
	if _, _, ev, sz := regionCache.Stats(); sz >= 0 {
		telemetry.CacheSize.WithLabelValues("region_enum").Set(float64(sz))
		telemetry.CacheEvictions.WithLabelValues("region_enum").Set(float64(ev))
	}
	return &canon
}

func NormalizeUnit(raw string) string {
	if raw == "" {
		return raw
	}
	keyLower := strings.ToLower(strings.TrimSpace(raw))
	if v, ok := unitCache.Get(keyLower); ok {
		telemetry.EnumCacheHits.WithLabelValues("unit", "any").Inc()
		return v
	}
	if canon, ok := normalization.NormalizeUnit(raw); ok {
		unitCache.Add(keyLower, canon)
		if _, _, ev, sz := unitCache.Stats(); sz >= 0 {
			telemetry.CacheSize.WithLabelValues("unit_enum").Set(float64(sz))
			telemetry.CacheEvictions.WithLabelValues("unit_enum").Set(float64(ev))
		}
		return canon
	}
	canon := CanonicalUnit(raw)
	unitCache.Add(keyLower, canon)
	if _, _, ev, sz := unitCache.Stats(); sz >= 0 {
		telemetry.CacheSize.WithLabelValues("unit_enum").Set(float64(sz))
		telemetry.CacheEvictions.WithLabelValues("unit_enum").Set(float64(ev))
	}
	return canon
}

// ApplyUnifiedNormalization mutates the FocusRecord with canonical enums (unified path only).
func ApplyUnifiedNormalization(provider string, fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	// For AWS, keep normalization lightweight (parity with legacy fast path)
	// to minimize per-row overhead while preserving semantics expected by
	// perf/parity guards.
	if provider == "aws" {
		NormalizeFocusRecord(fr)
		return
	}
	if fr.Region != nil {
		fr.Region = NormalizeRegion(provider, fr.Region)
	}
	fr.BillingCurrency = NormalizeCurrency(fr.BillingCurrency)
	if fr.UsageUnit != "" {
		fr.UsageUnit = NormalizeUnit(fr.UsageUnit)
	}
	if fr.PricingUnit != "" {
		fr.PricingUnit = NormalizeUnit(fr.PricingUnit)
	}
	fr.ChargeCategory = strings.TrimSpace(fr.ChargeCategory)
	fr.ChargeClass = strings.TrimSpace(fr.ChargeClass)
	fr.PricingCategory = strings.TrimSpace(fr.PricingCategory)
}

// NormalizeFocusRecord applies lightweight normalization for unified mapper parity.
func NormalizeFocusRecord(fr *types.FocusRecord) {
	if fr == nil {
		return
	}
	if fr.Region != nil {
		r := strings.TrimSpace(*fr.Region)
		r = strings.ToLower(r)
		fr.Region = &r
	}
	fr.BillingCurrency = strings.ToUpper(strings.TrimSpace(fr.BillingCurrency))
	if fr.UsageUnit != "" {
		fr.UsageUnit = CanonicalUnit(fr.UsageUnit)
	}
	if fr.PricingUnit != "" {
		fr.PricingUnit = CanonicalUnit(fr.PricingUnit)
	}
	fr.ChargeCategory = strings.TrimSpace(fr.ChargeCategory)
	fr.ChargeClass = strings.TrimSpace(fr.ChargeClass)
	fr.PricingCategory = strings.TrimSpace(fr.PricingCategory)
}

// Canonical unit constants and helpers
const (
	CUnitHours     = "Hours"
	CUnitGB        = "GB"
	CUnitMB        = "MB"
	CUnitVCPUHours = "vCPU-Hours"
)

func CanonicalUnit(u string) string {
	u = strings.TrimSpace(u)
	lu := strings.ToLower(u)
	switch lu {
	case "hours", "hour", "hrs", "hr":
		return CUnitHours
	case "gb", "gib":
		return CUnitGB
	case "mb", "mib":
		return CUnitMB
	case "vcpu", "vcpu-hours", "vcpu hour", "vcpu hours":
		return CUnitVCPUHours
	default:
		if isAllAlpha(lu) && len(u) > 1 && len(u) <= 6 {
			return strings.ToUpper(u)
		}
		return u
	}
}

func isAllAlpha(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return s != ""
}
