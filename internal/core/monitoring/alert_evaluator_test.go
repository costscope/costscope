package monitoring

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func TestDefaultAlertEvaluator_Evaluate(t *testing.T) {
	logger := logging.NewLogger(logging.LevelDebug)
	eval := NewDefaultAlertEvaluator(logger)

	m := &RealTimeMetrics{
		Performance: PerformanceMetrics{
			CPU:    CPUMetrics{UsagePercent: 95},
			Memory: MemoryMetrics{UsagePercent: 96},
			Application: AppPerformanceMetrics{
				ErrorRate: 6.0,
			},
		},
	}
	th := PerformanceThresholds{CPUWarning: 70, CPUCritical: 90, MemoryWarning: 80, MemoryCritical: 95, ErrorRateCritical: 5}

	got := eval.Evaluate(context.Background(), m, th)
	if len(got) < 3 {
		t.Fatalf("expected at least 3 alerts, got %d", len(got))
	}
}
