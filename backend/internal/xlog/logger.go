package xlog

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
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
	Nickname   string
	Body       string
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
	logger = logger.WithOptions(CoreWithUUID())

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

// uuidCore 包装 zapcore.Core，为每条日志添加唯一 UUID。
// CoreWithUUID 返回一个 zap.Option，为每条日志添加唯一 log_id 字段。
func CoreWithUUID() zap.Option {
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return &uuidCore{Core: c}
	})
}

type uuidCore struct {
	zapcore.Core
}

func (c *uuidCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	fields = append(fields, zap.String("log_id", uuid.NewString()))
	return c.Core.Write(entry, fields)
}

// HertzLogger 实现 hlog.FullLogger，代理到全局 zap.Logger，统一日志实例。
type HertzLogger struct {
	sugar *zap.SugaredLogger
}

// NewHertzLogger 返回一个代理到全局 zap.Logger 的 hlog.FullLogger。
func NewHertzLogger() *HertzLogger {
	return &HertzLogger{sugar: globalLogger.Sugar()}
}

// refreshSugar 确保 sugar 指向当前全局 logger（SetLevel/SetOutput 后 logger 可能被替换）。
func (h *HertzLogger) refreshSugar() {
	h.sugar = globalLogger.Sugar()
}

// — Logger —

func (h *HertzLogger) Trace(v ...interface{})            { h.sugar.Debug(v...) }
func (h *HertzLogger) Debug(v ...interface{})            { h.sugar.Debug(v...) }
func (h *HertzLogger) Info(v ...interface{})             { h.sugar.Info(v...) }
func (h *HertzLogger) Notice(v ...interface{})           { h.sugar.Warn(v...) }
func (h *HertzLogger) Warn(v ...interface{})             { h.sugar.Warn(v...) }
func (h *HertzLogger) Error(v ...interface{})            { h.sugar.Error(v...) }
func (h *HertzLogger) Fatal(v ...interface{})            { h.sugar.Fatal(v...) }

// — FormatLogger —

func (h *HertzLogger) Tracef(f string, v ...interface{})  { h.sugar.Debugf(f, v...) }
func (h *HertzLogger) Debugf(f string, v ...interface{})  { h.sugar.Debugf(f, v...) }
func (h *HertzLogger) Infof(f string, v ...interface{})   { h.sugar.Infof(f, v...) }
func (h *HertzLogger) Noticef(f string, v ...interface{}) { h.sugar.Warnf(f, v...) }
func (h *HertzLogger) Warnf(f string, v ...interface{})   { h.sugar.Warnf(f, v...) }
func (h *HertzLogger) Errorf(f string, v ...interface{})  { h.sugar.Errorf(f, v...) }
func (h *HertzLogger) Fatalf(f string, v ...interface{})  { h.sugar.Fatalf(f, v...) }

// — CtxLogger —

func (h *HertzLogger) CtxTracef(_ context.Context, f string, v ...interface{})  { h.sugar.Debugf(f, v...) }
func (h *HertzLogger) CtxDebugf(_ context.Context, f string, v ...interface{})  { h.sugar.Debugf(f, v...) }
func (h *HertzLogger) CtxInfof(_ context.Context, f string, v ...interface{})   { h.sugar.Infof(f, v...) }
func (h *HertzLogger) CtxNoticef(_ context.Context, f string, v ...interface{}) { h.sugar.Warnf(f, v...) }
func (h *HertzLogger) CtxWarnf(_ context.Context, f string, v ...interface{})   { h.sugar.Warnf(f, v...) }
func (h *HertzLogger) CtxErrorf(_ context.Context, f string, v ...interface{})  { h.sugar.Errorf(f, v...) }
func (h *HertzLogger) CtxFatalf(_ context.Context, f string, v ...interface{})  { h.sugar.Fatalf(f, v...) }

// — Control —

func (h *HertzLogger) SetLevel(level hlog.Level) {
	var lvl zapcore.Level
	switch level {
	case hlog.LevelTrace, hlog.LevelDebug:
		lvl = zapcore.DebugLevel
	case hlog.LevelInfo:
		lvl = zapcore.InfoLevel
	case hlog.LevelNotice, hlog.LevelWarn:
		lvl = zapcore.WarnLevel
	case hlog.LevelError:
		lvl = zapcore.ErrorLevel
	case hlog.LevelFatal:
		lvl = zapcore.FatalLevel
	default:
		lvl = zapcore.InfoLevel
	}
	_ = globalLogger.Core().Enabled(lvl)
	// 全局 logger 的级别在 Init 时已设定，此处记录但不实际变更以保持统一。
}

func (h *HertzLogger) SetOutput(w io.Writer) {
	// 输出目标由 Config.Format 决定（stdout），不在此处变更。
}
