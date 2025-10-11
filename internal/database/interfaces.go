package database

import (
	"context"
	"time"
)

// QueryExecutor is the minimal execution surface consumed by the analytics facade and DuckDB engine.
// Note: Prefer this over the broader AnalyticsEngine for new code paths.
type QueryExecutor interface {
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)
	ExecuteQueryWithParams(ctx context.Context, query string, params map[string]interface{}) (*QueryResult, error)
}

// Note: The previous broad AnalyticsEngine interface has been removed.
// New code should depend on small, purpose-built contracts like QueryExecutor
// and compose higher-level behavior via facades.

// QueryBuilder defines the core query surface always present in standard builds.
// Less frequently used operations (CTE/JOIN/HAVING, COUNT/EXPLAIN variants) are
// moved to an optional extended interface under build tag `qb_extended` to keep
// the default dependency surface minimal while preserving forward compatibility.
type QueryBuilder interface {
	// Basic operations
	Select(columns ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(condition string, args ...interface{}) QueryBuilder
	GroupBy(columns ...string) QueryBuilder
	OrderBy(column string, direction SortDirection) QueryBuilder
	Limit(count int) QueryBuilder
	Offset(count int) QueryBuilder

	// FOCUS-specific operations
	SelectCostMetrics() QueryBuilder
	FilterByProvider(provider string) QueryBuilder
	FilterByDateRange(start, end time.Time) QueryBuilder
	FilterByCostThreshold(threshold float64) QueryBuilder
	FilterByService(services ...string) QueryBuilder
	FilterByRegion(regions ...string) QueryBuilder
	FilterByAccount(accounts ...string) QueryBuilder
	FilterByTenant(tenantID string) QueryBuilder

	// Build
	Build() (string, []interface{}, error)
}

// FOCUSOptimizer defines the interface for FOCUS data optimization
type FOCUSOptimizer interface {
	// Schema optimization
	OptimizeSchema(ctx context.Context, schema FOCUSSchema) (FOCUSSchema, error)
	CreateOptimizedIndexes(ctx context.Context, tableName string) error

	// Query optimization
	OptimizeAggregationQuery(ctx context.Context, query string) (string, error)
	OptimizeJoinQuery(ctx context.Context, query string) (string, error)

	// Performance tuning
	AnalyzeTablePerformance(ctx context.Context, tableName string) (*TablePerformance, error)
	SuggestOptimizations(ctx context.Context, queryStats *QueryStats) ([]*OptimizationSuggestion, error)
}

// MLAnalytics defines the interface for machine learning analytics
type MLAnalytics interface {
	// Anomaly detection
	DetectCostAnomalies(ctx context.Context, data []*CostData) ([]*Anomaly, error)
	TrainAnomalyModel(ctx context.Context, trainingData []*CostData) error

	// Forecasting
	ForecastCosts(ctx context.Context, historicalData []*CostData, periods int) ([]*ForecastPoint, error)
	TrainForecastModel(ctx context.Context, trainingData []*CostData) error

	// Clustering
	ClusterServices(ctx context.Context, data []*ServiceMetrics) ([]*ServiceCluster, error)

	// Recommendations
	GenerateOptimizationRecommendations(ctx context.Context, data []*CostData) ([]*Recommendation, error)
}

// CacheManager defines the interface for query result caching
type CacheManager interface {
	// Basic cache operations
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error

	// Query-specific caching
	GetQueryResult(ctx context.Context, queryHash string) (*QueryResult, bool)
	SetQueryResult(ctx context.Context, queryHash string, result *QueryResult, ttl time.Duration) error

	// Cache management
	GetStats(ctx context.Context) (*CacheStats, error)
	Evict(ctx context.Context, pattern string) error
}

// ConnectionManager defines the interface for database connection management
type ConnectionManager interface {
	// Connection lifecycle
	GetConnection(ctx context.Context) (Connection, error)
	ReturnConnection(conn Connection) error
	CloseAll() error

	// Pool management
	GetPoolStats() *PoolStats
	ResizePool(size int) error

	// Health monitoring
	CheckConnections(ctx context.Context) error
}

// Connection defines the interface for individual database connections
type Connection interface {
	// Query execution
	Query(ctx context.Context, sql string, args ...interface{}) (*QueryResult, error)
	Exec(ctx context.Context, sql string, args ...interface{}) error

	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)

	// Connection state
	IsValid(ctx context.Context) bool
	Close() error
}

// Transaction defines the interface for database transactions
type Transaction interface {
	// Transaction operations
	Query(ctx context.Context, sql string, args ...interface{}) (*QueryResult, error)
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Commit() error
	Rollback() error
}
