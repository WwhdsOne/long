package httpapi

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
)

func registerPlayerPresenceRoutes(router route.IRouter, options Options) {
	router.POST("/api/player/presence", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		var body struct {
			Visible bool `json:"visible"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if options.Afk != nil {
			if err := options.Afk.ReportPresence(ctx, nickname, body.Visible); err != nil {
				if writeNicknameError(c, err) {
					return
				}
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "PRESENCE_REPORT_FAILED"})
				return
			}
		}

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.GET("/api/player/afk/settlement", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		if options.Afk == nil {
			writeJSON(c, consts.StatusOK, map[string]any{
				"kills":      0,
				"goldTotal":  0,
				"stoneTotal": 0,
				"startedAt":  0,
				"endedAt":    0,
			})
			return
		}
		writeJSON(c, consts.StatusOK, options.Afk.ConsumeSettlement(nickname))
	})
}
