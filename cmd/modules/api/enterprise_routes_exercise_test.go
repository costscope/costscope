package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/handlers"
	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/security"
)

// ExerciseAllModuleRoutes iterates the route groups and calls each handler with a
// minimal gin test context. It skips websocket/streaming endpoints which require
// upgrades. This is intentionally light: handlers may return errors or 4xx,
// which is still useful to exercise code paths without starting servers.
func TestExerciseAllModuleRoutes_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logging.NewLogger(logging.LevelInfo)
	reg := handlers.NewEnterpriseRegistry(logger, nil, nil)

	store := security.NewFileRBACStore(t.TempDir())
	rbac := security.NewRBACService(store, logger)

	groups := buildModuleRouteGroups(reg, logger, rbac)

	for _, g := range groups {
		for _, r := range g.Routes {
			// skip websocket/streaming handlers that require upgrade
			if strings.Contains(r.Path, "ws") || strings.Contains(g.BasePath, "stream") || strings.Contains(r.Path, "stream") {
				continue
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			method := r.Method
			if method == "" {
				method = http.MethodGet
			}
			c.Request = httptest.NewRequest(method, g.BasePath+r.Path, nil)

			// ensure handler runs without panic
			invoked := false
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						t.Fatalf("panic invoking handler %s %s: %v", method, g.BasePath+r.Path, rec)
					}
				}()
				r.Handler(c)
				invoked = true
			}()
			if !invoked {
				t.Fatalf("handler %s %s was not invoked", method, g.BasePath+r.Path)
			}
		}
	}
}
