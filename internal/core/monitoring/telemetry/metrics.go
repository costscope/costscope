package telemetry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Reports resolve-output metrics (MVP)
	ReportsResolveRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "reports",
		Name:      "resolve_output_requests_total",
		Help:      "Total report output path resolution requests",
	}, []string{"format"})

	ReportsResolveCollisions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "reports",
		Name:      "resolve_output_collisions_total",
		Help:      "Total detected collisions during output resolution (future use)",
	})

	ReportsResolveLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "reports",
		Name:      "resolve_output_latency_seconds",
		Help:      "Latency of report output path resolution",
		Buckets:   prometheus.DefBuckets,
	}, []string{"format", "source"})
	// Converter metrics
	ConverterRecords = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "converter",
		Name:      "records_total",
		Help:      "Total records processed by converters",
	}, []string{"provider", "mode", "status"})

	ConverterDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "converter",
		Name:      "duration_seconds",
		Help:      "Conversion duration in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"provider", "mode"})

	// Minimal observability task (TASK-OBS-BASE) additions
	ConversionDurationSimple = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "conversion_duration_seconds",
		Help:      "End-to-end conversion duration (seconds) per provider and path mode (legacy|unified)",
		Buckets:   prometheus.DefBuckets,
	}, []string{"provider", "path"})

	MapperRowsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "mapper_rows_total",
		Help:      "Total rows mapped into FOCUS records by provider and path (legacy|unified)",
	}, []string{"provider", "path"})

	ConversionActiveJobs = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "active_jobs",
		Help:      "Current number of active conversion jobs",
	})

	// Asynchronous conversion job lifecycle metrics (activation of previously dead code paths)
	ConversionJobsSubmitted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "jobs_submitted_total",
		Help:      "Total asynchronous conversion jobs submitted",
	})

	ConversionJobsCompleted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "jobs_completed_total",
		Help:      "Total asynchronous conversion jobs completed partitioned by outcome (success|failed|cancelled)",
	}, []string{"outcome"})

	ConversionJobDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "job_duration_seconds",
		Help:      "Duration of completed asynchronous conversion jobs in seconds (successful only)",
		Buckets:   prometheus.DefBuckets,
	})

	// Unified mapper focused performance metrics (parity/perf tracking)
	UnifiedMapperRows = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "unified_mapper",
		Name:      "rows_total",
		Help:      "Total rows processed (after successful mapping) by unified vs legacy paths",
	}, []string{"provider", "path"})

	UnifiedMapperDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "unified_mapper",
		Name:      "duration_seconds",
		Help:      "Wall-clock duration of conversion per provider/path (unified vs legacy)",
		Buckets:   prometheus.DefBuckets,
	}, []string{"provider", "path"})

	UnifiedMapperErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "unified_mapper",
		Name:      "errors_total",
		Help:      "Total mapping errors encountered by provider/path (unified vs legacy)",
	}, []string{"provider", "path"})

	// Classifier decision metrics (lightweight visibility of charge category choices)
	ClassifierDecisions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "classifier",
		Name:      "decisions_total",
		Help:      "Total charge category decisions by provider and path (legacy|unified)",
	}, []string{"provider", "path", "decision"})

	// Exporter metrics
	ExporterBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "exporter",
		Name:      "bytes_total",
		Help:      "Total bytes exported",
	}, []string{"target", "format"})

	ExporterDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "exporter",
		Name:      "duration_seconds",
		Help:      "Export duration in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"target", "format"})

	// Streaming metrics
	StreamingEvents = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "streaming",
		Name:      "events_total",
		Help:      "Total streaming events processed",
	}, []string{"job", "status"})

	StreamingLag = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "streaming",
		Name:      "lag_seconds",
		Help:      "Streaming processing lag in seconds",
	}, []string{"job"})

	// Observability expansion (TASK-OBS-EXP) additions
	ParquetRotationSize = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:  "costscope",
		Subsystem:  "parquet",
		Name:       "rotation_size_bytes",
		Help:       "Size in bytes of each rotated Parquet segment (post-finalization)",
		Objectives: map[float64]float64{0.5: 0.01, 0.9: 0.01, 0.99: 0.001},
	})

	StreamingBackpressure = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "streaming",
		Name:      "backpressure_total",
		Help:      "Total number of backpressure events (publish attempts found full buffer)",
	})

	MapperLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "mapper_latency_seconds",
		Help:      "Latency of mapping a chunk of provider billing rows to FOCUS records",
		Buckets:   prometheus.DefBuckets,
	}, []string{"provider", "path"})

	// Normalization cache metrics
	NormalizationCacheHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "normalization",
		Name:      "cache_hits_total",
		Help:      "Total cache hits for normalization lookups (region/unit)",
	}, []string{"type", "provider"})

	// Enum normalization helper specific cache hits (currency/region/unit) extracted to unified helper.
	// Note: No subsystem to satisfy exact requested metric name: costscope_enum_cache_hits_total
	EnumCacheHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Name:      "enum_cache_hits_total",
		Help:      "Total cache hits for enum normalization (currency|region|unit)",
	}, []string{"kind", "provider"})

	// Cache size / eviction gauges (tuning visibility) for provider enum/unit caches and unified mapper enums.
	CacheSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "cache",
		Name:      "entries",
		Help:      "Current number of entries in logical caches (normalization & enum caches)",
	}, []string{"cache"})

	CacheEvictions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "cache",
		Name:      "evictions_total",
		Help:      "Total evictions observed for caches (LRU based). Gauge mirrors internal monotonic counter.",
	}, []string{"cache"})

	// Drift analysis metrics (TASK-DRIFT-ADV ENHANCED)
	DriftChiSquare = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "drift",
		Name:      "chi_square",
		Help:      "Chi-square statistic per distribution type (charge|pricing)",
	}, []string{"distribution"})

	DriftPValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "drift",
		Name:      "p_value",
		Help:      "Minimum p-value across evaluated distributions",
	})

	DriftTrendSlope = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "drift",
		Name:      "trend_slope",
		Help:      "Trend slope (linear regression) for metrics (effective|usage|row)",
	}, []string{"metric"})

	DriftBucketDelta = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "drift",
		Name:      "bucket_delta_percentage_points",
		Help:      "Percentage point delta for cost/usage buckets vs baseline",
	}, []string{"metric", "bucket"})

	DriftAnalysesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "drift",
		Name:      "analyses_total",
		Help:      "Total drift analyses executed labeled by severity",
	}, []string{"severity"})

	DriftPercentileMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "drift",
		Name:      "percentile_max_relative_delta_pct",
		Help:      "Maximum relative percentile delta (effective|usage) in percent",
	}, []string{"metric"})

	// Health/readiness metric (optional, non-blocking): 1 when /health/ready reports ready, else 0.
	HealthReadiness = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "costscope",
		Subsystem: "health",
		Name:      "readiness",
		Help:      "Readiness status as a gauge (1=ready, 0=not ready)",
	})

	// Integration action (CLI registrar) instrumentation (TASK-INTEGRATION-METRICS)
	IntegrationActionCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "integration",
		Name:      "action_calls_total",
		Help:      "Total integration action invocations partitioned by action id, category and status (success|error)",
	}, []string{"action_id", "category", "status"})

	IntegrationActionDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "integration",
		Name:      "action_duration_seconds",
		Help:      "Duration of integration action handlers in seconds partitioned by action id, category and status",
		Buckets:   prometheus.DefBuckets,
	}, []string{"action_id", "category", "status"})

	// Error classification (TASK-INTEGRATION-ERROR-METRICS)
	IntegrationActionErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "integration",
		Name:      "action_errors_total",
		Help:      "Total integration action errors partitioned by action id, category and error_type (validation|not_found|conflict|timeout|unauthorized|other)",
	}, []string{"action_id", "category", "error_type"})

	// Authentication failure metrics (incremented on failed auth attempts)
	AuthFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "auth",
		Name:      "failures_total",
		Help:      "Total authentication failures partitioned by reason (missing_header|bad_format|validation_error|expired|issuer|forbidden_role|forbidden_scope)",
	}, []string{"reason"})

	// Azure discount normalization counter
	AzureDiscountNormalizations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "azure",
		Name:      "discount_normalizations_total",
		Help:      "Total Azure discount category normalizations applied (usage -> Discount or variant canonicalization)",
	}, []string{"provider", "path"})

	// Azure discount normalization skips (diagnostic when env disables normalization)
	AzureDiscountNormalizationSkips = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "azure",
		Name:      "discount_normalization_skips_total",
		Help:      "Total Azure discount normalizations skipped due to diagnostic disable flag",
	}, []string{"provider"})

	// Query builder build counter (core vs duckdb vs extended variants)
	QueryBuilderBuilds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "querybuilder",
		Name:      "build_total",
		Help:      "Total SQL query builds by builder type (focus|duckdb|extended)",
	}, []string{"type"})

	// RBAC authorization check metrics (added centralized permission path)
	RBACChecksTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "rbac",
		Name:      "checks_total",
		Help:      "Total RBAC permission checks partitioned by resource, action and outcome (allowed|denied)",
	}, []string{"resource", "action", "allowed"})

	// RBAC audit soft-deny metrics (incremented when audit mode allows a request that would otherwise be denied)
	RBACAuditSoftDenies = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "rbac",
		Name:      "audit_soft_denies_total",
		Help:      "Total RBAC soft denies (audit mode) partitioned by resource and action",
	}, []string{"resource", "action"})

	RBACCheckLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "rbac",
		Name:      "check_latency_seconds",
		Help:      "Latency of RBAC permission evaluations by resource and action",
		Buckets:   prometheus.DefBuckets,
	}, []string{"resource", "action"})

	// Provider manager metrics (migration from switch -> registry)
	ProviderRegistryFallbacks = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "providers",
		Name:      "registry_fallback_total",
		Help:      "Total provider creations that required legacy fallback path (should trend to zero before switch removal)",
	}, []string{"type"})

	// Post-conversion invariants (lightweight streaming aggregator) metrics
	InvariantsRows = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "invariants",
		Name:      "rows_total",
		Help:      "Total rows observed while computing post-conversion invariants (provider, path)",
	}, []string{"provider", "path"})
	InvariantsComputeDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costscope",
		Subsystem: "invariants",
		Name:      "compute_duration_seconds",
		Help:      "Duration of streaming invariants aggregation finalize per provider/path",
		Buckets:   prometheus.DefBuckets,
	}, []string{"provider", "path"})
	InvariantsFeatureRuns = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "invariants",
		Name:      "runs_total",
		Help:      "Total post-conversion invariant computations (enabled, baseline presence)",
	}, []string{"provider", "path", "baseline"})

	// Conversion persistence metrics (JobStore operations)
	ConversionPersistenceSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "persistence_success_total",
		Help:      "Total successful persistence operations for conversion job results",
	})

	ConversionPersistenceFailure = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "costscope",
		Subsystem: "conversion",
		Name:      "persistence_failure_total",
		Help:      "Total failed persistence operations for conversion job results",
	})
)

var registerOnce sync.Once

// Register registers the telemetry metrics with Prometheus registry (global default)
func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(ReportsResolveRequests)
		prometheus.MustRegister(ReportsResolveCollisions)
		prometheus.MustRegister(ReportsResolveLatency)
		prometheus.MustRegister(ConverterRecords)
		prometheus.MustRegister(ConverterDuration)
		prometheus.MustRegister(ConversionDurationSimple)
		prometheus.MustRegister(MapperRowsTotal)
		prometheus.MustRegister(ConversionActiveJobs)
		prometheus.MustRegister(ConversionJobsSubmitted)
		prometheus.MustRegister(ConversionJobsCompleted)
		prometheus.MustRegister(ConversionJobDuration)
		prometheus.MustRegister(UnifiedMapperRows)
		prometheus.MustRegister(UnifiedMapperDuration)
		prometheus.MustRegister(UnifiedMapperErrors)
		prometheus.MustRegister(ClassifierDecisions)
		prometheus.MustRegister(ExporterBytes)
		prometheus.MustRegister(ExporterDuration)
		prometheus.MustRegister(StreamingEvents)
		prometheus.MustRegister(StreamingLag)
		prometheus.MustRegister(ParquetRotationSize)
		prometheus.MustRegister(StreamingBackpressure)
		prometheus.MustRegister(MapperLatency)
		prometheus.MustRegister(NormalizationCacheHits)
		prometheus.MustRegister(EnumCacheHits)
		prometheus.MustRegister(CacheSize)
		prometheus.MustRegister(CacheEvictions)
		prometheus.MustRegister(DriftChiSquare)
		prometheus.MustRegister(DriftPValue)
		prometheus.MustRegister(DriftTrendSlope)
		prometheus.MustRegister(DriftBucketDelta)
		prometheus.MustRegister(DriftAnalysesTotal)
		prometheus.MustRegister(DriftPercentileMax)
		prometheus.MustRegister(HealthReadiness)
		prometheus.MustRegister(IntegrationActionCalls)
		prometheus.MustRegister(IntegrationActionDuration)
		prometheus.MustRegister(IntegrationActionErrors)
		prometheus.MustRegister(AuthFailures)
		prometheus.MustRegister(AzureDiscountNormalizations)
		prometheus.MustRegister(AzureDiscountNormalizationSkips)
		prometheus.MustRegister(RBACChecksTotal)
		prometheus.MustRegister(RBACAuditSoftDenies)
		prometheus.MustRegister(RBACCheckLatency)
		prometheus.MustRegister(ProviderRegistryFallbacks)
		prometheus.MustRegister(QueryBuilderBuilds)
		prometheus.MustRegister(InvariantsRows)
		prometheus.MustRegister(InvariantsComputeDuration)
		prometheus.MustRegister(InvariantsFeatureRuns)

		// Prime commonly used labeled series so they appear in /metrics even before first hits.
		// Keep cardinality minimal (provider="any").
		NormalizationCacheHits.WithLabelValues("region", "any").Add(0)
		NormalizationCacheHits.WithLabelValues("unit", "any").Add(0)
		EnumCacheHits.WithLabelValues("currency", "any").Add(0)
		EnumCacheHits.WithLabelValues("unit", "any").Add(0)
		EnumCacheHits.WithLabelValues("region", "any").Add(0)
		// Prime unified mapper series so they are visible in /metrics prior to first conversion.
		UnifiedMapperRows.WithLabelValues("any", "legacy").Add(0)
		UnifiedMapperDuration.WithLabelValues("any", "legacy").Observe(0)
		UnifiedMapperErrors.WithLabelValues("any", "legacy").Add(0)
		// Prime reports resolve metric (format="json" as a neutral exemplar)
		ReportsResolveRequests.WithLabelValues("json").Add(0)
		ReportsResolveLatency.WithLabelValues("json", "default").Observe(0)
		// Prime query builder build metric for focus core variant
		QueryBuilderBuilds.WithLabelValues("focus").Add(0)
		// Prime duckdb + extended variants so they appear pre-use
		QueryBuilderBuilds.WithLabelValues("duckdb").Add(0)
		QueryBuilderBuilds.WithLabelValues("extended").Add(0)
		// Prime RBAC metrics with a neutral placeholder
		RBACChecksTotal.WithLabelValues("_", "_", "allowed").Add(0)
		RBACChecksTotal.WithLabelValues("_", "_", "denied").Add(0)
		RBACAuditSoftDenies.WithLabelValues("_", "_").Add(0)
		RBACCheckLatency.WithLabelValues("_", "_").Observe(0)
		ProviderRegistryFallbacks.WithLabelValues("_").Add(0)
		InvariantsRows.WithLabelValues("any", "legacy").Add(0)
		InvariantsComputeDuration.WithLabelValues("any", "legacy").Observe(0)
		InvariantsFeatureRuns.WithLabelValues("any", "legacy", "no").Add(0)
		// Default readiness = 0 until first readiness check updates it.
		HealthReadiness.Set(0)
	})
}
