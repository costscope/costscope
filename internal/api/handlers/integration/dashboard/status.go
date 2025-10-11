package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusHandler struct{ logger *logging.Logger }

func NewStatusHandler(logger *logging.Logger) *StatusHandler { return &StatusHandler{logger: logger} }

// HandleStatus is a generated stub for integration action dashboard.status (category=dashboard)
func (h *StatusHandler) HandleStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.status"})
}
