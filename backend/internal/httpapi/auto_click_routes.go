package httpapi

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
)

func registerAutoClickRoutes(router route.IRouter, options Options) {
	router.GET("/api/auto-click", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		if options.AutoClick == nil {
			writeJSON(c, consts.StatusOK, AutoClickStatus{})
			return
		}
		writeJSON(c, consts.StatusOK, options.AutoClick.Status(nickname))
	})

	router.POST("/api/auto-click/start", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		if options.AutoClick == nil {
			writeJSON(c, consts.StatusServiceUnavailable, map[string]string{"error": "AUTO_CLICK_UNAVAILABLE"})
			return
		}

		var body struct {
			Slug string `json:"slug"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "挂机目标不能为空。",
		}) {
			return
		}

		status, err := options.AutoClick.Start(ctx, nickname, strings.TrimSpace(body.Slug))
		if err != nil {
			writeJSON(c, consts.StatusBadRequest, map[string]string{
				"error":   "INVALID_AUTO_CLICK_TARGET",
				"message": "挂机目标不能为空。",
			})
			return
		}
		writeJSON(c, consts.StatusOK, status)
	})

	router.POST("/api/auto-click/stop", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		if options.AutoClick == nil {
			writeJSON(c, consts.StatusOK, AutoClickStatus{})
			return
		}
		writeJSON(c, consts.StatusOK, options.AutoClick.Stop(nickname))
	})
}
