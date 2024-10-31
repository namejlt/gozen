package gozen

import (
	"context"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"google.golang.org/grpc/metadata"
	"os"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"time"
)

/*
*

span component
Golang程序使用范围是[5000, 6000)
*/
func InitSkyWalking() {
	if TracerDisabled() {
		return
	}
	var err error
	//skywalking oap server
	GReporter, err = reporter.NewGRPCReporter(TracerConfig.Tracer.Reporter.LocalAgentHostPort)
	if err != nil {
		LogErrorw(LogNameNet, "InitSkyWalking NewGRPCReporter",
			LogKNameCommonErr, err)
		return
	}
	//init tracer
	var opts []go2sky.TracerOption
	opts = append(opts, go2sky.WithReporter(GReporter))
	//服务副本
	serverHost, _ := os.Hostname()
	opts = append(opts, go2sky.WithInstance(serverHost))
	//抽样
	opts = append(opts, go2sky.WithSampler(TracerConfig.Tracer.Sampler.SamplingRate))
	GTracer, err = go2sky.NewTracer(TracerConfig.Tracer.ServiceName, opts...)
	if err != nil {
		LogErrorw(LogNameNet, "InitSkyWalking NewTracer",
			LogKNameCommonErr, err)
		return
	}
}

func CloseSkyWalking() {
	if GReporter != nil {
		GReporter.Close()
	}
}

func EntrySpan(peer string, operationName string, layer v3.SpanLayer) (span go2sky.Span, subCtx context.Context, err error) {
	ctx := context.Background()
	if GTracer == nil {
		return span, ctx, err
	}
	span, subCtx, err = GTracer.CreateEntrySpan(ctx, peer+operationName, func(headerKey string) (string, error) {
		return "", nil
	})
	if err != nil {
		LogErrorw(LogNameLogic, "Tracer CreateEntrySpan",
			LogKNameCommonErr, err,
			LogKNameCommonData, peer+operationName,
		)
		return span, ctx, err
	}
	span.SetSpanLayer(layer)
	span.SetPeer(peer)
	return
}

func LocalSpan(ctx context.Context, peer string, operationName string, layer v3.SpanLayer) (span go2sky.Span, subCtx context.Context, err error) {
	if GTracer == nil || ctx == nil {
		return span, ctx, err
	}
	span, subCtx, err = GTracer.CreateLocalSpan(ctx)
	if err != nil {
		LogErrorw(LogNameLogic, "Tracer CreateLocalSpan",
			LogKNameCommonErr, err,
			LogKNameCommonData, peer+operationName,
		)
		return span, ctx, err
	}
	span.SetSpanLayer(layer)
	span.SetPeer(peer)
	span.SetOperationName(peer + "/" + operationName)
	return
}

func ExitSpan(ctx context.Context, peer string, operationName string, layer v3.SpanLayer) (span go2sky.Span, err error) {
	if GTracer == nil || ctx == nil {
		return span, err
	}
	span, err = GTracer.CreateExitSpan(ctx, peer+"/"+operationName, peer, InjectorNull)
	if err != nil {
		LogErrorw(LogNameLogic, "Tracer CreateExitSpan",
			LogKNameCommonErr, err,
			LogKNameCommonData, peer+operationName,
		)
		return span, err
	}
	span.SetSpanLayer(layer)
	return
}

// 放数据到header内
func ExitSpanGRpc(ctx context.Context, peer string, operationName string, layer v3.SpanLayer) (span go2sky.Span, ctxValue context.Context, err error) {
	if GTracer == nil || ctx == nil {
		return span, ctxValue, err
	}
	ctxValue = context.Background()
	span, err = GTracer.CreateExitSpan(ctx, peer+"/"+operationName, peer, func(key, value string) error {
		ctxValue = metadata.AppendToOutgoingContext(ctxValue, key, value)
		return nil
	})
	if err != nil {
		LogErrorw(LogNameLogic, "Tracer ExitSpanGRpc",
			LogKNameCommonErr, err,
			LogKNameCommonData, peer+operationName,
		)
		return span, ctxValue, err
	}
	span.SetSpanLayer(layer)
	return
}

func SpanEnd(span go2sky.Span) {
	if span != nil {
		span.End()
	}
}

func SpanTag(span go2sky.Span, key go2sky.Tag, value string) {
	if span != nil {
		span.Tag(key, value)
	}
}

func SpanComponent(span go2sky.Span, component int32) {
	if span != nil {
		span.SetComponent(component)
	}
}

func SpanLog(span go2sky.Span, data ...string) {
	if span != nil {
		span.Log(time.Now(), data...)
	}
}

func SpanError(span go2sky.Span, data ...string) {
	if span != nil {
		span.Error(time.Now(), data...)
	}
}

func SpanErrorFast(span go2sky.Span, err error) {
	if span != nil {
		if err != nil {
			span.Error(time.Now(), "err", err.Error())
		}
	}
}
