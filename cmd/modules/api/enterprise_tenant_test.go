package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/costscope/costscope/internal/core/config"

	"github.com/gin-gonic/gin"
)

// Tests for multi-tenant middleware skeleton (TASK-MULTITENANT-SKEL).
// Ensures:
// 1. Disabled flag => middleware no-op, non-admin without header allowed.
// 2. Enabled flag + non-admin missing header => 400.
// 3. Enabled flag + non-admin with header => tenant_id set.
// 4. Enabled flag + admin without header => allowed (no tenant_id set).

func doTenantReq(r *gin.Engine, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/tenant-check", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestTenantMiddleware_Disabled_NoHeader_NoFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: false}}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("roles", []string{"user"}); c.Next() })
	r.Use(buildTenantMiddleware(cfg))
	r.GET("/tenant-check", func(c *gin.Context) {
		if _, has := c.Get("tenant_id"); has {
			t.Fatalf("tenant_id unexpectedly set while feature disabled")
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	w := doTenantReq(r, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestTenantMiddleware_Enabled_NonAdmin_NoHeader_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("roles", []string{"user"}); c.Next() })
	r.Use(buildTenantMiddleware(cfg))
	r.GET("/tenant-check", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	w := doTenantReq(r, nil)
	if w.Code != http.StatusBadRequest { // enforcement for non-admin
		t.Fatalf("expected 400 when missing tenant header, got %d", w.Code)
	}
}

func TestTenantMiddleware_Enabled_NonAdmin_WithHeader_SetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("roles", []string{"user"}); c.Next() })
	r.Use(buildTenantMiddleware(cfg))
	r.GET("/tenant-check", func(c *gin.Context) {
		v, ok := c.Get("tenant_id")
		if !ok || v.(string) != "tenant-123" {
			t.Fatalf("tenant_id not propagated; ok=%v val=%v", ok, v)
		}
		c.JSON(http.StatusOK, gin.H{"tenant_id": v})
	})
	w := doTenantReq(r, map[string]string{"X-Tenant-ID": "tenant-123"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestTenantMiddleware_Enabled_Admin_NoHeader_Allows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.ConsolidatedConfig{MultiTenant: config.UnifiedMultiTenantConfig{Enabled: true}}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("roles", []string{"admin"}); c.Next() })
	r.Use(buildTenantMiddleware(cfg))
	r.GET("/tenant-check", func(c *gin.Context) {
		if _, ok := c.Get("tenant_id"); ok {
			t.Fatalf("tenant_id should not be set when admin omitted header")
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	w := doTenantReq(r, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
