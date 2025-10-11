package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConfigureHandler struct{ logger *logging.Logger }

func NewConfigureHandler(logger *logging.Logger) *ConfigureHandler {
	return &ConfigureHandler{logger: logger}
}

// HandleConfigure is a generated stub for integration action dashboard.widget.configure (category=dashboard)
func (h *ConfigureHandler) HandleConfigure(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.widget.configure"})
}
