package httpapi

import (
	"context"
	"io/fs"
	"mime"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	publicfs "long"
)

func registerRealtimeRoutes(router route.IRouter, options Options) {
	if options.Events != nil {
		router.GET("/api/events", options.Events)
	}
	if options.RealtimeHub != nil {
		router.GET("/api/ws", newRealtimeSocketHandler(options))
	}
}

func registerStaticRoutes(engine *route.Engine, _ Options) {
	publicFS, err := fs.Sub(publicfs.FS, "public")
	if err != nil {
		return
	}

	// 显式 GET 通配路由，优先于 NoRoute，避免走 notFound 链路
	engine.GET("/*filepath", func(_ context.Context, c *app.RequestContext) {
		path := string(c.Param("filepath"))

		if strings.HasPrefix(path, "/api/") || path == "/api" {
			c.AbortWithStatus(consts.StatusNotFound)
			return
		}

		cleanedPath := strings.TrimPrefix(filepath.Clean(path), "/")
		if cleanedPath == "" {
			cleanedPath = "index.html"
		}

		// 先尝试读取对应文件
		if data, err := fs.ReadFile(publicFS, cleanedPath); err == nil {
			ext := filepath.Ext(cleanedPath)
			if ct := mime.TypeByExtension(ext); ct != "" {
				c.Response.Header.Set("Content-Type", ct)
			}
			c.Write(data)
			return
		}

		// SPA fallback：非 API 路径回退到 index.html
		if data, err := fs.ReadFile(publicFS, "index.html"); err == nil {
			c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
			c.Write(data)
			return
		}

		c.AbortWithStatus(consts.StatusNotFound)
	})

	// NoRoute 兜底：非 GET 请求或无匹配时
	engine.NoRoute(func(_ context.Context, c *app.RequestContext) {
		path := string(c.Path())

		if strings.HasPrefix(path, "/api") && path != "/api" {
			c.AbortWithStatus(consts.StatusNotFound)
			return
		}

		if data, err := fs.ReadFile(publicFS, "index.html"); err == nil {
			c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
			c.Write(data)
			return
		}

		c.AbortWithStatus(consts.StatusNotFound)
	})
}
