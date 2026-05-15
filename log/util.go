package log

import "time"

func UtilLogError(msg string) {
	Errorf(NameDefault, msg)
}

func UtilLogErrorf(format string, a ...any) {
	Errorf(NameDefault, format, a...)
}

func UtilLogInfo(msg string) {
	Infof(NameDefault, msg)
}

func UtilLogInfof(format string, a ...any) {
	Infof(NameDefault, format, a...)
}

func UtilLogDebug(msg string) {
	Debugf(NameDefault, msg)
}

func UtilLogDebugf(format string, a ...any) {
	Debugf(NameDefault, format, a...)
}

type Log struct {
}

func NewUtilLog() *Log {
	return &Log{}
}

func (l *Log) Error(format string, a ...any) {
	UtilLogErrorf(format, a...)
}

func (l *Log) Info(format string, a ...any) {
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
	Infow(NameLogic, "log time "+p.name,
		"log_time", s.String())
}
