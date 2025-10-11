package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/security"

	"github.com/gin-gonic/gin"
)

// helper to build RBAC service with a temp file-backed store
func newTestRBAC(t *testing.T, perms []security.Permission) *security.RBACService {
	t.Helper()
	dir := t.TempDir()
	store := security.NewFileRBACStore(dir)
	if err := store.Load(); err != nil { // should be empty
		t.Fatalf("load store: %v", err)
	}
	svc := security.NewRBACService(store, logging.NewLogger(logging.LevelError))
	role := security.Role{ // timestamp not critical
		Name:        "analyst",
		Description: "test role",
		Permissions: perms,
		CreatedAt:   time.Now().UTC(),
	}
	if err := store.AddRole(role); err != nil {
		t.Fatalf("add role: %v", err)
	}
	return svc
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	perms := []security.Permission{{Resource: security.ResourceFocus, Action: security.ActionConvert}}
	rbac := newTestRBAC(t, perms)

	router := gin.New()
	router.GET("/convert", RequirePermission(rbac, security.ResourceFocus, security.ActionConvert), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("allowed via single role header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/convert", nil)
		req.Header.Set("X-User-Roles", "analyst")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK || w.Body.String() != "ok" {
			t.Fatalf("expected 200 ok, got %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("denied when role missing permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/convert", nil)
		req.Header.Set("X-User-Roles", "viewer")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("allowed via multi-role header fallback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/convert", nil)
		req.Header.Set("X-User-Roles", "viewer, analyst,other")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestRequirePermission_OtherResourcesAndAuditMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Role with report generate and focus validate permissions
	perms := []security.Permission{{Resource: security.ResourceReports, Action: security.ActionGenerate}, {Resource: security.ResourceFocus, Action: security.ActionValidate}}
	rbac := newTestRBAC(t, perms)

	router := gin.New()
	router.GET("/reports/generate", RequirePermission(rbac, security.ResourceReports, security.ActionGenerate), func(c *gin.Context) {
		c.String(http.StatusOK, "gen")
	})
	router.GET("/focus/validate", RequirePermission(rbac, security.ResourceFocus, security.ActionValidate), func(c *gin.Context) {
		c.String(http.StatusOK, "val")
	})

	// Allowed: correct permission
	req := httptest.NewRequest(http.MethodGet, "/reports/generate", nil)
	req.Header.Set("X-User-Roles", "analyst")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "gen" {
		t.Fatalf("expected gen 200, got %d %s", w.Code, w.Body.String())
	}

	// Denied normally
	req2 := httptest.NewRequest(http.MethodGet, "/focus/validate", nil)
	req2.Header.Set("X-User-Roles", "viewer")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w2.Code)
	}

	// Audit mode: should pass through (200) with X-RBAC-Audit header set after denial point
	SetAuditModeForTests(true)
	t.Cleanup(func() { SetAuditModeForTests(false) })
	req3 := httptest.NewRequest(http.MethodGet, "/focus/validate", nil)
	req3.Header.Set("X-User-Roles", "viewer")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code == http.StatusForbidden {
		t.Fatalf("audit mode should not block, got 403")
	}
	if w3.Header().Get("X-RBAC-Audit") != "deny" {
		t.Fatalf("expected audit header, got %s", w3.Header().Get("X-RBAC-Audit"))
	}
}
