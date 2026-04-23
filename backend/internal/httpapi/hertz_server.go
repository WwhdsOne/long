package httpapi

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/pprof"
	"github.com/cloudwego/hertz/pkg/common/hlog"

)

// NewHertzServer 使用 Hertz 原生路由承载 API、SSE、静态资源与 pprof。
func NewHertzServer(addr string, options Options) *server.Hertz {
    hlog.SetLevel(hlog.LevelInfo) // 关键一行
	engine := server.Default(server.WithHostPorts(addr))
	pprof.Register(engine)
	registerRoutes(engine.Engine, options)
	return engine
}
