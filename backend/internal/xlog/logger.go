package xlog

import (
	"context"
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

// SystemLogEntry 表示一条系统日志记录。
type SystemLogEntry struct {
	Level     string
	Module    string
	Message   string
	RequestID string
	CreatedAt int64
}

var globalLogger = zap.NewNop()
var systemLogHook func(zapcore.Entry) error

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
	if systemLogHook != nil {
		logger = logger.WithOptions(zap.Hooks(systemLogHook))
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

// SetSystemLogHook 设置 warn/error 级别日志的附加落库钩子。
func SetSystemLogHook(hook func(zapcore.Entry) error) {
	systemLogHook = hook
}

// HookFromWriter 将系统日志写入器包装为 zap hook。
func HookFromWriter(writer interface {
	WriteSystemLog(context.Context, SystemLogEntry) error
}) func(zapcore.Entry) error {
	return func(entry zapcore.Entry) error {
		if writer == nil {
			return nil
		}
		if entry.Level < zapcore.WarnLevel {
			return nil
		}
		return writer.WriteSystemLog(context.Background(), SystemLogEntry{
			Level:     entry.Level.String(),
			Module:    entry.LoggerName,
			Message:   entry.Message,
			CreatedAt: entry.Time.Unix(),
		})
	}
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
