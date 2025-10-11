package api

import (
	"testing"

	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

func TestRegisterEnterpriseStructuredAuthRoutes_NoopExec(t *testing.T) {
	// Ensure the stubbed noop function for non-enterprise builds is callable and harmless
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	registerEnterpriseStructuredAuthRoutes(router, logger)
}
