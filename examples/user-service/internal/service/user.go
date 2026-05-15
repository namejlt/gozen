// Package service implements business logic for the user-service.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/namejlt/gozen/event"
	"github.com/namejlt/gozen/examples/user-service/internal/model"
	"github.com/namejlt/gozen/examples/user-service/internal/repository"
	"github.com/namejlt/gozen/log"
)

// UserService implements user-related business logic.
type UserService struct {
	repo      *repository.UserRepository
	eventBus  event.Bus
}

// NewUserService creates a UserService.
func NewUserService(repo *repository.UserRepository, bus event.Bus) *UserService {
	return &UserService{repo: repo, eventBus: bus}
}

// Create registers a new user.
func (s *UserService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
	u := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		log.L().Errorw("failed to create user", "username", req.Username, "error", err)
		return nil, errors.New("create user failed")
	}

	// Publish async event for downstream handlers (e.g. send welcome email).
	s.eventBus.Publish("user.created", u.ID, u.Email)

	log.L().Infow("user created", "id", u.ID, "username", u.Username)
	return model.NewUserResponse(u), nil
}

// GetByID retrieves a user.
func (s *UserService) GetByID(ctx context.Context, id int) (*model.UserResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.L().Warnw("user not found", "id", id)
		return nil, errors.New("user not found")
	}
	return model.NewUserResponse(u), nil
}

// List returns a paginated user list.
func (s *UserService) List(ctx context.Context, page, pageSize int) (*model.PaginatedResponse, error) {
	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		log.L().Errorw("list users failed", "error", err)
		return nil, errors.New("list users failed")
	}
	items := make([]*model.UserResponse, 0, len(users))
	for _, u := range users {
		items = append(items, model.NewUserResponse(&u))
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return &model.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Update modifies a user's fields.
func (s *UserService) Update(ctx context.Context, id int, req *model.UpdateUserRequest) (*model.UserResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	u.SetUpdatedTime(time.Now())

	if err := s.repo.Update(ctx, u); err != nil {
		log.L().Errorw("update user failed", "id", id, "error", err)
		return nil, errors.New("update failed")
	}
	log.L().Infow("user updated", "id", id)
	return model.NewUserResponse(u), nil
}

// Delete removes a user.
func (s *UserService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		log.L().Errorw("delete user failed", "id", id, "error", err)
		return errors.New("delete failed")
	}
	log.L().Infow("user deleted", "id", id)
	return nil
}
