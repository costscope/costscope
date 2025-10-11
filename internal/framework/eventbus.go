// Package framework - Event Bus for inter-module communication
package framework

import (
	"context"
	"fmt"
	"sync"
	"time"

	"local/costscope/internal/core/logging"
)

// Health status constants
const (
	HealthStatusHealthy  = "healthy"
	HealthStatusDegraded = "degraded"
	HealthStatusStopped  = "stopped"
)

// EventHandler is a function that handles events
type EventHandler func(event Event)

// Event represents an event in the system
type Event struct {
	Name      string                 `json:"name"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	ID        string                 `json:"id"`
}

// EventSubscription represents a subscription to events
type EventSubscription struct {
	ID      string
	Pattern string
	Handler EventHandler
	Once    bool
	Fired   bool
	Filter  func(Event) bool
}

// EventBus provides event-driven communication between components
type EventBus struct {
	subscribers map[string][]*EventSubscription
	mu          sync.RWMutex
	running     bool
	eventQueue  chan Event
	workers     int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]*EventSubscription),
		eventQueue:  make(chan Event, 1000),
		workers:     4,
	}
}

// Start starts the event bus
func (eb *EventBus) Start(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.running {
		return nil
	}

	eb.ctx, eb.cancel = context.WithCancel(ctx)

	// Start worker goroutines
	for i := 0; i < eb.workers; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}

	eb.running = true
	return nil
}

// Stop stops the event bus
func (eb *EventBus) Stop(ctx context.Context) error {
	eb.mu.Lock()
	if !eb.running {
		eb.mu.Unlock()
		return nil
	}

	eb.cancel()
	close(eb.eventQueue)
	eb.running = false
	eb.mu.Unlock()

	// Wait for workers to finish
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// worker processes events from the queue
func (eb *EventBus) worker(id int) {
	defer eb.wg.Done()

	for {
		select {
		case event, ok := <-eb.eventQueue:
			if !ok {
				return
			}
			eb.processEvent(event, id)
		case <-eb.ctx.Done():
			return
		}
	}
}

// processEvent processes a single event
func (eb *EventBus) processEvent(event Event, workerID int) {
	eb.mu.RLock()
	subscribers := eb.subscribers[event.Name]

	// Also check for wildcard subscribers
	wildcardSubscribers := eb.subscribers["*"]
	allSubscribers := make([]*EventSubscription, 0, len(subscribers)+len(wildcardSubscribers))
	allSubscribers = append(allSubscribers, subscribers...)
	allSubscribers = append(allSubscribers, wildcardSubscribers...)
	eb.mu.RUnlock()

	var toRemove []string
	for _, subscription := range allSubscribers {
		// Apply filter if present
		if subscription.Filter != nil && !subscription.Filter(event) {
			continue
		}

		// Handle event, but for one-time subscriptions ensure handler runs at most once
		alreadyFired := false
		if subscription.Once {
			// Check and set Fired under mutex to avoid races when multiple workers process events
			eb.mu.Lock()
			if subscription.Fired {
				alreadyFired = true
			} else {
				subscription.Fired = true
			}
			eb.mu.Unlock()
		}
		if alreadyFired {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logging.GetLogger().ErrorWithFields("Event handler panic", map[string]interface{}{"worker_id": workerID, "event": event.Name, "panic": r})
				}
			}()
			subscription.Handler(event)
		}()

		// Mark for removal if it's a one-time subscription
		if subscription.Once {
			toRemove = append(toRemove, subscription.ID)
		}
	}

	// Remove one-time subscriptions
	if len(toRemove) > 0 {
		eb.mu.Lock()
		for _, id := range toRemove {
			eb.removeSubscription(id)
		}
		eb.mu.Unlock()
	}
}

// Emit emits an event
func (eb *EventBus) Emit(name string, data map[string]interface{}) {
	if !eb.isRunning() {
		return
	}

	event := Event{
		Name:      name,
		Data:      data,
		Timestamp: time.Now(),
		Source:    "unknown",
		ID:        eb.generateEventID(),
	}

	select {
	case eb.eventQueue <- event:
	default:
		// Queue is full, drop event
		logging.GetLogger().WarnWithFields("Event queue full, dropping event", map[string]interface{}{"event": name})
	}
}

// EmitSync emits an event synchronously
func (eb *EventBus) EmitSync(name string, data map[string]interface{}) {
	event := Event{
		Name:      name,
		Data:      data,
		Timestamp: time.Now(),
		Source:    "unknown",
		ID:        eb.generateEventID(),
	}

	eb.processEvent(event, -1) // -1 indicates synchronous processing
}

// Subscribe subscribes to events by name
func (eb *EventBus) Subscribe(eventName string, handler EventHandler) string {
	return eb.SubscribeWithOptions(eventName, handler, SubscriptionOptions{})
}

// SubscriptionOptions provides options for event subscriptions
type SubscriptionOptions struct {
	Once   bool
	Filter func(Event) bool
}

// SubscribeWithOptions subscribes to events with options
func (eb *EventBus) SubscribeWithOptions(eventName string, handler EventHandler, options SubscriptionOptions) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscription := &EventSubscription{
		ID:      eb.generateSubscriptionID(),
		Pattern: eventName,
		Handler: handler,
		Once:    options.Once,
		Filter:  options.Filter,
	}

	eb.subscribers[eventName] = append(eb.subscribers[eventName], subscription)
	return subscription.ID
}

// SubscribeOnce subscribes to an event for one-time handling
func (eb *EventBus) SubscribeOnce(eventName string, handler EventHandler) string {
	return eb.SubscribeWithOptions(eventName, handler, SubscriptionOptions{Once: true})
}

// Unsubscribe removes a subscription
func (eb *EventBus) Unsubscribe(subscriptionID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.removeSubscription(subscriptionID)
}

// removeSubscription removes a subscription (internal, assumes lock is held)
func (eb *EventBus) removeSubscription(subscriptionID string) {
	for eventName, subscriptions := range eb.subscribers {
		for i, subscription := range subscriptions {
			if subscription.ID == subscriptionID {
				// Remove subscription from slice
				eb.subscribers[eventName] = append(subscriptions[:i], subscriptions[i+1:]...)
				return
			}
		}
	}
}

// UnsubscribeAll removes all subscriptions for an event
func (eb *EventBus) UnsubscribeAll(eventName string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	delete(eb.subscribers, eventName)
}

// Clear removes all subscriptions
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers = make(map[string][]*EventSubscription)
}

// GetSubscriptions returns all subscriptions for an event
func (eb *EventBus) GetSubscriptions(eventName string) []*EventSubscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subscriptions := eb.subscribers[eventName]
	result := make([]*EventSubscription, len(subscriptions))
	copy(result, subscriptions)
	return result
}

// GetAllSubscriptions returns all subscriptions
func (eb *EventBus) GetAllSubscriptions() map[string][]*EventSubscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	result := make(map[string][]*EventSubscription)
	for eventName, subscriptions := range eb.subscribers {
		result[eventName] = make([]*EventSubscription, len(subscriptions))
		copy(result[eventName], subscriptions)
	}
	return result
}

// isRunning checks if the event bus is running
func (eb *EventBus) isRunning() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.running
}

// generateEventID generates a unique event ID
func (eb *EventBus) generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// generateSubscriptionID generates a unique subscription ID
func (eb *EventBus) generateSubscriptionID() string {
	return fmt.Sprintf("sub_%d", time.Now().UnixNano())
}

// Health returns the health status of the event bus
func (eb *EventBus) Health() string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if !eb.running {
		return HealthStatusStopped
	}

	queueLength := len(eb.eventQueue)
	if queueLength > 800 {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// Stats returns statistics about the event bus
func (eb *EventBus) Stats() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	stats := map[string]interface{}{
		"running":           eb.running,
		"queue_length":      len(eb.eventQueue),
		"queue_capacity":    cap(eb.eventQueue),
		"workers":           eb.workers,
		"total_subscribers": 0,
	}

	totalSubscribers := 0
	for _, subscriptions := range eb.subscribers {
		totalSubscribers += len(subscriptions)
	}
	stats["total_subscribers"] = totalSubscribers

	return stats
}
