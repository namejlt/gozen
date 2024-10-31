package gozen

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

/**

日志 zap封装 本地文件记录

*/

const (
	LogNameDefault = "default"
	LogNameRedis   = "redis"
	LogNameMysql   = "mysql"
	LogNameMongodb = "mongodb"
	LogNameApi     = "api"
	LogNameAo      = "ao"
	LogNameGRpc    = "grpc"
	LogNameEs      = "es"
	LogNameTmq     = "tmq"
	LogNameAmq     = "amq"
	LogNameLogic   = "logic"
	LogNameFile    = "file"
	LogNameNet     = "net"
)

const (
	LogKNameCommonErr       = "log-common-err"
	LogKNameCommonFields    = "log-common-fields"
	LogKNameCommonCondition = "log-common-condition"
	LogKNameCommonAddress   = "log-common-address"
	LogKNameCommonName      = "log-common-name"
	LogKNameCommonCmd       = "log-common-cmd"
	LogKNameCommonData      = "log-common-data"
	LogKNameCommonDataType  = "log-common-data-type"
	LogKNameCommonKey       = "log-common-key"
	LogKNameCommonValue     = "log-common-value"
	LogKNameCommonUrl       = "log-common-url"
	LogKNameCommonNum       = "log-common-num"
	LogKNameCommonId        = "log-common-id"
	LogKNameCommonUid       = "log-common-uid"
	LogKNameCommonCode      = "log-common-code"
	LogKNameCommonLevel     = "log-common-level"
	LogKNameCommonCookie    = "log-common-cookie"
	LogKNameCommonReq       = "log-common-req"
	LogKNameCommonRes       = "log-common-res"
	LogKNameCommonTime      = "log-common-time"
	LogKNameCommonTenantId  = "log-common-tenant-id"
	LogKNameCommonRecordId  = "log-common-record-id"
	LogKNameCommonUniqueId  = "log-common-unique-id"

	LogKNameRedisKey  = "log-redis-key"
	LogKNameRedisData = "log-redis-data"

	LogKNameMysqlParam = "log-mysql-param"
	LogKNameMysqlData  = "log-mysql-data"

	LogKNameMongodbParam = "log-mongodb-param"
	LogKNameMongodbData  = "log-mongodb-data"

	LogKNameApiParam = "log-api-param"
	LogKNameApiUrl   = "log-api-url"
	LogKNameApiRes   = "log-api-res"

	LogKNameAoParam = "log-ao-param"
	LogKNameAoReq   = "log-ao-req"
	LogKNameAoRes   = "log-ao-res"

	LogKNameGRpcReq = "log-grpc-req"
	LogKNameGRpcRes = "log-grpc-res"

	LogKNameEsReq = "log-es-req"
	LogKNameEsRes = "log-es-res"

	LogKNameTmqTopic = "log-tmq-topic"
	LogKNameTmqReq   = "log-tmq-req"
	LogKNameTmqRes   = "log-tmq-res"
)

var (
	logKNameMap = map[interface{}]struct{}{
		LogKNameCommonErr:       {},
		LogKNameCommonFields:    {},
		LogKNameCommonCondition: {},
		LogKNameCommonAddress:   {},
		LogKNameCommonName:      {},
		LogKNameCommonData:      {},
		LogKNameCommonDataType:  {},
		LogKNameCommonKey:       {},
		LogKNameCommonValue:     {},
		LogKNameCommonUrl:       {},
		LogKNameCommonNum:       {},
		LogKNameCommonId:        {},
		LogKNameCommonUid:       {},
		LogKNameCommonCode:      {},
		LogKNameCommonLevel:     {},
		LogKNameCommonCookie:    {},
		LogKNameCommonReq:       {},
		LogKNameCommonRes:       {},
		LogKNameCommonTime:      {},
		LogKNameCommonTenantId:  {},
		LogKNameCommonRecordId:  {},
		LogKNameCommonUniqueId:  {},
		LogKNameRedisKey:        {},
		LogKNameRedisData:       {},
		LogKNameMysqlParam:      {},
		LogKNameMysqlData:       {},
		LogKNameMongodbParam:    {},
		LogKNameMongodbData:     {},
		LogKNameApiParam:        {},
		LogKNameApiUrl:          {},
		LogKNameApiRes:          {},
		LogKNameAoParam:         {},
		LogKNameAoReq:           {},
		LogKNameAoRes:           {},
		LogKNameGRpcReq:         {},
		LogKNameGRpcRes:         {},
		LogKNameEsReq:           {},
		LogKNameEsRes:           {},
		LogKNameTmqTopic:        {},
		LogKNameTmqReq:          {},
		LogKNameTmqRes:          {},
	}
)

var (
	logger      *zap.Logger
	logSugar    *zap.SugaredLogger
	logAppName  = "@log_name"
	logNameList = map[string]string{ //日志分类
		LogNameRedis:   LogNameRedis,
		LogNameMysql:   LogNameMysql,
		LogNameMongodb: LogNameMongodb,
		LogNameApi:     LogNameApi,
		LogNameAo:      LogNameAo,
		LogNameGRpc:    LogNameGRpc,
		LogNameEs:      LogNameEs,
		LogNameTmq:     LogNameTmq,
		LogNameAmq:     LogNameAmq,
		LogNameLogic:   LogNameLogic,
		LogNameFile:    LogNameFile,
		LogNameNet:     LogNameNet,
	}
)

// 初始化logger
func loggerInit() {
	var (
		err error
	)
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1), //忽略框架层代码跟踪
		zap.Fields(zap.String(logAppName, configProject.Log.Name)),
	}
	if configProject.Log.Debug {
		logger, err = zap.NewDevelopment(opts...)
	} else {
		cfg := zap.NewProductionConfig()

		// 日志路径 优先环境变量 其次conf
		var logPath string
		podLogPath := os.Getenv("POD_LOG_PATH")
		if podLogPath != "" {
			podLogPath, _ = strings.CutSuffix(podLogPath, "/")

			confLogPathArr := strings.Split(podLogPath, "/")
			var fileName string
			// fileName 为 podLogPath 的最后一个路径
			if len(confLogPathArr) > 0 {
				fileName = confLogPathArr[len(confLogPathArr)-1] + ".log"
			} else {
				fileName = configProject.Log.Name + ".log" //默认值
			}

			logPath = podLogPath + "/" + fileName
		} else {
			logPath = configProject.Log.Path
		}

		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    configProject.Log.MaxSize, // megabytes
			MaxBackups: configProject.Log.MaxBackups,
			MaxAge:     configProject.Log.MaxAge, // days
			Compress:   configProject.Log.Compress,
		})
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
		logger = zap.New(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(cfg.EncoderConfig),
				w,
				cfg.Level,
			),
			opts...,
		)
	}
	if err != nil {
		panic("logger init error:" + err.Error())
	}
	logSugar = logger.Sugar()
}

func getLogName(logName string) string {
	if v, ok := logNameList[logName]; ok {
		return v
	} else {
		return LogNameDefault
	}
}

func LogSync() {
	_ = logSugar.Sync()
}

// LogDebug debug debug模式下会记录
func LogDebug(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Debug(args)
}

func LogDebugw(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).Debugw(msg, logParamToStr(keysAndValues)...)
}

func LogDebugf(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Debugf(template, args...)
}

// LogInfo info 记录信息
func LogInfo(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Info(args)
}

func LogInfow(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).Infow(msg, logParamToStr(keysAndValues)...)
}

func LogInfof(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Infof(template, args...)
}

// LogWarn warn 有异常跟踪
func LogWarn(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Warn(args)
}

func LogWarnw(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).Warnw(msg, logParamToStr(keysAndValues)...)
}

func LogWarnf(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Warnf(template, args...)
}

// LogError error 有异常跟踪
func LogError(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Error(args)
}

func LogErrorw(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).Errorw(msg, logParamToStr(keysAndValues)...)
}

func LogErrorf(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Errorf(template, args...)
}

// LogDPanic dpanic debug模式下 当前协程panic
func LogDPanic(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).DPanic(args)
}

func LogDPanicw(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).DPanicw(msg, logParamToStr(keysAndValues)...)
}

func LogDPanicf(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).DPanicf(template, args...)
}

// LogPanic panic 当前协程panic
func LogPanic(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Panic(args)
}

func LogPanicw(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).Panicw(msg, logParamToStr(keysAndValues)...)
}

func LogPanicf(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Panicf(template, args...)
}

// LogFatal fatal 服务会终止
func LogFatal(logName string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Fatal(args)
}

func LogFatalw(logName string, msg string, keysAndValues ...interface{}) {
	logSugar.Named(getLogName(logName)).Fatalw(msg, logParamToStr(keysAndValues)...)
}

func LogFatalf(logName string, template string, args ...interface{}) {
	logSugar.Named(getLogName(logName)).Fatalf(template, args...)
}

// 格式化日志value
func logParamToStr(args []interface{}) (strArr []interface{}) {
	argsLen := len(args)
	if argsLen%2 != 0 {
		return args
	}
	for i := 0; i < argsLen; i = i + 2 {
		strArr = append(strArr, args[i])
		if _, ok := logKNameMap[args[i]]; ok {
			strArr = append(strArr, fmt.Sprint(args[i+1]))
		} else {
			strArr = append(strArr, args[i+1])
		}
	}
	return
}
