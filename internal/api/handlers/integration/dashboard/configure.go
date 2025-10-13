package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

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
