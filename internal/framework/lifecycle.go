// Package framework - Lifecycle Manager
package framework

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LifecyclePhase represents different phases of the lifecycle
type LifecyclePhase int

const (
	PhaseInitializing LifecyclePhase = iota
	PhaseStarting
	PhaseRunning
	PhaseStopping
	PhaseStopped
)

func (p LifecyclePhase) String() string {
	switch p {
	case PhaseInitializing:
		return "initializing"
	case PhaseStarting:
		return "starting"
	case PhaseRunning:
		return "running"
	case PhaseStopping:
		return "stopping"
	case PhaseStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// LifecycleComponent represents a component that participates in lifecycle management
type LifecycleComponent interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() string
}

// ComponentInfo holds information about a registered component
type ComponentInfo struct {
	Name      string
	Component LifecycleComponent
	Started   bool
	Priority  int
}

// Lifecycle manages the lifecycle of framework components
type Lifecycle struct {
	components map[string]*ComponentInfo
	phase      LifecyclePhase
	mu         sync.RWMutex
	startTime  time.Time
}

// NewLifecycle creates a new lifecycle manager
func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		components: make(map[string]*ComponentInfo),
		phase:      PhaseStopped,
	}
}

// Register registers a component for lifecycle management
func (l *Lifecycle) Register(name string, component LifecycleComponent, priority int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.components[name] = &ComponentInfo{
		Name:      name,
		Component: component,
		Started:   false,
		Priority:  priority,
	}
}

// Unregister removes a component from lifecycle management
func (l *Lifecycle) Unregister(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.components, name)
}

// Start starts all registered components in priority order
func (l *Lifecycle) Start(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.phase == PhaseRunning {
		return nil
	}

	l.phase = PhaseStarting
	l.startTime = time.Now()

	// Sort components by priority
	sortedComponents := l.getSortedComponents()

	// Start components in priority order
	for _, info := range sortedComponents {
		if err := info.Component.Start(ctx); err != nil {
			l.phase = PhaseStopped
			return fmt.Errorf("failed to start component '%s': %w", info.Name, err)
		}
		info.Started = true
	}

	l.phase = PhaseRunning
	return nil
}

// Stop stops all registered components in reverse priority order
func (l *Lifecycle) Stop(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.phase == PhaseStopped {
		return nil
	}

	l.phase = PhaseStopping

	// Sort components by priority (reverse order for stopping)
	sortedComponents := l.getSortedComponents()

	var errors []error

	// Stop components in reverse priority order
	for i := len(sortedComponents) - 1; i >= 0; i-- {
		info := sortedComponents[i]
		if info.Started {
			if err := info.Component.Stop(ctx); err != nil {
				errors = append(errors, fmt.Errorf("failed to stop component '%s': %w", info.Name, err))
			}
			info.Started = false
		}
	}

	l.phase = PhaseStopped

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}

// getSortedComponents returns components sorted by priority
func (l *Lifecycle) getSortedComponents() []*ComponentInfo {
	components := make([]*ComponentInfo, 0, len(l.components))
	for _, info := range l.components {
		components = append(components, info)
	}

	// Simple sort by priority (lower number = higher priority)
	for i := 0; i < len(components); i++ {
		for j := i + 1; j < len(components); j++ {
			if components[i].Priority > components[j].Priority {
				components[i], components[j] = components[j], components[i]
			}
		}
	}

	return components
}

// GetPhase returns the current lifecycle phase
func (l *Lifecycle) GetPhase() LifecyclePhase {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.phase
}

// IsRunning returns true if the lifecycle is in running phase
func (l *Lifecycle) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.phase == PhaseRunning
}

// GetComponents returns information about all registered components
func (l *Lifecycle) GetComponents() map[string]*ComponentInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]*ComponentInfo)
	for name, info := range l.components {
		result[name] = &ComponentInfo{
			Name:      info.Name,
			Component: info.Component,
			Started:   info.Started,
			Priority:  info.Priority,
		}
	}
	return result
}

// GetComponent returns information about a specific component
func (l *Lifecycle) GetComponent(name string) (*ComponentInfo, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	info, exists := l.components[name]
	if !exists {
		return nil, fmt.Errorf("component '%s' not found", name)
	}

	return &ComponentInfo{
		Name:      info.Name,
		Component: info.Component,
		Started:   info.Started,
		Priority:  info.Priority,
	}, nil
}

// Health returns the health status of the lifecycle manager
func (l *Lifecycle) Health() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	switch l.phase {
	case PhaseRunning:
		// Check if all components are healthy
		for _, info := range l.components {
			if info.Started && info.Component.Health() != HealthStatusHealthy {
				return HealthStatusDegraded
			}
		}
		return HealthStatusHealthy
	case PhaseStarting, PhaseStopping:
		return "transitioning"
	case PhaseStopped:
		return HealthStatusStopped
	default:
		return "unknown"
	}
}

// Stats returns statistics about the lifecycle manager
func (l *Lifecycle) Stats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	started := 0
	healthy := 0

	for _, info := range l.components {
		if info.Started {
			started++
		}
		if info.Component.Health() == HealthStatusHealthy {
			healthy++
		}
	}

	stats := map[string]interface{}{
		"phase":              l.phase.String(),
		"total_components":   len(l.components),
		"started_components": started,
		"healthy_components": healthy,
	}

	if !l.startTime.IsZero() {
		stats["uptime"] = time.Since(l.startTime).String()
	}

	return stats
}
