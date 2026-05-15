// Package gozen is a general-purpose business-development base framework (SDK)
// for Go.  It provides pluggable, interface-driven components for building
// production-grade microservices.
//
//	import "github.com/namejlt/gozen"
//
// ┌────────────────────── Architecture ──────────────────────────────────┐
// │                                                                      │
// │  gozen (root)          — framework entry point + documentation       │
// │                                                                      │
// │  errors/               — Error interface + code registry             │
// │  config/               — Config Manager + pluggable sources          │
// │     ├── FileSource                                                  │
// │     ├── NacosSource                                                 │
// │     └── Codec (YAML / JSON)                                         │
// │  log/                  — Logger interface + ZapLogger (default)      │
// │  trace/                — Tracer interface + SkyWalking / noop impl   │
// │  cache/                — Cache interface + Redis backend             │
// │  event/                — Bus interface + LocalEventBus               │
// │  database/             — Connector interfaces + GORM/default impls   │
// │     ├── mysql/         — MySQL / GORM                               │
// │     ├── mongodb/       — MongoDB                                     │
// │     ├── redis/         — Redis                                       │
// │     └── elasticsearch/ — Elasticsearch                               │
// │  transport/            — Transport layer                             │
// │     ├── http/          — HTTP Server + Gin router + Middleware       │
// │     └── grpc/          — gRPC Server + Connection Pool               │
// │  lifecycle/            — Application lifecycle management             │
// │  health/               — Health-check endpoints                     │
// │  metrics/              — Prometheus metrics                          │
// │  util/                 — General-purpose utility functions           │
// │  concurrent/           — Concurrency primitives                      │
// │  storage/              — Queue / Linked-list data structures         │
// │                                                                      │
// └──────────────────────────────────────────────────────────────────────┘
//
// Design principles:
//
//  1. Interface abstraction + default implementation
//     — every capability domain exposes an interface; the framework ships
//       a sensible default that can be swapped via dependency injection.
//
//  2. Explicit construction, no init() dependency
//     — components are constructed via New*() rather than implicit init().
//
//  3. Context propagation
//     — every I/O operation takes context.Context for tracing, deadlines,
//       and cancellation.
//
//  4. Option pattern
//     — optional configuration uses functional options (With*).
//
//  5. Concurrent safety
//     — all shared state is protected by sync primitives.
//
// ============================================================================
// Quick start
// ============================================================================
//
//	package main
//
//	import (
//		"context"
//
//		"github.com/namejlt/gozen"
//		"github.com/namejlt/gozen/config"
//		"github.com/namejlt/gozen/log"
//		"github.com/namejlt/gozen/lifecycle"
//		httptransport "github.com/namejlt/gozen/transport/http"
//		"github.com/namejlt/gozen/transport/http/middleware"
//	)
//
//	func main() {
//		// 1. Bootstrap config
//		mgr := config.NewManager()
//		mgr.AddSource(config.NewFileSource("./configs"))
//
//		var appCfg config.AppConfig
//		if err := mgr.Load("app", &appCfg); err != nil {
//			panic(err)
//		}
//
//		// 2. Bootstrap logging
//		zl, _ := log.NewZapLogger(log.Config{
//			Name:  "my-service",
//			Path:  "./logs/app.log",
//			Debug: true,
//		})
//		log.SetLogger(zl)
//
//		// 3. Build HTTP server
//		srv := httptransport.NewServer(
//			httptransport.WithAddr(":8080"),
//		)
//		srv.Router().GET("/ping", middlewarePing)
//
//		// 4. Run with lifecycle
//		app := lifecycle.New("my-service")
//		app.AddServer(srv)
//		app.AddHook(lifecycle.Hook{
//			Name:    "cleanup",
//			OnStop:  func(ctx context.Context) error { return nil },
//			Timeout: 5 * time.Second,
//		})
//		if err := app.Run(); err != nil {
//			log.L().Fatalw("app exited", "err", err)
//		}
//	}
package gozen

// Version is the framework version.
const Version = "2.0.0"
