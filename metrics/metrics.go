package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsOnce sync.Once
)

// HTTPMetrics Prometheus HTTP 指标
type HTTPMetrics struct {
	RequestTotal    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	InFlight        prometheus.Gauge
}

var defaultHTTPMetrics *HTTPMetrics

// InitMetrics 初始化并注册 Prometheus 指标
func InitMetrics() *HTTPMetrics {
	metricsOnce.Do(func() {
		defaultHTTPMetrics = &HTTPMetrics{
			RequestTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "http_requests_total",
					Help: "Total number of HTTP requests",
				},
				[]string{"method", "path", "status"},
			),
			RequestDuration: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "http_request_duration_ms",
					Help:    "HTTP request duration in milliseconds",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"method", "path"},
			),
			InFlight: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "http_requests_in_flight",
					Help: "Current number of in-flight HTTP requests",
				},
			),
		}
		prometheus.MustRegister(defaultHTTPMetrics.RequestTotal)
		prometheus.MustRegister(defaultHTTPMetrics.RequestDuration)
		prometheus.MustRegister(defaultHTTPMetrics.InFlight)
	})
	return defaultHTTPMetrics
}

// MetricsMiddleware Prometheus 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	m := InitMetrics()
	return func(c *gin.Context) {
		m.InFlight.Inc()
		defer m.InFlight.Dec()

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		m.RequestTotal.WithLabelValues(method, path, status).Inc()
		m.RequestDuration.WithLabelValues(method, path).Observe(float64(time.Since(start).Milliseconds()))
	}
}

// MetricsHandler 返回 /metrics 端点的 Handler
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
