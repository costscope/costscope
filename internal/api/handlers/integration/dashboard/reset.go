package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResetHandler struct{ logger *logging.Logger }

func NewResetHandler(logger *logging.Logger) *ResetHandler { return &ResetHandler{logger: logger} }

// HandleReset is a generated stub for integration action dashboard.config.reset (category=dashboard)
func (h *ResetHandler) HandleReset(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.config.reset"})
}
