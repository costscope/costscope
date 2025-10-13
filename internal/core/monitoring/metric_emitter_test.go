package monitoring

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func Test_LoggingMetricEmitter_Emit_nil(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	emitter := NewLoggingMetricEmitter(logger)
	// nil input should be a no-op and not panic
	if err := emitter.Emit(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error for nil metrics, got: %v", err)
	}
}

func Test_LoggingMetricEmitter_Emit_basic(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	emitter := NewLoggingMetricEmitter(logger)
	m := &RealTimeMetrics{
		Performance:      PerformanceMetrics{CPU: CPUMetrics{UsagePercent: 1.2}},
		ActiveAlerts:     0,
		HealthScore:      100,
		CollectionTimeMs: 5,
	}
	if err := emitter.Emit(context.Background(), m); err != nil {
		t.Fatalf("emit failed: %v", err)
	}
}
