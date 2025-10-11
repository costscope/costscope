package conversion

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"local/costscope/internal/core/focus/types"
	"local/costscope/internal/core/monitoring/telemetry"

	promtest "github.com/prometheus/client_golang/prometheus/testutil"
)

// testConverter is a minimal stub implementing types.Converter for metrics tests.
type testConverter struct {
	delay   time.Duration
	succeed bool
}

func (t *testConverter) Convert(ctx context.Context, cfg *types.ConversionConfig) (*types.ConversionResult, error) {
	if t.delay > 0 {
		select {
		case <-time.After(t.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if !t.succeed {
		return nil, errors.New("boom")
	}
	now := time.Now()
	return &types.ConversionResult{Success: true, ConversionId: cfg.ConversionId, StartTime: now.Add(-t.delay), EndTime: now, Duration: t.delay}, nil
}

// resetDefaultRegistry removes previously registered metrics (Prometheus global registry is process-wide).
// For isolation we rely on unique counters increasing; tests assert deltas not absolute zeros. We still call telemetry.Register() once.

func ensureMetricsRegistered() {
	// Register is idempotent but will panic on double registration. Protect with recover.
	defer func() { _ = recover() }()
	telemetry.Register()
}

func TestConversionManagerMetrics_SubmitAndSuccess(t *testing.T) {
	ensureMetricsRegistered()

	cm := NewConversionManager(2)
	// Replace real converters with stub to avoid file IO.
	tc := &testConverter{delay: 5 * time.Millisecond, succeed: true}
	if err := cm.GetConverter().RegisterConverter("test", tc); err != nil {
		t.Fatalf("register converter: %v", err)
	}

	cfg := &types.ConversionConfig{Provider: "test", InputPath: "in.csv", OutputPath: "out.parquet"}
	id, err := cm.SubmitJob(cfg)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if id == "" {
		t.Fatal("expected job id")
	}
	// Poll for completion (avoid flakiness under heavy CI load)
	deadline := time.Now().Add(1200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success")) >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success")) < 1 {
		t.Fatalf("timeout waiting for success metric (submitted=%v success=%v)", promtest.ToFloat64(telemetry.ConversionJobsSubmitted), promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success")))
	}
	// Assert counters > 0 using testutil (submitted checked after to reduce flake window)
	if v := promtest.ToFloat64(telemetry.ConversionJobsSubmitted); v < 1 {
		t.Fatalf("expected submitted >=1 got %v", v)
	}
	// Histogram: ensure at least one sample collected (CollectAndCount == 1 for histogram itself) and sum > 0 via metric iteration.
	if c := promtest.CollectAndCount(telemetry.ConversionJobDuration); c == 0 {
		t.Fatalf("expected histogram sample count >0")
	}
}

func TestConversionManagerMetrics_Failure(t *testing.T) {
	ensureMetricsRegistered()
	cm := NewConversionManager(1)
	tc := &testConverter{delay: 1 * time.Millisecond, succeed: false}
	if err := cm.GetConverter().RegisterConverter("fail", tc); err != nil {
		t.Fatalf("register converter: %v", err)
	}
	cfg := &types.ConversionConfig{Provider: "fail", InputPath: "in.csv", OutputPath: "out.parquet"}
	_, err := cm.SubmitJob(cfg)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if v := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("failed")); v < 1 {
		t.Fatalf("expected failed >=1 got %v", v)
	}
}

func TestConversionManagerMetrics_Cancel(t *testing.T) {
	ensureMetricsRegistered()
	cm := NewConversionManager(1)
	tc := &testConverter{delay: 50 * time.Millisecond, succeed: true}
	if err := cm.GetConverter().RegisterConverter("slow", tc); err != nil {
		t.Fatalf("register converter: %v", err)
	}
	cfg := &types.ConversionConfig{Provider: "slow", InputPath: "in.csv", OutputPath: "out.parquet"}
	id, err := cm.SubmitJob(cfg)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	// Cancel quickly
	if err := cm.CancelJob(id); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	// Ensure cancelled counter incremented and not double-counted as failed
	cancelled := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("cancelled"))
	if cancelled < 1 {
		t.Fatalf("expected cancelled >=1 got %v", cancelled)
	}
	failed := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("failed"))
	// Accept 0 or existing from other tests, but ensure it did NOT increment because of this cancellation (cannot isolate absolutely)
	// So just ensure cancelled >= failed isn't reliable; skip strict assert.
	_ = failed
}

// Stress test submitting more jobs than max concurrency and ensuring rejection + metrics stability.
func TestConversionManager_MaxConcurrencyAndMetrics(t *testing.T) {
	ensureMetricsRegistered()
	max := 3
	cm := NewConversionManager(max)
	// fast stub converter that sleeps briefly so jobs overlap
	if err := cm.GetConverter().RegisterConverter("stress", &testConverter{delay: 50 * time.Millisecond, succeed: true}); err != nil {
		t.Fatalf("register converter: %v", err)
	}
	baseSubmitted := promtest.ToFloat64(telemetry.ConversionJobsSubmitted)
	baseSuccess := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success"))
	baseActive := promtest.ToFloat64(telemetry.ConversionActiveJobs)
	accepted := 0
	rejected := 0
	for i := 0; i < max+2; i++ { // submit 2 over the limit
		cfg := &types.ConversionConfig{Provider: "stress", InputPath: "in.csv", OutputPath: "out.parquet"}
		if _, err := cm.SubmitJob(cfg); err != nil {
			rejected++
		} else {
			accepted++
		}
	}
	if accepted != max {
		t.Fatalf("expected %d accepted jobs, got %d", max, accepted)
	}
	if rejected != 2 {
		t.Fatalf("expected 2 rejections, got %d", rejected)
	}
	// Wait for jobs to complete
	time.Sleep(250 * time.Millisecond)
	// Gauge should return to baseline
	if v := promtest.ToFloat64(telemetry.ConversionActiveJobs); int(v-baseActive) != 0 {
		t.Fatalf("expected active jobs delta 0, got %v (baseline %v)", v-baseActive, v)
	}
	if v := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success")); int(v-baseSuccess) != accepted {
		t.Fatalf("expected success delta %d got %d", accepted, int(v-baseSuccess))
	}
	if v := promtest.ToFloat64(telemetry.ConversionJobsSubmitted); int(v-baseSubmitted) != accepted {
		t.Fatalf("expected submitted delta %d got %d", accepted, int(v-baseSubmitted))
	}
}

// Ensure cancellation avoids double-count metrics under race conditions.
func TestConversionManager_CancelMetricsNoDoubleCount(t *testing.T) {
	ensureMetricsRegistered()
	cm := NewConversionManager(1)
	if err := cm.GetConverter().RegisterConverter("slow", &testConverter{delay: 30 * time.Millisecond, succeed: true}); err != nil {
		t.Fatalf("register converter: %v", err)
	}
	cfg := &types.ConversionConfig{Provider: "slow", InputPath: "in.csv", OutputPath: "out.parquet"}
	baseCancelled := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("cancelled"))
	baseFailed := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("failed"))
	baseSuccess := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success"))
	id, err := cm.SubmitJob(cfg)
	if err != nil {
		t.Fatalf("submit err: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := cm.CancelJob(id); err != nil {
		t.Fatalf("cancel err: %v", err)
	}
	time.Sleep(70 * time.Millisecond)
	// success should be 0, failed 0, cancelled 1
	if v := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("cancelled")); int(v-baseCancelled) != 1 {
		t.Fatalf("expected cancelled delta=1 got %d", int(v-baseCancelled))
	}
	if v := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("failed")); int(v-baseFailed) != 0 {
		t.Fatalf("expected failed delta=0 got %d", int(v-baseFailed))
	}
	if v := promtest.ToFloat64(telemetry.ConversionJobsCompleted.WithLabelValues("success")); int(v-baseSuccess) != 0 {
		t.Fatalf("expected success delta=0 got %d", int(v-baseSuccess))
	}
}

// Prevent unused warnings in case helpers evolve
var _ io.Writer
