// Package trace provides a distributed-tracing abstraction backed by Apache
// SkyWalking (go2sky) by default.  The Tracer interface allows any APM
// system to be plugged in (Jaeger, OpenTelemetry, Zipkin, …).
package trace

import (
	"context"
)

// SpanLayer categorises the instrumented operation.
type SpanLayer int32

const (
	LayerUnknown      SpanLayer = 0
	LayerDatabase     SpanLayer = 1
	LayerRPCFramework  SpanLayer = 2
	LayerHTTP         SpanLayer = 3
	LayerMQ           SpanLayer = 4
	LayerCache        SpanLayer = 5
)

// Span represents a single tracing span.
type Span interface {
	End()
	SetTag(key, value string)
	SetError(err error)
	SetOperationName(name string)
}

// Tracer is the framework tracing abstraction.
type Tracer interface {
	// StartEntrySpan starts a server-side (entry) span.
	StartEntrySpan(ctx context.Context, operationName string, layer SpanLayer) (Span, context.Context)

	// StartLocalSpan starts a local (internal) span.
	StartLocalSpan(ctx context.Context, operationName string, layer SpanLayer) (Span, context.Context)

	// StartExitSpan starts a client-side (exit) span.
	StartExitSpan(ctx context.Context, peer, operationName string, layer SpanLayer) (Span, context.Context)

	// Inject propagates span context into carrier for downstream propagation.
	Inject(ctx context.Context, carrier map[string]string) error

	// Extract reconstructs a context from carrier (incoming request).
	Extract(ctx context.Context, carrier map[string]string) (context.Context, error)

	// Close flushes and shuts down the tracer.
	Close() error
}

// NoopTracer is a no-op implementation for environments where tracing is
// disabled.
type NoopTracer struct{}

func (NoopTracer) StartEntrySpan(ctx context.Context, op string, layer SpanLayer) (Span, context.Context) {
	return noopSpan{}, ctx
}
func (NoopTracer) StartLocalSpan(ctx context.Context, op string, layer SpanLayer) (Span, context.Context) {
	return noopSpan{}, ctx
}
func (NoopTracer) StartExitSpan(ctx context.Context, peer, op string, layer SpanLayer) (Span, context.Context) {
	return noopSpan{}, ctx
}
func (NoopTracer) Inject(ctx context.Context, carrier map[string]string) error  { return nil }
func (NoopTracer) Extract(ctx context.Context, carrier map[string]string) (context.Context, error) {
	return ctx, nil
}
func (NoopTracer) Close() error { return nil }

type noopSpan struct{}

func (noopSpan) End()                        {}
func (noopSpan) SetTag(string, string)       {}
func (noopSpan) SetError(error)              {}
func (noopSpan) SetOperationName(string)     {}
