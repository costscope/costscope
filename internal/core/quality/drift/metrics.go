package drift

import "github.com/costscope/costscope/internal/core/monitoring/telemetry"

// RecordMetrics publishes Prometheus metrics for a drift Report.
// Caller should invoke after successful Run (CLI/API) – kept separate to preserve purity of Run.
func RecordMetrics(r Report) {
	telemetry.DriftChiSquare.WithLabelValues("charge").Set(r.ChiSquare.ChargeCategory)
	telemetry.DriftChiSquare.WithLabelValues("pricing").Set(r.ChiSquare.PricingCategory)
	telemetry.DriftPValue.Set(r.ChiSquare.PValue)
	telemetry.DriftTrendSlope.WithLabelValues("effective").Set(r.Trend.EffectiveSlope)
	telemetry.DriftTrendSlope.WithLabelValues("usage").Set(r.Trend.UsageSlope)
	telemetry.DriftTrendSlope.WithLabelValues("row").Set(r.Trend.RowSlope)
	for b, v := range r.CostBucketDeltas.Effective {
		telemetry.DriftBucketDelta.WithLabelValues("effective", b).Set(v)
	}
	for b, v := range r.CostBucketDeltas.Usage {
		telemetry.DriftBucketDelta.WithLabelValues("usage", b).Set(v)
	}
	telemetry.DriftAnalysesTotal.WithLabelValues(r.ChiSquare.Severity).Inc()
	if r.Percentiles.MaxEffectiveDelta > 0 || r.Percentiles.ThresholdExceeded {
		telemetry.DriftPercentileMax.WithLabelValues("effective").Set(r.Percentiles.MaxEffectiveDelta)
	}
	if r.Percentiles.MaxUsageDelta > 0 || r.Percentiles.ThresholdExceeded {
		telemetry.DriftPercentileMax.WithLabelValues("usage").Set(r.Percentiles.MaxUsageDelta)
	}
}
