package dashboard

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InstallHandler struct{ logger *logging.Logger }

func NewInstallHandler(logger *logging.Logger) *InstallHandler {
	return &InstallHandler{logger: logger}
}

// HandleInstall is a generated stub for integration action dashboard.plugin.install (category=dashboard)
func (h *InstallHandler) HandleInstall(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.plugin.install"})
}
