# Project Structure Guide

## Recommended Layout for a Gozen-based Project

```
my-service/
├── main.go                     — Entry point: bootstrap + lifecycle.Run()
├── Dockerfile                  — Container build
├── Makefile                    — Build / test / lint targets
├── .golangci.yml               — Linter config
│
├── configs/                    — Application config files (YAML)
│   ├── app.yaml
│   └── db.yaml
│
├── internal/                   — Private application code
│   ├── model/                  — Domain types (entities, DTOs, request/response)
│   ├── repository/             — Data access layer (database queries)
│   ├── service/                — Business logic layer
│   └── handler/                — HTTP / gRPC handlers (parameter binding)
│
├── migrations/                 — Database migrations
├── proto/                      — Protobuf definitions (optional)
├── scripts/                    — Build / deploy scripts
└── tests/                      — Integration tests
```

## Standard Layer Responsibilities

### model/ — Domain Model

```go
// User represents the core user entity.
type User struct {
    ID        int64     `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"uniqueIndex"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest — API input.
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3"`
    Email    string `json:"email"    binding:"required,email"`
}

// UserResponse — API output.
type UserResponse struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}
```

### repository/ — Data Access

Uses gozen `database/*` interfaces:

```go
type UserRepository struct {
    db    mysql.Connector
    cache redis.Connector
}

func NewUserRepository(db mysql.Connector, cache redis.Connector) *UserRepository {
    return &UserRepository{db: db, cache: cache}
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    gdb, _ := r.db.Read(ctx)
    var u User
    if err := gdb.First(&u, id).Error; err != nil {
        return nil, err
    }
    return &u, nil
}
```

### service/ — Business Logic

```go
type UserService struct {
    repo *UserRepository
    bus  event.Bus
}

func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
    u := &User{Username: req.Username, Email: req.Email}
    if err := s.repo.Create(ctx, u); err != nil {
        log.L().Errorw("create user failed", "error", err)
        return nil, errors.New("create failed")
    }
    // Publish async event
    s.bus.Publish("user.created", u.ID)
    return NewUserResponse(u), nil
}
```

### handler/ — HTTP Handlers

```go
type UserHandler struct {
    svc *UserService
}

func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, util.NewError(1004, err.Error()))
        return
    }
    resp, err := h.svc.Create(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, util.NewError(1010, err.Error()))
        return
    }
    c.JSON(201, util.NewSuccess(resp))
}
```

## Migration from Legacy Structure

If you were using the old gozen structure (v1.x with dao/ / route/ / controller/ /
service/ / model/ directories in the project root):

| Old Path | New Path | Notes |
|----------|----------|-------|
| `controller/` | `internal/handler/` | Parameter binding + response |
| `service/*_logic.go` | `internal/service/` | Business logic |
| `dao/` | `internal/repository/` | Data access |
| `model/` | `internal/model/` | Domain types |
| `route/` | `server.go` or handler methods | Route registration in main |
| `pconst/` | `internal/model/` or dedicated | Constants |
| `middleware/` | `transport/http/middleware/` | Use framework middleware |
| `util/` | `util/` or `internal/` | Shared helpers |
