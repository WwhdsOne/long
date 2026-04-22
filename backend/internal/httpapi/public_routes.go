package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerPublicRoutes(router route.IRouter, options Options, stateView StateView) {
	router.GET("/api/health", func(_ context.Context, c *app.RequestContext) {
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.GET("/api/buttons", func(ctx context.Context, c *app.RequestContext) {
		state, err := stateView.GetState(ctx, resolvedPlayerNicknameForRead(ctx, c, options.PlayerAuthenticator))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, state)
	})

	router.GET("/api/shop", func(ctx context.Context, c *app.RequestContext) {
		state, err := stateView.GetState(ctx, resolvedPlayerNicknameForRead(ctx, c, options.PlayerAuthenticator))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SHOP_FETCH_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]any{
			"gems":              state.Gems,
			"ownedCosmetics":    state.OwnedCosmetics,
			"equippedCosmetics": state.EquippedCosmetics,
			"shopCatalog":       state.ShopCatalog,
		})
	})

	router.GET("/api/boss/history", func(ctx context.Context, c *app.RequestContext) {
		history, err := options.Store.ListBossHistory(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, history)
	})

	router.GET("/api/announcements/latest", func(ctx context.Context, c *app.RequestContext) {
		item, err := options.Store.GetLatestAnnouncement(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_FETCH_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, item)
	})

	router.GET("/api/announcements", func(ctx context.Context, c *app.RequestContext) {
		items, err := options.Store.ListAnnouncements(ctx, false)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, items)
	})

	router.GET("/api/messages", func(ctx context.Context, c *app.RequestContext) {
		page, err := options.Store.ListMessages(ctx, c.Query("cursor"), 50)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "MESSAGE_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, page)
	})

	router.POST("/api/messages", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
			Content  string `json:"content"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			if options.PlayerAuthenticator != nil {
				return
			}
			writeJSON(c, consts.StatusBadRequest, map[string]string{
				"error":   "INVALID_NICKNAME",
				"message": "昵称还没填好，先起个名字再发。",
			})
			return
		}

		message, err := options.Store.CreateMessage(ctx, nickname, body.Content)
		if err != nil {
			if writeNicknameError(c, err) || writeContentError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "MESSAGE_CREATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeMessageCreated,
			Nickname:  nickname,
			Timestamp: time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, message)
	})

	router.POST("/api/nickname/validate", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再试试。",
		}) {
			return
		}

		if err := options.Store.ValidateNickname(ctx, body.Nickname); err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "NICKNAME_VALIDATE_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}
