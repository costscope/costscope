package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RemoveHandler struct{ logger *logging.Logger }

func NewRemoveHandler(logger *logging.Logger) *RemoveHandler { return &RemoveHandler{logger: logger} }

// HandleRemove is a generated stub for integration action dashboard.widget.remove (category=dashboard)
func (h *RemoveHandler) HandleRemove(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.widget.remove"})
}
