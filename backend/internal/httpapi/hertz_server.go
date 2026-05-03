package httpapi

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/gzip"
	"github.com/hertz-contrib/pprof"

	"long/internal/xlog"
)

// NewHertzServer 使用 Hertz 原生路由承载 API、SSE、静态资源与 pprof。
// 日志统一代理到 xlog 全局实例，格式/级别/ UUID 与业务日志一致。
func NewHertzServer(addr string, options Options) *server.Hertz {
	hlog.SetLogger(xlog.NewHertzLogger())

	engine := server.Default(server.WithHostPorts(addr))
	engine.Use(AccessLogMiddleware(options.AccessLogQueue, options.PlayerAuthenticator))
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	pprof.Register(engine)
	registerRoutes(engine.Engine, options)
	return engine
}
