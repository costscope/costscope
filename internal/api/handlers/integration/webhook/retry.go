package webhook

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type RetryHandler struct{ logger *logging.Logger }

func NewRetryHandler(logger *logging.Logger) *RetryHandler { return &RetryHandler{logger: logger} }

// HandleRetry is a generated stub for integration action webhook.delivery.retry (category=webhook)
func (h *RetryHandler) HandleRetry(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.delivery.retry"})
}
