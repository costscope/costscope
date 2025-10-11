package database

import (
	"time"
)

// Core data types for DuckDB analytics engine

// QueryResult represents the result of a database query
type QueryResult struct {
	Columns  []string                 `json:"columns"`
	Data     []map[string]interface{} `json:"data"`
	Count    int                      `json:"count"`
	Metadata QueryMetadata            `json:"metadata"`
}

// QueryMetadata contains additional information about query execution
type QueryMetadata struct {
	ExecutionTime time.Duration `json:"execution_time"`
	RowsAffected  int64         `json:"rows_affected"`
	QueryHash     string        `json:"query_hash"`
	CacheHit      bool          `json:"cache_hit"`
}

// FOCUSSchema represents FOCUS data schema structure
type FOCUSSchema struct {
	Version string        `json:"version"`
	Tables  []FOCUSTable  `json:"tables"`
	Indexes []FOCUSIndex  `json:"indexes"`
	Options SchemaOptions `json:"options"`
}

// FOCUSTable represents a table in FOCUS schema
type FOCUSTable struct {
	Name        string        `json:"name"`
	Columns     []FOCUSColumn `json:"columns"`
	Partitions  []string      `json:"partitions"`
	Compression string        `json:"compression"`
	Indexes     []string      `json:"indexes"`
}

// FOCUSColumn represents a column in FOCUS table
type FOCUSColumn struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Nullable    bool        `json:"nullable"`
	Description string      `json:"description"`
	Constraints []string    `json:"constraints"`
	Default     interface{} `json:"default"`
}

// FOCUSIndex represents an index in FOCUS schema
type FOCUSIndex struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
	Unique  bool     `json:"unique"`
}

// SchemaOptions contains schema configuration options
type SchemaOptions struct {
	CompressionLevel  int    `json:"compression_level"`
	PartitionStrategy string `json:"partition_strategy"`
	IndexingStrategy  string `json:"indexing_strategy"`
	OptimizeForReads  bool   `json:"optimize_for_reads"`
	EnableStatistics  bool   `json:"enable_statistics"`
}

// Analytics types

// AnalyticsFilters defines filters for analytics queries
type AnalyticsFilters struct {
	StartDate   *time.Time        `json:"start_date"`
	EndDate     *time.Time        `json:"end_date"`
	Providers   []string          `json:"providers"`
	Services    []string          `json:"services"`
	Regions     []string          `json:"regions"`
	MinCost     *float64          `json:"min_cost"`
	MaxCost     *float64          `json:"max_cost"`
	Tags        map[string]string `json:"tags"`
	Accounts    []string          `json:"accounts"`
	ResourceIDs []string          `json:"resource_ids"`
	// TenantID optional; applied only when multi-tenant mode enabled. Empty => no tenant scoping enforced by this filter (caller/middleware may inject).
	TenantID string `json:"tenant_id,omitempty"`
}

// CostSummary represents aggregated cost information
type CostSummary struct {
	TotalCost   float64     `json:"total_cost"`
	Currency    string      `json:"currency"`
	Period      TimePeriod  `json:"period"`
	RecordCount int64       `json:"record_count"`
	AverageCost float64     `json:"average_cost"`
	MedianCost  float64     `json:"median_cost"`
	MinCost     float64     `json:"min_cost"`
	MaxCost     float64     `json:"max_cost"`
	StandardDev float64     `json:"standard_deviation"`
	Percentiles Percentiles `json:"percentiles"`
}

// ServiceCost represents cost data for a specific service
type ServiceCost struct {
	ServiceName string    `json:"service_name"`
	Provider    string    `json:"provider"`
	TotalCost   float64   `json:"total_cost"`
	Currency    string    `json:"currency"`
	RecordCount int64     `json:"record_count"`
	AverageCost float64   `json:"average_cost"`
	CostTrend   float64   `json:"cost_trend"` // Percentage change
	LastUpdated time.Time `json:"last_updated"`
}

// TrendData represents cost trend information
type TrendData struct {
	Timestamp   time.Time              `json:"timestamp"`
	Value       float64                `json:"value"`
	Granularity TimeGranularity        `json:"granularity"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TimeGranularity defines the granularity for time-based analytics
type TimeGranularity string

const (
	TimeGranularityHour  TimeGranularity = "hour"
	TimeGranularityDay   TimeGranularity = "day"
	TimeGranularityWeek  TimeGranularity = "week"
	TimeGranularityMonth TimeGranularity = "month"
	TimeGranularityYear  TimeGranularity = "year"
)

// TimePeriod represents a time period
type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Percentiles represents statistical percentiles
type Percentiles struct {
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

// Anomaly represents detected cost anomaly
type Anomaly struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Service      string                 `json:"service"`
	Provider     string                 `json:"provider"`
	ActualCost   float64                `json:"actual_cost"`
	ExpectedCost float64                `json:"expected_cost"`
	Deviation    float64                `json:"deviation"`
	Severity     AnomalySeverity        `json:"severity"`
	Confidence   float64                `json:"confidence"`
	Description  string                 `json:"description"`
	Context      map[string]interface{} `json:"context"`
}

// AnomalySeverity defines anomaly severity levels
type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "low"
	AnomalySeverityMedium   AnomalySeverity = "medium"
	AnomalySeverityHigh     AnomalySeverity = "high"
	AnomalySeverityCritical AnomalySeverity = "critical"
)

// ML Analytics types

// PredictiveConfig defines configuration for predictive analysis
type PredictiveConfig struct {
	Model          string                 `json:"model"`
	Parameters     map[string]interface{} `json:"parameters"`
	TrainingPeriod int                    `json:"training_period_days"`
	PredictionDays int                    `json:"prediction_days"`
	Confidence     float64                `json:"confidence_threshold"`
	Features       []string               `json:"features"`
}

// PredictionResult represents the result of predictive analysis
type PredictionResult struct {
	Predictions []PredictionPoint `json:"predictions"`
	Confidence  float64           `json:"confidence"`
	Model       string            `json:"model"`
	Features    []string          `json:"features"`
	Accuracy    float64           `json:"accuracy"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// PredictionPoint represents a single prediction point
type PredictionPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
	Confidence float64   `json:"confidence"`
}

// ForecastConfig defines configuration for cost forecasting
type ForecastConfig struct {
	Method          string                 `json:"method"`
	Horizon         int                    `json:"horizon_days"`
	Seasonality     bool                   `json:"include_seasonality"`
	Trends          bool                   `json:"include_trends"`
	ExternalFactors []string               `json:"external_factors"`
	Parameters      map[string]interface{} `json:"parameters"`
}

// ForecastResult represents the result of cost forecasting
type ForecastResult struct {
	Forecasts   []ForecastPoint `json:"forecasts"`
	Method      string          `json:"method"`
	Accuracy    float64         `json:"accuracy"`
	Seasonality bool            `json:"has_seasonality"`
	Trend       float64         `json:"trend"`
	GeneratedAt time.Time       `json:"generated_at"`
}

// ForecastPoint represents a single forecast point
type ForecastPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
	Confidence float64   `json:"confidence"`
}

// Performance types

// QueryPlan represents database query execution plan
type QueryPlan struct {
	Query         string          `json:"query"`
	EstimatedCost float64         `json:"estimated_cost"`
	EstimatedRows int64           `json:"estimated_rows"`
	Steps         []QueryPlanStep `json:"steps"`
	Indexes       []string        `json:"indexes_used"`
	Optimizations []string        `json:"optimizations"`
}

// QueryPlanStep represents a step in query execution plan
type QueryPlanStep struct {
	Operation    string  `json:"operation"`
	Table        string  `json:"table"`
	Cost         float64 `json:"cost"`
	Rows         int64   `json:"rows"`
	TimeEstimate float64 `json:"time_estimate_ms"`
}

// PerformanceStats represents database performance statistics
type PerformanceStats struct {
	QueriesExecuted   int64         `json:"queries_executed"`
	AverageQueryTime  time.Duration `json:"average_query_time"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
	ConnectionsActive int           `json:"connections_active"`
	MemoryUsage       int64         `json:"memory_usage_bytes"`
	DiskUsage         int64         `json:"disk_usage_bytes"`
	TableStats        []TableStats  `json:"table_stats"`
	TopSlowQueries    []SlowQuery   `json:"top_slow_queries"`
}

// TableStats represents statistics for a database table
type TableStats struct {
	TableName        string    `json:"table_name"`
	RowCount         int64     `json:"row_count"`
	SizeBytes        int64     `json:"size_bytes"`
	LastAnalyzed     time.Time `json:"last_analyzed"`
	IndexCount       int       `json:"index_count"`
	CompressionRatio float64   `json:"compression_ratio"`
}

// SlowQuery represents a slow query log entry
type SlowQuery struct {
	Query         string        `json:"query"`
	ExecutionTime time.Duration `json:"execution_time"`
	RowsReturned  int64         `json:"rows_returned"`
	Timestamp     time.Time     `json:"timestamp"`
	UserID        string        `json:"user_id"`
}

// Query Builder types

// SortDirection defines sort direction for queries
type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

// Optimization types

// TablePerformance represents performance metrics for a table
type TablePerformance struct {
	TableName       string        `json:"table_name"`
	QueryCount      int64         `json:"query_count"`
	AverageTime     time.Duration `json:"average_query_time"`
	SlowestQuery    time.Duration `json:"slowest_query"`
	IndexEfficiency float64       `json:"index_efficiency"`
	Suggestions     []string      `json:"suggestions"`
}

// QueryStats represents statistics for query analysis
type QueryStats struct {
	QueryHash      string        `json:"query_hash"`
	ExecutionCount int64         `json:"execution_count"`
	AverageTime    time.Duration `json:"average_time"`
	TotalTime      time.Duration `json:"total_time"`
	MaxTime        time.Duration `json:"max_time"`
	MinTime        time.Duration `json:"min_time"`
	LastExecuted   time.Time     `json:"last_executed"`
	TablesAccessed []string      `json:"tables_accessed"`
}

// OptimizationSuggestion represents a query optimization suggestion
type OptimizationSuggestion struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Effort      string  `json:"effort"`
	Confidence  float64 `json:"confidence"`
	Query       string  `json:"optimized_query"`
}

// Cache types

// CacheStats represents cache performance statistics
type CacheStats struct {
	HitCount   int64   `json:"hit_count"`
	MissCount  int64   `json:"miss_count"`
	HitRate    float64 `json:"hit_rate"`
	Size       int64   `json:"size_bytes"`
	EntryCount int     `json:"entry_count"`
	MaxSize    int64   `json:"max_size_bytes"`
	Evictions  int64   `json:"evictions"`
}

// Connection Pool types

// PoolStats represents connection pool statistics
type PoolStats struct {
	MaxConnections     int           `json:"max_connections"`
	ActiveConnections  int           `json:"active_connections"`
	IdleConnections    int           `json:"idle_connections"`
	WaitingRequests    int           `json:"waiting_requests"`
	TotalConnections   int64         `json:"total_connections_created"`
	ConnectionFailures int64         `json:"connection_failures"`
	AverageWaitTime    time.Duration `json:"average_wait_time"`
}

// ML-specific types

// CostData represents cost data for ML analysis
type CostData struct {
	Timestamp  time.Time              `json:"timestamp"`
	Provider   string                 `json:"provider"`
	Service    string                 `json:"service"`
	Region     string                 `json:"region"`
	Cost       float64                `json:"cost"`
	Currency   string                 `json:"currency"`
	ResourceID string                 `json:"resource_id"`
	Tags       map[string]string      `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ServiceMetrics represents metrics for service clustering
type ServiceMetrics struct {
	ServiceName  string    `json:"service_name"`
	Provider     string    `json:"provider"`
	TotalCost    float64   `json:"total_cost"`
	UsagePattern []float64 `json:"usage_pattern"`
	Seasonality  float64   `json:"seasonality"`
	Volatility   float64   `json:"volatility"`
	GrowthRate   float64   `json:"growth_rate"`
}

// ServiceCluster represents a cluster of similar services
type ServiceCluster struct {
	ClusterID       string         `json:"cluster_id"`
	Services        []string       `json:"services"`
	Centroid        ServiceMetrics `json:"centroid"`
	Characteristics []string       `json:"characteristics"`
	Confidence      float64        `json:"confidence"`
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      float64   `json:"estimated_savings"`
	Effort      string    `json:"implementation_effort"`
	Priority    string    `json:"priority"`
	Confidence  float64   `json:"confidence"`
	Actions     []string  `json:"recommended_actions"`
	CreatedAt   time.Time `json:"created_at"`
}
