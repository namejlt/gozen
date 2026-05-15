// Package log — backward-compatible shim.
//
// This file provides the legacy-style API where the first argument to every
// logging function is a log-name string (e.g. log.NameApi).  The canonical API
// uses log.L().Named(name).Infow(...) instead, but these package-level
// functions remain for callers that haven't migrated yet.
//
// All package-level state (global logger) is now managed by logger.go.
package log

// logNamed is a convenience helper that resolves the log-name to a
// named sub-logger and returns it.
func logNamed(logName string) Logger {
	return global.Named(logName)
}

// — Legacy wrapper: every call delegates through global.Named(name) —

func Debug(logName string, args ...any) {
	logNamed(logName).Debug(args...)
}
func Debugw(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).Debugw(msg, keysAndValues...)
}
func Debugf(logName string, template string, args ...any) {
	logNamed(logName).Debugf(template, args...)
}

func Info(logName string, args ...any) {
	logNamed(logName).Info(args...)
}
func Infow(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).Infow(msg, keysAndValues...)
}
func Infof(logName string, template string, args ...any) {
	logNamed(logName).Infof(template, args...)
}

func Warn(logName string, args ...any) {
	logNamed(logName).Warn(args...)
}
func Warnw(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).Warnw(msg, keysAndValues...)
}
func Warnf(logName string, template string, args ...any) {
	logNamed(logName).Warnf(template, args...)
}

func Error(logName string, args ...any) {
	logNamed(logName).Error(args...)
}
func Errorw(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).Errorw(msg, keysAndValues...)
}
func Errorf(logName string, template string, args ...any) {
	logNamed(logName).Errorf(template, args...)
}

func DPanic(logName string, args ...any) {
	logNamed(logName).DPanic(args...)
}
func DPanicw(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).DPanicw(msg, keysAndValues...)
}
func DPanicf(logName string, template string, args ...any) {
	logNamed(logName).DPanicf(template, args...)
}

func Panic(logName string, args ...any) {
	logNamed(logName).Panic(args...)
}
func Panicw(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).Panicw(msg, keysAndValues...)
}
func Panicf(logName string, template string, args ...any) {
	logNamed(logName).Panicf(template, args...)
}

func Fatal(logName string, args ...any) {
	logNamed(logName).Fatal(args...)
}
func Fatalw(logName string, msg string, keysAndValues ...any) {
	logNamed(logName).Fatalw(msg, keysAndValues...)
}
func Fatalf(logName string, template string, args ...any) {
	logNamed(logName).Fatalf(template, args...)
}
