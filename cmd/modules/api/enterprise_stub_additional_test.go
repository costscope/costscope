package api

import (
	"testing"

	"local/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

func TestRegisterEnterpriseStructuredAuthRoutes_NoOp(t *testing.T) {
	// Should be a no-op in non-enterprise build; ensure it doesn't panic
	r := gin.New()
	logger := logging.NewLogger(logging.LevelInfo)
	registerEnterpriseStructuredAuthRoutes(r, logger)
}

func TestRunAPIServer_MissingJWT_ReturnsError_Alternate(t *testing.T) {
	// Force missing jwtSecret and ensure runAPIServer returns an error
	prev := jwtSecret
	jwtSecret = ""
	defer func() { jwtSecret = prev }()

	cmd := BuildAPICommand()
	if err := runAPIServer(cmd, []string{}); err == nil {
		t.Fatalf("expected error when jwt secret missing, got nil")
	}
}
