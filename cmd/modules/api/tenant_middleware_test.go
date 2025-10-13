package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/core/config"
)

func TestBuildTenantMiddleware_RequiresTenantWhenMultiTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.ConsolidatedConfig{}
	cfg.MultiTenant.Enabled = true

	mw := buildTenantMiddleware(cfg)

	r := gin.New()
	r.Use(mw)
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest when tenant missing in multi-tenant mode, got %d", w.Code)
	}
}
