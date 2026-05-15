// Package grpc provides a gRPC transport layer with connection pooling.
//
// Typical usage:
//
//	srv := grpc.NewServer(grpc.WithAddr(":9090"))
//	pb.RegisterMyServiceServer(srv.Server(), &myImpl{})
//	srv.Run()
package grpc

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

// Server wraps a grpc.Server with graceful-shutdown semantics.
type Server struct {
	srv  *grpc.Server
	opts Options
}

// Options configures the gRPC server.
type Options struct {
	Addr string
}

// Option is a functional option.
type Option func(*Options)

// WithAddr sets the listen address.
func WithAddr(addr string) Option {
	return func(o *Options) { o.Addr = addr }
}

// NewServer creates a gRPC server with the given options.
func NewServer(opts ...Option) *Server {
	o := Options{Addr: ":9090"}
	for _, fn := range opts {
		fn(&o)
	}
	return &Server{
		srv:  grpc.NewServer(),
		opts: o,
	}
}

// Name implements lifecycle.Server.
func (s *Server) Name() string { return "grpc:" + s.opts.Addr }

// Server returns the underlying grpc.Server for service registration.
func (s *Server) Server() *grpc.Server { return s.srv }

// Run starts the gRPC server and blocks until SIGINT/SIGTERM.
func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return err
	}

	log.Println("[GRPC] listening on", s.opts.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		log.Println("[GRPC] shutting down...")
		s.srv.GracefulStop()
		log.Println("[GRPC] server exited")
	}
	return nil
}

// Stop triggers graceful stop.
func (s *Server) Stop() {
	s.srv.GracefulStop()
}
