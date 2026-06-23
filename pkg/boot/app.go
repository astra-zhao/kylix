// app.go — Top-level KylixBoot Application facade.
//
// Provides a default global application instance, mirroring Spring Boot's
// SpringApplication.run() entry point. Most apps will use the package-level
// shortcuts (boot.GET / boot.POST / boot.Run).
package boot

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// App is the central KylixBoot application object.
type App struct {
	Router    *Router
	Container *Container
	Config    *Config

	mu      sync.Mutex
	server  *http.Server
	running bool
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		Router:    NewRouter(),
		Container: NewContainer(),
		Config:    NewConfig(),
	}
}

// Default is the default singleton App used by package-level shortcuts.
var Default = NewApp()

// HTTP method registration shortcuts on the default app.
func GET(p string, h Handler, mws ...Middleware)    { Default.Router.GET(p, h, mws...) }
func POST(p string, h Handler, mws ...Middleware)   { Default.Router.POST(p, h, mws...) }
func PUT(p string, h Handler, mws ...Middleware)    { Default.Router.PUT(p, h, mws...) }
func DELETE(p string, h Handler, mws ...Middleware) { Default.Router.DELETE(p, h, mws...) }
func PATCH(p string, h Handler, mws ...Middleware)  { Default.Router.PATCH(p, h, mws...) }

// Use adds a global middleware to the default app.
func Use(mw Middleware) { Default.Router.Use(mw) }

// Register binds a singleton to the default DI container.
func Register(name string, factory func(*Container) interface{}) {
	Default.Container.Register(name, factory)
}

// RegisterInstance binds an existing instance to the default container.
func RegisterInstance(name string, instance interface{}) {
	Default.Container.RegisterInstance(name, instance)
}

// Resolve retrieves an instance from the default container.
func Resolve(name string) interface{} { return Default.Container.Resolve(name) }

// SetConfig stores a value in the default config.
func SetConfig(key string, value interface{}) { Default.Config.Set(key, value) }

// GetConfigString reads a string config value.
func GetConfigString(key, fallback string) string {
	return Default.Config.StringDefault(key, fallback)
}

// GetConfigInt reads an int config value.
func GetConfigInt(key string, fallback int) int {
	return Default.Config.IntDefault(key, fallback)
}

// Run starts the default app on the given port (blocks until interrupted).
func Run(port int) error {
	return Default.Run(port)
}

// Run starts this app's HTTP server.
func (a *App) Run(port int) error {
	addr := ":" + strconv.Itoa(port)
	a.mu.Lock()
	a.server = &http.Server{
		Addr:         addr,
		Handler:      a.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	a.running = true
	a.mu.Unlock()

	log.Printf("🚀 KylixBoot started on http://localhost%s", addr)
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("KylixBoot server error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server.
func (a *App) Stop() error {
	a.mu.Lock()
	srv := a.server
	a.running = false
	a.mu.Unlock()
	if srv != nil {
		return srv.Close()
	}
	return nil
}
