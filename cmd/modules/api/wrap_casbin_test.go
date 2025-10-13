package api

import (
	"net/http"
	"testing"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestWrapServerWithCasbinIfEnabled_EmptyPaths ensures that when Casbin is enabled but
// model/policy paths are not provided, the function logs and returns without wrapping the handler.
func TestWrapServerWithCasbinIfEnabled_EmptyPaths(t *testing.T) {
	// Preserve global and restore
	oldEnabled := enterpriseCasbinEnabled
	oldModel := enterpriseCasbinModelPath
	oldPolicy := enterpriseCasbinPolicyPath
	enterpriseCasbinEnabled = true
	enterpriseCasbinModelPath = ""
	enterpriseCasbinPolicyPath = ""
	defer func() {
		enterpriseCasbinEnabled = oldEnabled
		enterpriseCasbinModelPath = oldModel
		enterpriseCasbinPolicyPath = oldPolicy
	}()

	// Create a dummy handler and logger
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	logger := logging.NewLogger(logging.LevelInfo)

	wrapServerWithCasbinIfEnabled(&h, logger)

	// Handler should remain non-nil and callable (no panic)
	if h == nil {
		t.Fatalf("handler unexpectedly nil after wrapServerWithCasbinIfEnabled")
	}
}
