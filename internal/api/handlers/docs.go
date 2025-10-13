package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	docsb "github.com/costscope/costscope/internal/core/docs"
	"github.com/costscope/costscope/internal/core/logging"
)

// =====================================================================================
// Documentation Handler - OpenAPI/Swagger Documentation
// =====================================================================================

// DocsHandler provides API documentation endpoints
type DocsHandler struct {
	logger *logging.Logger
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler(logger *logging.Logger) *DocsHandler {
	return &DocsHandler{
		logger: logger,
	}
}

// GetDocumentation returns the OpenAPI specification
func (h *DocsHandler) GetDocumentation(c *gin.Context) {
	// TODO: Load actual OpenAPI spec from file
	base := docsb.GetBaseURL()
	spec := gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       "CostScope Enterprise API",
			"description": "Complete enterprise API for cloud cost management and FOCUS operations",
			"version":     "1.0.0",
		},
		"servers": []gin.H{
			{
				"url":         base + "/api/v1",
				"description": "Development server",
			},
		},
		"paths": gin.H{
			"/health": gin.H{
				"get": gin.H{
					"summary": "Health check",
					"responses": gin.H{
						"200": gin.H{
							"description": "Server is healthy",
						},
					},
				},
			},
		},
	}

	c.JSON(http.StatusOK, spec)
}

// ServeSwaggerUI serves the Swagger UI interface
func (h *DocsHandler) ServeSwaggerUI(c *gin.Context) {
	// TODO: Implement actual Swagger UI serving
	c.JSON(http.StatusOK, gin.H{
		"message": "Swagger UI would be served here",
		"path":    c.Param("filepath"),
	})
}
