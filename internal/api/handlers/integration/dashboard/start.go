package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type StartHandler struct{ logger *logging.Logger }

func NewStartHandler(logger *logging.Logger) *StartHandler { return &StartHandler{logger: logger} }

// HandleStart is a generated stub for integration action dashboard.start (category=dashboard)
func (h *StartHandler) HandleStart(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.start"})
}
