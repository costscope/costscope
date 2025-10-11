package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"local/costscope/internal/api/jobs"
	"local/costscope/internal/core/logging"
)

// =====================================================================================
// WebSocket Manager - Real-time Updates for FOCUS Operations
// =====================================================================================

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	JobID     string                 `json:"job_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	JobID  string
	Conn   *websocket.Conn
	Send   chan Message
	Logger *logging.Logger
}

// Manager manages WebSocket connections for real-time updates
type Manager struct {
	logger     *logging.Logger
	clients    map[string]*Client
	jobClients map[string][]*Client // Map job ID to clients
	mutex      sync.RWMutex
	upgrader   websocket.Upgrader
}

// NewManager creates a new WebSocket manager
func NewManager(logger *logging.Logger) *Manager {
	return &Manager{
		logger:     logger,
		clients:    make(map[string]*Client),
		jobClients: make(map[string][]*Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleConnectionHTTP handles new WebSocket connections using net/http and a given jobID.
// This mirrors HandleConnection (gin) but is framework-agnostic for simple mux integrations.
func (m *Manager) HandleConnectionHTTP(w http.ResponseWriter, r *http.Request, jobID string) {
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error(fmt.Sprintf("Failed to upgrade WebSocket connection: %s", err.Error()))
		return
	}

	// Create client
	clientID := fmt.Sprintf("%s_%d", jobID, time.Now().UnixNano())
	client := &Client{
		ID:     clientID,
		JobID:  jobID,
		Conn:   conn,
		Send:   make(chan Message, 256),
		Logger: m.logger,
	}

	// Register client
	m.registerClient(client)

	// Start client handlers
	go client.writePump(m)
	go client.readPump(m)

	// Send an initial welcome message
	select {
	case client.Send <- Message{Type: "welcome", JobID: jobID, Timestamp: time.Now(), Data: map[string]interface{}{"message": "connected"}}:
	default:
		// If channel is full for some reason, drop silently
	}

	m.logger.Info(fmt.Sprintf("WebSocket client connected: %s for job %s", clientID, jobID))
}

// HandleConnection handles new WebSocket connections
func (m *Manager) HandleConnection(c *gin.Context) {
	jobID := c.Param("jobID")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id is required"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := m.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		m.logger.Error(fmt.Sprintf("Failed to upgrade WebSocket connection: %s", err.Error()))
		return
	}

	// Create client
	clientID := fmt.Sprintf("%s_%d", jobID, time.Now().UnixNano())
	client := &Client{
		ID:     clientID,
		JobID:  jobID,
		Conn:   conn,
		Send:   make(chan Message, 256),
		Logger: m.logger,
	}

	// Register client
	m.registerClient(client)

	// Start client handlers
	go client.writePump(m)
	go client.readPump(m)

	m.logger.Info(fmt.Sprintf("WebSocket client connected: %s for job %s", clientID, jobID))
}

// registerClient registers a new client
func (m *Manager) registerClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.clients[client.ID] = client
	m.jobClients[client.JobID] = append(m.jobClients[client.JobID], client)
}

// unregisterClient unregisters a client
func (m *Manager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Remove from clients map
	delete(m.clients, client.ID)

	// Remove from job clients
	if clients, exists := m.jobClients[client.JobID]; exists {
		for i, c := range clients {
			if c.ID == client.ID {
				m.jobClients[client.JobID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		// Remove job entry if no clients left
		if len(m.jobClients[client.JobID]) == 0 {
			delete(m.jobClients, client.JobID)
		}
	}

	close(client.Send)
	if client.Conn != nil {
		if err := client.Conn.Close(); err != nil {
			m.logger.Error(fmt.Sprintf("Failed to close WebSocket connection: %s", err.Error()))
		}
	}

	m.logger.Info(fmt.Sprintf("WebSocket client disconnected: %s", client.ID))
}

// BroadcastJobUpdate broadcasts a job update to all connected clients for that job
func (m *Manager) BroadcastJobUpdate(jobID string, messageType string, data map[string]interface{}) {
	m.mutex.RLock()
	clients := m.jobClients[jobID]
	m.mutex.RUnlock()

	if len(clients) == 0 {
		return
	}

	message := Message{
		Type:      messageType,
		JobID:     jobID,
		Timestamp: time.Now(),
		Data:      data,
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, remove client
			m.unregisterClient(client)
		}
	}

	m.logger.Info(fmt.Sprintf("Broadcasted %s update to %d clients for job %s", messageType, len(clients), jobID))
}

// BroadcastJobProgress broadcasts job progress updates
func (m *Manager) BroadcastJobProgress(jobID string, progress *jobs.Progress) {
	data := map[string]interface{}{
		"progress": progress,
	}
	m.BroadcastJobUpdate(jobID, "progress", data)
}

// BroadcastJobStatus broadcasts job status changes
func (m *Manager) BroadcastJobStatus(jobID string, status jobs.JobStatus, result map[string]interface{}, error string) {
	data := map[string]interface{}{
		"status": status,
	}

	if result != nil {
		data["result"] = result
	}

	if error != "" {
		data["error"] = error
	}

	m.BroadcastJobUpdate(jobID, "status", data)
}

// GetClientCount returns the number of connected clients for a job
func (m *Manager) GetClientCount(jobID string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.jobClients[jobID])
}

// GetTotalClientCount returns the total number of connected clients
func (m *Manager) GetTotalClientCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.clients)
}

// =====================================================================================
// Client Methods
// =====================================================================================

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// readPump handles reading from the WebSocket connection
func (c *Client) readPump(manager *Manager) {
	defer func() {
		manager.unregisterClient(c)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger.Error(fmt.Sprintf("WebSocket error: %v", err))
			}
			break
		}
	}
}

// writePump handles writing to the WebSocket connection
func (c *Client) writePump(_ *Manager) { // TODO: Use manager for connection coordination
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Write message as JSON
			if err := json.NewEncoder(w).Encode(message); err != nil {
				c.Logger.Error(fmt.Sprintf("Failed to encode WebSocket message: %s", err.Error()))
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
