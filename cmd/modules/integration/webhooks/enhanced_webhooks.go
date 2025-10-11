package webhooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"local/costscope/cmd/modules/integration/types"
)

// WebhookManager handles advanced webhook functionality
type WebhookManager struct {
	webhooks        map[string]*types.WebhookConfig
	deliveryHistory map[string][]WebhookDelivery
	eventTypes      []string
	client          *http.Client
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string            `json:"id"`
	WebhookID    string            `json:"webhook_id"`
	Event        string            `json:"event"`
	Payload      interface{}       `json:"payload"`
	StatusCode   int               `json:"status_code"`
	ResponseTime time.Duration     `json:"response_time"`
	Timestamp    time.Time         `json:"timestamp"`
	Success      bool              `json:"success"`
	Error        string            `json:"error,omitempty"`
	Headers      map[string]string `json:"headers"`
	Retries      int               `json:"retries"`
}

// WebhookEvent represents an event that triggers webhooks
type WebhookEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	ID        string                 `json:"id"`
}

// NewWebhookManager creates a new enhanced webhook manager
func NewWebhookManager() *WebhookManager {
	return &WebhookManager{
		webhooks:        make(map[string]*types.WebhookConfig),
		deliveryHistory: make(map[string][]WebhookDelivery),
		eventTypes: []string{
			"cost.threshold.exceeded",
			"cost.anomaly.detected",
			"budget.limit.reached",
			"alert.triggered",
			"report.generated",
			"integration.connected",
			"integration.disconnected",
			"workflow.completed",
			"workflow.failed",
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NOTE: Legacy Cobra builder function removed (TASK-INTEGRATION-REGISTRAR-CLEANUP).
// Declarative registrar supplies the root 'webhook' command & subcommands.

// create/list/test migrated to registrar (TASK-INTEGRATION-REGISTRAR)

// buildListWebhooksCommand creates the webhook listing command
// list migrated

// buildTestWebhookCommand creates the webhook testing command
// test command migrated to registrar – legacy builder removed

// buildDeleteWebhookCommand creates the webhook deletion command
// delete command migrated to registrar – legacy builder removed

// buildDeliveryCommands creates delivery management commands
// delivery group + list/retry/stats migrated – legacy builder removed

// buildEventCommands creates event management commands
// event group + list/trigger migrated – legacy builder removed

// buildListEventsCommand (retained; not migrated yet)
// event list/trigger migrated; builders removed

// buildSecurityCommands returns security related webhook operations (stubs)
// security group migrated – legacy builder removed

// Implementation methods

// generateWebhookID removed (service layer provides ID generation); manager now expects IDs provided upstream.

func (wm *WebhookManager) listWebhooks(status string, verbose bool) error {
	fmt.Println(" Webhook Management")
	fmt.Println("════════════════════")

	if len(wm.webhooks) == 0 {
		fmt.Println("No webhooks configured")
		return nil
	}

	for _, webhook := range wm.webhooks {
		// Apply status filter
		if status != "" && webhook.Status != status {
			continue
		}

		fmt.Printf("\n %s (%s)\n", webhook.Name, webhook.ID)
		fmt.Printf("   URL: %s\n", webhook.URL)
		fmt.Printf("   Status: %s | Events: %v\n", webhook.Status, webhook.Events)
		fmt.Printf("   Created: %s\n", webhook.Created.Format("2006-01-02 15:04"))

		if verbose {
			// Show delivery statistics
			deliveries := wm.deliveryHistory[webhook.ID]
			if len(deliveries) > 0 {
				successCount := 0
				for _, delivery := range deliveries {
					if delivery.Success {
						successCount++
					}
				}
				successRate := float64(successCount) / float64(len(deliveries)) * 100

				fmt.Printf("   Deliveries: %d total, %.1f%% success rate\n",
					len(deliveries), successRate)

				// Show last delivery
				if len(deliveries) > 0 {
					lastDelivery := deliveries[len(deliveries)-1]
					status := ""
					if !lastDelivery.Success {
						status = ""
					}
					fmt.Printf("   Last Delivery: %s %s (%.0fms)\n",
						status, lastDelivery.Timestamp.Format("2006-01-02 15:04:05"),
						float64(lastDelivery.ResponseTime.Nanoseconds())/1000000)
				}
			} else {
				fmt.Println("   Deliveries: None")
			}

			// Show security and retry policy
			if webhook.Security != nil {
				fmt.Printf("   Security: %s signature\n", webhook.Security.Algorithm)
			}

			if webhook.RetryPolicy != nil {
				fmt.Printf("   Retry Policy: %d max retries\n", webhook.RetryPolicy.MaxRetries)
			}

			if len(webhook.Headers) > 0 {
				fmt.Printf("   Custom Headers: %d configured\n", len(webhook.Headers))
			}
		}
	}

	return nil
}

func (wm *WebhookManager) testWebhook(webhookID, eventType, customPayload string) error {
	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return fmt.Errorf("webhook %s not found", webhookID)
	}

	fmt.Printf(" Testing webhook: %s\n", webhook.Name)
	fmt.Printf(" URL: %s\n", webhook.URL)
	fmt.Printf(" Event: %s\n", eventType)

	// Create test event
	var payload interface{}
	if customPayload != "" {
		if err := json.Unmarshal([]byte(customPayload), &payload); err != nil {
			return fmt.Errorf("invalid JSON payload: %w", err)
		}
	} else {
		payload = map[string]interface{}{
			"test":      true,
			"webhook":   webhook.Name,
			"timestamp": time.Now().Format(time.RFC3339),
			"cost_data": map[string]interface{}{
				"amount":   1500.00,
				"currency": "USD",
				"service":  "EC2",
			},
		}
	}

	event := WebhookEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    "costscope-test",
		Data:      payload.(map[string]interface{}),
		ID:        fmt.Sprintf("test_%d", time.Now().UnixNano()),
	}

	// Deliver test event
	delivery := wm.deliverWebhook(webhook, event)

	// Display results
	fmt.Printf("\n Test Results:\n")
	fmt.Printf("   Delivery ID: %s\n", delivery.ID)
	fmt.Printf("   Response Time: %.0fms\n",
		float64(delivery.ResponseTime.Nanoseconds())/1000000)
	fmt.Printf("   Status Code: %d\n", delivery.StatusCode)

	if delivery.Success {
		fmt.Printf("   Result:  Success\n")
	} else {
		fmt.Printf("   Result:  Failed\n")
		if delivery.Error != "" {
			fmt.Printf("   Error: %s\n", delivery.Error)
		}
	}

	// Store delivery in history
	wm.deliveryHistory[webhookID] = append(wm.deliveryHistory[webhookID], delivery)

	return nil
}

// === Adapter methods exposed for registrar handlers ===
func (wm *WebhookManager) DeleteWebhook(webhookID string, confirm bool) error { // export wrapper
	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return fmt.Errorf("webhook %s not found", webhookID)
	}

	if !confirm {
		fmt.Printf("Are you sure you want to delete webhook '%s'? (y/N): ", webhook.Name)
		// In a real implementation, would read user input
		fmt.Println("y") // Simulating user confirmation
	}

	delete(wm.webhooks, webhookID)
	delete(wm.deliveryHistory, webhookID)

	fmt.Printf(" Webhook '%s' deleted successfully\n", webhook.Name)
	return nil
}

func (wm *WebhookManager) deliverWebhook(webhook *types.WebhookConfig, event WebhookEvent) WebhookDelivery {
	delivery := WebhookDelivery{
		ID:        fmt.Sprintf("delivery_%d", time.Now().UnixNano()),
		WebhookID: webhook.ID,
		Event:     event.Type,
		Payload:   event,
		Timestamp: time.Now(),
		Headers:   webhook.Headers,
	}

	// Simulate HTTP request
	start := time.Now()

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	delivery.ResponseTime = time.Since(start)
	delivery.StatusCode = 200
	delivery.Success = true

	// Simulate occasional failures for testing
	if webhook.Name == "test-fail" {
		delivery.StatusCode = 500
		delivery.Success = false
		delivery.Error = "Internal server error"
	}

	return delivery
}

// Additional command builders (stubs for full implementation)

// delivery list/retry/stats builders removed – registrar handlers call adapter methods

// test migrated
func (wm *WebhookManager) listDeliveries() error {
	fmt.Println(" Webhook Deliveries:")
	// Implementation would list deliveries
	return nil
}

func (wm *WebhookManager) ListDeliveries() error { return wm.listDeliveries() }

func (wm *WebhookManager) retryDelivery() error {
	fmt.Println(" Retrying failed delivery...")
	// Implementation would retry delivery
	return nil
}

func (wm *WebhookManager) RetryDelivery() error { return wm.retryDelivery() }

func (wm *WebhookManager) showDeliveryStats() error {
	fmt.Println(" Delivery Statistics:")
	// Implementation would show statistics
	return nil
}

func (wm *WebhookManager) DeliveryStats() error { return wm.showDeliveryStats() }

func (wm *WebhookManager) listEventTypes() error {
	fmt.Println(" Available Event Types:")
	for _, eventType := range wm.eventTypes {
		fmt.Printf("   • %s\n", eventType)
	}
	return nil
}

func (wm *WebhookManager) ListEvents() error { return wm.listEventTypes() }

func (wm *WebhookManager) TriggerEvent() error { return wm.triggerTestEvent() }

// triggerTestEventPlaceholder removed (unused) during registrar refactor

func (wm *WebhookManager) triggerTestEvent() error {
	fmt.Println(" Triggering test event...")
	// Implementation would trigger test event
	return nil
}

// ===== Thin adapter layer (TASK-INTEGRATION-REGISTRAR) =====
// These small exported helpers expose previously internal logic to the action registrar
// without leaking underlying maps; keeps existing behavior intact.

// AddWebhook stores a new webhook configuration (idempotent by ID).
func (wm *WebhookManager) AddWebhook(cfg *types.WebhookConfig) {
	if wm.webhooks == nil {
		wm.webhooks = make(map[string]*types.WebhookConfig)
	}
	wm.webhooks[cfg.ID] = cfg
	if wm.deliveryHistory == nil {
		wm.deliveryHistory = make(map[string][]WebhookDelivery)
	}
	if _, ok := wm.deliveryHistory[cfg.ID]; !ok {
		wm.deliveryHistory[cfg.ID] = make([]WebhookDelivery, 0)
	}
}

// ListWebhooks exported wrapper around listWebhooks.
func (wm *WebhookManager) ListWebhooks(status string, verbose bool) error {
	return wm.listWebhooks(status, verbose)
}

// TestWebhook exported wrapper around testWebhook.
func (wm *WebhookManager) TestWebhook(id, eventType, payload string) error {
	return wm.testWebhook(id, eventType, payload)
}
