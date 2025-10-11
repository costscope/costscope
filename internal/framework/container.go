// Package framework - Dependency Injection Container
package framework

import (
	"fmt"
	"reflect"
	"sync"
)

// HealthStatusEmpty is returned by components that report an empty / uninitialized state.
const HealthStatusEmpty = "empty"

// ServiceFactory is a function that creates a service instance
type ServiceFactory func(c *Container) interface{}

// ServiceDescriptor describes how to create a service
type ServiceDescriptor struct {
	Name      string
	Type      reflect.Type
	Factory   ServiceFactory
	Instance  interface{}
	Singleton bool
}

// Container provides dependency injection functionality
type Container struct {
	services map[string]*ServiceDescriptor
	mu       sync.RWMutex
}

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return &Container{
		services: make(map[string]*ServiceDescriptor),
	}
}

// RegisterSingleton registers a singleton service instance
func (c *Container) RegisterSingleton(name string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = &ServiceDescriptor{
		Name:      name,
		Type:      reflect.TypeOf(instance),
		Instance:  instance,
		Singleton: true,
	}
}

// RegisterFactory registers a service factory
func (c *Container) RegisterFactory(name string, factory ServiceFactory) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = &ServiceDescriptor{
		Name:      name,
		Factory:   factory,
		Singleton: false,
	}
}

// RegisterSingletonFactory registers a singleton service factory
func (c *Container) RegisterSingletonFactory(name string, factory ServiceFactory) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = &ServiceDescriptor{
		Name:      name,
		Factory:   factory,
		Singleton: true,
	}
}

// Get retrieves a service by name
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	descriptor, exists := c.services[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	// Return singleton instance if available
	if descriptor.Singleton && descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	// Create new instance using factory
	if descriptor.Factory != nil {
		instance := descriptor.Factory(c)

		// Store singleton instance
		if descriptor.Singleton {
			c.mu.Lock()
			descriptor.Instance = instance
			c.mu.Unlock()
		}

		return instance, nil
	}

	// Return existing instance
	if descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	return nil, fmt.Errorf("no factory or instance available for service '%s'", name)
}

// GetAs retrieves a service by name and casts it to the specified type
func (c *Container) GetAs(name string, target interface{}) error {
	service, err := c.Get(name)
	if err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	targetType := targetValue.Elem().Type()

	if !serviceValue.Type().AssignableTo(targetType) {
		return fmt.Errorf("service '%s' of type %v is not assignable to %v",
			name, serviceValue.Type(), targetType)
	}

	targetValue.Elem().Set(serviceValue)
	return nil
}

// Removed MustGet: callers should use Get/GetAs and handle errors explicitly.

// Has checks if a service is registered
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.services[name]
	return exists
}

// List returns all registered service names
func (c *Container) List() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}
	return names
}

// Unregister removes a service from the container
func (c *Container) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.services, name)
}

// Clear removes all services from the container
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[string]*ServiceDescriptor)
}

// Inject performs dependency injection on a struct
func (c *Container) Inject(target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Check for inject tag
		injectTag := field.Tag.Get("inject")
		if injectTag == "" {
			continue
		}

		// Get service name (use tag value or field name)
		serviceName := injectTag
		if serviceName == "true" || serviceName == "" {
			serviceName = field.Name
		}

		// Get and inject service
		service, err := c.Get(serviceName)
		if err != nil {
			return fmt.Errorf("failed to inject field '%s': %w", field.Name, err)
		}

		serviceValue := reflect.ValueOf(service)
		if !serviceValue.Type().AssignableTo(fieldValue.Type()) {
			return fmt.Errorf("service '%s' of type %v is not assignable to field '%s' of type %v",
				serviceName, serviceValue.Type(), field.Name, fieldValue.Type())
		}

		fieldValue.Set(serviceValue)
	}

	return nil
}

// GetServiceDescriptor returns the service descriptor for a given name
func (c *Container) GetServiceDescriptor(name string) (*ServiceDescriptor, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	descriptor, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	return descriptor, nil
}

// Health returns the health status of the container
func (c *Container) Health() string {
	c.mu.RLock()
	empty := len(c.services) == 0
	c.mu.RUnlock()
	if empty {
		return HealthStatusEmpty
	}
	return HealthStatusHealthy
}
