package log

// Config holds the bootstrap configuration for ZapLogger.
type Config struct {
	Name       string // app name injected into every log line
	Path       string // log file path (ignored when POD_LOG_PATH env is set)
	Debug      bool   // if true, use human-readable Development encoder instead of JSON
	MaxSize    int    // MB before rotation
	MaxAge     int    // days to keep old logs
	MaxBackups int    // number of rotated files to keep
	Compress   bool   // gzip rotated files
}
