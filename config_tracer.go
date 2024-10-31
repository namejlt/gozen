package gozen

import (
	"sync"
)

var (
	tracerConfigMux sync.Mutex
	TracerConfig    *ConfigTracer
)

type ConfigTracer struct {
	Tracer ConfigTracerData `yaml:"Tracer"`
}

type ConfigTracerData struct {
	ServiceName string                   `yaml:"ServiceName"`
	Disabled    bool                     `yaml:"Disabled"`
	Sampler     ConfigTracerSamplerData  `yaml:"Sampler"`
	Reporter    ConfigTracerReporterData `yaml:"Reporter"`
}

type ConfigTracerSamplerData struct {
	SamplingRate float64 `yaml:"SamplingRate"`
}

type ConfigTracerReporterData struct {
	LogSpans            bool   `yaml:"LogSpans"`
	BufferFlushInterval int    `yaml:"BufferFlushInterval"`
	LocalAgentHostPort  string `yaml:"LocalAgentHostPort"`
}

func configTracerInit() {
	if TracerConfig == nil {
		configFileName := "tracer"
		tracerConfigMux.Lock()
		defer tracerConfigMux.Unlock()
		defaultConfig := configTracerGetDefault()
		if cfp.configPathExist(configFileName) {
			TracerConfig = &ConfigTracer{}
			err := cfp.configGet("tracer", nil, TracerConfig, defaultConfig)
			if err != nil {
				panic("configTracerInit error:" + err.Error())
			}
		} else {
			TracerConfig = defaultConfig
		}
	}
}

func configTracerClear() {
	appConfigMux.Lock()
	defer appConfigMux.Unlock()
	appConfig = nil
}

func configTracerGetDefault() *ConfigTracer {
	c := new(ConfigTracer)
	c.Tracer = ConfigTracerData{}
	c.Tracer.Disabled = true //默认关闭
	c.Tracer.ServiceName = "app_test"
	return c
}

func TracerDisabled() bool {
	return TracerConfig.Tracer.Disabled
}
