// Command user-service is a complete web business application example built
// on the gozen framework. It demonstrates:
//
//   - Configuration via config.FileSource
//   - Structured logging with ZapLogger
//   - HTTP server with gin middleware stack
//   - MySQL database via database/mysql.Connector
//   - Redis cache via database/redis.Connector
//   - Event bus for async pub/sub
//   - Layered architecture (handler → service → repository)
//   - Lifecycle management for graceful shutdown
//   - Health check endpoints
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/namejlt/gozen/config"
	"github.com/namejlt/gozen/database/mysql"
	"github.com/namejlt/gozen/database/redis"
	"github.com/namejlt/gozen/event"
	"github.com/namejlt/gozen/examples/user-service/internal/handler"
	"github.com/namejlt/gozen/examples/user-service/internal/repository"
	"github.com/namejlt/gozen/examples/user-service/internal/service"
	"github.com/namejlt/gozen/lifecycle"
	"github.com/namejlt/gozen/log"
	httptransport "github.com/namejlt/gozen/transport/http"
)

func main() {
	if err := run(); err != nil {
		log.L().Fatalw("application terminated with error", "error", err)
	}
}

func run() error {
	// ============================================================
	// 1. Bootstrap configuration
	// ============================================================
	cfgMgr := config.NewManager()
	cfgMgr.AddSource(config.NewFileSource("./configs"))

	var appCfg config.AppConfig
	if err := cfgMgr.Load("app", &appCfg); err != nil {
		return fmt.Errorf("load app config: %w", err)
	}

	var dbCfg config.DBConfig
	if err := cfgMgr.Load("db", &dbCfg); err != nil {
		return fmt.Errorf("load db config: %w", err)
	}

	// ============================================================
	// 2. Bootstrap logging
	// ============================================================
	zl, err := log.NewZapLogger(log.Config{
		Name:  "user-service",
		Path:  "./logs/app.log",
		Debug: true,
	})
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	log.SetLogger(zl)
	log.L().Infow("starting user-service", "env", appCfg.Env)

	// ============================================================
	// 3. Bootstrap database connectors
	// ============================================================

	// MySQL — build DSN from config
	mysqlWriteDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		dbCfg.Mysql.Write.User,
		dbCfg.Mysql.Write.Password,
		dbCfg.Mysql.Write.Address,
		dbCfg.Mysql.Write.Port,
		dbCfg.Mysql.DbName,
	)
	mysqlReadDSN := mysqlWriteDSN // use same DSN for read/write if single node
	if len(dbCfg.Mysql.Reads) > 0 {
		r := dbCfg.Mysql.Reads[0]
		mysqlReadDSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
			r.User, r.Password, r.Address, r.Port, dbCfg.Mysql.DbName)
	}

	myCfg := mysql.DBConfig{
		MaxIdleConns: dbCfg.Mysql.Pool.MinCap,
		MaxOpenConns: dbCfg.Mysql.Pool.MaxCap,
		MaxLifetime:  time.Duration(dbCfg.Mysql.Pool.LifeTimeout) * time.Millisecond,
		MaxIdleTime:  time.Duration(dbCfg.Mysql.Pool.IdleTimeout) * time.Millisecond,
	}
	mysqlConn, err := mysql.NewGormConnector(mysqlReadDSN, mysqlWriteDSN, myCfg)
	if err != nil {
		return fmt.Errorf("mysql connect: %w", err)
	}
	log.L().Infow("mysql connected")

	// Redis
	var redisConn *redis.SingleConnector
	if len(dbCfg.Redis.Address) > 0 {
		redisConn, err = redis.NewSingleConnector(redis.Config{
			Address:  dbCfg.Redis.Address,
			Password: dbCfg.Redis.Password,
			DB:       dbCfg.Redis.DB,
		})
		if err != nil {
			return fmt.Errorf("redis connect: %w", err)
		}
		log.L().Infow("redis connected")
	}

	// ============================================================
	// 4. Bootstrap event bus
	// ============================================================
	bus := event.NewLocalBus(20)
	bus.Subscribe("user.created", handleUserCreated)
	defer bus.Stop()

	// ============================================================
	// 5. Build application layers
	// ============================================================
	userRepo := repository.NewUserRepository(mysqlConn, redisConn)
	userSvc := service.NewUserService(userRepo, bus)
	userHdr := handler.NewUserHandler(userSvc)

	// ============================================================
	// 6. Build HTTP server
	// ============================================================
	srv := httptransport.NewServer(
		httptransport.WithAddr(appCfg.Host),
		httptransport.WithMode("release"),
	)
	r := srv.Router()

	// Health check / monitor / metrics
	// These are auto-wired via NewRouter — we just add business routes.
	api := r.Group("/api/v1")
	userHdr.RegisterRoutes(api)

	// Simple ping route
	r.GET("/ping", pingHandler)

	// ============================================================
	// 7. Run with lifecycle manager
	// ============================================================
	app := lifecycle.New("user-service")
	app.AddServer(srv)

	// Register shutdown hooks for database connections
	app.AddHook(lifecycle.Hook{
		Name: "mysql-close",
		OnStop: func(ctx context.Context) error {
			return mysqlConn.Close()
		},
		Timeout: 10 * time.Second,
	})
	if redisConn != nil {
		app.AddHook(lifecycle.Hook{
			Name: "redis-close",
			OnStop: func(ctx context.Context) error {
				return redisConn.Close()
			},
			Timeout: 5 * time.Second,
		})
	}

	log.L().Infow("user-service ready", "addr", appCfg.Host)
	return app.Run()
}

// pingHandler responds to health/liveness checks.
func pingHandler(c *gin.Context) {
	c.JSON(200, map[string]any{"message": "pong"})
}

// handleUserCreated is an async event handler for user creation.
func handleUserCreated(userID int, email string) {
	log.L().Infow("async: user created event received",
		"user_id", userID,
		"email", email,
	)
	// In a real app: send welcome email, index for search, etc.
}
