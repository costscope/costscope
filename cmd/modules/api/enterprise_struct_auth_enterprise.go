//go:build enterprise
// +build enterprise

package api

import (
	"net/http"

	"local/costscope/internal/api/middleware"
	"local/costscope/internal/core/logging"

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
