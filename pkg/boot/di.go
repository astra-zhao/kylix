// di.go — Dependency Injection container for KylixBoot.
//
// In v3.1.0, the container is a simple map-based registry. Future versions
// will integrate with [Inject] attributes for compile-time auto-wiring.
package boot

import (
	"fmt"
	"reflect"
	"sync"
)

// Lifetime defines how the container manages an instance.
type Lifetime int

const (
	// Singleton — one shared instance for the application lifetime.
	Singleton Lifetime = iota
	// Transient — a new instance every Resolve().
	Transient
)

// Container holds registered components and their factories.
type Container struct {
	mu         sync.RWMutex
	bindings   map[string]*binding
	singletons map[string]interface{}
}

type binding struct {
	name     string
	factory  func(c *Container) interface{}
	lifetime Lifetime
}

// NewContainer creates an empty DI container.
func NewContainer() *Container {
	return &Container{
		bindings:   map[string]*binding{},
		singletons: map[string]interface{}{},
	}
}

// Register binds a name to a singleton factory.
func (c *Container) Register(name string, factory func(*Container) interface{}) {
	c.bind(name, factory, Singleton)
}

// RegisterTransient binds a name to a factory that returns a new instance each time.
func (c *Container) RegisterTransient(name string, factory func(*Container) interface{}) {
	c.bind(name, factory, Transient)
}

// RegisterInstance binds a name to a pre-constructed instance.
func (c *Container) RegisterInstance(name string, instance interface{}) {
	c.mu.Lock()
	c.singletons[name] = instance
	c.bindings[name] = &binding{
		name:     name,
		factory:  func(*Container) interface{} { return instance },
		lifetime: Singleton,
	}
	c.mu.Unlock()
}

func (c *Container) bind(name string, factory func(*Container) interface{}, lt Lifetime) {
	c.mu.Lock()
	c.bindings[name] = &binding{name: name, factory: factory, lifetime: lt}
	c.mu.Unlock()
}

// Resolve returns an instance of the named component.
// Panics if not registered (use TryResolve for safe access).
func (c *Container) Resolve(name string) interface{} {
	inst, ok := c.TryResolve(name)
	if !ok {
		panic(fmt.Sprintf("boot.DI: no binding for %q", name))
	}
	return inst
}

// TryResolve returns (instance, true) if registered, (nil, false) otherwise.
func (c *Container) TryResolve(name string) (interface{}, bool) {
	c.mu.RLock()
	b, ok := c.bindings[name]
	if !ok {
		c.mu.RUnlock()
		return nil, false
	}
	if b.lifetime == Singleton {
		if existing, exists := c.singletons[name]; exists {
			c.mu.RUnlock()
			return existing, true
		}
	}
	c.mu.RUnlock()

	inst := b.factory(c)

	if b.lifetime == Singleton {
		c.mu.Lock()
		c.singletons[name] = inst
		c.mu.Unlock()
	}
	return inst, true
}

// Inject populates fields of a struct pointer by looking up matching DI names.
// Each field whose name matches a registered binding will be set.
//
// Used internally by the annotation processor for [Inject] fields.
func (c *Container) Inject(target interface{}) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("boot.Inject: target must be a struct pointer")
	}
	elem := v.Elem()
	t := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if !field.CanSet() {
			continue
		}
		name := t.Field(i).Name
		if inst, ok := c.TryResolve(name); ok {
			rv := reflect.ValueOf(inst)
			if rv.Type().AssignableTo(field.Type()) {
				field.Set(rv)
			}
		}
	}
	return nil
}
