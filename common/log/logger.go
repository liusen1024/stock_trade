package log

import (
	"stock/common/env"

	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

// init 日志初始化
// TODO 区分dev & prod， 日志时间格式改进，定期写文件
func init() {
	var logger *zap.Logger
	var err error
	e := env.GlobalEnv()
	if e != nil && e.IsProd() {
		logger, err = zap.NewProduction(zap.AddCallerSkip(1))
	} else {
		// 默认情况下都用开发状态
		logger, err = zap.NewDevelopment(zap.AddCallerSkip(1))
	}
	if err != nil {
		panic(err)
	}
	sugar = logger.Sugar()
}

// Debug debug log
func Debug(args ...interface{}) {
	sugar.Debug(args...)
}

// Debugf debug log
func Debugf(msg string, args ...interface{}) {
	sugar.Debugf(msg, args...)
}

// Info info log
func Info(args ...interface{}) {
	sugar.Info(args...)
}

// Infof info log
func Infof(msg string, args ...interface{}) {
	sugar.Infof(msg, args...)
}

// Warn warn log
func Warn(args ...interface{}) {
	sugar.Warn(args...)
}

// Warnf warn log
func Warnf(msg string, args ...interface{}) {
	sugar.Warnf(msg, args...)
}

// Error error log
func Error(args ...interface{}) {
	sugar.Error(args...)
}

// Errorf error log
func Errorf(msg string, args ...interface{}) {
	sugar.Errorf(msg, args...)
}

// Panic panic
func Panic(args ...interface{}) {
	sugar.Panic(args...)
}

// Sugar returns sugar log
func Sugar() *zap.SugaredLogger {
	return sugar
}
