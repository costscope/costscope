package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// streamRespond writes a standard JSON payload for simple id/status responses.
// Keeping it unexported and colocated avoids repetitive boilerplate in handlers.
func streamRespond(c *gin.Context, id, status string) {
	c.JSON(http.StatusOK, gin.H{"id": id, "status": status})
}
