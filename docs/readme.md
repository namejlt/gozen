# Gozen Framework — Architecture Guide

## Design Philosophy

Gozen is an interface-driven business-development base framework for Go.
Every capability domain exports a **pure interface** and ships a **sensible default
implementation** that can be swapped via dependency injection.

## Package Map

```
gozen/ (root)           — Entry point: gozen.go (documentation) + gozen.Version
├── config/             — Pluggable configuration centre
│   ├── Manager         — Load/Watch orchestrator
│   ├── Source          — Interface for config backends
│   │   ├── FileSource  — YAML/JSON files in a local directory
│   │   └── NacosSource — Remote Nacos config centre
│   └── Codec           — Marshal/Unmarshal interface (YAML / JSON)
├── errors/             — Unified error handling
│   ├── Error           — Interface: Code() / Message() / Unwrap()
│   ├── AppError        — Default implementation
│   └── Registry        — Code→Message lookup with hot-reload
├── log/                — Structured logging
│   ├── Logger          — Interface: Debugw / Infow / Errorw / etc.
│   ├── ZapLogger       — Default: uber-go/zap + lumberjack rotation
│   └── Config          — Log level / path / rotation settings
├── trace/              — Distributed tracing
│   ├── Tracer          — Interface: StartEntrySpan / StartExitSpan
│   ├── NoopTracer      — Default (disabled)
│   ├── SkyWalking      — Optional Apache SkyWalking (go2sky) implementation
│   └── Config          — Tracer bootstrap config
├── cache/              — Cache abstraction
│   └── Cache           — Interface: Get / Set / Del / Exists
├── event/              — In-process event bus
│   ├── Bus             — Interface: Subscribe / Publish / Stop
│   └── LocalEventBus   — Default: async pub/sub with concurrency limit
├── database/           — Data access layer
│   ├── mysql/          — Connector + Repository interfaces, GormConnector
│   ├── mongodb/        — Connector interface, MongoConnector
│   ├── redis/          — Connector interface, SingleConnector / ClusterConnector
│   └── elasticsearch/  — Connector interface, ES7Connector
├── transport/          — Network transport layer
│   ├── http/           — HTTP server (gin-based)
│   │   ├── Server      — Graceful shutdown, lifecycle.Server
│   │   ├── Router      — Factory with default middleware stack
│   │   └── middleware/ — Recovery / RequestID / AccessLog / CORS
│   └── grpc/           — gRPC server + connection pool
│       ├── Server      — Graceful shutdown, lifecycle.Server
│       └── pool/       — Connection pool with round-robin LB
├── lifecycle/          — Application orchestration
│   └── Application     — Server + Hook management, signal handling
├── health/             — Health-check endpoints (/health, /ready, /live)
├── metrics/            — Prometheus metrics (counter, histogram, in-flight)
├── util/               — General-purpose helpers
│   ├── response.go     — API response utilities
│   ├── pagination.go   — PageRequest / PageResult
│   ├── conv.go         — Type conversion
│   ├── slice.go        — Slice operations
│   └── ...
├── concurrent/         — Concurrency primitives
└── storage/            — Linked list / queue data structures
```

## Layer Interaction

```
        HTTP / gRPC Request
               │
               ▼
  ┌─────────────────────────┐
  │  transport/http/        │  ← Middleware (Recovery, RequestID, AccessLog, CORS)
  │  transport/grpc/        │
  └────────┬────────────────┘
           │
           ▼
  ┌─────────────────────────┐
  │  handler (user code)    │  ← Parameter binding, response writing
  └────────┬────────────────┘
           │
           ▼
  ┌─────────────────────────┐
  │  service (user code)    │  ← Business logic, event publishing
  └────────┬────────────────┘
           │
           ▼
  ┌─────────────────────────┐
  │  repository (user code) │  ← Data access via database/* interfaces
  └────────┬────────────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
  database/   cache/       ← Connector interfaces
  (MySQL)     (Redis)
```

## Typical Bootstrap Sequence

```go
func main() {
    // 1. Config
    mgr := config.NewManager()
    mgr.AddSource(config.NewFileSource("./configs"))
    var cfg config.AppConfig
    mgr.Load("app", &cfg)

    // 2. Logger
    zl, _ := log.NewZapLogger(log.Config{Name: "myapp", Path: "./logs", Debug: true})
    log.SetLogger(zl)

    // 3. Database
    conn, _ := mysql.NewGormConnector(readDSN, writeDSN, mysql.DBConfig{...})

    // 4. Event bus
    bus := event.NewLocalBus(20)

    // 5. HTTP server
    srv := httptransport.NewServer(httptransport.WithAddr(":8080"))
    srv.Router().GET("/ping", pingHandler)

    // 6. Lifecycle
    app := lifecycle.New("myapp")
    app.AddServer(srv)
    app.Run()
}
```

## Design Rules

| Rule | Explanation |
|------|-------------|
| Interface first | Every domain has an exported interface |
| Default impl | Framework ships one, swap via SetXxx() or constructor injection |
| No implicit init | All components are explicitly constructed |
| Context propagation | Every I/O operation receives `context.Context` |
| Option pattern | Config uses functional options: `WithAddr(":8080")` |
| Concurrent safe | Shared state is protected by `sync.Mutex`, `sync.RWMutex`, or `atomic` |
| No package globals | Logger, tracer, etc. are instance-based |
