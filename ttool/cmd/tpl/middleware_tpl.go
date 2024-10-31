package tpl

var (
	MiddlewareDirName   = "middleware"
	MiddlewareFilesName = []string{
		"monitor_grpc.go",
		"monitor_http.go",
		"signature.go",
	}
	MiddlewareFilesContent = []string{
		MiddlewareMonitorGRpcGo,
		MiddlewareMonitorHttpGo,
		MiddlewareSignatureGo,
	}
)

var (
	MiddlewareMonitorGRpcGo = `package middleware

import (
	"github.com/namejlt/gozen"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

func Register(options *[]grpc.ServerOption) {
	*options = append(*options, grpc.ChainUnaryInterceptor(
		grpcMonitor(),
		gozen.MiddlewareGRpcUnaryInterceptorTracer(),
	))
}

func grpcMonitor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		timeStart := time.Now()
		resp, err = handler(ctx, req)
		monitorGrpcReport(req, info, resp, err, timeStart)
		return
	}
}

func monitorGrpcReport(req interface{}, info *grpc.UnaryServerInfo, resp interface{}, err error, timeStart time.Time) {
	dur := time.Now().Sub(timeStart)
	durMill := dur.Nanoseconds() / 1000000
	// grpc 上报 info.FullMethod, "", err == nil, durMill
	logConfig := gozen.ConfigAppGet("KMonitorLogTimeMill")
	lc, ok := logConfig.(int64)
	if !ok {
		lc = 1000
	}
	if durMill >= lc {
		gozen.UtilLogErrorf("method:%s,req:%v,time:%d", info.FullMethod, req, durMill)
	}
}

`
	MiddlewareMonitorHttpGo = `package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Monitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Now().Sub(start)
		result, exist := c.Get("result")
		if !exist || c.Writer.Status() == 404 {
			return
		}
		status := result.(bool)
		path := c.Request.URL.Path
		go monitorReport(path, status, duration)
	}
}

func monitorReport(path string, status bool, duration time.Duration) {
	// http 上报
}

`
	MiddlewareSignatureGo = `package middleware

import (
	"regexp"

	"github.com/namejlt/gozen"
	"github.com/gin-gonic/gin"
)

//验证签名
func Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		/*
		   1、获取请求url，是否符合app uri
		   2、获取签名，解密，匹配token，取得timestamp
		   3、本地加密对应匹配
		   4、根据时间戳以及加密匹配情况判断是否继续
		*/
		path := c.Request.URL.Path
		reg := regexp.MustCompile("app/(.*)")
		if reg.MatchString(path) {
			if !gozen.UtilSignCheckSign(c, "") {
				//todo 转发到错误返回
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
`
)
