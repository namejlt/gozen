# Swagger API Documentation

The framework supports auto-generated OpenAPI / Swagger documentation
via [swaggo](https://github.com/swaggo/swag).

## 1. Install Swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## 2. Annotate Your Code

Add declarative comments to your handler functions:

```go
// package handler

// CreateUser
// @Summary     Create a new user
// @Description Register a user with username and email
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body CreateUserRequest true "User info"
// @Success     201 {object} util.SuccessResponse{data=UserResponse}
// @Failure     400 {object} util.ErrorResponse
// @Router      /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
    // ...
}
```

## 3. Generate Docs

```bash
# From your project root
swag init -g main.go -o docs/swagger
```

This creates `docs/swagger/docs.go`, `docs/swagger/swagger.json`, and
`docs/swagger/swagger.yaml`.

## 4. Import Generated Docs in main.go

```go
package main

import (
    _ "my-service/docs/swagger"
)
```

## 5. Enable Swagger in Config

```yaml
# configs/app.yaml
docs: swagger
```

## 6. Access the UI

```
http://localhost:8080/swagger/index.html
```

## Reference

- [swaggo/gin-swagger](https://github.com/swaggo/gin-swagger)
- [Declarative Comments Format](https://github.com/swaggo/swag#declarative-comments-format)
