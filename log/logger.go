// Package log provides a structured-logging abstraction for the gozen framework.
//
// The Logger interface is the primary API; the default implementation is based on
// uber-go/zap with lumberjack rotation (ZapLogger).  Users may supply their own
// implementation via log.SetLogger.
//
// Standard field keys are exported as constants (KName*), shared by all transports
// and DAOs so that log output remains consistent across the entire application.
package log

// Logger is the structured-logging interface used throughout the framework.
// All methods are safe for concurrent use.
type Logger interface {
	Debug(args ...any)
	Debugw(msg string, keysAndValues ...any)
	Debugf(template string, args ...any)

	Info(args ...any)
	Infow(msg string, keysAndValues ...any)
	Infof(template string, args ...any)

	Warn(args ...any)
	Warnw(msg string, keysAndValues ...any)
	Warnf(template string, args ...any)

	Error(args ...any)
	Errorw(msg string, keysAndValues ...any)
	Errorf(template string, args ...any)

	DPanic(args ...any)
	DPanicw(msg string, keysAndValues ...any)
	DPanicf(template string, args ...any)

	Panic(args ...any)
	Panicw(msg string, keysAndValues ...any)
	Panicf(template string, args ...any)

	Fatal(args ...any)
	Fatalw(msg string, keysAndValues ...any)
	Fatalf(template string, args ...any)

	Named(name string) Logger
	With(keysAndValues ...any) Logger
	Sync() error
}

// global is the package-wide logger instance.  It defaults to a no-op logger
// and is replaced by SetLogger (typically during framework initialisation).
var global Logger = &noopLogger{}

// SetLogger replaces the global logger.  It is NOT safe for concurrent use
// with other log functions — call it once during startup.
func SetLogger(l Logger) { global = l }

// L returns the current global logger.
func L() Logger { return global }

// Sync flushes the global logger.
func Sync() error { return global.Sync() }

// noopLogger is the default implementation used before a real logger is set.
// It silently discards all log messages.
type noopLogger struct{}

func (n *noopLogger) Debug(args ...any)                    {}
func (n *noopLogger) Debugw(msg string, kv ...any)         {}
func (n *noopLogger) Debugf(template string, args ...any)  {}
func (n *noopLogger) Info(args ...any)                     {}
func (n *noopLogger) Infow(msg string, kv ...any)          {}
func (n *noopLogger) Infof(template string, args ...any)   {}
func (n *noopLogger) Warn(args ...any)                     {}
func (n *noopLogger) Warnw(msg string, kv ...any)          {}
func (n *noopLogger) Warnf(template string, args ...any)   {}
func (n *noopLogger) Error(args ...any)                    {}
func (n *noopLogger) Errorw(msg string, kv ...any)         {}
func (n *noopLogger) Errorf(template string, args ...any)  {}
func (n *noopLogger) DPanic(args ...any)                   {}
func (n *noopLogger) DPanicw(msg string, kv ...any)        {}
func (n *noopLogger) DPanicf(template string, args ...any) {}
func (n *noopLogger) Panic(args ...any)                    {}
func (n *noopLogger) Panicw(msg string, kv ...any)         {}
func (n *noopLogger) Panicf(template string, args ...any)  {}
func (n *noopLogger) Fatal(args ...any)                    {}
func (n *noopLogger) Fatalw(msg string, kv ...any)         {}
func (n *noopLogger) Fatalf(template string, args ...any)  {}
func (n *noopLogger) Named(name string) Logger             { return n }
func (n *noopLogger) With(kv ...any) Logger                { return n }
func (n *noopLogger) Sync() error                          { return nil }
