package production

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

func Test_runStep_wrappedError(t *testing.T) {
	logger := logging.NewLogger(logging.LevelDebug)
	wantErr := errors.New("underlying failure")

	spec := StepSpec[int]{
		Name: "failing_step",
		Run: func(ctx context.Context) (int, error) {
			return 0, wantErr
		},
		ErrWrap: "step failed: %w",
	}

	_, err := runStep(context.Background(), logger, spec)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	// Expect the wrapped message to include original error string
	if !strings.Contains(err.Error(), wantErr.Error()) {
		t.Fatalf("wrapped error does not include underlying error: %v", err)
	}
}
