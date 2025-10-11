package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/websocket"
	"local/costscope/internal/core/logging"
)

// =====================================================================================
// WebSocket Handler - Real-time Communication
// =====================================================================================

// WebSocketHandler provides WebSocket connection management
type WebSocketHandler struct {
	logger    *logging.Logger
	wsManager *websocket.Manager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(logger *logging.Logger, wsManager *websocket.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		logger:    logger,
		wsManager: wsManager,
	}
}

// Connect establishes a WebSocket connection
func (h *WebSocketHandler) Connect(c *gin.Context) {
	// Extract user information from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required for WebSocket connection",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	h.wsManager.HandleConnection(c)

	h.logger.InfoWithFields("WebSocket connection established", map[string]interface{}{
		"user_id": userID,
	})
}
