package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// helper to build a minimal metrics object that will trigger CPU critical alert
func makeHighCPUMetrics() *RealTimeMetrics {
	return &RealTimeMetrics{
		Performance: PerformanceMetrics{
			CPU:         CPUMetrics{UsagePercent: 95.0},
			Memory:      MemoryMetrics{UsagePercent: 40.0},
			Application: AppPerformanceMetrics{ErrorRate: 0.1},
		},
	}
}

func TestBasicMonitoringService_StartStop(t *testing.T) {
	logger := logging.GetLogger()
	svc := NewBasicMonitoringService(logger, nil, nil)

	// run fast to exercise loops
	cfg := *svc.config
	cfg.MetricsInterval = 10 * time.Millisecond
	cfg.HealthCheckInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.StartRealTimeMonitoring(ctx, &cfg); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(35 * time.Millisecond)

	if err := svc.StopRealTimeMonitoring(ctx); err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	// ensure metrics were collected at least once
	m, err := svc.GetRealTimeMetrics(ctx)
	if err != nil {
		t.Fatalf("get metrics failed: %v", err)
	}
	if m == nil {
		t.Fatalf("expected non-nil metrics after start/stop cycle")
	}
}

func TestBasicMonitoringService_GetRealTimeMetrics_CollectsWhenNil(t *testing.T) {
	svc := NewBasicMonitoringService(logging.GetLogger(), nil, nil)
	ctx := context.Background()
	m, err := svc.GetRealTimeMetrics(ctx)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}
	if m == nil {
		t.Fatalf("expected non-nil metrics")
	}
}

func TestBasicMonitoringService_ProcessAlerts_NoDuplicateAndNotify(t *testing.T) {
	svc := NewBasicMonitoringService(logging.GetLogger(), nil, nil)
	svc.config.AlertingEnabled = true

	// seed high CPU to trigger alert generation
	svc.mu.Lock()
	svc.realTimeMetrics = makeHighCPUMetrics()
	svc.mu.Unlock()

	ctx := context.Background()
	svc.processAlerts(ctx)
	svc.processAlerts(ctx) // run twice; should not duplicate

	svc.mu.RLock()
	defer svc.mu.RUnlock()
	if len(svc.activeAlerts) == 0 {
		t.Fatalf("expected at least one active alert")
	}
	// verify no duplicates for the same type when still active
	seen := map[string]bool{}
	for _, a := range svc.activeAlerts {
		if seen[a.Type] {
			t.Fatalf("duplicate alert type detected: %s", a.Type)
		}
		seen[a.Type] = true
	}
}

func TestBasicMonitoringService_ResolveAlert(t *testing.T) {
	svc := NewBasicMonitoringService(logging.GetLogger(), nil, nil)
	ctx := context.Background()

	// add one alert
	a := &Alert{ID: "a1", Type: "cpu_critical", Status: AlertStatusActive}
	svc.mu.Lock()
	svc.activeAlerts = append(svc.activeAlerts, a)
	svc.mu.Unlock()

	if err := svc.ResolveAlert(ctx, "a1"); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if err := svc.ResolveAlert(ctx, "missing"); err == nil {
		t.Fatalf("expected error resolving missing alert")
	}
}

func TestBasicMonitoringService_GetDashboardData_Composes(t *testing.T) {
	svc := NewBasicMonitoringService(logging.GetLogger(), nil, nil)
	ctx := context.Background()

	d1, err := svc.GetDashboardData(ctx)
	if err != nil || d1 == nil {
		t.Fatalf("first dashboard generation failed: %v", err)
	}
	d2, err := svc.GetDashboardData(ctx)
	if err != nil || d2 == nil {
		t.Fatalf("second dashboard fetch failed: %v", err)
	}
	// basic sanity checks (no strict caching guarantee required by contract)
	if d1.SystemOverview.TotalComponents == 0 || d2.SystemOverview.TotalComponents == 0 {
		t.Fatalf("expected non-zero components in dashboard overview")
	}
}

func TestBasicMonitoringService_ConfigUpdate(t *testing.T) {
	svc := NewBasicMonitoringService(logging.GetLogger(), nil, nil)
	ctx := context.Background()
	newCfg := *svc.config
	newCfg.AlertingEnabled = !svc.config.AlertingEnabled
	if err := svc.UpdateMonitoringConfig(ctx, &newCfg); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, _ := svc.GetMonitoringConfig(ctx)
	if got.AlertingEnabled != newCfg.AlertingEnabled {
		t.Fatalf("config not applied")
	}
}

func TestBasicMonitoringService_StartTwiceAndStopWhenNotRunning(t *testing.T) {
	svc := NewBasicMonitoringService(logging.GetLogger(), nil, nil)
	ctx := context.Background()

	if err := svc.StopRealTimeMonitoring(ctx); err == nil {
		t.Fatalf("expected error when stopping not running service")
	}

	if err := svc.StartRealTimeMonitoring(ctx, nil); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := svc.StartRealTimeMonitoring(ctx, nil); err == nil {
		t.Fatalf("expected error on duplicate start")
	}
	if err := svc.StopRealTimeMonitoring(ctx); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}
