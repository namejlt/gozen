// Package http provides an HTTP transport layer built on gin-gonic/gin.
//
// Typical usage:
//
//	srv := http.NewServer(http.WithAddr(":8080"))
//	srv.Router().GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
//	srv.Run()          // blocks until signal
//	srv.Shutdown(ctx)  // manual shutdown
package http

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Server wraps an http.Server + gin.Engine with graceful-shutdown semantics.
type Server struct {
	engine   *gin.Engine
	srv      *http.Server
	opts     ServerOptions
	shutdown []ShutdownFunc

	mu       sync.Mutex
	running  bool
}

// ShutdownFunc is a callback executed during graceful shutdown.
// ctx is cancelled after the specified timeout.
type ShutdownFunc struct {
	Fn      func(ctx context.Context)
	Timeout time.Duration
}

// ServerOptions configures the HTTP server.
type ServerOptions struct {
	Addr    string
	Mode    string // "debug" or "release"
	Timeout int    // shutdown timeout in seconds, default 10
}

// ServerOption is a functional option.
type ServerOption func(*ServerOptions)

// WithAddr sets the listen address.
func WithAddr(addr string) ServerOption {
	return func(o *ServerOptions) { o.Addr = addr }
}

// WithMode sets gin mode ("debug" / "release" / "test").
func WithMode(mode string) ServerOption {
	return func(o *ServerOptions) { o.Mode = mode }
}

// WithTimeout sets the graceful-shutdown timeout (seconds).
func WithTimeout(sec int) ServerOption {
	return func(o *ServerOptions) { o.Timeout = sec }
}

// NewServer creates an HTTP server with the given options.
func NewServer(opts ...ServerOption) *Server {
	o := ServerOptions{
		Addr:    ":8080",
		Mode:    "release",
		Timeout: 10,
	}
	for _, fn := range opts {
		fn(&o)
	}
	gin.SetMode(o.Mode)
	engine := gin.New()
	return &Server{
		engine: engine,
		opts:   o,
	}
}

// Name implements lifecycle.Server.
func (s *Server) Name() string { return "http:" + s.opts.Addr }

// Router returns the underlying gin.Engine for route registration.
func (s *Server) Router() *gin.Engine { return s.engine }

// RegisterShutdown adds a shutdown callback.  Safe for concurrent use.
func (s *Server) RegisterShutdown(fn func(ctx context.Context), timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shutdown = append(s.shutdown, ShutdownFunc{Fn: fn, Timeout: timeout})
}

// Run starts the server and blocks until SIGINT / SIGTERM, then shuts down.
func (s *Server) Run() error {
	s.srv = &http.Server{Addr: s.opts.Addr, Handler: s.engine}
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		return s.shutdownSequence()
	}
}

// Shutdown triggers graceful shutdown manually.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.shutdownSequence()
}

func (s *Server) shutdownSequence() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	s.running = false

	// Run registered shutdown callbacks.
	for _, sf := range s.shutdown {
		var wg sync.WaitGroup
		wg.Add(1)
		ctx, cancel := context.WithTimeout(context.Background(), sf.Timeout)
		go sf.Fn(ctx)
		select {
		case <-ctx.Done():
			cancel()
			wg.Done()
		case <-time.After(sf.Timeout + 1*time.Second):
			cancel()
			wg.Done()
		}
		wg.Wait()
	}

	// Shutdown the HTTP server.
	tctx, tcancel := context.WithTimeout(context.Background(), time.Duration(s.opts.Timeout)*time.Second)
	defer tcancel()
	return s.srv.Shutdown(tctx)
}
