package webhook

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type TestHandler struct{ logger *logging.Logger }

func NewTestHandler(logger *logging.Logger) *TestHandler { return &TestHandler{logger: logger} }

// HandleTest is a generated stub for integration action webhook.test (category=webhook)
func (h *TestHandler) HandleTest(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.test"})
}
