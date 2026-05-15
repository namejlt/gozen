// Package repository implements data access for the user-service.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/namejlt/gozen/database/mysql"
	"github.com/namejlt/gozen/database/redis"
	"github.com/namejlt/gozen/examples/user-service/internal/model"
)

// UserRepository handles user persistence (MySQL + Redis cache).
type UserRepository struct {
	db    mysql.Connector
	cache redis.Connector
}

// NewUserRepository creates a UserRepository.
func NewUserRepository(db mysql.Connector, cache redis.Connector) *UserRepository {
	return &UserRepository{db: db, cache: cache}
}

// Create inserts a new user and returns it with the generated ID.
func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	db, err := r.db.Write(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}
	now := time.Now()
	u.InitTime(now)
	return db.Table("users").Create(u).Error
}

// GetByID retrieves a user by ID (tries cache first, then DB).
func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	// Try cache first
	cached, err := r.getFromCache(ctx, id)
	if err == nil && cached != nil {
		return cached, nil
	}

	db, err := r.db.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db: %w", err)
	}
	var u model.User
	if err := db.Table("users").First(&u, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	_ = r.setCache(ctx, &u) // best-effort cache write
	return &u, nil
}

// List returns a paginated list of users.
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	db, err := r.db.Read(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("get db: %w", err)
	}
	var total int64
	if err := db.Table("users").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var users []model.User
	if err := db.Table("users").
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// Update persists changes to an existing user.
func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	db, err := r.db.Write(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}
	u.SetUpdatedTime(time.Now())

	// Invalidate cache on update
	_ = r.delCache(ctx, u.ID)

	return db.Table("users").Where("id = ?", u.ID).Updates(u).Error
}

// Delete soft-deletes (or hard-deletes) a user.
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	db, err := r.db.Write(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}
	_ = r.delCache(ctx, id)
	return db.Table("users").Delete(&model.User{}, id).Error
}

// — cache helpers -----------------------------------------------------------

func cacheKey(id int) string { return fmt.Sprintf("user:%d", id) }

func (r *UserRepository) getFromCache(ctx context.Context, id int) (*model.User, error) {
	if r.cache == nil {
		return nil, nil
	}
	data, err := r.cache.Client().Get(ctx, cacheKey(id)).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var uc model.UserCache
	if err := json.Unmarshal(data, &uc); err != nil {
		return nil, err
	}
	return userFromCache(&uc), nil
}

func (r *UserRepository) setCache(ctx context.Context, u *model.User) error {
	if r.cache == nil {
		return nil
	}
	data, err := json.Marshal(u.ToCache())
	if err != nil {
		return err
	}
	return r.cache.Client().Set(ctx, cacheKey(u.ID), data, 15*time.Minute).Err()
}

func (r *UserRepository) delCache(ctx context.Context, id int) error {
	if r.cache == nil {
		return nil
	}
	return r.cache.Client().Del(ctx, cacheKey(id)).Err()
}

func userFromCache(uc *model.UserCache) *model.User {
	return &model.User{
		BaseModel: mysql.BaseModel{
			ID:        uc.ID,
			CreatedAt: uc.CreatedAt,
		},
		Username: uc.Username,
		Email:    uc.Email,
		Status:   uc.Status,
	}
}


