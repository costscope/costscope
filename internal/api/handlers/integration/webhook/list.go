package webhook

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListHandler struct{ logger *logging.Logger }

func NewListHandler(logger *logging.Logger) *ListHandler { return &ListHandler{logger: logger} }

// HandleList is a generated stub for integration action webhook.list (category=webhook)
func (h *ListHandler) HandleList(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.list"})
}
