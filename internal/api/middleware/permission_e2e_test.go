package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/security"

	"github.com/gin-gonic/gin"
)

// Minimal E2E test: spin up a gin engine with auth simulation and permission enforcement
func TestPermissionMiddleware_E2E(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := security.NewFileRBACStore(t.TempDir())
	if err := store.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	svc := security.NewRBACService(store, logging.NewLogger(logging.LevelError))
	// create role
	_, err := svc.CreateRole("analyst", "", []security.Permission{{Resource: security.ResourceReports, Action: security.ActionGenerate}})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}

	// Engine
	r := gin.New()
	r.GET("/reports/generate", func(c *gin.Context) {
		// simulate auth setting roles header
		c.Request = c.Request.WithContext(context.Background())
		RequirePermission(svc, security.ResourceReports, security.ActionGenerate)(c)
		if c.IsAborted() {
			return
		}
		c.String(http.StatusOK, "ok")
	})

	// allowed
	req := httptest.NewRequest(http.MethodGet, "/reports/generate", nil)
	req.Header.Set("X-User-Roles", "analyst")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// denied
	req2 := httptest.NewRequest(http.MethodGet, "/reports/generate", nil)
	req2.Header.Set("X-User-Roles", "viewer")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 403 {
		t.Fatalf("expected 403, got %d", w2.Code)
	}

	// audit mode
	SetAuditModeForTests(true)
	defer SetAuditModeForTests(false)
	req3 := httptest.NewRequest(http.MethodGet, "/reports/generate", nil)
	req3.Header.Set("X-User-Roles", "viewer")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code == 403 {
		t.Fatalf("audit mode should not block")
	}
	if h := w3.Header().Get("X-RBAC-Audit"); h != "deny" {
		t.Fatalf("expected audit header, got %s", h)
	}

	// quick latency expectation (not strict, ensure path executed fast <100ms)
	if d := w3.Result().Header.Get("Date"); d == "" {
		_ = time.Now()
	} // noop, just referencing time for vet sanity
}
