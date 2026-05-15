// Package middleware provides standard HTTP middleware for gin-gonic/gin.
// Each middleware is exported as a constructor returning gin.HandlerFunc.
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// — Recovery ----------------------------------------------------------------

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// Recovery returns a middleware that recovers from panics, logs the stack
// trace to stderr, and returns 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := stack(3)
				httpRequest, _ := httputil.DumpRequest(c.Request, true)
				fmt.Fprintf(os.Stderr, "[PANIC] %v\n%s\nrequest: %s\nform: %v\n",
					err, stack, httpRequest, c.Request.Form)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func source(lines [][]byte, n int) []byte {
	n--
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	return bytes.Replace(name, centerDot, dot, -1)
}

// — RequestID ---------------------------------------------------------------

const (
	headerRequestID = "X-Request-Id"
	ctxKeyRequestID = "RequestId"
)

// RequestID generates or propagates a unique request ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(ctxKeyRequestID, rid)
		c.Header(headerRequestID, rid)
		c.Next()
	}
}

// GetRequestID extracts the request ID from the gin context.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if rid, ok := v.(string); ok {
			return rid
		}
	}
	return ""
}

// — AccessLog ---------------------------------------------------------------

// AccessLog logs structured request information after each request.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		var body string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			data, err := io.ReadAll(c.Request.Body)
			if err == nil {
				body = string(data)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
			}
		}

		c.Next()

		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		fmt.Fprintf(os.Stdout,
			"[ACCESS] %s %s %d %dms ip=%s rid=%s",
			method, path, status, latency, clientIP, GetRequestID(c),
		)
		if body != "" && status >= 400 {
			fmt.Fprintf(os.Stdout, " body=%s", body)
		}
		fmt.Fprintln(os.Stdout)
	}
}

// — CORS --------------------------------------------------------------------

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           string
}

// DefaultCORSConfig returns a permissive default.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           "43200",
	}
}

// CORS returns a CORS middleware with optional config override.
func CORS(config ...CORSConfig) gin.HandlerFunc {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	allowOriginFunc := buildAllowOriginFunc(cfg.AllowOrigins)
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}
		allowOrigin := allowOriginFunc(origin)
		if allowOrigin == "" {
			c.Next()
			return
		}
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Expose-Headers", exposeHeaders)
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Max-Age", cfg.MaxAge)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func buildAllowOriginFunc(origins []string) func(string) string {
	if len(origins) == 1 && origins[0] == "*" {
		return func(string) string { return "*" }
	}
	set := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		set[strings.TrimSpace(o)] = struct{}{}
	}
	return func(origin string) string {
		if _, ok := set[origin]; ok {
			return origin
		}
		return ""
	}
}

// — Circuit Breaker (minimal) -----------------------------------------------

// CircuitState represents the state of a circuit breaker.
type CircuitState int32

const (
	CircuitClosed   CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)
