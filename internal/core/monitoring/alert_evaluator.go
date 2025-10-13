package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// DefaultAlertEvaluator is a stateless evaluator that inspects metrics and emits alert candidates.
type DefaultAlertEvaluator struct {
	logger *logging.Logger
}

func NewDefaultAlertEvaluator(logger *logging.Logger) *DefaultAlertEvaluator {
	if logger == nil {
		logger = logging.GetLogger()
	}
	return &DefaultAlertEvaluator{logger: logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "alerts"})}
}

// Evaluate inspects metrics against thresholds and returns alert candidates (no side-effects).
func (e *DefaultAlertEvaluator) Evaluate(ctx context.Context, metrics *RealTimeMetrics, th PerformanceThresholds) []*Alert {
	if metrics == nil {
		return nil
	}
	out := make([]*Alert, 0, 4)

	// CPU
	if metrics.Performance.CPU.UsagePercent > th.CPUCritical {
		out = append(out, e.new("cpu_critical", "CPU usage exceeded critical threshold", "critical"))
	} else if metrics.Performance.CPU.UsagePercent > th.CPUWarning {
		out = append(out, e.new("cpu_warning", "CPU usage exceeded warning threshold", "warning"))
	}

	// Memory
	if metrics.Performance.Memory.UsagePercent > th.MemoryCritical {
		out = append(out, e.new("memory_critical", "Memory usage exceeded critical threshold", "critical"))
	} else if metrics.Performance.Memory.UsagePercent > th.MemoryWarning {
		out = append(out, e.new("memory_warning", "Memory usage exceeded warning threshold", "warning"))
	}

	// Error rate
	if metrics.Performance.Application.ErrorRate > th.ErrorRateCritical {
		out = append(out, e.new("error_rate_critical", "Error rate exceeded critical threshold", "critical"))
	}

	if len(out) > 0 {
		e.logger.Debug(fmt.Sprintf("evaluated %d alert candidates", len(out)))
	}
	return out
}

func (e *DefaultAlertEvaluator) new(t, desc, sev string) *Alert {
	now := time.Now()
	return &Alert{
		ID:          fmt.Sprintf("%s_%d", t, now.Unix()),
		Type:        t,
		Severity:    sev,
		Source:      "monitoring",
		Component:   "system",
		Title:       fmt.Sprintf("%s Alert", t),
		Description: desc,
		Status:      AlertStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        map[string]string{"source": "monitoring"},
	}
}
