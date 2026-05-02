package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

type bossResourceReader interface {
	GetBossResources(context.Context) (core.BossResources, error)
}

func registerPublicRoutes(router route.IRouter, options Options, stateView StateView) {
	messageStore := MessageStore(options.Store)
	if options.MessageStore != nil {
		messageStore = options.MessageStore
	}
	registerShopRoutes(router, options)

	router.GET("/api/health", func(_ context.Context, c *app.RequestContext) {
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.GET("/api/online-count", func(_ context.Context, c *app.RequestContext) {
		count := 0
		if options.RealtimeHub != nil {
			count = options.RealtimeHub.SubscriberCount()
		}
		writeJSON(c, consts.StatusOK, map[string]int{"count": count})
	})

	router.GET("/api/boss/history", func(ctx context.Context, c *app.RequestContext) {
		history, err := options.Store.ListBossHistory(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, history)
	})

	router.GET("/api/boss/resources", func(ctx context.Context, c *app.RequestContext) {
		reader := bossResourceReader(options.Store)
		if cachedReader, ok := stateView.(bossResourceReader); ok {
			reader = cachedReader
		}

		resources, err := reader.GetBossResources(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_RESOURCES_FETCH_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, resources)
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
		page, err := messageStore.ListMessages(ctx, c.Query("cursor"), 50)
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

		message, err := messageStore.CreateMessage(ctx, nickname, body.Content)
		if err != nil {
			if writeNicknameError(c, err) || writeContentError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "MESSAGE_CREATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:      core.StateChangeMessageCreated,
			Nickname:  nickname,
			Timestamp: time.Now().Unix(),
		})
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "message.created",
			Nickname:  nickname,
			Payload: map[string]any{
				"message_id": message.ID,
			},
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
