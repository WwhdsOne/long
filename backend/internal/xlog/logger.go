package xlog

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 控制全局日志器初始化。
type Config struct {
	Level         string
	Format        string
	IncludeCaller bool
}

var globalLogger = zap.NewNop()

// Init 初始化并替换全局日志器。
func Init(cfg Config) (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	loggerConfig.DisableCaller = !cfg.IncludeCaller

	if strings.EqualFold(strings.TrimSpace(cfg.Format), "console") {
		loggerConfig.Encoding = "console"
	} else {
		loggerConfig.Encoding = "json"
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	globalLogger = logger
	zap.ReplaceGlobals(logger)
	return logger, nil
}

// L 返回当前全局日志器。
func L() *zap.Logger {
	return globalLogger
}

// Sync 刷新全局日志器缓冲。
func Sync() error {
	return globalLogger.Sync()
}

// Err 是错误字段快捷方法。
func Err(err error) zap.Field {
	return zap.Error(err)
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
