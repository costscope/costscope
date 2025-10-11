package monitoring

import (
	"context"
	"fmt"
)

// collectPerformanceMetrics was extracted from service.go to reduce LOC without changing behavior.
func (bms *BasicMonitoringService) collectPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	resourceMetrics, err := bms.metricsCollector.CollectResourceMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect resource metrics: %w", err)
	}

	appMetrics, err := bms.metricsCollector.CollectApplicationMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect application metrics: %w", err)
	}

	performance := &PerformanceMetrics{
		CPU:     resourceMetrics.CPU,
		Memory:  resourceMetrics.Memory,
		Disk:    resourceMetrics.Disk,
		Network: resourceMetrics.Network,
		Application: AppPerformanceMetrics{
			RequestCount:   int64(appMetrics.RequestsPerSecond * 60),
			SuccessRate:    100.0 - appMetrics.ErrorRate,
			ErrorRate:      appMetrics.ErrorRate,
			ResponseTime:   appMetrics.ResponseTime,
			ThroughputRPS:  appMetrics.RequestsPerSecond,
			MemoryUsage:    resourceMetrics.Memory.UsagePercent,
			GoroutineCount: 150,
			GCPauseTime:    2.5,
		},
		Database: DatabaseMetrics{
			ConnectionCount:   25,
			ActiveQueries:     8,
			SlowQueries:       2,
			QueryResponseTime: 15.5,
			LockWaitTime:      2.1,
			TransactionRate:   125.5,
			CacheHitRate:      95.2,
			TableSize:         2.5,
			IndexUsage:        88.7,
		},
		API: APIMetrics{
			TotalRequests:      12500,
			SuccessfulRequests: 12000,
			FailedRequests:     500,
			AverageLatency:     45.2,
			P95Latency:         125.8,
			P99Latency:         250.3,
			RateLimitHits:      15,
		},
		Throughput: ThroughputMetrics{
			RequestsPerSecond:     appMetrics.RequestsPerSecond,
			BytesPerSecond:        1024 * 1024 * 2.5,
			TransactionsPerSecond: 85.5,
			Peak:                  150.0,
			Average:               appMetrics.RequestsPerSecond,
			Minimum:               25.0,
		},
		Latency: appMetrics.ResponseTime,
		Concurrency: ConcurrencyMetrics{
			ActiveConnections:     appMetrics.ConcurrentUsers,
			MaxConnections:        1000,
			QueueLength:           5,
			ThreadPoolUtilization: 65.2,
			WorkerUtilization:     72.8,
			ConcurrencyLevel:      0.75,
		},
		PerformanceScore: bms.calculatePerformanceScore(resourceMetrics),
		Grade:            bms.calculatePerformanceGrade(resourceMetrics),
	}

	return performance, nil
}
