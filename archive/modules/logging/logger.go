package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultRollingLogger = &lumberjack.Logger{
	Filename:   "server.log",
	MaxSize:    500,
	MaxBackups: 3,
	MaxAge:     30,
}

func ReplaceDefaultRollingLogger(logger *lumberjack.Logger) {
	defaultRollingLogger = logger
}

func GetDefaultRollingLogger() *lumberjack.Logger {
	return defaultRollingLogger
}

func ReplaceDefaultFileName(filename string) {
	defaultRollingLogger.Filename = filename
}

func NewRollingLogger(filename string, maxSize int, maxBackups int, maxAge int) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}
}

type LoggerConfig struct {
	Syncs   []zapcore.WriteSyncer
	Level   zapcore.Level
	Encoder zapcore.Encoder
}

func NewLoggerConfig(syncs []zapcore.WriteSyncer, level zapcore.Level, encoder zapcore.Encoder) LoggerConfig {
	return LoggerConfig{
		Syncs:   syncs,
		Level:   level,
		Encoder: encoder,
	}
}

func (c LoggerConfig) ReplaceGlobalLogger() {
	cores := make([]zapcore.Core, len(c.Syncs))
	for i, sync := range c.Syncs {
		cores[i] = zapcore.NewCore(c.Encoder, sync, c.Level)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
}
