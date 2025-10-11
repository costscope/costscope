package webhook

import (
	"local/costscope/internal/core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateHandler struct{ logger *logging.Logger }

func NewCreateHandler(logger *logging.Logger) *CreateHandler { return &CreateHandler{logger: logger} }

// HandleCreate is a generated stub for integration action webhook.create (category=webhook)
func (h *CreateHandler) HandleCreate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented", "action_id": "webhook.create"})
}
