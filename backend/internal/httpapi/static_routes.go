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

	// 显式为内嵌的静态文件注册路由，确保不走 NoRoute 链路
	serveFile := func(path string) app.HandlerFunc {
		return func(_ context.Context, c *app.RequestContext) {
			data, err := fs.ReadFile(publicFS, path)
			if err != nil {
				c.AbortWithStatus(consts.StatusNotFound)
				return
			}
			ext := filepath.Ext(path)
			if ct := mime.TypeByExtension(ext); ct != "" {
				c.Response.Header.Set("Content-Type", ct)
			}
			c.Write(data)
		}
	}

	// 根路径和独立文件用精确路由
	engine.GET("/", func(_ context.Context, c *app.RequestContext) {
		data, err := fs.ReadFile(publicFS, "index.html")
		if err != nil {
			c.AbortWithStatus(consts.StatusNotFound)
			return
		}
		c.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
		c.Write(data)
	})
	engine.GET("/favicon.svg", serveFile("favicon.svg"))
	engine.GET("/icons.svg", serveFile("icons.svg"))

	// /assets、/images、/effects 等目录用通配路由
	engine.GET("/*filepath", func(_ context.Context, c *app.RequestContext) {
		path := string(c.Param("filepath"))

		if strings.HasPrefix(path, "/api/") || path == "/api" {
			c.AbortWithStatus(consts.StatusNotFound)
			return
		}

		cleanedPath := strings.TrimPrefix(filepath.Clean(path), "/")
		if cleanedPath == "" {
			c.Redirect(consts.StatusMovedPermanently, []byte("/"))
			return
		}

		if data, err := fs.ReadFile(publicFS, cleanedPath); err == nil {
			ext := filepath.Ext(cleanedPath)
			if ct := mime.TypeByExtension(ext); ct != "" {
				c.Response.Header.Set("Content-Type", ct)
			}
			c.Write(data)
			return
		}

		// SPA fallback
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
