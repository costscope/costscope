//go:build !enterprise

package streaming

import (
	"context"
	"testing"

	"github.com/costscope/costscope/internal/core/enterprise"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestEnterpriseStreamingEngineStub_Disabled(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	eng := NewEnterpriseStreamingEngine(logger)
	if _, err := eng.StartStreamingOperation(context.Background(), &StreamingOperationRequest{OperationID: "op1"}); !enterprise.IsDisabled(err) {
		t.Fatalf("expected disabled error, got %v", err)
	}
}
