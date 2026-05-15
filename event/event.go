// Package event provides an in-process publish-subscribe event bus.
// The EventBus interface supports topic-based pub/sub with async dispatch
// and concurrency control.
package event

// EventBus defines the pub/sub contract.
type Bus interface {
	// Subscribe registers a handler for a topic.
	Subscribe(topic string, fn any)

	// Unsubscribe removes a handler from a topic.
	Unsubscribe(topic string, fn any)

	// Publish enqueues an event for async dispatch.
	Publish(topic string, args ...any)

	// Stop shuts down the event bus.
	Stop()
}

// BusConfig configures a Bus.
type BusConfig struct {
	MaxConcurrent int // max concurrent goroutines handling events
}
