package integration

import (
	"fmt"
	"sync"
	"time"
)

// AuditLogger handles audit logging for integration activities
type AuditLogger struct {
	mu   sync.RWMutex
	logs []AuditLog
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]AuditLog, 0),
	}
}

// Log records an audit event
func (a *AuditLogger) Log(action, system, details string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	log := AuditLog{
		ID:        fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Action:    action,
		System:    system,
		User:      "system", // In real implementation, get from context
		Timestamp: time.Now(),
		Details:   map[string]interface{}{"message": details},
		Success:   true,
	}

	a.logs = append(a.logs, log)

	// Keep only last 1000 logs
	if len(a.logs) > 1000 {
		a.logs = a.logs[len(a.logs)-1000:]
	}
}

// LogError records an error audit event
func (a *AuditLogger) LogError(action, system, errorMsg string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	log := AuditLog{
		ID:        fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Action:    action,
		System:    system,
		User:      "system",
		Timestamp: time.Now(),
		Details:   map[string]interface{}{"message": "Operation failed"},
		Success:   false,
		Error:     errorMsg,
	}

	a.logs = append(a.logs, log)

	// Keep only last 1000 logs
	if len(a.logs) > 1000 {
		a.logs = a.logs[len(a.logs)-1000:]
	}
}

// GetLogs returns recent audit logs
func (a *AuditLogger) GetLogs(limit int) []AuditLog {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 || limit > len(a.logs) {
		limit = len(a.logs)
	}

	result := make([]AuditLog, limit)
	copy(result, a.logs[len(a.logs)-limit:])

	return result
}

// MetricsCollector handles metrics collection for integrations
type MetricsCollector struct {
	mu          sync.RWMutex
	connections map[string]*PerformanceMetrics
	startTime   time.Time
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		connections: make(map[string]*PerformanceMetrics),
		startTime:   time.Now(),
	}
}

// RecordConnection records a successful connection
func (m *MetricsCollector) RecordConnection(systemName string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[systemName]; !exists {
		m.connections[systemName] = &PerformanceMetrics{
			RequestsPerSecond: 0.0,
			AverageLatency:    "0ms",
			P95Latency:        "0ms",
			P99Latency:        "0ms",
			ErrorRate:         0.0,
			ThroughputMB:      0.0,
			ConcurrentUsers:   1,
			MemoryUsage:       "0MB",
			CPUUsage:          0.0,
		}
	}

	if !success {
		m.connections[systemName].ErrorRate += 0.1
	}
}

// RecordDisconnection records a disconnection
func (m *MetricsCollector) RecordDisconnection(systemName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.connections[systemName]; exists {
		metrics.ConcurrentUsers = 0
	}
}

// GetMetrics returns performance metrics for a system
func (m *MetricsCollector) GetMetrics(systemName string) *PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics, exists := m.connections[systemName]; exists {
		// Return a copy
		return &PerformanceMetrics{
			RequestsPerSecond: metrics.RequestsPerSecond,
			AverageLatency:    metrics.AverageLatency,
			P95Latency:        metrics.P95Latency,
			P99Latency:        metrics.P99Latency,
			ErrorRate:         metrics.ErrorRate,
			ThroughputMB:      metrics.ThroughputMB,
			ConcurrentUsers:   metrics.ConcurrentUsers,
			MemoryUsage:       metrics.MemoryUsage,
			CPUUsage:          metrics.CPUUsage,
		}
	}

	return nil
}

// GetAllMetrics returns all performance metrics
func (m *MetricsCollector) GetAllMetrics() map[string]*PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*PerformanceMetrics)
	for name, metrics := range m.connections {
		result[name] = &PerformanceMetrics{
			RequestsPerSecond: metrics.RequestsPerSecond,
			AverageLatency:    metrics.AverageLatency,
			P95Latency:        metrics.P95Latency,
			P99Latency:        metrics.P99Latency,
			ErrorRate:         metrics.ErrorRate,
			ThroughputMB:      metrics.ThroughputMB,
			ConcurrentUsers:   metrics.ConcurrentUsers,
			MemoryUsage:       metrics.MemoryUsage,
			CPUUsage:          metrics.CPUUsage,
		}
	}

	return result
}
