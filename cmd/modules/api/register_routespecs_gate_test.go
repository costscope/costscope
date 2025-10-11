package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

func TestRegisterRouteSpecs_SkipsWhenGateDisabled(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	mux := http.NewServeMux()

	// Route guarded by TEST_GATE
	specs := []RouteSpec{
		{Method: http.MethodGet, Path: "/secret", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }), FeatureGate: "TEST_GATE"},
	}

	// ensure gate disabled
	if err := os.Setenv("TEST_GATE", "0"); err != nil {
		t.Fatalf("failed to set TEST_GATE: %v", err)
	}
	registerRouteSpecs(mux, specs, nil, logger)

	// request should 404 because route skipped
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secret", nil)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for skipped route, got %d", rr.Code)
	}

	// enable gate
	if err := os.Setenv("TEST_GATE", "1"); err != nil {
		t.Fatalf("failed to set TEST_GATE: %v", err)
	}
	mux2 := http.NewServeMux()
	registerRouteSpecs(mux2, specs, nil, logger)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/secret", nil)
	mux2.ServeHTTP(rr2, req2)
	if rr2.Code == http.StatusNotFound {
		t.Fatalf("expected route to be registered when gate enabled")
	}
}

func TestRegisterDebugRoutes_WhenEnabled_AddsCacheStats(t *testing.T) {
	// enable flag
	prev := enableCacheStats
	enableCacheStats = true
	defer func() { enableCacheStats = prev }()

	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	router := gin.New()
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	// simple auth that marks the user as admin
	auth := func(c *gin.Context) {
		c.Set("roles", []string{"admin"})
		c.Next()
	}

	registerDebugRoutes(router, reg, auth)

	// request should be handled (not 404)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/cache-stats", nil)
	router.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Fatalf("expected /debug/cache-stats to be registered when enabled")
	}
}

func TestRegisterEnterpriseStructuredAuthRoutes_Noop_Call(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	router := gin.New()
	// Should be a no-op in non-enterprise build; just ensure it doesn't panic
	registerEnterpriseStructuredAuthRoutes(router, logger)
}
