package drift

// Advanced distribution & cost drift detection (TASK-DRIFT-ADV).
//
// Responsibilities:
// 1. Compute chi-square statistic between current and baseline categorical distributions.
// 2. Compute cost bucket deltas (user-friendly segmentation of effective_cost and usage_quantity).
// 3. Produce a trend summary (simple moving deltas across supplied historical snapshots).
// 4. Emit a JSON friendly report structure consumed by CLI / API.
//
// Design Notes:
// - We deliberately do NOT pull in heavy math/stat dependencies; chi-square is computed inline.
// - For small expected counts (<5) we fallback to combining into an "_other" bucket to keep
//   approximation validity reasonable.
// - Trend analysis stays lightweight: caller can pass prior N snapshots; we compute slope via
//   ordinary least squares on (index,value) for selected numeric metrics.
// - All logging to be performed by caller; this package is pure/side-effect free.

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// Distribution represents categorical frequencies (raw counts, not percentages).
type Distribution map[string]float64

// Snapshot represents a historical point-in-time dataset summary used for trends.
type Snapshot struct {
	TimestampUnix int64              `json:"timestamp_unix"`
	RowCount      int64              `json:"row_count"`
	SumEffective  float64            `json:"sum_effective_cost"`
	SumList       float64            `json:"sum_list_cost"`
	SumUsage      float64            `json:"sum_usage_quantity"`
	ChargeDist    map[string]float64 `json:"charge_category_dist,omitempty"` // counts
	PricingDist   map[string]float64 `json:"pricing_category_dist,omitempty"`
}

// CostBuckets groups cost / usage magnitudes for coarse drift signalling.
type CostBuckets struct {
	EffectiveCounts map[string]int64 `json:"effective_cost_counts"`
	UsageCounts     map[string]int64 `json:"usage_quantity_counts"`
}

// Report is the output structure returned by Run.
type Report struct {
	ChiSquare struct {
		ChargeCategory   float64 `json:"charge_category_chi_square"`
		PricingCategory  float64 `json:"pricing_category_chi_square"`
		DegreesOfFreedom int     `json:"degrees_of_freedom"`
		PValue           float64 `json:"p_value"`
		// ThresholdExceeded indicates p < alpha (default alpha provided by caller) for any distribution.
		ThresholdExceeded bool   `json:"threshold_exceeded"`
		Severity          string `json:"severity"` // none|low|medium|high based on p-value bands
	} `json:"chi_square"`
	CostBucketDeltas struct {
		Effective map[string]float64 `json:"effective_cost_bucket_delta_pct"`
		Usage     map[string]float64 `json:"usage_quantity_bucket_delta_pct"`
	} `json:"cost_bucket_deltas"`
	Percentiles struct {
		BaselineEffective map[string]float64 `json:"baseline_effective_percentiles,omitempty"`
		CurrentEffective  map[string]float64 `json:"current_effective_percentiles,omitempty"`
		DeltaEffectivePct map[string]float64 `json:"delta_effective_percentiles_relative_pct,omitempty"`
		BaselineUsage     map[string]float64 `json:"baseline_usage_percentiles,omitempty"`
		CurrentUsage      map[string]float64 `json:"current_usage_percentiles,omitempty"`
		DeltaUsagePct     map[string]float64 `json:"delta_usage_percentiles_relative_pct,omitempty"`
		MaxEffectiveDelta float64            `json:"max_effective_percentile_relative_delta,omitempty"`
		MaxUsageDelta     float64            `json:"max_usage_percentile_relative_delta,omitempty"`
		ThresholdExceeded bool               `json:"percentile_threshold_exceeded,omitempty"`
	} `json:"percentiles"`
	Trend struct {
		EffectiveSlope float64 `json:"effective_cost_slope"`
		UsageSlope     float64 `json:"usage_quantity_slope"`
		RowSlope       float64 `json:"row_count_slope"`
	} `json:"trend"`
	Notes []string `json:"notes,omitempty"`
}

// Config controls drift computation.
type Config struct {
	Alpha                    float64   // significance threshold (default 0.01)
	MinExpectedCount         float64   // minimum expected frequency before bucket folding (default 5)
	BucketSchema             []float64 // optional ascending boundaries for bucketization; default used if empty
	Percentiles              []float64 // optional percentile list (0-1). Default: 0.5,0.9,0.95,0.99
	PercentileDriftThreshold float64   // relative threshold (abs delta) to flag warning (default 0.01 = 1%)
}

// Run performs advanced drift analysis.
// baselineDist / currentDist are raw count distributions (not percentages).
// Prior snapshots must be ordered oldest->newest (excluding current). Current snapshot is handled via currentDist + row/aggregate counts.
func Run(cfg Config,
	baselineCharge, currentCharge Distribution,
	baselinePricing, currentPricing Distribution,
	baselineBuckets, currentBuckets CostBuckets,
	snapshots []Snapshot, current Snapshot,
	baselineEffectiveVals, currentEffectiveVals, baselineUsageVals, currentUsageVals []float64,
) (Report, error) {
	if cfg.Alpha <= 0 {
		cfg.Alpha = 0.01
	}
	if cfg.MinExpectedCount <= 0 {
		cfg.MinExpectedCount = 5
	}
	if len(cfg.Percentiles) == 0 {
		cfg.Percentiles = []float64{0.5, 0.9, 0.95, 0.99}
	}
	if cfg.PercentileDriftThreshold <= 0 {
		cfg.PercentileDriftThreshold = 0.01
	}
	var rep Report

	// 1. Chi-square for charge & pricing.
	chiCharge, dofCharge, pCharge, noteCharge := chiSquareSafe(baselineCharge, currentCharge, cfg.MinExpectedCount)
	chiPricing, dofPricing, pPricing, notePricing := chiSquareSafe(baselinePricing, currentPricing, cfg.MinExpectedCount)
	rep.ChiSquare.ChargeCategory = chiCharge
	rep.ChiSquare.PricingCategory = chiPricing
	// Use minimal df for summary (if differ choose max to be conservative) and min p-value.
	rep.ChiSquare.DegreesOfFreedom = maxInt(dofCharge, dofPricing)
	rep.ChiSquare.PValue = math.Min(pCharge, pPricing)
	if rep.ChiSquare.PValue < cfg.Alpha {
		rep.ChiSquare.ThresholdExceeded = true
	}
	rep.ChiSquare.Severity = classifySeverity(rep.ChiSquare.PValue)
	if noteCharge != "" {
		rep.Notes = append(rep.Notes, noteCharge)
	}
	if notePricing != "" {
		rep.Notes = append(rep.Notes, notePricing)
	}

	// 2. Cost bucket deltas (percentage point change relative to baseline bucket share).
	rep.CostBucketDeltas.Effective = bucketDeltaPct(baselineBuckets.EffectiveCounts, currentBuckets.EffectiveCounts)
	rep.CostBucketDeltas.Usage = bucketDeltaPct(baselineBuckets.UsageCounts, currentBuckets.UsageCounts)

	// 3. Trend analysis including current (append current to snapshots copy).
	timeline := append([]Snapshot{}, snapshots...)
	timeline = append(timeline, current)
	effVals := make([]float64, len(timeline))
	usageVals := make([]float64, len(timeline))
	rowVals := make([]float64, len(timeline))
	for i, s := range timeline {
		effVals[i] = s.SumEffective
		usageVals[i] = s.SumUsage
		rowVals[i] = float64(s.RowCount)
	}
	rep.Trend.EffectiveSlope = slope(effVals)
	rep.Trend.UsageSlope = slope(usageVals)
	rep.Trend.RowSlope = slope(rowVals)

	// 4. Percentile comparison (KS-like heuristic via selected percentiles)
	if len(baselineEffectiveVals) > 0 && len(currentEffectiveVals) > 0 {
		be := append([]float64{}, baselineEffectiveVals...)
		ce := append([]float64{}, currentEffectiveVals...)
		bu := append([]float64{}, baselineUsageVals...)
		cu := append([]float64{}, currentUsageVals...)
		sort.Float64s(be)
		sort.Float64s(ce)
		sort.Float64s(bu)
		sort.Float64s(cu)
		rep.Percentiles.BaselineEffective = map[string]float64{}
		rep.Percentiles.CurrentEffective = map[string]float64{}
		rep.Percentiles.DeltaEffectivePct = map[string]float64{}
		rep.Percentiles.BaselineUsage = map[string]float64{}
		rep.Percentiles.CurrentUsage = map[string]float64{}
		rep.Percentiles.DeltaUsagePct = map[string]float64{}
		var maxEff, maxUse float64
		for _, p := range cfg.Percentiles {
			if p <= 0 || p >= 1 {
				continue
			}
			label := fmt.Sprintf("p%g", p*100)
			bvE := percentile(be, p)
			cvE := percentile(ce, p)
			bvU := percentile(bu, p)
			cvU := percentile(cu, p)
			rep.Percentiles.BaselineEffective[label] = bvE
			rep.Percentiles.CurrentEffective[label] = cvE
			rep.Percentiles.BaselineUsage[label] = bvU
			rep.Percentiles.CurrentUsage[label] = cvU
			var relEff, relUse float64
			if bvE != 0 {
				relEff = math.Abs(cvE-bvE) / math.Abs(bvE)
			} else if cvE != 0 { // treat baseline zero, current non-zero as 100% drift
				relEff = 1
			}
			if bvU != 0 {
				relUse = math.Abs(cvU-bvU) / math.Abs(bvU)
			} else if cvU != 0 {
				relUse = 1
			}
			rep.Percentiles.DeltaEffectivePct[label] = relEff * 100
			rep.Percentiles.DeltaUsagePct[label] = relUse * 100
			if relEff > maxEff {
				maxEff = relEff
			}
			if relUse > maxUse {
				maxUse = relUse
			}
		}
		rep.Percentiles.MaxEffectiveDelta = maxEff * 100
		rep.Percentiles.MaxUsageDelta = maxUse * 100
		if maxEff > cfg.PercentileDriftThreshold || maxUse > cfg.PercentileDriftThreshold {
			rep.Percentiles.ThresholdExceeded = true
			rep.Notes = append(rep.Notes, fmt.Sprintf("percentile_drift_exceeded eff=%.2f%% use=%.2f%% threshold=%.2f%%", maxEff*100, maxUse*100, cfg.PercentileDriftThreshold*100))
		}
	}

	return rep, nil
}

// chiSquareSafe computes chi-square with bucket folding for low expected counts.
func chiSquareSafe(base, cur Distribution, minExpected float64) (chi float64, dof int, p float64, note string) {
	// Merge key set
	keys := map[string]struct{}{}
	var baseTotal, curTotal float64
	for k, v := range base {
		keys[k] = struct{}{}
		baseTotal += v
	}
	for k, v := range cur {
		keys[k] = struct{}{}
		curTotal += v
	}
	if baseTotal == 0 || curTotal == 0 {
		return 0, 0, 1, "empty_distribution"
	}
	type row struct {
		k    string
		b, c float64
	}
	rows := make([]row, 0, len(keys))
	for k := range keys {
		rows = append(rows, row{k, base[k], cur[k]})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].k < rows[j].k })
	// Expected for current given baseline proportions: e_i = total_cur * (b_i/base_total)
	// We'll fold small expected buckets into _other.
	var othersB, othersC float64
	filtered := make([]row, 0, len(rows))
	for _, r := range rows {
		exp := curTotal * (r.b / baseTotal)
		if exp < minExpected {
			othersB += r.b
			othersC += r.c
			continue
		}
		filtered = append(filtered, r)
	}
	if othersB > 0 || othersC > 0 {
		filtered = append(filtered, row{"_other", othersB, othersC})
	}
	if len(filtered) <= 1 {
		return 0, 0, 1, "insufficient_degrees_of_freedom"
	}
	for _, r := range filtered {
		expected := curTotal * (r.b / baseTotal)
		if expected <= 0 {
			continue
		}
		diff := r.c - expected
		chi += (diff * diff) / expected
	}
	dof = len(filtered) - 1
	// Approximate p-value using survival function of chi-square via incomplete gamma (simplified).
	p = chiSquarePValue(chi, dof)
	return
}

// chiSquarePValue approximates the upper-tail p-value.
func chiSquarePValue(chi float64, k int) float64 {
	if k <= 0 {
		return 1
	}
	// Using regularized gamma Q(k/2, chi/2). For small implementation scope we use a series / continued fraction hybrid.
	x := chi / 2
	a := float64(k) / 2
	return gammaQ(a, x)
}

// gammaQ = 1 - P(a,x); we implement using Lentz's algorithm for the continued fraction when x >= a+1 else series.
func gammaQ(a, x float64) float64 {
	if x <= 0 {
		return 1
	}
	if x < a+1 { // series for P then 1-P
		sum := 1 / a
		term := sum
		for n := 1; n < 100; n++ {
			term *= x / (a + float64(n))
			sum += term
			if math.Abs(term) < 1e-12 {
				break
			}
		}
		p := sum * math.Exp(-x+a*math.Log(x)-lgamma(a))
		if p > 1 {
			p = 1
		}
		if p < 0 {
			p = 0
		}
		return 1 - p
	}
	// Continued fraction for Q directly
	eps := 1e-12
	F := 0.0
	C := 1e30
	D := 0.0
	for i := 1; i < 200; i++ {
		var an float64
		if i%2 == 1 { // odd
			n := (i + 1) / 2
			an = float64(n) * (a - float64(n))
		} else {
			n := i / 2
			an = -float64(n) * (float64(n) - a)
		}
		D = 1 + an*D
		if math.Abs(D) < eps {
			D = eps
		}
		C = 1 + an/C
		if math.Abs(C) < eps {
			C = eps
		}
		D = 1 / D
		delta := C * D
		F *= delta
		if math.Abs(delta-1) < eps {
			break
		}
	}
	return math.Exp(-x+a*math.Log(x)-lgamma(a)) * F
}

// lgamma wrapper (math.Lgamma returns value, sign). We only need the natural log of Gamma.
func lgamma(x float64) float64 { v, _ := math.Lgamma(x); return v }

func bucketDeltaPct(base, cur map[string]int64) map[string]float64 {
	out := make(map[string]float64)
	var totalBase, totalCur float64
	for _, v := range base {
		totalBase += float64(v)
	}
	for _, v := range cur {
		totalCur += float64(v)
	}
	if totalBase == 0 || totalCur == 0 {
		return out
	}
	keys := map[string]struct{}{}
	for k := range base {
		keys[k] = struct{}{}
	}
	for k := range cur {
		keys[k] = struct{}{}
	}
	for k := range keys {
		bp := float64(base[k]) / totalBase * 100
		cp := float64(cur[k]) / totalCur * 100
		out[k] = cp - bp
	}
	return out
}

func slope(vals []float64) float64 {
	n := float64(len(vals))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range vals {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

// BuildCostBuckets converts raw values into predefined magnitude buckets (powers of 10 style plus fine-grained low range).
func BuildCostBuckets(effectiveValues, usageValues []float64, schema []float64) (CostBuckets, error) {
	if len(effectiveValues) == 0 && len(usageValues) == 0 {
		return CostBuckets{}, errors.New("no values")
	}
	if len(schema) == 0 { // default boundaries
		schema = []float64{0.01, 0.1, 1, 10, 100, 1000, 10000}
	}
	// copy & sort schema
	boundaries := append([]float64{}, schema...)
	sort.Float64s(boundaries)
	bk := func(v float64) string {
		av := math.Abs(v)
		prev := 0.0
		for _, b := range boundaries {
			if av < b {
				return rangeLabel(prev, b)
			}
			prev = b
		}
		return ">=" + humanNum(boundaries[len(boundaries)-1])
	}
	eff := map[string]int64{}
	use := map[string]int64{}
	for _, v := range effectiveValues {
		eff[bk(v)]++
	}
	for _, v := range usageValues {
		use[bk(v)]++
	}
	return CostBuckets{EffectiveCounts: eff, UsageCounts: use}, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// -------- helpers for bucket labeling & severity --------
func rangeLabel(a, b float64) string {
	if a == 0 {
		return "<" + humanNum(b)
	}
	return humanNum(a) + "-" + humanNum(b)
}

func humanNum(v float64) string {
	switch {
	case v >= 1000000:
		return fmt.Sprintf("%gM", v/1000000)
	case v >= 10000:
		return fmt.Sprintf("%gk", v/1000)
	case v >= 1000:
		return fmt.Sprintf("%gk", v/1000)
	default:
		// limit precision without trailing zeros noise
		return trimFloat(v)
	}
}

func trimFloat(v float64) string { return fmt.Sprintf("%g", v) }

func classifySeverity(p float64) string {
	switch {
	case p < 0.001:
		return "high"
	case p < 0.01:
		return "medium"
	case p < 0.05:
		return "low"
	default:
		return "none"
	}
}

// percentile returns the value at quantile q (0<q<1) using nearest-rank after linear position interpolation.
func percentile(sorted []float64, q float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if q <= 0 {
		return sorted[0]
	}
	if q >= 1 {
		return sorted[n-1]
	}
	pos := q * float64(n-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sorted[lo]
	}
	frac := pos - float64(lo)
	return sorted[lo] + (sorted[hi]-sorted[lo])*frac
}
