package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/costscope/costscope/internal/api/jobs"
	"github.com/costscope/costscope/internal/core/logging"
)

func TestNewManager(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	manager := NewManager(logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.clients)
	assert.NotNil(t, manager.jobClients)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, 1024, manager.upgrader.ReadBufferSize)
	assert.Equal(t, 1024, manager.upgrader.WriteBufferSize)
}

func TestManagerClientCounts(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	manager := NewManager(logger)

	// Initially no clients
	assert.Equal(t, 0, manager.GetTotalClientCount())
	assert.Equal(t, 0, manager.GetClientCount("job1"))

	// Create test clients
	client1 := &Client{
		ID:     "client1",
		JobID:  "job1",
		Conn:   nil, // Mock connection for testing
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	client2 := &Client{
		ID:     "client2",
		JobID:  "job1",
		Conn:   nil,
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	client3 := &Client{
		ID:     "client3",
		JobID:  "job2",
		Conn:   nil,
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	// Register clients
	manager.registerClient(client1)
	manager.registerClient(client2)
	manager.registerClient(client3)

	// Check counts
	assert.Equal(t, 3, manager.GetTotalClientCount())
	assert.Equal(t, 2, manager.GetClientCount("job1"))
	assert.Equal(t, 1, manager.GetClientCount("job2"))

	// Unregister a client
	manager.unregisterClient(client1)
	assert.Equal(t, 2, manager.GetTotalClientCount())
	assert.Equal(t, 1, manager.GetClientCount("job1"))
	assert.Equal(t, 1, manager.GetClientCount("job2"))

	// Unregister all clients from job1
	manager.unregisterClient(client2)
	assert.Equal(t, 1, manager.GetTotalClientCount())
	assert.Equal(t, 0, manager.GetClientCount("job1"))
	assert.Equal(t, 1, manager.GetClientCount("job2"))
}

func TestMessageStructure(t *testing.T) {
	message := Message{
		Type:      "progress",
		JobID:     "job_123",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"progress":    50,
			"total_files": 100,
			"status":      "processing",
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON deserialization
	var deserializedMessage Message
	err = json.Unmarshal(jsonData, &deserializedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message.Type, deserializedMessage.Type)
	assert.Equal(t, message.JobID, deserializedMessage.JobID)
	// JSON unmarshaling converts numbers to float64
	assert.Equal(t, float64(50), deserializedMessage.Data["progress"])
}

func TestBroadcastJobUpdate(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	manager := NewManager(logger)

	// Create test clients for the same job
	client1 := &Client{
		ID:     "client1",
		JobID:  "job1",
		Conn:   nil,
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	client2 := &Client{
		ID:     "client2",
		JobID:  "job1",
		Conn:   nil,
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	// Register clients
	manager.registerClient(client1)
	manager.registerClient(client2)

	// Broadcast update
	testData := map[string]interface{}{
		"status":   "processing",
		"progress": 25,
	}
	manager.BroadcastJobUpdate("job1", "progress", testData)

	// Check that both clients received the message
	select {
	case msg := <-client1.Send:
		assert.Equal(t, "progress", msg.Type)
		assert.Equal(t, "job1", msg.JobID)
		assert.Equal(t, 25, msg.Data["progress"])
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 should have received message")
	}

	select {
	case msg := <-client2.Send:
		assert.Equal(t, "progress", msg.Type)
		assert.Equal(t, "job1", msg.JobID)
		assert.Equal(t, 25, msg.Data["progress"])
	case <-time.After(100 * time.Millisecond):
		t.Error("Client2 should have received message")
	}
}

func TestBroadcastJobProgress(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	manager := NewManager(logger)

	client := &Client{
		ID:     "client1",
		JobID:  "job1",
		Conn:   nil,
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	manager.registerClient(client)

	// Create progress data
	progress := &jobs.Progress{
		Current:    75,
		Total:      100,
		Percentage: 75.0,
		Message:    "Processing file data.csv",
		Stage:      "data_processing",
		UpdatedAt:  time.Now(),
	}

	manager.BroadcastJobProgress("job1", progress)

	// Verify message received
	select {
	case msg := <-client.Send:
		assert.Equal(t, "progress", msg.Type)
		assert.Equal(t, "job1", msg.JobID)
		assert.NotNil(t, msg.Data["progress"])
	case <-time.After(100 * time.Millisecond):
		t.Error("Client should have received progress message")
	}
}

func TestBroadcastJobStatus(t *testing.T) {
	logger := logging.NewLogger(logging.LevelInfo)
	manager := NewManager(logger)

	client := &Client{
		ID:     "client1",
		JobID:  "job1",
		Conn:   nil,
		Send:   make(chan Message, 256),
		Logger: logger,
	}

	manager.registerClient(client)

	// Test successful status
	result := map[string]interface{}{
		"files_processed": 150,
		"output_file":     "/tmp/result.parquet",
	}
	manager.BroadcastJobStatus("job1", jobs.StatusCompleted, result, "")

	select {
	case msg := <-client.Send:
		assert.Equal(t, "status", msg.Type)
		assert.Equal(t, "job1", msg.JobID)
		assert.Equal(t, jobs.StatusCompleted, msg.Data["status"])
		assert.Equal(t, 150, msg.Data["result"].(map[string]interface{})["files_processed"])
		assert.Nil(t, msg.Data["error"])
	case <-time.After(100 * time.Millisecond):
		t.Error("Client should have received status message")
	}

	// Test error status
	manager.BroadcastJobStatus("job1", jobs.StatusFailed, nil, "File not found")

	select {
	case msg := <-client.Send:
		assert.Equal(t, "status", msg.Type)
		assert.Equal(t, jobs.StatusFailed, msg.Data["status"])
		assert.Equal(t, "File not found", msg.Data["error"])
		assert.Nil(t, msg.Data["result"])
	case <-time.After(100 * time.Millisecond):
		t.Error("Client should have received error status message")
	}
}

func TestWebSocketUpgrade(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test server with WebSocket support
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple WebSocket upgrade test
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				t.Logf("Failed to close connection: %v", err)
			}
		}()

		// Echo server for testing
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				t.Logf("Failed to write message: %v", err)
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Test WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)

	if conn != nil {
		// Test sending and receiving a message
		testMessage := "test message"
		err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
		assert.NoError(t, err)

		_, receivedMessage, err := conn.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, testMessage, string(receivedMessage))

		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}
}

func TestMessageTypes(t *testing.T) {
	testCases := []struct {
		name        string
		messageType string
		jobID       string
		data        map[string]interface{}
		shouldPass  bool
	}{
		{
			name:        "Progress Message",
			messageType: "progress",
			jobID:       "job_123",
			data:        map[string]interface{}{"progress": 50, "total": 100},
			shouldPass:  true,
		},
		{
			name:        "Status Message",
			messageType: "status",
			jobID:       "job_456",
			data:        map[string]interface{}{"status": "completed"},
			shouldPass:  true,
		},
		{
			name:        "Error Message",
			messageType: "error",
			jobID:       "job_789",
			data:        map[string]interface{}{"error": "File not found"},
			shouldPass:  true,
		},
		{
			name:        "Empty Type",
			messageType: "",
			jobID:       "job_000",
			data:        map[string]interface{}{},
			shouldPass:  false,
		},
		{
			name:        "Empty Job ID",
			messageType: "progress",
			jobID:       "",
			data:        map[string]interface{}{},
			shouldPass:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := Message{
				Type:      tc.messageType,
				JobID:     tc.jobID,
				Data:      tc.data,
				Timestamp: time.Now(),
			}

			if tc.shouldPass {
				assert.NotEmpty(t, message.Type)
				assert.NotEmpty(t, message.JobID)
				assert.NotNil(t, message.Data)
				assert.False(t, message.Timestamp.IsZero())
			} else {
				hasValidType := message.Type != ""
				hasValidJobID := message.JobID != ""
				assert.False(t, hasValidType && hasValidJobID)
			}
		})
	}
}
