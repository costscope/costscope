//go:build duckdb

package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/database"
	"local/costscope/internal/database/performance"
)

// DuckDBEngine provides a DuckDB-backed implementation of the database execution
// and analytics utilities used by the facade and higher layers.
type DuckDBEngine struct {
	db           *sql.DB
	dbPath       string
	logger       *logging.Logger
	connPool     *ConnectionPool
	cache        *QueryCache
	queryBuilder database.QueryBuilder
	// optimizer       database.FOCUSOptimizer  // TODO: Implement without circular imports
	// mlAnalytics     database.MLAnalytics     // TODO: Implement without circular imports
	performanceEngine *performance.PerformanceEngine
	mutex             sync.RWMutex
	isConnected       bool
	config            *Config
}

// Config holds DuckDB engine configuration
type Config struct {
	DatabasePath     string        `json:"database_path"`
	MaxConnections   int           `json:"max_connections"`
	CacheSize        int           `json:"cache_size"`
	QueryTimeout     time.Duration `json:"query_timeout"`
	EnableLogging    bool          `json:"enable_logging"`
	MemoryLimit      string        `json:"memory_limit"`
	ThreadCount      int           `json:"thread_count"`
	EnableExtensions bool          `json:"enable_extensions"`
	TempDirectory    string        `json:"temp_directory"`
}

// DefaultConfig returns default DuckDB configuration
func DefaultConfig() *Config {
	return &Config{
		DatabasePath:     "./data/costscope_analytics.db",
		MaxConnections:   10,
		CacheSize:        1000,
		QueryTimeout:     5 * time.Minute,
		EnableLogging:    true,
		MemoryLimit:      "2GB",
		ThreadCount:      4,
		EnableExtensions: true,
		TempDirectory:    "./tmp",
	}
}

// NewDuckDBEngine creates a new DuckDB analytics engine
func NewDuckDBEngine(config *Config) (*DuckDBEngine, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(config.DatabasePath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(config.TempDirectory, 0750); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	engine := &DuckDBEngine{
		dbPath: config.DatabasePath,
		logger: logging.NewLogger(logging.LevelInfo),
		config: config,
	}

	// Initialize components
	if err := engine.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return engine, nil
}

// initializeComponents initializes internal components
func (e *DuckDBEngine) initializeComponents() error {
	// Initialize connection pool
	pool, err := NewConnectionPool(e.config.MaxConnections, e.dbPath, e.config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	e.connPool = pool

	// Initialize query cache
	e.cache = NewQueryCache(e.config.CacheSize)

	// Initialize query builder
	e.queryBuilder = NewDuckDBQueryBuilder()

	// Initialize optimizer - TODO: Fix circular imports
	// e.optimizer = NewFOCUSOptimizer(e)

	// Initialize ML analytics - TODO: Fix circular imports
	// e.mlAnalytics = NewMLAnalytics(e)

	// Initialize performance engine
	perfConfig := &performance.PerformanceConfig{
		Enabled: true,
		Memory: &performance.MemoryConfig{
			MaxMemoryMB:     2048, // Match DuckDB memory limit
			GCThresholdMB:   1024,
			MonitorEnabled:  true,
			MonitorInterval: 30 * time.Second,
		},
		Parallel: &performance.ParallelConfig{
			WorkerCount:      e.config.ThreadCount,
			QueueSize:        1000,
			MemoryLimitMB:    1024,
			EnableMonitoring: true,
			WorkerTimeout:    e.config.QueryTimeout,
		},
		Cache: &performance.CacheConfig{
			MaxSize:         e.config.CacheSize * 2, // Larger cache for performance
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 10 * time.Minute,
			Monitor: &performance.CacheMonitorConfig{
				Enabled:       true,
				Interval:      1 * time.Minute,
				HitRateTarget: 0.8,
			},
		},
	}
	e.performanceEngine = performance.NewPerformanceEngine(perfConfig)

	return nil
}

// Connect establishes connection to DuckDB
func (e *DuckDBEngine) Connect() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.isConnected {
		return nil
	}

	// Create connection string with configuration
	connStr := e.buildConnectionString()

	db, err := sql.Open("duckdb", connStr)
	if err != nil {
		return fmt.Errorf("failed to open DuckDB connection: %w", err)
	}

	// Configure connection
	db.SetMaxOpenConns(e.config.MaxConnections)
	db.SetMaxIdleConns(e.config.MaxConnections / 2)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping DuckDB: %w", err)
	}

	e.db = db
	e.isConnected = true

	// Initialize database schema and extensions
	if err := e.initializeDatabase(ctx); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Start performance engine
	if e.performanceEngine != nil {
		if err := e.performanceEngine.Start(ctx); err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to start performance engine: %v", err))
		} else {
			e.logger.Info("Performance engine started successfully")
		}
	}

	e.logger.Info(fmt.Sprintf("Connected to DuckDB analytics engine: %s, memory=%s, threads=%d",
		e.dbPath, e.config.MemoryLimit, e.config.ThreadCount))

	return nil
}

// buildConnectionString creates DuckDB connection string with configuration
func (e *DuckDBEngine) buildConnectionString() string {
	// For file-based database
	return e.dbPath + "?access_mode=read_write"
}

// initializeDatabase sets up database configuration and extensions
func (e *DuckDBEngine) initializeDatabase(ctx context.Context) error {
	// Set memory limit
	if _, err := e.db.ExecContext(ctx, fmt.Sprintf("SET memory_limit='%s'", e.config.MemoryLimit)); err != nil {
		e.logger.Warn("Failed to set memory limit: " + err.Error())
	}

	// Set thread count
	if _, err := e.db.ExecContext(ctx, fmt.Sprintf("SET threads=%d", e.config.ThreadCount)); err != nil {
		e.logger.Warn("Failed to set thread count: " + err.Error())
	}

	// Set temp directory
	if _, err := e.db.ExecContext(ctx, fmt.Sprintf("SET temp_directory='%s'", e.config.TempDirectory)); err != nil {
		e.logger.Warn("Failed to set temp directory: " + err.Error())
	}

	// Install and load extensions if enabled
	if e.config.EnableExtensions {
		if err := e.installExtensions(ctx); err != nil {
			e.logger.Warn("Failed to install some extensions: " + err.Error())
		}
	}

	// Create FOCUS schema
	if err := e.createFOCUSSchema(ctx); err != nil {
		return fmt.Errorf("failed to create FOCUS schema: %w", err)
	}

	return nil
}

// installExtensions installs required DuckDB extensions
//
//nolint:unparam // Extension installation placeholder for future functionality
func (e *DuckDBEngine) installExtensions(ctx context.Context) error {
	extensions := []string{
		"parquet",      // For Parquet file support
		"json",         // For JSON data processing
		"httpfs",       // For HTTP file system access
		"postgres",     // For PostgreSQL compatibility
		"autocomplete", // For query autocompletion
	}

	for _, ext := range extensions {
		// Install extension
		installQuery := fmt.Sprintf("INSTALL %s", ext)
		if _, err := e.db.ExecContext(ctx, installQuery); err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to install extension %s: %s", ext, err.Error()))
			continue
		}

		// Load extension
		loadQuery := fmt.Sprintf("LOAD %s", ext)
		if _, err := e.db.ExecContext(ctx, loadQuery); err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to load extension %s: %s", ext, err.Error()))
			continue
		}

		e.logger.Info("Extension loaded successfully: " + ext)
	}

	return nil
}

// createFOCUSSchema creates the FOCUS data schema
func (e *DuckDBEngine) createFOCUSSchema(ctx context.Context) error {
	schema := `
		-- Create FOCUS cost data table with optimized schema
		CREATE TABLE IF NOT EXISTS focus_cost_data (
			-- Core FOCUS fields
			billing_period_start TIMESTAMP NOT NULL,
			billing_period_end TIMESTAMP NOT NULL,
			charge_period_start TIMESTAMP NOT NULL,
			charge_period_end TIMESTAMP NOT NULL,
			
			-- Provider information
			provider_id VARCHAR(100) NOT NULL,
			publisher_id VARCHAR(100),
			publisher_name VARCHAR(200),
			
			-- Account information  
			billing_account_id VARCHAR(100) NOT NULL,
			billing_account_name VARCHAR(200),
			sub_account_id VARCHAR(100),
			sub_account_name VARCHAR(200),
			
			-- Service information
			service_category VARCHAR(100),
			service_name VARCHAR(200) NOT NULL,
			service_id VARCHAR(100),
			
			-- Resource information
			resource_id VARCHAR(500),
			resource_name VARCHAR(500),
			resource_type VARCHAR(100),
			
			-- Location information
			availability_zone VARCHAR(100),
			region VARCHAR(100),
			
			-- Cost information
			billing_currency VARCHAR(10) NOT NULL,
			effective_cost DECIMAL(20,8) NOT NULL,
			list_cost DECIMAL(20,8),
			pricing_unit VARCHAR(50),
			pricing_quantity DECIMAL(20,8),
			
			-- Usage information
			usage_unit VARCHAR(50),
			usage_quantity DECIMAL(20,8),
			
			-- Metadata
			invoice_issuer_id VARCHAR(100),
			tags JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Create optimized indexes for performance
		CREATE INDEX IF NOT EXISTS idx_focus_billing_period ON focus_cost_data(billing_period_start, billing_period_end);
		CREATE INDEX IF NOT EXISTS idx_focus_charge_period ON focus_cost_data(charge_period_start, charge_period_end);
		CREATE INDEX IF NOT EXISTS idx_focus_provider ON focus_cost_data(provider_id);
		CREATE INDEX IF NOT EXISTS idx_focus_service ON focus_cost_data(service_name);
		CREATE INDEX IF NOT EXISTS idx_focus_account ON focus_cost_data(billing_account_id);
		CREATE INDEX IF NOT EXISTS idx_focus_cost ON focus_cost_data(effective_cost);
		CREATE INDEX IF NOT EXISTS idx_focus_region ON focus_cost_data(region);
		CREATE INDEX IF NOT EXISTS idx_focus_resource_type ON focus_cost_data(resource_type);

		-- Create aggregated views for performance
		CREATE VIEW IF NOT EXISTS focus_daily_costs AS
		SELECT 
			DATE_TRUNC('day', charge_period_start) as date,
			provider_id,
			service_name,
			region,
			billing_currency,
			SUM(effective_cost) as daily_cost,
			COUNT(*) as record_count
		FROM focus_cost_data
		GROUP BY 1, 2, 3, 4, 5;

		CREATE VIEW IF NOT EXISTS focus_monthly_costs AS
		SELECT 
			DATE_TRUNC('month', charge_period_start) as month,
			provider_id,
			service_name,
			billing_currency,
			SUM(effective_cost) as monthly_cost,
			COUNT(*) as record_count
		FROM focus_cost_data
		GROUP BY 1, 2, 3, 4;

		-- Create materialized view for top services (requires periodic refresh)
		CREATE TABLE IF NOT EXISTS focus_top_services AS
		SELECT 
			service_name,
			provider_id,
			SUM(effective_cost) as total_cost,
			COUNT(*) as record_count,
			AVG(effective_cost) as avg_cost,
			MIN(charge_period_start) as first_seen,
			MAX(charge_period_start) as last_seen,
			CURRENT_TIMESTAMP as last_updated
		FROM focus_cost_data
		GROUP BY service_name, provider_id
		ORDER BY total_cost DESC;
	`

	_, err := e.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create FOCUS schema: %w", err)
	}

	e.logger.Info("FOCUS schema created successfully")
	return nil
}

// Close closes the database connection
func (e *DuckDBEngine) Close() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.isConnected || e.db == nil {
		return nil
	}

	// Close connection pool
	if e.connPool != nil {
		e.connPool.Close()
	}

	// Clear cache
	if e.cache != nil {
		e.cache.Clear()
	}

	// Stop performance engine
	if e.performanceEngine != nil {
		if err := e.performanceEngine.Stop(); err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to stop performance engine: %v", err))
		} else {
			e.logger.Info("Performance engine stopped successfully")
		}
	}

	// Close main connection
	if err := e.db.Close(); err != nil {
		return fmt.Errorf("failed to close DuckDB connection: %w", err)
	}

	e.isConnected = false
	e.logger.Info("DuckDB analytics engine closed")
	return nil
}

// Health checks the health of the database connection
func (e *DuckDBEngine) Health(ctx context.Context) error {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if !e.isConnected || e.db == nil {
		return fmt.Errorf("database not connected")
	}

	// Test with a simple query
	var result int
	err := e.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}

// ExecuteQuery executes a SQL query and returns results
func (e *DuckDBEngine) ExecuteQuery(ctx context.Context, query string) (*database.QueryResult, error) {
	return e.ExecuteQueryWithParams(ctx, query, nil)
}

// ExecuteQueryWithParams executes a SQL query with parameters
func (e *DuckDBEngine) ExecuteQueryWithParams(ctx context.Context, query string, params map[string]interface{}) (*database.QueryResult, error) {
	startTime := time.Now()

	// Generate query hash for caching
	queryHash := e.cache.generateQueryHash(query, params)

	// Check enhanced cache first (performance engine)
	if e.performanceEngine != nil && e.performanceEngine.IsEnabled() {
		if cachedResult, found := e.performanceEngine.CacheGet(queryHash); found {
			if result, ok := cachedResult.(*database.QueryResult); ok {
				result.Metadata.CacheHit = true
				result.Metadata.ExecutionTime = time.Since(startTime)
				e.logger.Debug("Query result served from performance cache")
				return result, nil
			}
		}
	}

	// Check legacy cache
	if cached, found := e.cache.Get(queryHash); found {
		if result, ok := cached.(*database.QueryResult); ok {
			result.Metadata.CacheHit = true
			result.Metadata.ExecutionTime = time.Since(startTime)
			return result, nil
		}
	}

	// For complex queries, use parallel processing if available
	if e.performanceEngine != nil && e.performanceEngine.IsEnabled() && e.isComplexQuery(query) {
		result, err := e.executeQueryWithPerformanceEngine(ctx, query, params, queryHash, startTime)
		if err == nil {
			return result, nil
		}
		// Fallback to regular execution on error
		e.logger.Warn(fmt.Sprintf("Performance engine execution failed, falling back: %v", err))
	}

	// Execute query normally
	result, err := e.executeQueryStandard(ctx, query, params, queryHash, startTime)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// isComplexQuery determines if a query is complex enough to benefit from parallel processing
func (e *DuckDBEngine) isComplexQuery(query string) bool {
	// Simple heuristics - check for JOINs, aggregations, subqueries
	complexKeywords := []string{"JOIN", "GROUP BY", "ORDER BY", "HAVING", "WITH", "UNION", "SUBQUERY"}
	queryUpper := strings.ToUpper(query)

	for _, keyword := range complexKeywords {
		if strings.Contains(queryUpper, keyword) {
			return true
		}
	}

	// Check query length - longer queries are often more complex
	return len(query) > 500
}

// executeQueryWithPerformanceEngine executes query using performance optimizations
func (e *DuckDBEngine) executeQueryWithPerformanceEngine(ctx context.Context, query string, params map[string]interface{}, queryHash string, startTime time.Time) (*database.QueryResult, error) {
	// Monitor memory usage during execution
	if e.performanceEngine != nil {
		_ = e.performanceEngine.OptimizeMemory(ctx)
	}

	// Create a job for parallel execution
	job := performance.Job{
		ID: queryHash,
		Data: map[string]interface{}{
			"query":  query,
			"params": params,
		},
		Priority: performance.PriorityNormal,
		Processor: func(data interface{}) (interface{}, error) {
			return e.executeQueryStandard(ctx, query, params, queryHash, startTime)
		},
		Timeout: 30 * time.Second,
	}

	results, err := e.performanceEngine.ExecuteParallel(ctx, []performance.Job{job})
	if err != nil {
		return nil, fmt.Errorf("parallel execution failed: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results from parallel execution")
	}

	result := results[0]
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert result to QueryResult
	if queryResult, ok := result.Result.(*database.QueryResult); ok {
		// Cache in performance engine
		_ = e.performanceEngine.CacheSet(queryHash, queryResult)
		return queryResult, nil
	}

	return nil, fmt.Errorf("unexpected result type from parallel execution")
}

// executeQueryStandard executes query using standard DuckDB processing
func (e *DuckDBEngine) executeQueryStandard(ctx context.Context, query string, params map[string]interface{}, queryHash string, startTime time.Time) (*database.QueryResult, error) {
	// Log execution details when parameters are provided
	if len(params) > 0 && e.logger != nil {
		e.logger.Debug("Executing query with parameters")
	}

	// Use queryHash and startTime for potential future metrics
	_ = queryHash
	_ = startTime

	// Execute query
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Process results
	result, err := e.processQueryRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process query results: %w", err)
	}

	// Set metadata
	result.Metadata = database.QueryMetadata{
		ExecutionTime: time.Since(startTime),
		QueryHash:     queryHash,
		CacheHit:      false,
	}

	// Cache result in both caches
	e.cache.Set(queryHash, result, 5*time.Minute)
	if e.performanceEngine != nil && e.performanceEngine.IsEnabled() {
		_ = e.performanceEngine.CacheSet(queryHash, result)
	}

	return result, nil
}

// processQueryRows processes SQL query rows into QueryResult
func (e *DuckDBEngine) processQueryRows(rows *sql.Rows) (*database.QueryResult, error) {
	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Process rows
	var data []map[string]interface{}
	for rows.Next() {
		// Create value containers
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to map
		row := make(map[string]interface{})
		for i, col := range columns {
			value := values[i]
			// Handle byte arrays (common with DuckDB)
			if b, ok := value.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = value
			}
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return &database.QueryResult{
		Columns: columns,
		Data:    data,
		Count:   len(data),
	}, nil
}

// LoadFOCUSData loads FOCUS data from a Parquet file
func (e *DuckDBEngine) LoadFOCUSData(ctx context.Context, filePath string) error {
	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("FOCUS file does not exist: %s", filePath)
	}

	// Validate file path to prevent injection
	if !filepath.IsAbs(filePath) {
		return fmt.Errorf("file path must be absolute: %s", filePath)
	}

	// Create load query using parameterized approach (DuckDB supports this for file paths)
	query := `
		INSERT INTO focus_cost_data 
		SELECT * FROM read_parquet(?)
		WHERE effective_cost IS NOT NULL 
		AND billing_period_start IS NOT NULL
		AND service_name IS NOT NULL
	`

	startTime := time.Now()
	result, err := e.db.ExecContext(ctx, query, filePath)
	if err != nil {
		return fmt.Errorf("failed to load FOCUS data: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	loadTime := time.Since(startTime)

	e.logger.Info(fmt.Sprintf("FOCUS data loaded successfully: file=%s, rows=%d, time=%s",
		filePath, rowsAffected, loadTime.String()))

	// Refresh materialized views
	go e.refreshMaterializedViews(ctx)

	return nil
}

// refreshMaterializedViews refreshes materialized views in background
func (e *DuckDBEngine) refreshMaterializedViews(ctx context.Context) {
	query := `
		DROP TABLE IF EXISTS focus_top_services;
		CREATE TABLE focus_top_services AS
		SELECT 
			service_name,
			provider_id,
			SUM(effective_cost) as total_cost,
			COUNT(*) as record_count,
			AVG(effective_cost) as avg_cost,
			MIN(charge_period_start) as first_seen,
			MAX(charge_period_start) as last_seen,
			CURRENT_TIMESTAMP as last_updated
		FROM focus_cost_data
		GROUP BY service_name, provider_id
		ORDER BY total_cost DESC;
	`

	_, err := e.db.ExecContext(ctx, query)
	if err != nil {
		e.logger.Error("Failed to refresh materialized views: " + err.Error())
	} else {
		e.logger.Info("Materialized views refreshed successfully")
	}
}

// CreateFOCUSTable creates a FOCUS table with specified schema
func (e *DuckDBEngine) CreateFOCUSTable(ctx context.Context, tableName string, schema database.FOCUSSchema) error {
	// Generate CREATE TABLE statement from schema
	query, err := e.generateCreateTableQuery(tableName, schema)
	if err != nil {
		return fmt.Errorf("failed to generate CREATE TABLE query: %w", err)
	}

	_, err = e.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create FOCUS table: %w", err)
	}

	e.logger.Info(fmt.Sprintf("FOCUS table created successfully: %s (v%s)", tableName, schema.Version))

	return nil
}

// generateCreateTableQuery generates CREATE TABLE SQL from schema
//
//nolint:unparam // Future error handling in schema generation
func (e *DuckDBEngine) generateCreateTableQuery(tableName string, schema database.FOCUSSchema) (string, error) {
	// This is a simplified implementation
	// In production, you'd want more sophisticated schema generation
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName)

	// Validate schema information for future use
	if schema.Version == "" && e.logger != nil {
		e.logger.Warn("Schema version not specified, using default")
	}

	// Add columns (simplified for this example)
	query += `
		billing_period_start TIMESTAMP NOT NULL,
		billing_period_end TIMESTAMP NOT NULL,
		effective_cost DECIMAL(20,8) NOT NULL,
		service_name VARCHAR(200) NOT NULL,
		provider_id VARCHAR(100) NOT NULL
	`

	query += ")"
	return query, nil
}

// GetPerformanceStats returns database performance statistics
func (e *DuckDBEngine) GetPerformanceStats(ctx context.Context) (*database.PerformanceStats, error) {
	stats := &database.PerformanceStats{
		CacheHitRate: e.cache.GetHitRate(),
	}

	// Get connection pool stats
	if poolStats := e.connPool.GetStats(); poolStats != nil {
		stats.ConnectionsActive = poolStats.ActiveConnections
	}

	// Get database size information
	var dbSize int64
	err := e.db.QueryRowContext(ctx, "SELECT database_size FROM duckdb_settings WHERE name='database_size'").Scan(&dbSize)
	if err == nil {
		stats.DiskUsage = dbSize
	}

	return stats, nil
}

// GetDatabasePath returns the database file path
func (e *DuckDBEngine) GetDatabasePath() string {
	return e.dbPath
}
