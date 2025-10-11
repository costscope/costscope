package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EnableHandler struct{ logger *logging.Logger }

func NewEnableHandler(logger *logging.Logger) *EnableHandler { return &EnableHandler{logger: logger} }

// HandleEnable is a generated stub for integration action dashboard.plugin.enable (category=dashboard)
func (h *EnableHandler) HandleEnable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.plugin.enable"})
}
