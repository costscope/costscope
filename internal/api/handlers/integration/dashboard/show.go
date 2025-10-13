package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type ShowHandler struct{ logger *logging.Logger }

func NewShowHandler(logger *logging.Logger) *ShowHandler { return &ShowHandler{logger: logger} }

// HandleShow is a generated stub for integration action dashboard.config.show (category=dashboard)
func (h *ShowHandler) HandleShow(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.config.show"})
}
