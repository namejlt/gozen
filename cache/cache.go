// Package cache provides a unified Cache interface and a default Redis-backed
// implementation.  The interface is intentionally simple so that in-memory,
// Redis, or Memcached backends can be swapped transparently.
package cache

import (
	"context"
	"time"
)

// Cache is the fundamental key-value store abstraction.
type Cache interface {
	// Get retrieves a value.  Returns nil, nil if key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with an optional TTL.  ttl <= 0 means no expiry.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Del removes one or more keys.
	Del(ctx context.Context, keys ...string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Close cleans up resources.
	Close() error
}

// CacheConfig holds common cache parameters.
type CacheConfig struct {
	DefaultTTL time.Duration
	KeyPrefix  string
}
