package quality

// Lightweight streaming aggregator for invariants to avoid re-reading the
// output file (used in post-conversion invariants computation). It mirrors
// the logic in ComputeInvariants but operates incrementally per record.
//
// NOTE: Keep this minimal; any heavy / comparative logic stays in
// CompareInvariants / ComputeInvariants to preserve test coverage parity.
// Additional optional counters (distinct services/resources) can be added
// later behind additive JSON fields.

import (
	focustypes "local/costscope/internal/core/focus/types"
	"time"
)

// InvariantsAggregator accumulates metrics row-by-row.
type InvariantsAggregator struct {
	rowCount         int
	sumEffectiveCost float64
	sumListCost      float64
	sumUsageQuantity float64

	chargeCat    map[string]float64
	pricingCat   map[string]float64
	providerDist map[string]float64

	negativeUsageAllowed    int
	negativeUsageViolations int

	// Distinct trackers (cheap set via map keys)
	distinctResources map[string]struct{}
	distinctServices  map[string]struct{}
}

// NewInvariantsAggregator constructs a new streaming aggregator.
func NewInvariantsAggregator() *InvariantsAggregator {
	return &InvariantsAggregator{
		chargeCat:         make(map[string]float64),
		pricingCat:        make(map[string]float64),
		providerDist:      make(map[string]float64),
		distinctResources: make(map[string]struct{}),
		distinctServices:  make(map[string]struct{}),
	}
}

// Add ingests a FocusRecord into the aggregated invariants.
func (a *InvariantsAggregator) Add(r focustypes.FocusRecord) {
	a.rowCount++
	a.sumEffectiveCost += r.EffectiveCost
	a.sumListCost += r.ListCost
	a.sumUsageQuantity += r.UsageQuantity
	a.chargeCat[r.ChargeCategory]++
	a.pricingCat[r.PricingCategory]++
	a.providerDist[r.ProviderName]++
	if r.UsageQuantity < 0 {
		if r.ChargeCategory == focustypes.ChargeCategories.Credit || r.ChargeCategory == focustypes.ChargeCategories.Adjustment {
			a.negativeUsageAllowed++
		} else {
			a.negativeUsageViolations++
		}
	}
	if r.ResourceId != "" {
		a.distinctResources[r.ResourceId] = struct{}{}
	}
	if r.ServiceName != "" {
		a.distinctServices[r.ServiceName] = struct{}{}
	}
}

// Produce converts accumulated state into InvariantMetrics (percentage
// distributions computed here). Violations slice left empty; comparison
// happens separately. Added metadata for distinct counters.
func (a *InvariantsAggregator) Produce() InvariantMetrics {
	m := InvariantMetrics{
		RowCount:                    a.rowCount,
		SumEffectiveCost:            a.sumEffectiveCost,
		SumListCost:                 a.sumListCost,
		SumUsageQuantity:            a.sumUsageQuantity,
		ChargeCategoryDistribution:  make(map[string]float64, len(a.chargeCat)),
		PricingCategoryDistribution: make(map[string]float64, len(a.pricingCat)),
		ProviderDistribution:        make(map[string]float64, len(a.providerDist)),
		NegativeUsageAllowedCount:   a.negativeUsageAllowed,
		NegativeUsageViolationCount: a.negativeUsageViolations,
		GeneratedAt:                 time.Now().UTC(),
		Metadata: map[string]string{
			"source":             "post_conversion",
			"distinct_resources": intToString(len(a.distinctResources)),
			"distinct_services":  intToString(len(a.distinctServices)),
		},
	}
	if a.rowCount == 0 {
		return m
	}
	toPct := func(src map[string]float64, dst map[string]float64) {
		for k, v := range src {
			dst[k] = (v / float64(a.rowCount)) * 100.0
		}
	}
	toPct(a.chargeCat, m.ChargeCategoryDistribution)
	toPct(a.pricingCat, m.PricingCategoryDistribution)
	toPct(a.providerDist, m.ProviderDistribution)
	return m
}

// Helper to avoid pulling in fmt for single use; tiny optimization.
func intToString(v int) string {
	// simple itoa for non-negative ints
	if v == 0 {
		return "0"
	}
	// length calc
	n := v
	digits := 0
	for n > 0 {
		n /= 10
		digits++
	}
	b := make([]byte, digits)
	n = v
	for i := digits - 1; i >= 0; i-- {
		b[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(b)
}
