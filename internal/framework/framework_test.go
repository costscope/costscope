package framework

import (
	"context"
	"testing"
	"time"
)

func TestFrameworkBasic(t *testing.T) {
	// Create framework
	fw := NewFramework()

	// Start framework
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := fw.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start framework: %v", err)
	}

	// Check if framework is started
	if !fw.IsStarted() {
		t.Fatal("Framework should be started")
	}

	// Check health
	health := fw.Health()
	if health.Status != HealthStatusHealthy {
		t.Fatalf("Expected healthy status, got %s", health.Status)
	}

	// Test components
	if health.Components["container"] != HealthStatusHealthy {
		t.Fatal("Container should be healthy")
	}

	if health.Components["eventBus"] != HealthStatusHealthy {
		t.Fatal("EventBus should be healthy")
	}

	// Stop framework
	err = fw.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop framework: %v", err)
	}

	// Check if framework is stopped
	if fw.IsStarted() {
		t.Fatal("Framework should be stopped")
	}
}

func TestContainerBasic(t *testing.T) {
	container := NewContainer()

	// Test singleton registration
	testService := "test-service"
	container.RegisterSingleton("test", testService)

	// Test service retrieval
	service, err := container.Get("test")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	if service != testService {
		t.Fatal("Retrieved service doesn't match")
	}

	// Test non-existent service
	_, err = container.Get("non-existent")
	if err == nil {
		t.Fatal("Should return error for non-existent service")
	}

	// Test service list
	services := container.List()
	found := false
	for _, name := range services {
		if name == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Service 'test' should be in the list")
	}
}

func TestEventBusBasic(t *testing.T) {
	eventBus := NewEventBus()

	// Start event bus
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := eventBus.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	// Test event subscription and emission
	received := make(chan bool, 1)
	eventBus.Subscribe("test.event", func(event Event) {
		if event.Name == "test.event" {
			received <- true
		}
	})

	// Give time for subscription to register
	time.Sleep(100 * time.Millisecond)

	// Emit event
	eventBus.Emit("test.event", map[string]interface{}{
		"message": "test message",
	})

	// Wait for event to be received
	select {
	case <-received:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Event was not received")
	}

	// Stop event bus
	err = eventBus.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

func TestPluginLoader(t *testing.T) {
	loader := NewPluginLoader()
	container := NewContainer()

	ctx := context.Background()

	// Load built-in plugins
	err := loader.LoadPlugins(ctx, container)
	if err != nil {
		t.Fatalf("Failed to load plugins: %v", err)
	}

	// Check if plugins are loaded
	plugins := loader.ListPlugins()
	if len(plugins) == 0 {
		t.Fatal("No plugins loaded")
	}

	// Check for expected plugins
	expectedPlugins := []string{"analytics", "reporting", "monitoring"}
	for _, expected := range expectedPlugins {
		found := false
		for _, plugin := range plugins {
			if plugin.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Plugin '%s' not found", expected)
		}
	}

	// Test plugin retrieval
	analyticsPlugin, err := loader.GetPlugin("analytics")
	if err != nil {
		t.Fatalf("Failed to get analytics plugin: %v", err)
	}

	if analyticsPlugin.Name() != "analytics" {
		t.Fatal("Analytics plugin name mismatch")
	}

	// Unload plugins
	err = loader.UnloadPlugins(ctx)
	if err != nil {
		t.Fatalf("Failed to unload plugins: %v", err)
	}
}

func BenchmarkEventBus(b *testing.B) {
	eventBus := NewEventBus()
	ctx := context.Background()

	err := eventBus.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start event bus: %v", err)
	}
	defer func() {
		if err := eventBus.Stop(ctx); err != nil {
			b.Errorf("Failed to stop event bus: %v", err)
		}
	}()

	// Subscribe to events
	eventBus.Subscribe("benchmark.event", func(event Event) {
		// Do nothing, just receive
	})

	b.ResetTimer()

	// Benchmark event emission
	for i := 0; i < b.N; i++ {
		eventBus.Emit("benchmark.event", map[string]interface{}{
			"iteration": i,
		})
	}
}

func BenchmarkContainer(b *testing.B) {
	container := NewContainer()

	// Register a service
	container.RegisterSingleton("benchmark-service", "test")

	b.ResetTimer()

	// Benchmark service retrieval
	for i := 0; i < b.N; i++ {
		_, err := container.Get("benchmark-service")
		if err != nil {
			b.Fatalf("Failed to get service: %v", err)
		}
	}
}
