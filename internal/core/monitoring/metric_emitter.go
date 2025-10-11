package monitoring

import (
	"context"

	"local/costscope/internal/core/logging"
)

// LoggingMetricEmitter outputs a compact summary into unified logs.
type LoggingMetricEmitter struct {
	logger *logging.Logger
}

func NewLoggingMetricEmitter(logger *logging.Logger) *LoggingMetricEmitter {
	if logger == nil {
		logger = logging.GetLogger()
	}
	return &LoggingMetricEmitter{logger: logger.WithFields(map[string]interface{}{"component": "monitoring", "subcomponent": "metrics"})}
}

func (e *LoggingMetricEmitter) Emit(ctx context.Context, m *RealTimeMetrics) error {
	if m == nil {
		return nil
	}
	fields := map[string]interface{}{
		"cpu":        m.Performance.CPU.UsagePercent,
		"mem":        m.Performance.Memory.UsagePercent,
		"alerts":     m.ActiveAlerts,
		"health":     m.HealthScore,
		"collection": m.CollectionTimeMs,
	}
	e.logger.InfoWithFields("metrics_emit", fields)
	return nil
}

// helper to support future sinks (noop for now)
func (e *LoggingMetricEmitter) String() string { return "LoggingMetricEmitter" }
