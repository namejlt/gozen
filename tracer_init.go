package gozen

import (
	"github.com/SkyAPM/go2sky/propagation"
	"runtime"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	//链路日志节点

	//server 节点
	SpanGRpcServer   = "GRpcServer"
	SpanHttpServer   = "HttpServer"
	SpanScriptServer = "ScriptServer"

	//service 节点
	SpanServiceOpen = "ServiceOpen" //公开逻辑
	SpanServiceBase = "ServiceBase" //基础逻辑

	//middleware 节点
	SpanMiddlewareLogic = "MiddlewareLogic"

	//dao 节点
	SpanDaoMysql   = "DaoMySQL"
	SpanDaoRedis   = "DaoRedis"
	SpanDaoMongoDb = "DaoMongoDB"
	SpanDaoApi     = "DaoAPI"
	SpanDaoAo      = "DaoAO"
	SpanDaoEs      = "DaoES"
	SpanDaoGRpc    = "DaoGRpc"
	SpanDaoMQ      = "DaoMQ"
)

// 组件ID Golang程序使用范围是[5000, 6000)
const (
	ComponentIDX int32 = 5000 + iota
)

const (
	TracerSw8  = propagation.Header
	TracerSw8C = propagation.HeaderCorrelation
)

// span layer
const (
	TracerSpanLayerUnknown      = v3.SpanLayer_Unknown
	TracerSpanLayerDatabase     = v3.SpanLayer_Database
	TracerSpanLayerRPCFramework = v3.SpanLayer_RPCFramework
	TracerSpanLayerHttp         = v3.SpanLayer_Http
	TracerSpanLayerMQ           = v3.SpanLayer_MQ
	TracerSpanLayerCache        = v3.SpanLayer_Cache
)

func InjectorNull(headerKey, headerValue string) error {
	return nil
}

// skip 3 父级函数
func RunFuncNameUp() string {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

// skip 2 当前函数
func RunFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func RunFuncNameSkip(skip int) string {
	pc := make([]uintptr, 1)
	runtime.Callers(skip, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}
