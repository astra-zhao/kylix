// server.go — HTTP server wrapping Router with graceful lifecycle.
package boot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is a pre-configured HTTP server with optional TLS, graceful shutdown.
type Server struct {
	Router      *Router
	Addr        string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewServer creates a server on the given port with a fresh Router.
func NewServer(port int) *Server {
	return &Server{
		Router:       NewRouter(),
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// Run starts the server and blocks until SIGINT/SIGTERM.
func (s *Server) Run() error {
	httpServer := &http.Server{
		Addr:         s.Addr,
		Handler:      s.Router,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("KylixBoot server listening on %s", s.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(ctx)
}