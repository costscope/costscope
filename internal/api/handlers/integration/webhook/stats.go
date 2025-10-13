package webhook

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct{ logger *logging.Logger }

func NewStatsHandler(logger *logging.Logger) *StatsHandler { return &StatsHandler{logger: logger} }

// HandleStats is a generated stub for integration action webhook.delivery.stats (category=webhook)
func (h *StatsHandler) HandleStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.delivery.stats"})
}
