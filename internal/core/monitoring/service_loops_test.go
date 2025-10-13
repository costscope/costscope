package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestBasicMonitoringService_Loops exercise metricsCollectionLoop, healthCheckLoop and alertProcessingLoop early stop paths.
func TestBasicMonitoringService_Loops(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError) // reduce log noise
	svc := NewBasicMonitoringService(logger, nil, nil)
	cfg := *svc.config
	cfg.MetricsInterval = 5 * time.Millisecond
	cfg.HealthCheckInterval = 5 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := svc.StartRealTimeMonitoring(ctx, &cfg); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Allow a few ticks
	time.Sleep(20 * time.Millisecond)
	if err := svc.StopRealTimeMonitoring(ctx); err != nil {
		t.Fatalf("stop: %v", err)
	}
	// Second stop should error (exercise branch)
	if err := svc.StopRealTimeMonitoring(ctx); err == nil {
		t.Fatalf("expected error on second stop")
	}
}
