package monitoring

import (
	"context"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestDefaultHealthChecker_System(t *testing.T) {
	hc := NewDefaultHealthChecker(logging.NewLogger(logging.LevelDebug))
	st := hc.System(context.Background())
	if st == nil || st.HealthScore <= 0 {
		t.Fatalf("expected non-nil system health with positive score")
	}
}

func TestDefaultHealthChecker_Component(t *testing.T) {
	hc := NewDefaultHealthChecker(logging.NewLogger(logging.LevelDebug))
	h, err := hc.Component(context.Background(), "api")
	if err != nil || h == nil {
		t.Fatalf("expected api component health, got err=%v", err)
	}
}
