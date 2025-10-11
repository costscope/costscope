package enterprise

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerOnce sync.Once

	featureInvocations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "costscope",
			Subsystem: "enterprise",
			Name:      "feature_invocations_total",
			Help:      "Total enterprise feature invocations (allowed=false for stub calls).",
		},
		[]string{"feature", "allowed"},
	)
	featureErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "costscope",
			Subsystem: "enterprise",
			Name:      "feature_errors_total",
			Help:      "Enterprise feature errors by feature and kind.",
		},
		[]string{"feature", "error_kind"},
	)
	featureDisabled = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "costscope",
			Subsystem: "enterprise",
			Name:      "disabled_total",
			Help:      "Counts stub (disabled) feature constructor/use occurrences.",
		},
		[]string{"feature"},
	)
)

// ensureRegistered registers metrics exactly once.
func ensureRegistered() {
	registerOnce.Do(func() {
		prometheus.MustRegister(featureInvocations, featureErrors, featureDisabled)
	})
}

// ObserveInvocation records an enterprise feature invocation. allowed=false for stub paths.
func ObserveInvocation(feature string, allowed bool) {
	ensureRegistered()
	featureInvocations.WithLabelValues(feature, boolToStr(allowed)).Inc()
}

// ObserveError records an enterprise feature error.
func ObserveError(feature, kind string) {
	ensureRegistered()
	featureErrors.WithLabelValues(feature, kind).Inc()
}

// ObserveDisabled increments the disabled counter for a feature (constructor-time or first use).
func ObserveDisabled(feature string) {
	ensureRegistered()
	featureDisabled.WithLabelValues(feature).Inc()
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
