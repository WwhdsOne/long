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

func registerEventRoutes(router route.IRouter, options Options) {
	if options.Events != nil {
		router.GET("/api/events", options.Events)
	}
}

func registerStaticRoutes(engine *route.Engine, options Options) {
	if options.PublicDir == "" {
		return
	}

	indexFile := filepath.Join(options.PublicDir, "index.html")

	// 只挂明确的静态资源目录，避免把所有路由都吃掉
	assetsDir := filepath.Join(options.PublicDir, "assets")
	if stat, err := os.Stat(assetsDir); err == nil && stat.IsDir() {
		engine.Static("/assets", assetsDir)
	}

	// 常见静态文件单独挂
	faviconFile := filepath.Join(options.PublicDir, "favicon.ico")
	if stat, err := os.Stat(faviconFile); err == nil && !stat.IsDir() {
		engine.StaticFile("/favicon.ico", faviconFile)
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

		// 真存在的静态文件直接返回
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			app.ServeFile(c, target)
			return
		}

		// 其他前端路由全部回退到 index.html
		app.ServeFile(c, indexFile)
	})
}
