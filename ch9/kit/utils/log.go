package utils

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func NewLoggerServer() {
	logger = NewLogger(
		SetAppName("go-kit"),
		SetDevelopment(true),
		SetLevel(zap.DebugLevel),
	)
}

func GetLogger() *zap.Logger {
	return logger
}
