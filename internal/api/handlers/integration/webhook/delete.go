package webhook

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteHandler struct{ logger *logging.Logger }

func NewDeleteHandler(logger *logging.Logger) *DeleteHandler { return &DeleteHandler{logger: logger} }

// HandleDelete is a generated stub for integration action webhook.delete (category=webhook)
func (h *DeleteHandler) HandleDelete(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.delete"})
}
