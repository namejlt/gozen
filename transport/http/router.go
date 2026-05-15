package http

import (
	"os"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/contrib/expvar"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"

	"github.com/namejlt/gozen/health"
	"github.com/namejlt/gozen/metrics"
	"github.com/namejlt/gozen/transport/http/middleware"
)

// NewRouter creates a gin.Engine pre-configured with the standard
// middleware stack and framework routes (monitor, metrics, health, swagger).
//
//	param[0] = swagger instance name (optional)
func NewRouter(isDev bool, swaggerName ...string) *gin.Engine {
	if isDev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.CORS())
	r.Use(metrics.MetricsMiddleware())

	if isDev {
		ginpprof.Wrap(r)
		r.Use(gin.Logger())
	}

	routeMonitor(r)
	routeMetrics(r)
	health.RouteHealth(r)

	if len(swaggerName) > 0 && swaggerName[0] != "" {
		r.GET("/swagger/*any", gs.WrapHandler(swaggerfiles.Handler,
			gs.InstanceName(swaggerName[0])))
	}

	r.GET("/debug/vars", expvar.Handler())

	return r
}

func routeMonitor(r *gin.Engine) {
	r.GET("/monitor", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code_version": getEnv("CODE_VERSION", "unknown"),
			"build_time":   getEnv("BUILD_TIME", "unknown"),
			"system_time":  time.Now().Format("2006-01-02 15:04:05"),
		})
	})
}

func routeMetrics(r *gin.Engine) {
	r.GET("/metrics", metrics.MetricsHandler())
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
