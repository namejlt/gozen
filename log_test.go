package gozen

import (
	"go.uber.org/zap"
	"testing"
)

func Test_logError(t *testing.T) {
	LogError(LogNameDefault, 1, "2", map[string]interface{}{"aa": "aa"}, []uint8{1, 2, 3})
	LogErrorw(LogNameDefault, "this is content", zap.String("name", "jialongtian"), zap.Int("age", 10))
	LogErrorf(LogNameDefault, "my name is %s, age is %d", "jialongtian", 10)
}

func Test_logInfo(t *testing.T) {
	LogInfo(LogNameDefault, 1, "2", map[string]interface{}{"aa": "aa"}, []uint8{1, 2, 3})
	LogInfow(LogNameDefault, "this is content", zap.String("name", "jialongtian"), zap.Int("age", 10))
	LogInfof(LogNameDefault, "my name is %s, age is %d", "jialongtian", 10)
}

func Test_logWarn(t *testing.T) {
	LogWarn(LogNameDefault, 1, "2", map[string]interface{}{"aa": "aa"}, []uint8{1, 2, 3})
	LogWarnw(LogNameDefault, "this is content", zap.String("name", "jialongtian"), zap.Int("age", 10))
	LogWarnf(LogNameDefault, "my name is %s, age is %d", "jialongtian", 10)
}

func Test_logDebug(t *testing.T) {
	LogDebug(LogNameDefault, 1, "2", map[string]interface{}{"aa": "aa"}, []uint8{1, 2, 3})
	LogDebugw(LogNameDefault, "this is content", zap.String("name", "jialongtian"), zap.Int("age", 10))
	LogDebugf(LogNameDefault, "my name is %s, age is %d", "jialongtian", 10)
}
