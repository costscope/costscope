package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterGinRouteGroups_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	parent := r.Group("/api")

	groups := []GinRouteGroup{
		{
			BasePath: "/test",
			Routes: []GinRoute{
				{Method: "GET", Path: "/ping", Handler: func(c *gin.Context) { c.String(200, "ok") }},
			},
		},
	}

	RegisterGinRouteGroups(parent, groups)

	// Make a request to registered route
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200 OK from registered route, got %d", w.Code)
	}
}
