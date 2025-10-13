package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/handlers"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/security"
)

// TestExerciseModuleRoutes_WithMinimalPayloads iterates all module routes and
// invokes them with minimal JSON bodies for POST/PUT methods to exercise
// deeper branches without starting network servers.
func TestExerciseModuleRoutes_WithMinimalPayloads(t *testing.T) {
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

			var body string
			if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
				// Provide small useful payloads for common handlers
				switch {
				case strings.Contains(r.Path, "validate"):
					body = `{"input_path":"/tmp/input.csv"}`
				case strings.Contains(r.Path, "connect"):
					body = `{"credentials":{"token":"x"}}`
				case strings.Contains(r.Path, "recommendations"):
					body = `{}`
				default:
					body = `{}`
				}
				c.Request = httptest.NewRequest(method, g.BasePath+r.Path, strings.NewReader(body))
				c.Request.Header.Set("Content-Type", "application/json")
			} else {
				c.Request = httptest.NewRequest(method, g.BasePath+r.Path, nil)
			}

			// Ensure no panics and handler is invoked
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
