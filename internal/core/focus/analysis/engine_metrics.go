package analysis

import "github.com/prometheus/client_golang/prometheus"

var (
	focusEngineInvocations = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "costscope_focus_engine_invocations_total", Help: "Focus engine phase invocations"},
		[]string{"component", "status"},
	)
	focusEnginePhaseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "costscope_focus_engine_phase_duration_seconds", Help: "Focus engine phase duration"},
		[]string{"component"},
	)
)

func init() {
	prometheus.MustRegister(focusEngineInvocations, focusEnginePhaseDuration)
}
