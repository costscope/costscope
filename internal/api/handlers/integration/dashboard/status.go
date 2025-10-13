package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type StatusHandler struct{ logger *logging.Logger }

func NewStatusHandler(logger *logging.Logger) *StatusHandler { return &StatusHandler{logger: logger} }

// HandleStatus is a generated stub for integration action dashboard.status (category=dashboard)
func (h *StatusHandler) HandleStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.status"})
}
