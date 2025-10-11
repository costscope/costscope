package webhook

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TriggerHandler struct{ logger *logging.Logger }

func NewTriggerHandler(logger *logging.Logger) *TriggerHandler {
	return &TriggerHandler{logger: logger}
}

// HandleTrigger is a generated stub for integration action webhook.event.trigger (category=webhook)
func (h *TriggerHandler) HandleTrigger(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.event.trigger"})
}
