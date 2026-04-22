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

	engine.StaticFS("/", &app.FS{Root: options.PublicDir, GenerateIndexPages: false})
	indexFile := filepath.Join(options.PublicDir, "index.html")

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
		if cleanedPath == "/" {
			app.ServeFile(c, indexFile)
			return
		}

		target := filepath.Join(options.PublicDir, cleanedPath)
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			app.ServeFile(c, target)
			return
		}

		app.ServeFile(c, indexFile)
	})
}
