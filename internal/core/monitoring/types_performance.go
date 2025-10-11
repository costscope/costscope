package monitoring

import "time"

// PerformanceMetrics captures summarized performance state used by the service and APIs
type PerformanceMetrics struct {
	CPU              CPUMetrics            `json:"cpu"`
	Memory           MemoryMetrics         `json:"memory"`
	Disk             DiskMetrics           `json:"disk"`
	Network          NetworkMetrics        `json:"network"`
	Application      AppPerformanceMetrics `json:"application"`
	Database         DatabaseMetrics       `json:"database"`
	API              APIMetrics            `json:"api"`
	Throughput       ThroughputMetrics     `json:"throughput"`
	Latency          LatencyMetrics        `json:"latency"`
	Concurrency      ConcurrencyMetrics    `json:"concurrency"`
	PerformanceScore int                   `json:"performance_score"`
	Grade            string                `json:"grade"`
}

// PerformanceTrends provides aggregate trend analysis over a time window
type PerformanceTrends struct {
	TimeRange          time.Duration `json:"time_range"`
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	CPUTrend           TrendData     `json:"cpu_trend"`
	MemoryTrend        TrendData     `json:"memory_trend"`
	LatencyTrend       TrendData     `json:"latency_trend"`
	ThroughputTrend    TrendData     `json:"throughput_trend"`
	ErrorRateTrend     TrendData     `json:"error_rate_trend"`
	PerformanceSummary string        `json:"performance_summary"`
	Recommendations    []string      `json:"recommendations"`
	Anomalies          []Anomaly     `json:"anomalies"`
}

// AppPerformanceMetrics details app-layer perf signals (moved from types.go for modularity)
type AppPerformanceMetrics struct {
	RequestCount   int64          `json:"request_count"`
	SuccessRate    float64        `json:"success_rate"`
	ErrorRate      float64        `json:"error_rate"`
	ResponseTime   LatencyMetrics `json:"response_time"`
	ThroughputRPS  float64        `json:"throughput_rps"`
	MemoryUsage    float64        `json:"memory_usage"`
	GoroutineCount int            `json:"goroutine_count"`
	GCPauseTime    float64        `json:"gc_pause_time"`
}

// DatabaseMetrics captures database system performance stats
type DatabaseMetrics struct {
	ConnectionCount   int     `json:"connection_count"`
	ActiveQueries     int     `json:"active_queries"`
	SlowQueries       int     `json:"slow_queries"`
	QueryResponseTime float64 `json:"query_response_time"`
	LockWaitTime      float64 `json:"lock_wait_time"`
	TransactionRate   float64 `json:"transaction_rate"`
	CacheHitRate      float64 `json:"cache_hit_rate"`
	TableSize         float64 `json:"table_size"`
	IndexUsage        float64 `json:"index_usage"`
}

// APIMetrics aggregates endpoint-level stats
type APIMetrics struct {
	TotalRequests      int64                      `json:"total_requests"`
	SuccessfulRequests int64                      `json:"successful_requests"`
	FailedRequests     int64                      `json:"failed_requests"`
	AverageLatency     float64                    `json:"average_latency"`
	P95Latency         float64                    `json:"p95_latency"`
	P99Latency         float64                    `json:"p99_latency"`
	RateLimitHits      int64                      `json:"rate_limit_hits"`
	EndpointMetrics    map[string]EndpointMetrics `json:"endpoint_metrics"`
}

// EndpointMetrics defines HTTP endpoint-level stats
type EndpointMetrics struct {
	Path           string  `json:"path"`
	Method         string  `json:"method"`
	RequestCount   int64   `json:"request_count"`
	ErrorCount     int64   `json:"error_count"`
	AverageLatency float64 `json:"average_latency"`
	P95Latency     float64 `json:"p95_latency"`
	P99Latency     float64 `json:"p99_latency"`
	ThroughputRPS  float64 `json:"throughput_rps"`
}

// ThroughputMetrics captures rates across subsystems
type ThroughputMetrics struct {
	RequestsPerSecond     float64 `json:"requests_per_second"`
	BytesPerSecond        float64 `json:"bytes_per_second"`
	TransactionsPerSecond float64 `json:"transactions_per_second"`
	MessagesPerSecond     float64 `json:"messages_per_second"`
	Peak                  float64 `json:"peak"`
	Average               float64 `json:"average"`
	Minimum               float64 `json:"minimum"`
}

// LatencyMetrics provides latency distribution percentiles
type LatencyMetrics struct {
	Average           float64 `json:"average"`
	Median            float64 `json:"median"`
	P95               float64 `json:"p95"`
	P99               float64 `json:"p99"`
	P999              float64 `json:"p999"`
	Min               float64 `json:"min"`
	Max               float64 `json:"max"`
	StandardDeviation float64 `json:"standard_deviation"`
}

// ConcurrencyMetrics reflects system concurrency pressure
type ConcurrencyMetrics struct {
	ActiveConnections     int     `json:"active_connections"`
	MaxConnections        int     `json:"max_connections"`
	QueueLength           int     `json:"queue_length"`
	ThreadPoolUtilization float64 `json:"thread_pool_utilization"`
	WorkerUtilization     float64 `json:"worker_utilization"`
	ConcurrencyLevel      float64 `json:"concurrency_level"`
}
