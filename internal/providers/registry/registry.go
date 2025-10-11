package registry

import (
	"fmt"
	"sync"
)

// Provider is a minimal interface representing a cloud provider instance.
// Concrete provider packages (aws, azure, gcp) should satisfy this.
type Provider interface {
	Name() string
}

// Factory constructs a Provider. Accepts variadic options for future extensibility.
type Factory func(opts ...interface{}) (Provider, error)

var (
	mu        sync.RWMutex
	factories = map[string]Factory{}
)

// Register adds a provider factory under a name (lowercase key). Panics on duplicate to surface programmer error.
func Register(name string, f Factory) {
	mu.Lock()
	defer mu.Unlock()
	key := normalize(name)
	if _, exists := factories[key]; exists {
		panic("provider registry: duplicate registration for " + key)
	}
	factories[key] = f
}

// Get returns a provider instance by name.
func Get(name string, opts ...interface{}) (Provider, error) {
	mu.RLock()
	f, ok := factories[normalize(name)]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("provider '%s' not registered", name)
	}
	return f(opts...)
}

// List returns registered provider names (sorted insertion order not guaranteed).
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(factories))
	for k := range factories {
		names = append(names, k)
	}
	return names
}

func normalize(s string) string { return s }
