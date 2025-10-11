package monitoring

import "testing"

func TestLoggingMetricEmitter_String(t *testing.T) {
	e := &LoggingMetricEmitter{}
	s := e.String()
	if s == "" {
		t.Fatalf("expected non-empty string")
	}
}
