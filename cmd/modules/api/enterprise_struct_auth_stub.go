//go:build !enterprise
// +build !enterprise

package api

import (
	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

// Intentional stub (enterprise gating): structured auth routes unavailable in community build.
func registerEnterpriseStructuredAuthRoutes(_ *gin.Engine, _ *logging.Logger) {}
