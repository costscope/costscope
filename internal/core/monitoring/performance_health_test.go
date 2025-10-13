package monitoring

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/production"
)

// TestMonitoringService_PerformanceAndHealthChecks covers GetPerformanceMetrics (collect path) and RunHealthChecks.
func TestMonitoringService_PerformanceAndHealthChecks(t *testing.T) {
	svc := NewBasicMonitoringService(logging.NewLogger(logging.LevelError), nil, nil)
	ctx := context.Background()

	pm, err := svc.GetPerformanceMetrics(ctx)
	if err != nil || pm == nil {
		t.Fatalf("performance metrics collect failed: %v", err)
	}
	if pm.CPU.Cores == 0 {
		t.Fatalf("expected non-zero CPU cores")
	}

	// Run health checks for a subset of components
	results, err := svc.RunHealthChecks(ctx, []string{"api", "cache", "missing"})
	if err != nil || results == nil {
		t.Fatalf("RunHealthChecks failed: %v", err)
	}
	if results.TotalChecks != 3 {
		t.Fatalf("expected 3 total checks, got %d", results.TotalChecks)
	}
	if len(results.FailedChecks) == 0 { // 'missing' should fail
		t.Fatalf("expected at least one failed check")
	}
}

// TestMonitoringService_NilHealthChecker exercises error branches when healthChecker removed.
func TestMonitoringService_NilHealthChecker(t *testing.T) {
	svc := NewBasicMonitoringService(logging.NewLogger(logging.LevelError), nil, nil)
	ctx := context.Background()
	// Remove health checker to hit error paths
	svc.healthChecker = nil
	if _, err := svc.GetComponentHealth(ctx, "api"); err == nil {
		t.Fatalf("expected error when healthChecker nil")
	}
	if _, err := svc.RunHealthChecks(ctx, []string{"api"}); err == nil {
		t.Fatalf("expected error when healthChecker nil")
	}
}

// TestMonitoringService_GetSupportedChannels ensures the trivial getter is covered.
func TestMonitoringService_GetSupportedChannels(t *testing.T) {
	ns := NewBasicNotificationService(logging.NewLogger(logging.LevelError))
	ch := ns.GetSupportedChannels()
	if len(ch) == 0 {
		t.Fatalf("expected supported channels")
	}
}

// TestMetricsCollector_ConvertProductionMetrics covers convertProductionToApplicationMetrics path.
func TestMetricsCollector_ConvertProductionMetrics(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	mc := NewBasicMetricsCollector(logger, nil, nil)
	prod := &production.ProductionSystemMetrics{
		Performance:      production.PerformanceMetrics{ThroughputOpsPerSec: 1000, NetworkLatencyMs: 10},
		SystemHealth:     production.SystemHealthStatus{ErrorRate: 0.5},
		TotalFeatures:    10,
		TotalCommands:    5,
		TotalEndpoints:   3,
		ReadinessScore:   80,
		CompletionLevel:  "beta",
		ProcessingTimeMs: 123,
		ProductionReady:  true,
	}
	app := mc.convertProductionToApplicationMetrics(prod)
	if app == nil || app.RequestsPerSecond == 0 {
		t.Fatalf("conversion failed")
	}
	if _, ok := app.CustomMetrics["readiness_score"]; !ok {
		t.Fatalf("expected readiness_score in custom metrics")
	}
}
