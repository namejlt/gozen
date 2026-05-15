// Package handler implements HTTP handlers for the user-service.
//
// Each handler method is responsible for:
//  1. Binding and validating input parameters
//  2. Calling service layer
//  3. Writing structured responses via util helpers
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/namejlt/gozen/errors"
	"github.com/namejlt/gozen/examples/user-service/internal/model"
	"github.com/namejlt/gozen/examples/user-service/internal/service"
	"github.com/namejlt/gozen/util"
)

// UserHandler wires HTTP routes to the user service.
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes registers all user endpoints under the given router group.
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/users", h.Create)
	rg.GET("/users", h.List)
	rg.GET("/users/:id", h.GetByID)
	rg.PUT("/users/:id", h.Update)
	rg.DELETE("/users/:id", h.Delete)
}

// Create
// @Summary     Create a new user
// @Description Register a user with username and email
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body model.CreateUserRequest true "User info"
// @Success     201 {object} util.SuccessResponse{data=model.UserResponse}
// @Failure     400 {object} util.ErrorResponse
// @Router      /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, util.NewError(errors.CodeParamsIncomplete, "invalid request: "+err.Error()))
		return
	}
	resp, svcErr := h.svc.Create(c.Request.Context(), &req)
	if svcErr != nil {
		c.JSON(500, util.NewError(errors.CodeServerBusy, svcErr.Error()))
		return
	}
	c.JSON(201, util.NewSuccess(resp))
}

// List
// @Summary     List users
// @Description Get paginated user list
// @Tags        users
// @Produce     json
// @Param       page     query int false "Page number (default 1)"
// @Param       page_size query int false "Page size (default 20, max 100)"
// @Success     200 {object} util.SuccessResponse{data=model.PaginatedResponse}
// @Router      /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	resp, svcErr := h.svc.List(c.Request.Context(), page, pageSize)
	if svcErr != nil {
		c.JSON(500, util.NewError(errors.CodeServerBusy, svcErr.Error()))
		return
	}
	c.JSON(200, util.NewSuccess(resp))
}

// GetByID
// @Summary     Get a user by ID
// @Description Retrieve a single user
// @Tags        users
// @Produce     json
// @Param       id  path int true "User ID"
// @Success     200 {object} util.SuccessResponse{data=model.UserResponse}
// @Failure     404 {object} util.ErrorResponse
// @Router      /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, util.NewError(errors.CodeParamsIncomplete, "invalid user id"))
		return
	}
	resp, svcErr := h.svc.GetByID(c.Request.Context(), id)
	if svcErr != nil {
		c.JSON(404, util.NewError(errors.CodeBusinessError, svcErr.Error()))
		return
	}
	c.JSON(200, util.NewSuccess(resp))
}

// Update
// @Summary     Update a user
// @Description Modify user fields (partial update)
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id      path int           true "User ID"
// @Param       request body model.UpdateUserRequest true "Fields to update"
// @Success     200 {object} util.SuccessResponse{data=model.UserResponse}
// @Router      /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, util.NewError(errors.CodeParamsIncomplete, "invalid user id"))
		return
	}
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, util.NewError(errors.CodeParamsIncomplete, "invalid request: "+err.Error()))
		return
	}
	resp, svcErr := h.svc.Update(c.Request.Context(), id, &req)
	if svcErr != nil {
		c.JSON(500, util.NewError(errors.CodeServerBusy, svcErr.Error()))
		return
	}
	c.JSON(200, util.NewSuccess(resp))
}

// Delete
// @Summary     Delete a user
// @Description Remove a user by ID
// @Tags        users
// @Produce     json
// @Param       id  path int true "User ID"
// @Success     200 {object} util.SuccessResponse
// @Router      /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, util.NewError(errors.CodeParamsIncomplete, "invalid user id"))
		return
	}
	if svcErr := h.svc.Delete(c.Request.Context(), id); svcErr != nil {
		c.JSON(500, util.NewError(errors.CodeServerBusy, svcErr.Error()))
		return
	}
	c.JSON(200, util.NewSuccess(map[string]any{"id": id, "deleted": true}))
}
