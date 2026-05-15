package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthStatus 健康检查状态
type HealthStatus string

const (
	HealthStatusUp   HealthStatus = "up"
	HealthStatusDown HealthStatus = "down"
)

// HealthCheck 单个检查项
type HealthCheck struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
}

// HealthReport 健康检查报告
type HealthReport struct {
	Status  HealthStatus            `json:"status"`
	Version string                  `json:"version,omitempty"`
	Uptime  string                  `json:"uptime,omitempty"`
	Checks  map[string]*HealthCheck `json:"checks,omitempty"`
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) *HealthCheck
}

var (
	healthCheckers   []HealthChecker
	healthCheckersMu sync.RWMutex
	startTime        = time.Now()
)

// RegisterHealthChecker 注册健康检查器
func RegisterHealthChecker(checker HealthChecker) {
	healthCheckersMu.Lock()
	defer healthCheckersMu.Unlock()
	healthCheckers = append(healthCheckers, checker)
}

func runHealthChecks(ctx context.Context) map[string]*HealthCheck {
	healthCheckersMu.RLock()
	checkers := make([]HealthChecker, len(healthCheckers))
	copy(checkers, healthCheckers)
	healthCheckersMu.RUnlock()

	checks := make(map[string]*HealthCheck, len(checkers))
	for _, c := range checkers {
		checks[c.Name()] = c.Check(ctx)
	}
	return checks
}

func healthHandler(c *gin.Context) {
	checks := runHealthChecks(c.Request.Context())
	overall := HealthStatusUp
	for _, check := range checks {
		if check.Status != HealthStatusUp {
			overall = HealthStatusDown
			break
		}
	}
	c.JSON(http.StatusOK, &HealthReport{
		Status:  overall,
		Version: getEnv("CODE_VERSION", "unknown"),
		Uptime:  time.Since(startTime).String(),
	})
}

func healthReadyHandler(c *gin.Context) {
	checks := runHealthChecks(c.Request.Context())
	overall := HealthStatusUp
	for _, check := range checks {
		if check.Status != HealthStatusUp {
			overall = HealthStatusDown
			break
		}
	}
	status := http.StatusOK
	if overall == HealthStatusDown {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, &HealthReport{
		Status: overall,
		Checks: checks,
	})
}

func healthLiveHandler(c *gin.Context) {
	c.JSON(http.StatusOK, &HealthReport{Status: HealthStatusUp})
}

// RouteHealth 注册健康检查路由
func RouteHealth(r *gin.Engine) {
	r.GET("/health", healthHandler)
	r.GET("/health/ready", healthReadyHandler)
	r.GET("/health/live", healthLiveHandler)
}

// getEnv 获取环境变量，不存在则返回默认值
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

// ========= 内置健康检查器 =========

type mysqlHealthChecker struct {
	name string
	db   *gorm.DB
}

func (m *mysqlHealthChecker) Name() string { return m.name }

func (m *mysqlHealthChecker) Check(ctx context.Context) *HealthCheck {
	sqlDB, err := m.db.DB()
	if err != nil {
		return &HealthCheck{Status: HealthStatusDown, Message: err.Error()}
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return &HealthCheck{Status: HealthStatusDown, Message: err.Error()}
	}
	stats := sqlDB.Stats()
	return &HealthCheck{
		Status:  HealthStatusUp,
		Message: fmt.Sprintf("open=%d in_use=%d idle=%d", stats.OpenConnections, stats.InUse, stats.Idle),
	}
}

// RegisterMySQLHealthCheck 自动注册 MySQL 健康检查
func RegisterMySQLHealthCheck(name string, db *gorm.DB) {
	if db != nil {
		RegisterHealthChecker(&mysqlHealthChecker{name: name, db: db})
	}
}
