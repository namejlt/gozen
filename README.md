# Gozen

**Go通用业务开发底层基座框架** — Interface-driven, production-grade Go SDK for microservices.

[![Go Reference](https://pkg.go.dev/badge/github.com/namejlt/gozen)](https://pkg.go.dev/github.com/namejlt/gozen)
[![Go Report Card](https://goreportcard.com/badge/github.com/namejlt/gozen)](https://goreportcard.com/report/github.com/namejlt/gozen)

## Features

- **Pluggable Configuration** — File, Nacos, or custom sources via `config.Source` interface
- **Structured Logging** — `log.Logger` interface with Zap(lumberjack) default
- **Distributed Tracing** — `trace.Tracer` interface with Apache SkyWalking implementation
- **Database Abstractions** — `database/mysql`, `database/redis`, `database/mongodb`, `database/elasticsearch`
- **HTTP Transport** — Gin-based server with middleware (Recovery, RequestID, AccessLog, CORS)
- **gRPC Transport** — Server + connection pool with round-robin load balancing
- **Event Bus** — In-process pub/sub with async dispatch
- **Unified Errors** — `errors.Error` interface with code registry + hot-reload
- **Lifecycle Management** — Graceful startup/shutdown with signal handling
- **Health Checks** — `/health`, `/health/ready`, `/health/live` endpoints
- **Prometheus Metrics** — Request count, duration, and in-flight gauges
- **Concurrency Primitives** — Semaphore-based concurrency limiter

## Quick Start

```go
package main

import (
    "context"
    "time"

    "github.com/namejlt/gozen/config"
    "github.com/namejlt/gozen/log"
    "github.com/namejlt/gozen/lifecycle"
    httptransport "github.com/namejlt/gozen/transport/http"
)

func main() {
    // Config
    mgr := config.NewManager()
    mgr.AddSource(config.NewFileSource("./configs"))
    var appCfg config.AppConfig
    mgr.Load("app", &appCfg)

    // Logger
    zl, _ := log.NewZapLogger(log.Config{Name: "my-service", Debug: true})
    log.SetLogger(zl)

    // HTTP Server
    srv := httptransport.NewServer(httptransport.WithAddr(":8080"))
    srv.Router().GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // Run
    app := lifecycle.New("my-service")
    app.AddServer(srv)
    if err := app.Run(); err != nil {
        log.L().Fatalw("exited", "error", err)
    }
}
```

## Full Example

See [examples/user-service](./examples/user-service) for a complete web business
project with MySQL, Redis, event bus, and layered architecture.

## Documentation

- [Architecture Guide](./docs/readme.md)
- [Configuration Guide](./docs/config.md)
- [Project Structure Guide](./docs/project.md)
- [Debug Variables](./docs/debug_vars.md)
- [Swagger Guide](./docs/swaggo.md)

## License

MIT
