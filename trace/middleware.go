package trace

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/namejlt/gozen/log"
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
			log.Errorw(log.NameApi, "MiddlewareHttp CreateEntrySpan",
				log.KNameCommonErr, err)
			c.Next()
			return
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func MiddlewareGRpcUnaryInterceptorTracer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
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
			log.Errorw(log.NameApi, "MiddlewareGRpc CreateEntrySpan",
				log.KNameCommonErr, err)
		}
		defer SpanEnd(spanStart)
		resp, err = handler(ctxSub, req) //业务处理
		return
	}
}
