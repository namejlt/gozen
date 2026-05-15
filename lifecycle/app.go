// Package lifecycle provides application lifecycle management — startup
// orchestration, graceful shutdown, and signal handling.  It unifies the
// shutdown logic from the old gin.go / grpc.go into a single abstraction.
package lifecycle

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Application is the top-level app container.  It starts and stops
// registered services in dependency order.
type Application struct {
	name      string
	hooks     []Hook
	servers   []Server
	startOnce sync.Once
	cancel    context.CancelFunc
}

// Server is a runnable service (HTTP, gRPC, cron, …).
type Server interface {
	// Name returns a human-readable name for logging.
	Name() string

	// Run starts the server.  It should block until Shutdown is called
	// or the server encounters a fatal error.
	Run() error

	// Shutdown gracefully stops the server.
	Shutdown(ctx context.Context) error
}

// Hook is a lifecycle callback.
type Hook struct {
	// Name is a human-readable name.
	Name string

	// OnStart runs during startup, after all servers start.
	OnStart func(ctx context.Context) error

	// OnStop runs during shutdown, after servers stop.
	OnStop func(ctx context.Context) error

	// Timeout for OnStop.
	Timeout time.Duration
}

// New creates an Application.
func New(name string) *Application {
	return &Application{name: name}
}

// AddServer registers a server.  Servers are started in registration order
// and shut down in reverse order.
func (a *Application) AddServer(s Server) {
	a.servers = append(a.servers, s)
}

// AddHook registers a lifecycle hook.
func (a *Application) AddHook(h Hook) {
	a.hooks = append(a.hooks, h)
}

// Run starts all servers and hooks, then blocks until SIGINT/SIGTERM.
func (a *Application) Run() error {
	var err error
	a.startOnce.Do(func() {
		err = a.start()
	})
	if err != nil {
		return err
	}

	// Wait for signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit
	log.Printf("[lifecycle] received signal %v, shutting down…", sig)

	return a.shutdown()
}

// start launches everything.
func (a *Application) start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	// Start servers.
	for _, s := range a.servers {
		log.Printf("[lifecycle] starting server %s", s.Name())
		go func(srv Server) {
			if err := srv.Run(); err != nil {
				log.Printf("[lifecycle] server %s exited: %v", srv.Name(), err)
			}
		}(s)
	}

	// Run OnStart hooks.
	for _, h := range a.hooks {
		if h.OnStart != nil {
			if err := h.OnStart(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// shutdown stops everything in reverse order.
func (a *Application) shutdown() error {
	if a.cancel != nil {
		a.cancel()
	}

	var wg sync.WaitGroup

	// Shutdown servers in reverse.
	for i := len(a.servers) - 1; i >= 0; i-- {
		s := a.servers[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.Shutdown(ctx); err != nil {
				log.Printf("[lifecycle] server %s shutdown error: %v", s.Name(), err)
			}
		}()
	}

	// Run OnStop hooks in parallel.
	for _, h := range a.hooks {
		if h.OnStop != nil {
			h := h
			wg.Add(1)
			go func() {
				defer wg.Done()
				timeout := h.Timeout
				if timeout == 0 {
					timeout = 5 * time.Second
				}
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				if err := h.OnStop(ctx); err != nil {
					log.Printf("[lifecycle] hook %s error: %v", h.Name, err)
				}
			}()
		}
	}

	wg.Wait()
	log.Printf("[lifecycle] %s shut down gracefully", a.name)
	return nil
}
