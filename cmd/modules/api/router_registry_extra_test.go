package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterGinRouteGroups_MethodMapping_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// create a child group via engine to get a valid *gin.RouterGroup
	g := router.Group("/api")

	groups := []GinRouteGroup{
		{
			BasePath: "/test",
			Routes: []GinRoute{
				{Method: http.MethodGet, Path: "/g", Handler: func(c *gin.Context) {}},
				{Method: http.MethodPost, Path: "/p", Handler: func(c *gin.Context) {}},
				{Method: http.MethodPut, Path: "/u", Handler: func(c *gin.Context) {}},
			},
		},
	}

	// Ensure registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterGinRouteGroups panicked: %v", r)
		}
	}()
	RegisterGinRouteGroups(g, groups)
}
