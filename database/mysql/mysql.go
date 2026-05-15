// Package mysql provides a MySQL database abstraction built on GORM.
package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connector manages a MySQL connection pool.  Implementations must support
// read/write splitting.
type Connector interface {
	// Read returns a read-only *gorm.DB handle.
	Read(ctx context.Context) (*gorm.DB, error)

	// Write returns a read-write *gorm.DB handle.
	Write(ctx context.Context) (*gorm.DB, error)

	// Close shuts down all connections.
	Close() error
}

// Repository provides generic CRUD operations over a named table.
// This is a minimal interface — real repositories embed or extend it.
type Repository interface {
	// Insert creates a new row.
	Insert(ctx context.Context, entity any) error

	// Find queries rows into a slice.
	Find(ctx context.Context, condition string, dest any, sort string, offset, limit int) error

	// First queries a single row by condition.
	First(ctx context.Context, condition string, dest any, sort string) error

	// Update modifies matching rows.
	Update(ctx context.Context, condition string, updates any) error

	// Delete removes matching rows.
	Delete(ctx context.Context, condition string, model any) error
}

// BaseModel is the standard model struct for MySQL tables.
type BaseModel struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InitTime sets both CreatedAt and UpdatedAt.
func (m *BaseModel) InitTime(t time.Time) {
	m.CreatedAt = t
	m.UpdatedAt = t
}

// SetUpdatedTime updates the UpdatedAt timestamp.
func (m *BaseModel) SetUpdatedTime(t time.Time) { m.UpdatedAt = t }

// — Default Connector implementation (GORM) — -------------------------------

// GormConnector implements Connector using GORM + go-sql-driver/mysql.
type GormConnector struct {
	read  *gorm.DB
	write *gorm.DB
}

// DBConfig holds MySQL connection parameters.
type DBConfig struct {
	DSN          string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
	MaxIdleTime  time.Duration
}

// NewGormConnector creates a Connector with separate read/write DSNs.
// If writeDSN is empty, readDSN is used for both.
func NewGormConnector(readDSN, writeDSN string, cfg DBConfig) (*GormConnector, error) {
	if readDSN == "" {
		return nil, errors.New("mysql: empty read DSN")
	}
	rdb, err := openDB(readDSN, cfg)
	if err != nil {
		return nil, fmt.Errorf("mysql read: %w", err)
	}
	wdb := rdb
	if writeDSN != "" && writeDSN != readDSN {
		wdb, err = openDB(writeDSN, cfg)
		if err != nil {
			return nil, fmt.Errorf("mysql write: %w", err)
		}
	}
	return &GormConnector{read: rdb, write: wdb}, nil
}

func openDB(dsn string, cfg DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	return db, nil
}

// Read implements Connector.
func (g *GormConnector) Read(ctx context.Context) (*gorm.DB, error) {
	return g.read.WithContext(ctx), nil
}

// Write implements Connector.
func (g *GormConnector) Write(ctx context.Context) (*gorm.DB, error) {
	return g.write.WithContext(ctx), nil
}

// Close implements Connector.
func (g *GormConnector) Close() error {
	var errs []error
	if db, e := g.read.DB(); e == nil {
		errs = append(errs, db.Close())
	}
	if g.write != g.read {
		if db, e := g.write.DB(); e == nil {
			errs = append(errs, db.Close())
		}
	}
	return errors.Join(errs...)
}
