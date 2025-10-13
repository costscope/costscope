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

func TestPermissionWrappedHandlers_RunWithoutRBACDeny(t *testing.T) {
	// enable audit mode so RequirePermission soft-denies instead of aborting
	middleware.SetAuditModeForTests(true)
	defer middleware.SetAuditModeForTests(false)

	logger := logging.NewLogger(logging.LevelInfo)
	// minimal handlers registry using mocks
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)
	// file-backed RBAC store used by buildModuleRouteGroups; pass a minimal RBACService
	rbac := security.NewRBACService(security.NewFileRBACStore("data/security"), logger)

	g := gin.New()
	v1 := g.Group("/api/v1")
	v1.Use(func(c *gin.Context) { c.Next() }) // noop auth for tests
	// register groups
	groups := buildModuleRouteGroups(reg, logger, rbac)
	RegisterGinRouteGroups(v1, groups)

	// Call focus validate (permission-wrapped POST /focus/validate)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/focus/validate", strings.NewReader(`{}`))
	g.ServeHTTP(rr, req)
	if rr.Code != 200 && rr.Code != 400 {
		t.Fatalf("unexpected status for focus validate: %d", rr.Code)
	}

	// Call providers test (permission-wrapped POST /providers/aws/test)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/providers/aws/test", strings.NewReader(`{}`))
	g.ServeHTTP(rr2, req2)
	if rr2.Code != 200 && rr2.Code != 400 && rr2.Code != 404 {
		t.Fatalf("unexpected status for providers test: %d", rr2.Code)
	}
}
