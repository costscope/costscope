package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/security"
)

func TestBuildModuleRouteGroups_HandlerPermissionAllow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	store := security.NewFileRBACStore(t.TempDir())
	rbac := security.NewRBACService(store, logger)

	// create a role that allows focus:convert
	perms := []security.Permission{{Resource: security.ResourceFocus, Action: security.ActionConvert}}
	if _, err := rbac.CreateRole("focus-convert-role", "test", perms); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	groups := buildModuleRouteGroups(reg, logger, rbac)

	// find the focus convert closure route
	var handler gin.HandlerFunc
	for _, g := range groups {
		if g.BasePath == "/focus" {
			for _, r := range g.Routes {
				if r.Path == "/convert" {
					handler = r.Handler
					break
				}
			}
		}
	}
	if handler == nil {
		t.Fatal("convert handler not found")
	}

	// build gin context and set header to role
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/focus/convert", nil)
	c.Request.Header.Set("X-User-Roles", "focus-convert-role")

	handler(c)
	// when allowed the handler may accept and return Accepted or 200; ensure not aborted with 403
	if w.Code == http.StatusForbidden {
		t.Fatalf("expected allowed role to not be forbidden")
	}
}

func TestBuildModuleRouteGroups_HandlerPermissionDeny(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	store := security.NewFileRBACStore(t.TempDir())
	rbac := security.NewRBACService(store, logger)

	// create a role that does NOT allow focus:convert
	perms := []security.Permission{{Resource: security.ResourceFocus, Action: security.ActionValidate}}
	if _, err := rbac.CreateRole("focus-validate-role", "test", perms); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	groups := buildModuleRouteGroups(reg, logger, rbac)

	// find the focus convert closure route
	var handler gin.HandlerFunc
	for _, g := range groups {
		if g.BasePath == "/focus" {
			for _, r := range g.Routes {
				if r.Path == "/convert" {
					handler = r.Handler
					break
				}
			}
		}
	}
	if handler == nil {
		t.Fatal("convert handler not found")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/focus/convert", nil)
	c.Request.Header.Set("X-User-Roles", "focus-validate-role")

	handler(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden when role lacks permission, got %d", w.Code)
	}
}
