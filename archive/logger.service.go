package common

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger
var rollingLogger *lumberjack.Logger = &lumberjack.Logger{
	Filename:   "server.log",
	MaxSize:    100,
	MaxBackups: 3,
	MaxAge:     30,
}

type loggerContextKey struct{}

func SetMaxSize(maxSize int) *lumberjack.Logger {
	rollingLogger.MaxSize = maxSize
	return rollingLogger
}

func SetMaxBackups(maxBackups int) *lumberjack.Logger {
	rollingLogger.MaxBackups = maxBackups
	return rollingLogger
}

func SetMaxAge(maxAge int) *lumberjack.Logger {
	rollingLogger.MaxAge = maxAge
	return rollingLogger
}

func InitializeLogger() {
	if IsDev() {
		initializeDevLogger()
	} else {
		initializeProdLogger()
	}
	if logger != nil {
		logger.Info("logger initialized")
	}
}
func initializeDevLogger() {
	logger = zap.Must(zap.NewDevelopment())
}

func initializeProdLogger() {
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(pe)
	consoleEncoder := zapcore.NewConsoleEncoder(pe)
	level := zap.InfoLevel
	rollingLogsSync := zapcore.AddSync(rollingLogger)
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, rollingLogsSync, level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)
	logger = zap.New(core)
}

func GetLogger() *zap.Logger {
	if logger == nil {
		InitializeLogger()
	}
	return logger
}

func LoggerFromCtx(ctx context.Context) *zap.Logger {
	ctxLogger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	if !ok || ctxLogger == nil {
		return GetLogger()
	}
	return ctxLogger
}

func SetLoggerToCtx(ctx context.Context, logger *zap.Logger) context.Context {
	ctxLogger := LoggerFromCtx(ctx)
	if ctxLogger == logger {
		return ctx
	}
	return context.WithValue(ctx, loggerContextKey{}, logger)
}
