package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

func registerAdminBossRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/boss/activate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body core.BossUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		boss, err := options.Store.ActivateBoss(ctx, body)
		if err != nil {
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 必须配置至少一个部位。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_ACTIVATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.activate",
			TargetType:  "boss",
			TargetID:    boss.ID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "boss.activated",
			BossID:    boss.ID,
			Payload: map[string]any{
				"name": boss.Name,
			},
		})
		writeJSON(c, consts.StatusOK, boss)
	})

	router.POST("/api/admin/boss/deactivate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := options.Store.DeactivateBoss(ctx); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_DEACTIVATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.deactivate",
			TargetType:  "boss",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "boss.deactivated",
			Payload: map[string]any{
				"operator": options.AdminAuthenticator.Username(),
			},
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/loot", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			BossID string               `json:"bossId"`
			Loot   []core.BossLootEntry `json:"loot"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SetBossLoot(ctx, body.BossID, body.Loot); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_LOOT_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/boss/pool", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body core.BossTemplateUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SaveBossTemplate(ctx, body); err != nil {
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 模板必须配置至少一个部位。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
			return
		}

		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.template.create",
			TargetType:  "boss_template",
			TargetID:    body.ID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/pool/:templateId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body core.BossTemplateUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.ID = c.Param("templateId")

		if err := options.Store.SaveBossTemplate(ctx, body); err != nil {
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 模板必须配置至少一个部位。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
			return
		}

		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.template.update",
			TargetType:  "boss_template",
			TargetID:    body.ID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.DELETE("/api/admin/boss/pool/:templateId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := options.Store.DeleteBossTemplate(ctx, c.Param("templateId")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_DELETE_FAILED"})
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.template.delete",
			TargetType:  "boss_template",
			TargetID:    c.Param("templateId"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/pool/:templateId/loot", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			Loot []core.BossLootEntry `json:"loot"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SetBossTemplateLoot(ctx, c.Param("templateId"), body.Loot); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_LOOT_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/boss/cycle/enable", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		boss, err := options.Store.SetBossCycleEnabled(ctx, true)
		if err != nil {
			if errors.Is(err, core.ErrBossPoolEmpty) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_POOL_EMPTY",
					"message": "Boss 池还是空的，先加模板再开启循环。",
				})
				return
			}
			if errors.Is(err, core.ErrBossCycleQueueEmpty) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_CYCLE_QUEUE_EMPTY",
					"message": "请先在 Boss 池里配置循环队列。",
				})
				return
			}
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 模板缺少部位，请先修正模板。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_ENABLE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.cycle.enable",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, boss)
	})

	router.PUT("/api/admin/boss/cycle/queue", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			TemplateIDs []string `json:"templateIds"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		queue, err := options.Store.SetBossCycleQueue(ctx, body.TemplateIDs)
		if err != nil {
			if errors.Is(err, core.ErrBossTemplateNotFound) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_TEMPLATE_NOT_FOUND",
					"message": "循环队列里包含不存在的 Boss 模板。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_QUEUE_SAVE_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]any{
			"ok":          true,
			"templateIds": queue,
		})
	})

	router.POST("/api/admin/boss/cycle/disable", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		boss, err := options.Store.SetBossCycleEnabled(ctx, false)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_DISABLE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.cycle.disable",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, boss)
	})

	router.GET("/api/admin/boss/history", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		page, pageSize, ok := parseAdminPageParams(c)
		if !ok {
			return
		}

		var historyReader AdminBossHistoryReader = options.Store
		if options.AdminBossHistoryReader != nil {
			historyReader = options.AdminBossHistoryReader
		}

		history, err := historyReader.ListAdminBossHistoryPage(ctx, page, pageSize)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, history)
	})
}
