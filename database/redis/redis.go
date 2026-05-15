// Package redis provides a Redis cache / data-structure abstraction.
package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Connector manages a Redis client or cluster.
type Connector interface {
	// Client returns the underlying *redis.Client / *redis.ClusterClient.
	Client() redis.Cmdable

	// Close closes all connections.
	Close() error

	// Ping checks connectivity.
	Ping(ctx context.Context) error
}

// Config holds Redis connection parameters.
type Config struct {
	Address        []string // single or multi-node (cluster uses multiple)
	Password       string
	DB             int
	Cluster        bool
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ConnectTimeout time.Duration
	PoolMaxActive  int
	PoolMinIdle    int
	PoolIdleTimeout time.Duration
}

// SingleConnector wraps *redis.Client for standalone mode.
type SingleConnector struct {
	client *redis.Client
}

// NewSingleConnector creates a standalone Redis connector.
func NewSingleConnector(cfg Config) (*SingleConnector, error) {
	if len(cfg.Address) == 0 {
		cfg.Address = []string{"localhost:6379"}
	}
	opts := &redis.Options{
		Addr:         cfg.Address[0],
		Password:     cfg.Password,
		DB:           cfg.DB,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolMaxActive,
		MinIdleConns: cfg.PoolMinIdle,
		IdleTimeout:  cfg.PoolIdleTimeout,
	}
	client := redis.NewClient(opts)
	return &SingleConnector{client: client}, nil
}

func (s *SingleConnector) Client() redis.Cmdable         { return s.client }
func (s *SingleConnector) Close() error                   { return s.client.Close() }
func (s *SingleConnector) Ping(ctx context.Context) error  { return s.client.Ping(ctx).Err() }

// ClusterConnector wraps *redis.ClusterClient for cluster mode.
type ClusterConnector struct {
	client *redis.ClusterClient
}

// NewClusterConnector creates a cluster-mode Redis connector.
func NewClusterConnector(cfg Config) (*ClusterConnector, error) {
	if len(cfg.Address) == 0 {
		return nil, nil // invalid
	}
	opts := &redis.ClusterOptions{
		Addrs:        cfg.Address,
		Password:     cfg.Password,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolMaxActive,
		MinIdleConns: cfg.PoolMinIdle,
		IdleTimeout:  cfg.PoolIdleTimeout,
	}
	client := redis.NewClusterClient(opts)
	return &ClusterConnector{client: client}, nil
}

func (c *ClusterConnector) Client() redis.Cmdable         { return c.client }
func (c *ClusterConnector) Close() error                   { return c.client.Close() }
func (c *ClusterConnector) Ping(ctx context.Context) error  { return c.client.Ping(ctx).Err() }

// — ZSet helper ——————————————————————————————————————————————

// ZSetEntry is a single element in a sorted set.
type ZSetEntry struct {
	Key   string
	Score string
}
