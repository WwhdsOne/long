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

	serveEmbedFile := func(c *app.RequestContext, path string) {
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

	// SPA fallback: 非 /api/ 路径尝试读取对应文件，不存在则回退到 index.html
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

		cleanedPath := strings.TrimPrefix(filepath.Clean("/"+strings.TrimPrefix(path, "/")), "/")

		if cleanedPath != "" {
			if _, err := fs.Stat(publicFS, cleanedPath); err == nil {
				serveEmbedFile(c, cleanedPath)
				return
			}
		}

		// SPA fallback
		serveEmbedFile(c, "index.html")
	})
}
