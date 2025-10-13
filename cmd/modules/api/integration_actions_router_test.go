package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	integration "github.com/costscope/costscope/cmd/modules/integration"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// ensure every non-group ActionSpec has a corresponding POST route
func TestBuildIntegrationActionRoutes_CoversAllSpecs(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	routes := buildIntegrationActionRoutes(logger)
	have := map[string]struct{}{}
	for _, r := range routes {
		if r.Method == http.MethodPost {
			have[r.Path] = struct{}{}
		}
	}
	for _, s := range integration.BuildDefaultActionSpecs() {
		if s.Group {
			continue
		}
		path := "/" + s.Category
		if len(s.Parents) > 0 {
			for _, p := range s.Parents {
				path += "/" + p
			}
		}
		path += "/" + s.Use
		if _, ok := have[path]; !ok {
			t.Fatalf("missing route for action %s at path %s", s.ID, path)
		}
	}
}

// smoke test: POST to a known integration endpoint should return 501 from stub/fallback
func TestIntegrationActionRoute_Stub501(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	gin.SetMode(gin.TestMode)
	e := gin.New()
	// mimic enterprise router layout: /api/v1 + group /integration
	v1 := e.Group("/api/v1")
	RegisterGinRouteGroups(v1, []GinRouteGroup{{BasePath: "/integration", Routes: buildIntegrationActionRoutes(logger)}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration/webhook/create", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
	// response should be JSON with at least status and action_id
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if payload["status"] != "not_implemented" {
		t.Fatalf("unexpected status: %v", payload["status"])
	}
	if payload["action_id"] == nil {
		t.Fatalf("missing action_id in response")
	}
}

// grouped action path should include parent segments and return 501 stub
func TestIntegrationActionRoute_GroupedPath_Stub501(t *testing.T) {
	logger := logging.NewLogger(logging.LevelError)
	gin.SetMode(gin.TestMode)
	e := gin.New()
	v1 := e.Group("/api/v1")
	RegisterGinRouteGroups(v1, []GinRouteGroup{{BasePath: "/integration", Routes: buildIntegrationActionRoutes(logger)}})

	// webhook.delivery.retry should be exposed at /webhook/delivery/retry
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration/webhook/delivery/retry", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}
