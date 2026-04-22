package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	playerauth "long/internal/playerauth"
)

func registerPlayerAuthRoutes(router route.IRouter, options Options) {
	if options.PlayerAuthenticator == nil {
		return
	}

	router.POST("/api/player/auth/login", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
			Password string `json:"password"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		token, nickname, err := options.PlayerAuthenticator.Login(ctx, body.Nickname, body.Password)
		if err != nil {
			switch {
			case writeNicknameError(c, err):
				return
			case errors.Is(err, playerauth.ErrInvalidPassword):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "INVALID_PASSWORD",
					"message": "密码不能为空。",
				})
			case errors.Is(err, playerauth.ErrInvalidCredentials):
				writeJSON(c, consts.StatusUnauthorized, map[string]string{
					"error":   "INVALID_CREDENTIALS",
					"message": "昵称或密码不对。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "PLAYER_LOGIN_FAILED"})
			}
			return
		}

		setPlayerSessionCookie(c, token)
		writeJSON(c, consts.StatusOK, map[string]any{
			"authenticated": true,
			"nickname":      nickname,
		})
	})

	router.POST("/api/player/auth/logout", func(_ context.Context, c *app.RequestContext) {
		clearPlayerSessionCookie(c)
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.GET("/api/player/auth/session", func(ctx context.Context, c *app.RequestContext) {
		nickname := authenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if nickname == "" {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]any{
			"authenticated": true,
			"nickname":      nickname,
		})
	})
}
