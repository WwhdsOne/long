package xlog

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

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

// AccessLogEntry 表示一条 HTTP 请求日志。
type AccessLogEntry struct {
	Method     string
	Path       string
	StatusCode int
	LatencyMs  int64
	ClientIP   string
	UserAgent  string
	CreatedAt  int64
}

var globalLogger = zap.NewNop()
var systemLogHook func(zapcore.Entry) error

// Init 初始化并替换全局日志器。
func Init(cfg Config) (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
	loggerConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
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

	// 包装 uuidCore，为每条日志添加唯一 UUID
	logger = logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return &uuidCore{Core: c}
	}))

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

// generateUUID 生成一个简单的 UUID v4。
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// uuidCore 包装 zapcore.Core，为每条日志添加唯一 UUID。
type uuidCore struct {
	zapcore.Core
}

func (c *uuidCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// 在日志写入前添加 UUID 字段
	fields = append(fields, zap.String("log_id", generateUUID()))
	return c.Core.Write(entry, fields)
}
