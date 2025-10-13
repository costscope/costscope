package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type DisableHandler struct{ logger *logging.Logger }

func NewDisableHandler(logger *logging.Logger) *DisableHandler {
	return &DisableHandler{logger: logger}
}

// HandleDisable is a generated stub for integration action dashboard.plugin.disable (category=dashboard)
func (h *DisableHandler) HandleDisable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.plugin.disable"})
}
