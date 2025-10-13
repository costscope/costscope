package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/security"
)

func TestInvokeAllModuleRoutes_WithAllAccessRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	store := security.NewFileRBACStore(t.TempDir())
	rbac := security.NewRBACService(store, logger)

	// Build a permissive role that grants the main actions used by module routes.
	perms := []security.Permission{
		{Resource: security.ResourceFocus, Action: security.ActionConvert},
		{Resource: security.ResourceFocus, Action: security.ActionValidate},
		{Resource: security.ResourceProviders, Action: security.ActionConnect},
		{Resource: security.ResourceAnalytics, Action: security.ActionForecast},
		{Resource: security.ResourceAnalytics, Action: security.ActionDetectAnomalies},
		{Resource: security.ResourceAnalytics, Action: security.ActionRecommendations},
		{Resource: security.ResourceAnalytics, Action: security.ActionTrends},
		{Resource: security.ResourceAnalytics, Action: security.ActionTrainModel},
		{Resource: security.ResourceReports, Action: security.ActionGenerate},
		{Resource: security.ResourceStreaming, Action: security.ActionCreateJob},
		{Resource: security.ResourceStreaming, Action: security.ActionStartJob},
		{Resource: security.ResourceStreaming, Action: security.ActionStopJob},
		{Resource: security.ResourceStreaming, Action: security.ActionDeleteJob},
	}
	if _, err := rbac.CreateRole("all-access", "test", perms); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	groups := buildModuleRouteGroups(reg, logger, rbac)

	for _, g := range groups {
		for _, r := range g.Routes {
			// skip websocket/streaming handlers that require upgrade
			if r.Path == "/jobs/:jobID" && g.BasePath == "/ws" {
				continue
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			method := r.Method
			if method == "" {
				method = http.MethodGet
			}
			c.Request = httptest.NewRequest(method, g.BasePath+r.Path, nil)
			c.Request.Header.Set("X-User-Roles", "all-access")

			// invoke and ensure no panic
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						t.Fatalf("panic invoking handler %s %s: %v", method, g.BasePath+r.Path, rec)
					}
				}()
				r.Handler(c)
			}()
		}
	}
}
