// Package elasticsearch provides an Elasticsearch client abstraction.
//
// The Connector interface wraps the olivere/elastic v7 client.  For v3/v5/v6
// compatibility, wrap the respective client behind the same interface.
package elasticsearch

import (
	"context"
	"time"
)

// Connector manages an Elasticsearch client.
type Connector interface {
	// Index indexes a document (create / replace).
	Index(ctx context.Context, index, id, typeName string, body any) error

	// Update partially updates a document.
	Update(ctx context.Context, index, id string, doc any) error

	// Delete removes a document.
	Delete(ctx context.Context, index, id string) error

	// Search executes a search query.
	Search(ctx context.Context, index string, query any) ([]byte, error)

	// BulkIndex indexes multiple documents in one request.
	BulkIndex(ctx context.Context, index string, docs map[string]any) error

	// Close closes idle connections.
	Close() error
}

// Config holds Elasticsearch connection parameters.
type Config struct {
	Address          []string
	TransportMaxIdle int
	Timeout          time.Duration
	// Sniff controls whether the client sniffs the cluster state.
	Sniff bool
	// Healthcheck controls periodic health checks.
	Healthcheck bool
}
