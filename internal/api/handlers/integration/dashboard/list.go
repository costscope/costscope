package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type ListHandler struct{ logger *logging.Logger }

func NewListHandler(logger *logging.Logger) *ListHandler { return &ListHandler{logger: logger} }

// HandleList is a generated stub for integration action dashboard.widget.list (category=dashboard)
func (h *ListHandler) HandleList(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.widget.list"})
}
