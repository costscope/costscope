//go:build duckdb

package duckdb

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ConnectionPool manages DuckDB connections
type ConnectionPool struct {
	connections chan *sql.DB
	maxSize     int
	currentSize int
	connStr     string
	config      *Config
	mutex       sync.Mutex
	stats       *PoolStats
}

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

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxSize int, connStr string, config *Config) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		connections: make(chan *sql.DB, maxSize),
		maxSize:     maxSize,
		connStr:     connStr,
		config:      config,
		stats: &PoolStats{
			MaxConnections: maxSize,
		},
	}
	return pool, nil
}

// GetConnection gets a connection from the pool
func (cp *ConnectionPool) GetConnection() (*sql.DB, error) {
	start := time.Now()

	select {
	case conn := <-cp.connections:
		cp.updateWaitTime(time.Since(start))
		cp.stats.IdleConnections--
		cp.stats.ActiveConnections++
		return conn, nil
	default:
		// Create new connection if pool is empty and under limit
		cp.mutex.Lock()
		if cp.currentSize < cp.maxSize {
			conn, err := cp.createConnection()
			if err != nil {
				cp.stats.ConnectionFailures++
				cp.mutex.Unlock()
				return nil, err
			}
			cp.currentSize++
			cp.stats.TotalConnections++
			cp.stats.ActiveConnections++
			cp.mutex.Unlock()
			cp.updateWaitTime(time.Since(start))
			return conn, nil
		}
		cp.mutex.Unlock()

		// Wait for a connection to become available
		cp.stats.WaitingRequests++
		conn := <-cp.connections
		cp.stats.WaitingRequests--
		cp.stats.IdleConnections--
		cp.stats.ActiveConnections++
		cp.updateWaitTime(time.Since(start))
		return conn, nil
	}
}

// ReturnConnection returns a connection to the pool
func (cp *ConnectionPool) ReturnConnection(conn *sql.DB) {
	cp.stats.ActiveConnections--

	select {
	case cp.connections <- conn:
		cp.stats.IdleConnections++
	default:
		// Pool is full, close the connection
		_ = conn.Close()
		cp.mutex.Lock()
		cp.currentSize--
		cp.mutex.Unlock()
	}
}

// createConnection creates a new database connection
func (cp *ConnectionPool) createConnection() (*sql.DB, error) {
	conn, err := sql.Open("duckdb", cp.connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create DuckDB connection: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping DuckDB connection: %w", err)
	}

	return conn, nil
}

// updateWaitTime updates average wait time statistics
func (cp *ConnectionPool) updateWaitTime(waitTime time.Duration) {
	// Simple exponential moving average
	if cp.stats.AverageWaitTime == 0 {
		cp.stats.AverageWaitTime = waitTime
	} else {
		cp.stats.AverageWaitTime = time.Duration(
			float64(cp.stats.AverageWaitTime)*0.9 + float64(waitTime)*0.1,
		)
	}
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() *PoolStats {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Create a copy to avoid race conditions
	stats := *cp.stats
	stats.IdleConnections = len(cp.connections)
	return &stats
}

// Close closes all connections in the pool
func (cp *ConnectionPool) Close() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	close(cp.connections)
	for conn := range cp.connections {
		_ = conn.Close()
	}
	cp.currentSize = 0
	cp.stats.ActiveConnections = 0
	cp.stats.IdleConnections = 0
}
