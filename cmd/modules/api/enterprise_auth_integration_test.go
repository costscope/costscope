//go:build enterprise
// +build enterprise

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"local/costscope/internal/api/middleware"
	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// TestEnterpriseRequireRoleRoute verifies the live admin-only route wired with structured AuthMiddleware.
func TestEnterpriseRequireRoleRoute(t *testing.T) {
	t.Parallel()

	// Setup router and register the production wiring helper
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())

	logger := logging.NewLogger(logging.LevelDebug)
	enterpriseJwtSecret = strings.Repeat("s", 48)
	enterpriseJwtIssuer = "costscope"
	registerEnterpriseStructuredAuthRoutes(r, logger)

	// Create a matching auth helper for token minting
	auth := middleware.NewAuthMiddleware(logger, enterpriseJwtSecret, enterpriseJwtIssuer)

	// Admin token should pass
	tokenAdmin, err := auth.GenerateToken("u1", "alice", "a@example.com", []string{"admin"}, nil, time.Minute)
	if err != nil {
		t.Fatalf("token gen: %v", err)
	}
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/role-check", nil)
	req1.Header.Set("Authorization", "Bearer "+tokenAdmin)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("admin expected 200, got %d: %s", w1.Code, w1.Body.String())
	}

	// Non-admin token should be forbidden
	tokenUser, err := auth.GenerateToken("u2", "bob", "b@example.com", []string{"user"}, nil, time.Minute)
	if err != nil {
		t.Fatalf("token gen: %v", err)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/role-check", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenUser)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("user expected 403, got %d: %s", w2.Code, w2.Body.String())
	}
}
