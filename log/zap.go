package log

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger implements Logger backed by uber-go/zap with file rotation.
// It is safe for concurrent use.
type ZapLogger struct {
	zap    *zap.SugaredLogger
	app    string
	names  map[string]string // logName → zap named suffix
}

// nameList maps the framework's log-name constants to the names used
// in zap.Named() calls.
var nameList = map[string]string{
	NameRedis:   NameRedis,
	NameMysql:   NameMysql,
	NameMongodb: NameMongodb,
	NameApi:     NameApi,
	NameAo:      NameAo,
	NameGRpc:    NameGRpc,
	NameEs:      NameEs,
	NameTmq:     NameTmq,
	NameAmq:     NameAmq,
	NameLogic:   NameLogic,
	NameFile:    NameFile,
	NameNet:     NameNet,
}

// NewZapLogger creates a ZapLogger from Config.  It returns the concrete
// implementation so callers that need the *ZapLogger for further customisation
// can access it directly; the standard path is log.SetLogger(zl).
func NewZapLogger(cfg Config) (*ZapLogger, error) {
	core, err := buildCore(cfg)
	if err != nil {
		return nil, err
	}

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("@log_name", cfg.Name)),
	}
	if !cfg.Debug {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	zl := &ZapLogger{
		zap:   zap.New(core, opts...).Sugar(),
		app:   cfg.Name,
		names: nameList,
	}
	return zl, nil
}

func buildCore(cfg Config) (zapcore.Core, error) {
	var (
		encoder zapcore.Encoder
		ws      zapcore.WriteSyncer
		level   zapcore.LevelEnabler
	)

	prodCfg := zap.NewProductionConfig()
	prodCfg.EncoderConfig.TimeKey = "ts"
	prodCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if cfg.Debug {
		devCfg := zap.NewDevelopmentConfig()
		devCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(devCfg.EncoderConfig)
		ws = zapcore.AddSync(os.Stdout)
		level = zap.DebugLevel
	} else {
		encoder = zapcore.NewJSONEncoder(prodCfg.EncoderConfig)
		path := resolveLogPath(cfg)
		ws = zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
		level = prodCfg.Level
	}

	return zapcore.NewCore(encoder, ws, level), nil
}

// resolveLogPath respects POD_LOG_PATH env-var override.
func resolveLogPath(cfg Config) string {
	podLogPath := os.Getenv("POD_LOG_PATH")
	if podLogPath == "" {
		return cfg.Path
	}
	podLogPath = strings.TrimSuffix(podLogPath, "/")
	parts := strings.Split(podLogPath, "/")
	name := cfg.Name + ".log"
	if len(parts) > 0 && parts[len(parts)-1] != "" {
		name = parts[len(parts)-1] + ".log"
	}
	return podLogPath + "/" + name
}

// — Logger interface —

func (l *ZapLogger) resolveName(logName string) string {
	if v, ok := l.names[logName]; ok {
		return v
	}
	return NameDefault
}

func (l *ZapLogger) Debug(args ...any)                     { l.zap.Debug(args...) }
func (l *ZapLogger) Info(args ...any)                      { l.zap.Info(args...) }
func (l *ZapLogger) Warn(args ...any)                      { l.zap.Warn(args...) }
func (l *ZapLogger) Error(args ...any)                     { l.zap.Error(args...) }
func (l *ZapLogger) DPanic(args ...any)                    { l.zap.DPanic(args...) }
func (l *ZapLogger) Panic(args ...any)                     { l.zap.Panic(args...) }
func (l *ZapLogger) Fatal(args ...any)                     { l.zap.Fatal(args...) }

func (l *ZapLogger) Debugw(msg string, kv ...any)          { l.zap.Debugw(msg, l.sanitiseKV(kv)...) }
func (l *ZapLogger) Infow(msg string, kv ...any)           { l.zap.Infow(msg, l.sanitiseKV(kv)...) }
func (l *ZapLogger) Warnw(msg string, kv ...any)           { l.zap.Warnw(msg, l.sanitiseKV(kv)...) }
func (l *ZapLogger) Errorw(msg string, kv ...any)          { l.zap.Errorw(msg, l.sanitiseKV(kv)...) }
func (l *ZapLogger) DPanicw(msg string, kv ...any)         { l.zap.DPanicw(msg, l.sanitiseKV(kv)...) }
func (l *ZapLogger) Panicw(msg string, kv ...any)          { l.zap.Panicw(msg, l.sanitiseKV(kv)...) }
func (l *ZapLogger) Fatalw(msg string, kv ...any)          { l.zap.Fatalw(msg, l.sanitiseKV(kv)...) }

func (l *ZapLogger) Debugf(tmpl string, args ...any)       { l.zap.Debugf(tmpl, args...) }
func (l *ZapLogger) Infof(tmpl string, args ...any)        { l.zap.Infof(tmpl, args...) }
func (l *ZapLogger) Warnf(tmpl string, args ...any)        { l.zap.Warnf(tmpl, args...) }
func (l *ZapLogger) Errorf(tmpl string, args ...any)       { l.zap.Errorf(tmpl, args...) }
func (l *ZapLogger) DPanicf(tmpl string, args ...any)      { l.zap.DPanicf(tmpl, args...) }
func (l *ZapLogger) Panicf(tmpl string, args ...any)       { l.zap.Panicf(tmpl, args...) }
func (l *ZapLogger) Fatalf(tmpl string, args ...any)       { l.zap.Fatalf(tmpl, args...) }

func (l *ZapLogger) Named(name string) Logger {
	named := l.resolveName(name)
	return &ZapLogger{zap: l.zap.Named(named), app: l.app, names: l.names}
}

func (l *ZapLogger) With(kv ...any) Logger {
	return &ZapLogger{zap: l.zap.With(kv...), app: l.app, names: l.names}
}

func (l *ZapLogger) Sync() error { return l.zap.Sync() }

// sanitiseKV converts known log-key values to strings (for consistency).
func (l *ZapLogger) sanitiseKV(args []any) []any {
	if len(args)%2 != 0 {
		return args
	}
	out := make([]any, 0, len(args))
	for i := 0; i < len(args); i += 2 {
		out = append(out, args[i])
		if _, ok := logKNameSet[args[i]]; ok {
			out = append(out, fmt.Sprint(args[i+1]))
		} else {
			out = append(out, args[i+1])
		}
	}
	return out
}

// Keep the old global-passthrough API working.  After the user calls
// NewZapLogger + SetLogger, these will dispatch to the instance.
// The raw `log.go` file (original) can be deprecated in favour of this.

// OldInit is a compat shim for the old Init function signature.
func OldInit(cfg Config) error {
	zl, err := NewZapLogger(cfg)
	if err != nil {
		return err
	}
	SetLogger(zl)
	return nil
}
