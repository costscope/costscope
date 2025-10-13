package api

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestBuildModuleRouteGroups_Exhaustive(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	groups := buildModuleRouteGroups(reg, logger, nil)
	if len(groups) < 6 {
		t.Fatalf("expected at least 6 groups, got %d", len(groups))
	}
	// ensure common bases exist
	bases := map[string]bool{}
	for _, g := range groups {
		bases[g.BasePath] = true
	}
	want := []string{"/focus", "/providers", "/analytics", "/reports", "/streaming", "/monitoring", "/integration"}
	for _, b := range want {
		if !bases[b] {
			t.Fatalf("expected base path %s to be present", b)
		}
	}
}

func TestRunEnhancedAPI_DirectCall(t *testing.T) {
	// call runEnhancedAPI directly; it should return nil in test mode
	cmd := &cobra.Command{}
	if err := runEnhancedAPI(cmd, []string{}); err != nil {
		t.Fatalf("runEnhancedAPI returned error: %v", err)
	}
}

func TestRegisterEnterpriseStructuredAuthRoutes_Noop(t *testing.T) {
	// Should be a no-op in non-enterprise build
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	// Should not panic
	registerEnterpriseStructuredAuthRoutes(r, logger)
}
