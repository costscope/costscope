package api

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/api/middleware"
	"local/costscope/internal/api/websocket"
	"local/costscope/internal/core/security"
)

// Exercise a few inline closures in buildModuleRouteGroups without starting a server.
func TestBuildModuleRouteGroups_InlineHandlers(t *testing.T) {
	// Enable audit-mode so RequirePermission soft-allows execution in tests
	middleware.SetAuditModeForTests(true)
	t.Cleanup(func() { middleware.SetAuditModeForTests(false) })

	logger := testLogger()
	ws := websocket.NewManager(logger)
	reg := handlers.NewEnterpriseRegistry(logger, nil, ws)

	// Initialize a simple RBAC service pointing at a temp dir store
	store := security.NewFileRBACStore(t.TempDir())
	_ = store.Load()
	rbacSvc := security.NewRBACService(store, logger)

	groups := buildModuleRouteGroups(reg, logger, rbacSvc)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	v1 := engine.Group("/api/v1")
	RegisterGinRouteGroups(v1, groups)

	// Focus validate
	req := httptest.NewRequest("POST", "/api/v1/focus/validate", bytes.NewBufferString("{}"))
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code >= 500 {
		t.Fatalf("focus/validate returned server error %d", w.Code)
	}

	// Provider test
	req2 := httptest.NewRequest("POST", "/api/v1/providers/aws/test", bytes.NewBufferString("{}"))
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	if w2.Code >= 500 {
		t.Fatalf("providers/:provider/test returned server error %d", w2.Code)
	}

	// Analytics model train
	req3 := httptest.NewRequest("POST", "/api/v1/analytics/models/1/train", bytes.NewBufferString("{}"))
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	if w3.Code >= 500 {
		t.Fatalf("analytics model train returned server error %d", w3.Code)
	}
}
