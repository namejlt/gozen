package gozen

import (
	"github.com/SkyAPM/go2sky"
	"time"
)

var (
	GReporter go2sky.Reporter
	GTracer   *go2sky.Tracer
)

const (
	gRPCInvokeTimeout       = 5 * time.Second
	appJsonDebugVarsPrefix  = "DebugVarsPrefix"
	appJsonEnv              = "Env"
	appJsonDocs             = "Docs"
	appJsonDocsInstanceName = "DocsInstanceName"
	appJsonHost             = "Host"
)
