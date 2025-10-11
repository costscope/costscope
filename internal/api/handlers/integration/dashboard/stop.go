package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StopHandler struct{ logger *logging.Logger }

func NewStopHandler(logger *logging.Logger) *StopHandler { return &StopHandler{logger: logger} }

// HandleStop is a generated stub for integration action dashboard.stop (category=dashboard)
func (h *StopHandler) HandleStop(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.stop"})
}
