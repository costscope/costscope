package dashboard

import (
	"net/http"

	"github.com/costscope/costscope/internal/core/logging"

	"github.com/gin-gonic/gin"
)

type AddHandler struct{ logger *logging.Logger }

func NewAddHandler(logger *logging.Logger) *AddHandler { return &AddHandler{logger: logger} }

// HandleAdd is a generated stub for integration action dashboard.widget.add (category=dashboard)
func (h *AddHandler) HandleAdd(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "dashboard.widget.add"})
}
