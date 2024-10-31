package gozen

import (
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func MiddlewareHttp() gin.HandlerFunc {
	if TracerDisabled() {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		requestApi := SpanHttpServer + c.Request.Method + c.Request.URL.Path
		spanStart, ctx, err := GTracer.CreateEntrySpan(c.Request.Context(), requestApi, func(headerKey string) (string, error) {
			return c.Request.Header.Get(headerKey), nil //从header获取tracer信息
		})
		defer SpanEnd(spanStart)
		if err != nil {
			LogErrorw(LogNameLogic, "MiddlewareHttp CreateEntrySpan",
				LogKNameCommonErr, err)
			c.Next()
			return
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func MiddlewareGRpcUnaryInterceptorTracer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		if TracerDisabled() {
			resp, err = handler(ctx, req) //业务处理
			return
		}
		requestApi := SpanGRpcServer + info.FullMethod
		spanStart, ctxSub, err := GTracer.CreateEntrySpan(ctx, requestApi, func(headerKey string) (string, error) {
			md, ok := metadata.FromIncomingContext(ctx) //从header获取tracer信息
			var str string
			if ok {
				if v, ok := md[headerKey]; ok {
					if len(v) > 0 {
						str = v[0]
					}
				}
			}
			return str, nil
		})
		if err != nil {
			LogErrorw(LogNameLogic, "MiddlewareGRpc CreateEntrySpan",
				LogKNameCommonErr, err)
		}
		defer SpanEnd(spanStart)
		resp, err = handler(ctxSub, req) //业务处理
		return
	}
}
