package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/api/middleware"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/security"
)

// TestBuildModuleRouteGroups_PermissionWrappedHandlers calls two permission-wrapped
// handlers directly to exercise the middleware composition and handler invocation.
func TestBuildModuleRouteGroups_PermissionWrappedHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	store := security.NewFileRBACStore(t.TempDir())
	rbac := security.NewRBACService(store, logger)

	// Enable audit mode so permission middleware will soft-deny but continue the chain
	middleware.SetAuditModeForTests(true)
	groups := buildModuleRouteGroups(reg, logger, rbac)

	// Find and invoke /focus/validate and /providers/:provider/connect handlers
	var focusHandler, providersConnectHandler func(c *gin.Context)
	for _, g := range groups {
		for _, r := range g.Routes {
			full := g.BasePath + r.Path
			if strings.Contains(full, "/focus/validate") {
				focusHandler = r.Handler
			}
			if strings.Contains(full, "/providers/:provider/connect") || strings.Contains(full, "/providers/:provider/test") {
				providersConnectHandler = r.Handler
			}
		}
	}
	if focusHandler == nil || providersConnectHandler == nil {
		t.Fatalf("expected handlers to be present")
	}

	// Invoke focus validate - send minimal JSON body required by handler
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/focus/validate", strings.NewReader(`{"input_path":"/data/input.csv"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	focusHandler(c)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 from focus validate got %d body=%q", w.Code, w.Body.String())
	}

	// Invoke providers connect
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/providers/aws/test", nil)
	providersConnectHandler(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 from providers connect got %d", w2.Code)
	}
}
