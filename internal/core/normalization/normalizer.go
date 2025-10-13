package normalization

import (
	"strings"
	"time"

	"github.com/costscope/costscope/internal/core/cache"
	"github.com/costscope/costscope/internal/core/monitoring/telemetry"
)

// regionVariants maps provider -> variant(lowercased trimmed) -> canonical region code
var regionVariants = map[string]map[string]string{
	"aws": {
		"us-east-1":             "us-east-1",
		"us east 1":             "us-east-1",
		"us east (n. virginia)": "us-east-1",
		"use1":                  "us-east-1",
		"us-east1":              "us-east-1",
		"us-west-2":             "us-west-2",
		"us west 2":             "us-west-2",
		"uswest2":               "us-west-2",
		// Additional commercial regions
		"us-west-1":      "us-west-1",
		"us west 1":      "us-west-1",
		"eu-central-1":   "eu-central-1",
		"eu central 1":   "eu-central-1",
		"ap-northeast-1": "ap-northeast-1",
		"ap northeast 1": "ap-northeast-1",
		"ap-south-1":     "ap-south-1",
		"ap south 1":     "ap-south-1",
		"sa-east-1":      "sa-east-1",
		"sa east 1":      "sa-east-1",
		"eu-west-1":      "eu-west-1",
		"eu west 1":      "eu-west-1",
		"eu-west1":       "eu-west-1",
		"euw1":           "eu-west-1",
		"ap-southeast-1": "ap-southeast-1",
		"ap southeast 1": "ap-southeast-1",
		"apsoutheast1":   "ap-southeast-1",
		// GovCloud
		"us-gov-west-1":    "us-gov-west-1",
		"us gov west 1":    "us-gov-west-1",
		"us-gov-east-1":    "us-gov-east-1",
		"us gov east 1":    "us-gov-east-1",
		"govcloud us-west": "us-gov-west-1",
		"govcloud us-east": "us-gov-east-1",
		// China
		"cn-north-1":     "cn-north-1",
		"cn north 1":     "cn-north-1",
		"cn-northwest-1": "cn-northwest-1",
		"cn northwest 1": "cn-northwest-1",
	},
	"azure": {
		"eastus":       "eastus",
		"east us":      "eastus",
		"east-us":      "eastus",
		"westus2":      "westus2",
		"west us 2":    "westus2",
		"westeurope":   "westeurope",
		"west europe":  "westeurope",
		"northeurope":  "northeurope",
		"north europe": "northeurope",
		"centralus":    "centralus",
		"central us":   "centralus",
		// Additional public
		"uksouth":        "uksouth",
		"uk south":       "uksouth",
		"canadacentral":  "canadacentral",
		"canada central": "canadacentral",
		"australiaeast":  "australiaeast",
		"australia east": "australiaeast",
		// Sovereign / gov / china
		"chinaeast":            "chinaeast",
		"china east":           "chinaeast",
		"chinanorth":           "chinanorth",
		"china north":          "chinanorth",
		"chinaeast2":           "chinaeast2",
		"china east 2":         "chinaeast2",
		"chinaeast2(stage)":    "chinaeast2", // occasional suffix stripping
		"usgovvirginia":        "usgovvirginia",
		"us gov virginia":      "usgovvirginia",
		"usgovarizona":         "usgovarizona",
		"us gov arizona":       "usgovarizona",
		"germanywestcentral":   "germanywestcentral",
		"germany west central": "germanywestcentral",
		"francesouth":          "francesouth",
		"france south":         "francesouth",
	},
	"gcp": {
		"us-central1":       "us-central1",
		"us central 1":      "us-central1",
		"us central (iowa)": "us-central1",
		"europe-west1":      "europe-west1",
		"europe west 1":     "europe-west1",
		"europe-west 1":     "europe-west1",
		"asia-southeast1":   "asia-southeast1",
		"asia southeast 1":  "asia-southeast1",
		"asia-southeast 1":  "asia-southeast1",
		// Newer regions
		"us-east5":              "us-east5",
		"us east 5":             "us-east5",
		"europe-west8":          "europe-west8",
		"europe west 8":         "europe-west8",
		"asia-south2":           "asia-south2",
		"asia south 2":          "asia-south2",
		"australia-southeast2":  "australia-southeast2",
		"australia southeast 2": "australia-southeast2",
	},
}

var unitVariants = map[string]string{
	"hour":       "Hours",
	"hours":      "Hours",
	"hr":         "Hours",
	"hrs":        "Hours",
	"gb":         "GB",
	"gib":        "GB",
	"gbyte":      "GB",
	"gbytes":     "GB",
	"gb-hours":   "GB-Hours",
	"gb hours":   "GB-Hours",
	"gib-hours":  "GB-Hours",
	"vcpu-hours": "vCPU-Hours",
	"vcpu hours": "vCPU-Hours",
	"cpu-hours":  "vCPU-Hours",
	"request":    "Requests",
	"requests":   "Requests",
	"quantity":   "Quantity",
}

var (
	regionCache = cache.NewLRU(4096)
	unitCache   = cache.NewLRU(1024)
)

// statsProvider is a minimal interface to query cache stats for metrics updates.
type statsProvider interface {
	Stats() (hits, misses, evicts uint64, size int)
}

// setCacheMetrics updates size and eviction gauges for a given cache label.
func setCacheMetrics(sp statsProvider, label string) {
	_, _, ev, sz := sp.Stats()
	telemetry.CacheSize.WithLabelValues(label).Set(float64(sz))
	telemetry.CacheEvictions.WithLabelValues(label).Set(float64(ev))
}

// setCacheMetricsIfValid conditionally updates metrics when size is non-negative.
func setCacheMetricsIfValid(sp statsProvider, label string) {
	_, _, ev, sz := sp.Stats()
	if sz >= 0 {
		telemetry.CacheSize.WithLabelValues(label).Set(float64(sz))
		telemetry.CacheEvictions.WithLabelValues(label).Set(float64(ev))
	}
}

// PreWarm loads all known variants into the cache for warm start cache hits.
func PreWarm() {
	for prov, m := range regionVariants {
		for variant, canon := range m {
			regionCache.Add(prov+"|"+variant, canon)
		}
	}
	for v, canon := range unitVariants {
		unitCache.Add(v, canon)
	}
	// Expose initial sizes.
	setCacheMetrics(regionCache, "region_normalizer")
	setCacheMetrics(unitCache, "unit_normalizer")
}

// StartCacheMetricsRefresher launches a background goroutine to periodically refresh
// cache size/eviction gauges in case future code mutates caches via alternative paths.
// interval <= 0 disables refresh.
func StartCacheMetricsRefresher(interval time.Duration, stopCh <-chan struct{}) {
	if interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				setCacheMetricsIfValid(regionCache, "region_normalizer")
				setCacheMetricsIfValid(unitCache, "unit_normalizer")
			case <-stopCh:
				return
			}
		}
	}()
}

// RegionCacheStats exposes stats for region normalization cache.
func RegionCacheStats() (hits, misses, evicts uint64, size int) {
	return regionCache.Stats()
}

// UnitCacheStats exposes stats for unit normalization cache.
func UnitCacheStats() (hits, misses, evicts uint64, size int) {
	return unitCache.Stats()
}

// NormalizeRegion returns canonical region for provider & raw input (case/format agnostic).
// provider may be empty: all provider dictionaries are searched.
func NormalizeRegion(provider, raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw, false
	}
	p := strings.ToLower(strings.TrimSpace(provider))
	keyLower := strings.ToLower(raw)
	cacheKey := p + "|" + keyLower
	if v, ok := regionCache.Get(cacheKey); ok {
		telemetry.NormalizationCacheHits.WithLabelValues("region", providerLabel(p)).Inc()
		return v, true
	}
	if p != "" {
		if dict, ok := regionVariants[p]; ok {
			if canon, ok2 := dict[keyLower]; ok2 {
				regionCache.Add(cacheKey, canon)
				setCacheMetrics(regionCache, "region_normalizer")
				return canon, true
			}
		}
	}
	for prov, dict := range regionVariants {
		if canon, ok := dict[keyLower]; ok {
			regionCache.Add(prov+"|"+keyLower, canon)
			setCacheMetrics(regionCache, "region_normalizer")
			return canon, true
		}
	}
	if looksCanonicalRegion(keyLower) {
		regionCache.Add(cacheKey, raw)
		setCacheMetrics(regionCache, "region_normalizer")
		return raw, true
	}
	return raw, false
}

// NormalizeUnit returns canonical unit string.
func NormalizeUnit(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw, false
	}
	key := strings.ToLower(raw)
	if v, ok := unitCache.Get(key); ok {
		telemetry.NormalizationCacheHits.WithLabelValues("unit", "any").Inc()
		return v, true
	}
	if canon, ok := unitVariants[key]; ok {
		unitCache.Add(key, canon)
		setCacheMetrics(unitCache, "unit_normalizer")
		return canon, true
	}
	return raw, false
}

func providerLabel(p string) string {
	if p == "" {
		return "any"
	}
	return p
}

func looksCanonicalRegion(lower string) bool {
	if strings.Count(lower, "-") >= 1 && (strings.HasPrefix(lower, "us") || strings.HasPrefix(lower, "eu") || strings.HasPrefix(lower, "ap") || strings.HasPrefix(lower, "asia") || strings.HasPrefix(lower, "europe") || strings.HasPrefix(lower, "cn")) {
		return true
	}
	if lower == "eastus" || lower == "westus2" || lower == "westeurope" || lower == "northeurope" || lower == "centralus" || lower == "usgovvirginia" || lower == "usgovarizona" || lower == "chinaeast" || lower == "chinanorth" {
		return true
	}
	return false
}
