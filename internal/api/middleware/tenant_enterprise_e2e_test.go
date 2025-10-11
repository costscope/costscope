//go:build enterprise
// +build enterprise

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local/costscope/internal/core/config"
	"local/costscope/internal/core/multitenant"

	"github.com/gin-gonic/gin"
)

// This test verifies that when multi-tenancy is enabled and no X-Tenant-ID header is provided,
// the tenant is derived from JWTClaims.TenantID placed in the context by RequireAuth() middleware.
func TestTenant_FromJWTClaims_EndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}

	am := NewAuthMiddleware(nil, "test-secret-very-long-for-jwt", "issuer")

	// Create a token with standard fields first.
	tok, err := am.GenerateToken("u1", "alice", "a@example", []string{"admin"}, []string{"read"}, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	r := gin.New()
	// First, auth to populate jwt_claims in the Gin context
	r.Use(am.RequireAuth())
	// Then, set TenantContext to extract tenant from jwt_claims.TenantID
	r.Use(func(c *gin.Context) {
		// After RequireAuth has run, inject TenantID onto the claims object stored at jwt_claims key.
		if v, ok := c.Get("jwt_claims"); ok && v != nil {
			if claims, ok2 := v.(*JWTClaims); ok2 {
				claims.TenantID = "tenant-e2e"
			}
		}
		c.Next()
	})
	r.Use(TenantContext(cfg))
	r.GET("/p", func(c *gin.Context) {
		// Verify Gin context
		if gv, ok := c.Get("tenant_id"); !ok || gv.(string) != "tenant-e2e" {
			t.Fatalf("gin tenant_id mismatch: %v", gv)
		}
		// Verify request context
		if cv := c.Request.Context().Value(multitenant.ContextKeyTenantID); cv == nil || cv.(string) != "tenant-e2e" {
			t.Fatalf("context tenant_id mismatch: %v", cv)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/p", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	// Note: no X-Tenant-ID header set
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
