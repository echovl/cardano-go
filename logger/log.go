package logger

import (
	"os"

	"go.uber.org/zap"
)

var defaultLogger *zap.SugaredLogger
var homePath = os.Getenv("HOME")

const fallbackLogFilePath = "./.cardano_wallet_log"

func init() {
	config := zap.NewDevelopmentConfig()
	if homePath != "" {
		config.OutputPaths = []string{homePath + "/.cardano_wallet_log"}
	} else {
		config.OutputPaths = []string{fallbackLogFilePath}
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	defaultLogger = logger.Sugar()
}

func Infof(template string, args ...interface{}) {
	defaultLogger.Infof(template, args...)
}

func Errorf(template string, args ...interface{}) {
	defaultLogger.Errorf(template, args...)
}

func Warnf(template string, args ...interface{}) {
	defaultLogger.Warnf(template, args...)
}

func Infow(msg string, keyValues ...interface{}) {
	defaultLogger.Infow(msg, keyValues...)
}

func Errorw(msg string, keyValues ...interface{}) {
	defaultLogger.Errorw(msg, keyValues...)
}

func Warnw(msg string, keyValues ...interface{}) {
	defaultLogger.Warnw(msg, keyValues...)
}
