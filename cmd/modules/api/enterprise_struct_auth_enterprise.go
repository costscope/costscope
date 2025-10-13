//go:build enterprise
// +build enterprise

package api

import (
	"net/http"

	"github.com/costscope/costscope/internal/api/middleware"
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// registerEnterpriseStructuredAuthRoutes wires a small admin-only route using the structured AuthMiddleware.
func registerEnterpriseStructuredAuthRoutes(router *gin.Engine, logger *logging.Logger) {
	structuredAuth := middleware.NewAuthMiddleware(logger, enterpriseJwtSecret, enterpriseJwtIssuer)
	admin := router.Group("/api/v1/admin")
	admin.Use(structuredAuth.RequireAuth(), structuredAuth.RequireRole("admin"))
	admin.GET("/role-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
