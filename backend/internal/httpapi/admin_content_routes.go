package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerAdminContentRoutes(router route.IRouter, options Options) {
	messageStore := MessageStore(options.Store)
	if options.MessageStore != nil {
		messageStore = options.MessageStore
	}

	router.GET("/api/admin/announcements", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		items, err := options.Store.ListAnnouncements(ctx, true)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, items)
	})

	router.POST("/api/admin/announcements", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.AnnouncementUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		item, err := options.Store.SaveAnnouncement(ctx, body)
		if err != nil {
			if writeContentError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_SAVE_FAILED"})
			return
		}
		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeAnnouncementChanged,
			Timestamp: time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, item)
	})

	router.DELETE("/api/admin/announcements/:id", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := options.Store.DeleteAnnouncement(ctx, c.Param("id")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_DELETE_FAILED"})
			return
		}
		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeAnnouncementChanged,
			Timestamp: time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.GET("/api/admin/messages", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		page, err := messageStore.ListMessages(ctx, c.Query("cursor"), 50)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "MESSAGE_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, page)
	})

	router.DELETE("/api/admin/messages/:id", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := messageStore.DeleteMessage(ctx, c.Param("id")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "MESSAGE_DELETE_FAILED"})
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "message.delete",
			TargetType:  "message",
			TargetID:    c.Param("id"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "message.deleted",
			Payload: map[string]any{
				"message_id": c.Param("id"),
				"operator":   options.AdminAuthenticator.Username(),
			},
		})
		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeMessageDeleted,
			Timestamp: time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/oss/sts", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		if options.OSSSigner == nil {
			writeJSON(c, consts.StatusServiceUnavailable, map[string]string{
				"error":   "OSS_NOT_CONFIGURED",
				"message": "OSS 直传还没配置，先手动填图片 URL。",
			})
			return
		}

		policy, err := options.OSSSigner.CreatePolicy(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "OSS_POLICY_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, policy)
	})
}
