package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/api/middleware"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/security"
)

func TestMorePermissionWrappedHandlers_RunWithoutRBACDeny(t *testing.T) {
	middleware.SetAuditModeForTests(true)
	defer middleware.SetAuditModeForTests(false)

	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	rbac := security.NewRBACService(security.NewFileRBACStore("data/security"), logger)

	g := gin.New()
	v1 := g.Group("/api/v1")
	v1.Use(func(c *gin.Context) { c.Next() })
	RegisterGinRouteGroups(v1, buildModuleRouteGroups(reg, logger, rbac))

	// 1) Analytics model train (permission wrapped)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/analytics/models/foo/train", strings.NewReader(`{}`))
	g.ServeHTTP(rr, req)
	if rr.Code != 200 && rr.Code != 202 && rr.Code != 400 {
		t.Fatalf("unexpected status for analytics train: %d", rr.Code)
	}

	// 2) Reports generate (permission wrapped)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/reports/generate", strings.NewReader(`{}`))
	g.ServeHTTP(rr2, req2)
	if rr2.Code != 200 && rr2.Code != 202 && rr2.Code != 400 {
		t.Fatalf("unexpected status for reports generate: %d", rr2.Code)
	}

	// 3) Streaming create job (permission wrapped)
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/v1/streaming/jobs", strings.NewReader(`{}`))
	g.ServeHTTP(rr3, req3)
	if rr3.Code != 200 && rr3.Code != 202 && rr3.Code != 400 {
		t.Fatalf("unexpected status for streaming create: %d", rr3.Code)
	}
}
