// Package model defines domain types for the user-service example.
package model

import (
	"time"

	"github.com/namejlt/gozen/database/mysql"
)

// User represents a user entity in the system.
type User struct {
	mysql.BaseModel
	Username string `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Email    string `json:"email"    gorm:"size:128;not null"`
	Phone    string `json:"phone"    gorm:"size:32"`
	Avatar   string `json:"avatar"   gorm:"size:256"`
	Status   int    `json:"status"   gorm:"default:1"`
}

// UserCache is the Redis cache representation.
type UserCache struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToCache converts a User to a UserCache.
func (u *User) ToCache() *UserCache {
	return &UserCache{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

// CreateUserRequest is the API request body for creating a user.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Email    string `json:"email"    binding:"required,email"`
	Phone    string `json:"phone"`
}

// UpdateUserRequest is the API request body for updating a user.
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Phone  string `json:"phone"`
	Status *int   `json:"status"`
}

// UserResponse is the API response body for a user.
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUserResponse builds a response from a User model.
func NewUserResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		Avatar:    u.Avatar,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// PaginatedResponse wraps a paginated list response.
type PaginatedResponse struct {
	Items      any   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}
