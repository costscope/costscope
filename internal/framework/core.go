// Package framework provides the core extensible framework for CostScope
package framework

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Framework represents the core extensible framework
type Framework struct {
	container    *Container
	eventBus     *EventBus
	pluginLoader *PluginLoader
	lifecycle    *Lifecycle
	mu           sync.RWMutex
	started      bool
}

// NewFramework creates a new framework instance
func NewFramework() *Framework {
	return &Framework{
		container:    NewContainer(),
		eventBus:     NewEventBus(),
		pluginLoader: NewPluginLoader(),
		lifecycle:    NewLifecycle(),
	}
}

// Start initializes and starts the framework
func (f *Framework) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.started {
		return nil
	}

	// Initialize lifecycle
	if err := f.lifecycle.Start(ctx); err != nil {
		return fmt.Errorf("failed to start lifecycle: %w", err)
	}

	// Register core services
	f.registerCoreServices()

	// Load plugins
	if err := f.pluginLoader.LoadPlugins(ctx, f.container); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Start event bus
	if err := f.eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Emit framework started event
	f.eventBus.Emit("framework.started", map[string]interface{}{
		"timestamp": time.Now(),
		"version":   GetVersion(),
	})

	f.started = true
	return nil
}

// Stop gracefully shuts down the framework
func (f *Framework) Stop(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.started {
		return nil
	}

	// Emit framework stopping event
	f.eventBus.Emit("framework.stopping", map[string]interface{}{
		"timestamp": time.Now(),
	})

	// Stop components in reverse order
	if err := f.eventBus.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop event bus: %w", err)
	}

	if err := f.pluginLoader.UnloadPlugins(ctx); err != nil {
		return fmt.Errorf("failed to unload plugins: %w", err)
	}

	if err := f.lifecycle.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop lifecycle: %w", err)
	}

	f.started = false
	return nil
}

// Container returns the dependency injection container
func (f *Framework) Container() *Container {
	return f.container
}

// EventBus returns the event bus
func (f *Framework) EventBus() *EventBus {
	return f.eventBus
}

// PluginLoader returns the plugin loader
func (f *Framework) PluginLoader() *PluginLoader {
	return f.pluginLoader
}

// Lifecycle returns the lifecycle manager
func (f *Framework) Lifecycle() *Lifecycle {
	return f.lifecycle
}

// IsStarted returns true if the framework is started
func (f *Framework) IsStarted() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.started
}

// registerCoreServices registers core framework services
func (f *Framework) registerCoreServices() {
	// Register framework components
	f.container.RegisterSingleton("framework", f)
	f.container.RegisterSingleton("container", f.container)
	f.container.RegisterSingleton("eventBus", f.eventBus)
	f.container.RegisterSingleton("pluginLoader", f.pluginLoader)
	f.container.RegisterSingleton("lifecycle", f.lifecycle)

	// Register factories
	f.container.RegisterFactory("logger", func(c *Container) interface{} {
		return NewFrameworkLogger()
	})

	f.container.RegisterFactory("config", func(c *Container) interface{} {
		return NewFrameworkConfig()
	})
}

// GetVersion returns the framework version
func GetVersion() string {
	return "1.0.0"
}

// Health represents framework health status
type Health struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Health returns the current health status of the framework
func (f *Framework) Health() *Health {
	f.mu.RLock()
	defer f.mu.RUnlock()

	health := &Health{
		Status:     HealthStatusHealthy,
		Components: make(map[string]string),
		Timestamp:  time.Now(),
	}

	if !f.started {
		health.Status = HealthStatusStopped
		return health
	}

	// Check component health
	health.Components["container"] = HealthStatusHealthy
	health.Components["eventBus"] = f.eventBus.Health()
	health.Components["pluginLoader"] = f.pluginLoader.Health()
	health.Components["lifecycle"] = f.lifecycle.Health()

	// Check if any component is unhealthy
	for _, status := range health.Components {
		if status != HealthStatusHealthy {
			health.Status = HealthStatusDegraded
			break
		}
	}

	return health
}
