// Package mongodb provides a MongoDB database abstraction.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connector manages a MongoDB connection.
type Connector interface {
	// Collection returns a named collection.
	Collection(name string) *mongo.Collection

	// Client returns the underlying mongo.Client.
	Client() *mongo.Client

	// Database returns the current *mongo.Database.
	Database() *mongo.Database

	// Close disconnects the client.
	Close(ctx context.Context) error
}

// Config holds MongoDB connection parameters.
type Config struct {
	URI          string
	Database     string
	MaxPoolSize  uint64
	MinPoolSize  uint64
	ReadOption   string // primary / secondaryPreferred / nearest
	Timeout      time.Duration
	MaxIdleTime  time.Duration
}

// MongoConnector is the default Connector implementation using the official driver.
type MongoConnector struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewConnector creates a MongoConnector from Config.
func NewConnector(cfg Config) (*MongoConnector, error) {
	opts := options.Client().ApplyURI(cfg.URI)
	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}
	if cfg.MaxIdleTime > 0 {
		opts.SetMaxConnIdleTime(cfg.MaxIdleTime)
	}
	if cfg.Timeout > 0 {
		opts.SetConnectTimeout(cfg.Timeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.Database)
	return &MongoConnector{client: client, db: db}, nil
}

func (c *MongoConnector) Collection(name string) *mongo.Collection  { return c.db.Collection(name) }
func (c *MongoConnector) Client() *mongo.Client                      { return c.client }
func (c *MongoConnector) Database() *mongo.Database                  { return c.db }
func (c *MongoConnector) Close(ctx context.Context) error            { return c.client.Disconnect(ctx) }

// — BaseModel types ——————————————————————————————————————————————

// BaseModel defines standard timestamps for MongoDB documents.
type BaseModel struct {
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}

// ModelWithAutoID uses an auto-increment int64 ID.
type ModelWithAutoID struct {
	ID        int64     `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}

// InitTime sets both timestamps.
func (m *ModelWithAutoID) InitTime(t time.Time) { m.CreatedAt = t; m.UpdatedAt = t }

// SetUpdatedTime sets UpdatedAt.
func (m *ModelWithAutoID) SetUpdatedTime(t time.Time) { m.UpdatedAt = t }

// ModelWithObjectID uses a MongoDB ObjectID.
type ModelWithObjectID struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

// SetObjectID generates a new ObjectID if not set.
func (m *ModelWithObjectID) SetObjectID() {
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
}
