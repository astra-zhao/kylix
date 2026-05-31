package stdlib

import (
	"fmt"
	"reflect"
	"sync"
)

// Service lifetime
type ServiceLifetime int

const (
	Singleton ServiceLifetime = iota
	Transient
	Scoped
)

// ServiceDescriptor describes a registered service
type ServiceDescriptor struct {
	Name     string
	Type     reflect.Type
	Factory  interface{}
	Lifetime ServiceLifetime
	Instance interface{} // For singletons
}

// Container is a simple dependency injection container
type Container struct {
	services map[string]*ServiceDescriptor
	scoped   map[string]interface{} // Scoped instances
	mu       sync.RWMutex
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{
		services: make(map[string]*ServiceDescriptor),
		scoped:   make(map[string]interface{}),
	}
}

// Register registers a service with a factory function
func (c *Container) Register(name string, factory interface{}, lifetime ServiceLifetime) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function, got %v", factoryType.Kind())
	}

	if factoryType.NumOut() != 1 {
		return fmt.Errorf("factory must return exactly one value, got %d", factoryType.NumOut())
	}

	c.services[name] = &ServiceDescriptor{
		Name:     name,
		Type:     factoryType.Out(0),
		Factory:  factory,
		Lifetime: lifetime,
	}

	return nil
}

// RegisterSingleton registers a singleton service (one instance for the entire application)
func (c *Container) RegisterSingleton(name string, factory interface{}) error {
	return c.Register(name, factory, Singleton)
}

// RegisterTransient registers a transient service (new instance each time)
func (c *Container) RegisterTransient(name string, factory interface{}) error {
	return c.Register(name, factory, Transient)
}

// RegisterScoped registers a scoped service (one instance per scope)
func (c *Container) RegisterScoped(name string, factory interface{}) error {
	return c.Register(name, factory, Scoped)
}

// RegisterInstance registers an existing instance as a singleton
func (c *Container) RegisterInstance(name string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = &ServiceDescriptor{
		Name:     name,
		Type:     reflect.TypeOf(instance),
		Lifetime: Singleton,
		Instance: instance,
	}
}

// Resolve resolves a service by name
func (c *Container) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	descriptor, exists := c.services[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service '%s' not registered", name)
	}

	switch descriptor.Lifetime {
	case Singleton:
		if descriptor.Instance == nil {
			// Create the singleton instance
			c.mu.Lock()
			defer c.mu.Unlock()

			// Double-check after acquiring write lock
			if descriptor.Instance == nil {
				instance, err := c.createInstance(descriptor)
				if err != nil {
					return nil, err
				}
				descriptor.Instance = instance
			}
		}
		return descriptor.Instance, nil

	case Transient:
		// Always create a new instance
		return c.createInstance(descriptor)

	case Scoped:
		// Check if we have a scoped instance
		c.mu.RLock()
		instance, exists := c.scoped[name]
		c.mu.RUnlock()

		if exists {
			return instance, nil
		}

		// Create and cache the scoped instance
		c.mu.Lock()
		defer c.mu.Unlock()

		// Double-check after acquiring write lock
		instance, exists = c.scoped[name]
		if exists {
			return instance, nil
		}

		instance, err := c.createInstance(descriptor)
		if err != nil {
			return nil, err
		}
		c.scoped[name] = instance
		return instance, nil

	default:
		return nil, fmt.Errorf("unknown service lifetime: %v", descriptor.Lifetime)
	}
}

// ResolveType resolves a service by type
func (c *Container) ResolveType(target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetType := targetValue.Elem().Type()

	c.mu.RLock()
	var descriptor *ServiceDescriptor
	for _, desc := range c.services {
		if desc.Type == targetType {
			descriptor = desc
			break
		}
	}
	c.mu.RUnlock()

	if descriptor == nil {
		return fmt.Errorf("no service registered for type %v", targetType)
	}

	instance, err := c.Resolve(descriptor.Name)
	if err != nil {
		return err
	}

	targetValue.Elem().Set(reflect.ValueOf(instance))
	return nil
}

// createInstance creates an instance using the factory function
func (c *Container) createInstance(descriptor *ServiceDescriptor) (interface{}, error) {
	factoryValue := reflect.ValueOf(descriptor.Factory)

	// Check if factory needs the container as a parameter
	factoryType := factoryValue.Type()
	var args []reflect.Value

	if factoryType.NumIn() == 1 && factoryType.In(0) == reflect.TypeOf((*Container)(nil)) {
		args = []reflect.Value{reflect.ValueOf(c)}
	} else if factoryType.NumIn() > 0 {
		// Try to resolve dependencies
		args = make([]reflect.Value, factoryType.NumIn())
		for i := 0; i < factoryType.NumIn(); i++ {
			paramType := factoryType.In(i)
			// Find a service that matches this type
			var found bool
			for _, desc := range c.services {
				if desc.Type == paramType {
					instance, err := c.Resolve(desc.Name)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve dependency: %v", err)
					}
					args[i] = reflect.ValueOf(instance)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("cannot resolve parameter of type %v", paramType)
			}
		}
	}

	results := factoryValue.Call(args)
	if len(results) == 0 {
		return nil, fmt.Errorf("factory returned no values")
	}

	return results[0].Interface(), nil
}

// CreateScope creates a new scope for scoped services
func (c *Container) CreateScope() *Container {
	return &Container{
		services: c.services,
		scoped:   make(map[string]interface{}),
	}
}

// Has checks if a service is registered
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.services[name]
	return exists
}

// GetServiceNames returns all registered service names
func (c *Container) GetServiceNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}
	return names
}
