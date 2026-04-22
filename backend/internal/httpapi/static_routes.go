package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
)

func registerRealtimeRoutes(router route.IRouter, options Options) {
	if options.Events != nil {
		router.GET("/api/events", options.Events)
	}
	if options.RealtimeHub != nil {
		router.GET("/api/ws", newRealtimeSocketHandler(options))
	}
}

func registerStaticRoutes(engine *route.Engine, options Options) {
	if options.PublicDir == "" {
		return
	}

	indexFile := filepath.Join(options.PublicDir, "index.html")

	// 这些前缀请求里已经自带了目录名，所以 root 用 public 即可
	engine.Static("/assets", options.PublicDir)
	engine.Static("/images", options.PublicDir)

	// 单文件静态资源
	for _, name := range []string{"favicon.ico", "favicon.svg", "icons.svg"} {
		target := filepath.Join(options.PublicDir, name)
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			engine.StaticFile("/"+name, target)
		}
	}

	// SPA fallback
	engine.NoRoute(func(_ context.Context, c *app.RequestContext) {
		path := string(c.Path())

		if strings.HasPrefix(path, "/api/") || path == "/api" {
			c.AbortWithStatus(consts.StatusNotFound)
			return
		}

		if !c.IsGet() && !c.IsHead() {
			c.AbortWithStatus(consts.StatusNotFound)
			return
		}

		cleanedPath := filepath.Clean("/" + strings.TrimPrefix(path, "/"))
		target := filepath.Join(options.PublicDir, cleanedPath)

		// 如果真有对应文件，就直接返回
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			app.ServeFile(c, target)
			return
		}

		// 否则按 SPA 路由处理，回退到 index.html
		app.ServeFile(c, indexFile)
	})
}
