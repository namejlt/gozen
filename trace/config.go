package trace

// TracerConfig holds SkyWalking-compatible tracer bootstrap parameters.
type TracerConfig struct {
	ServiceName string
	Disabled    bool
	Reporter    struct {
		LocalAgentHostPort string
	}
	Sampler struct {
		SamplingRate float64
	}
}

// DefaultConfig returns a sensible disabled-by-default configuration.
func DefaultConfig() TracerConfig {
	return TracerConfig{Disabled: true}
}

// DefaultConfig returns a copy of the current package-level config.
// (Call SetConfig before InitSkyWalking.)
var (
	pkgConfig = DefaultConfig()
)

// SetConfig replaces the package-level config.
func SetConfig(cfg TracerConfig) { pkgConfig = cfg }

// Config returns the current config.
func Config() TracerConfig { return pkgConfig }

// TracerDisabled returns true if tracing is disabled.
func TracerDisabled() bool { return pkgConfig.Disabled }
