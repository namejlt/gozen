package gozen

import "time"

func UtilLogError(msg string) {
	LogErrorf(LogNameDefault, msg)
}

func UtilLogErrorf(format string, a ...interface{}) {
	LogErrorf(LogNameDefault, format, a...)
}

func UtilLogInfo(msg string) {
	LogInfof(LogNameDefault, msg)
}

func UtilLogInfof(format string, a ...interface{}) {
	LogInfof(LogNameDefault, format, a...)
}

func UtilLogDebug(msg string) {
	LogDebugf(LogNameDefault, msg)
}

func UtilLogDebugf(format string, a ...interface{}) {
	LogDebugf(LogNameDefault, format, a...)
}

type Log struct {
}

func NewUtilLog() *Log {
	return &Log{}
}

func (l *Log) Error(format string, a ...interface{}) {
	UtilLogErrorf(format, a...)
}

func (l *Log) Info(format string, a ...interface{}) {
	UtilLogInfof(format, a...)
}

var (
	subTimeLogSwitch bool
)

func SetLogTimeSwitch(b bool) {
	subTimeLogSwitch = b
}

type LogTime struct {
	name  string
	start time.Time
}

func NewLogTime(name string) (p *LogTime) {
	p = new(LogTime)
	if !subTimeLogSwitch {
		return
	}
	p.name = name
	p.start = time.Now()
	return
}

func (p *LogTime) LogEnd() {
	if !subTimeLogSwitch {
		return
	}
	s := time.Now().Sub(p.start)
	LogInfow(LogNameLogic, "log time "+p.name,
		"log_time", s.String())
}
