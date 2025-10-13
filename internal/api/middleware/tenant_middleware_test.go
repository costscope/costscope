package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/config"
	"github.com/costscope/costscope/internal/core/multitenant"

	"github.com/gin-gonic/gin"
)

func TestTenantContext_NoOpWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: false}}
	r.Use(TenantContext(cfg))
	r.GET("/", func(c *gin.Context) {
		if _, ok := c.Get("tenant_id"); ok {
			t.Fatalf("tenant_id should not be set when disabled")
		}
		// also ensure request context has no value
		if v := c.Request.Context().Value(multitenant.ContextKeyTenantID); v != nil {
			t.Fatalf("context tenant should be nil when disabled")
		}
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Tenant-ID", "t-a")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

func TestTenantContext_SetsWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}
	r.Use(TenantContext(cfg))
	r.GET("/", func(c *gin.Context) {
		v, ok := c.Get("tenant_id")
		if !ok {
			t.Fatalf("tenant_id not set in gin context")
		}
		if v.(string) != "tenant-abc" {
			t.Fatalf("unexpected gin tenant: %v", v)
		}
		// Check request context propagation
		cv := c.Request.Context().Value(multitenant.ContextKeyTenantID)
		if cv == nil || cv.(string) != "tenant-abc" {
			t.Fatalf("unexpected ctx tenant: %v", cv)
		}
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Tenant-ID", "tenant-abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}

func TestTenantContext_UsesJWTClaimsWhenHeaderMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}

	// Inject jwt_claims with tenant_id map before TenantContext runs
	r.Use(func(c *gin.Context) {
		c.Set("jwt_claims", map[string]any{"tenant_id": "tenant-from-claims"})
		c.Next()
	})
	r.Use(TenantContext(cfg))
	r.GET("/", func(c *gin.Context) {
		v, ok := c.Get("tenant_id")
		if !ok || v.(string) != "tenant-from-claims" {
			t.Fatalf("tenant_id not derived from jwt_claims map: %v", v)
		}
		cv := c.Request.Context().Value(multitenant.ContextKeyTenantID)
		if cv == nil || cv.(string) != "tenant-from-claims" {
			t.Fatalf("context tenant not derived from jwt_claims: %v", cv)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
