//go:build enterprise
// +build enterprise

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/middleware"
	"local/costscope/internal/core/logging"
)

// setupEnterpriseRouterForTest creates a minimal enterprise router with debug route enabled
func setupEnterpriseRouterForTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	// enable debug endpoint
	enableCacheStats = true

	// minimal logger
	_ = logging.NewLogger(logging.LevelWarn)

	// auth middleware
	jwtSecret := "abcdefghijklmnopqrstuvwxyz012345" // 33 bytes
	am := middleware.NewAuthMiddleware(logging.NewLogger(logging.LevelWarn), jwtSecret, "costscope")

	r := gin.New()
	r.Use(middleware.RequestID())

	// Register a local protected route equivalent to /debug/cache-stats
	dbg := r.Group("/debug")
	dbg.Use(am.RequireAuth())
	dbg.Use(middleware.RBAC("admin", "system_admin"))
	dbg.GET("/cache-stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

func TestEnterpriseDebugRBAC_DenyAllow(t *testing.T) {
	r := setupEnterpriseRouterForTest(t)

	// Prepare JWTs
	am := middleware.NewAuthMiddleware(logging.NewLogger(logging.LevelWarn), "abcdefghijklmnopqrstuvwxyz012345", "costscope")
	adminToken, _ := am.GenerateToken("u1", "admin", "a@x", []string{"admin"}, []string{"read:all"}, time.Hour)
	userToken, _ := am.GenerateToken("u2", "user", "u@x", []string{"user"}, []string{"read:all"}, time.Hour)

	// 1) Deny anonymous
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/debug/cache-stats", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous, got %d", w.Code)
	}

	// 2) Deny authenticated non-admin
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/debug/cache-stats", nil)
	req2.Header.Set("Authorization", "Bearer "+userToken)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", w2.Code)
	}

	// 3) Allow admin
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/debug/cache-stats", nil)
	req3.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d body=%s", w3.Code, w3.Body.String())
	}
}
