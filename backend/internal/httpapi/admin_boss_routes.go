package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerAdminBossRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/boss/activate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.BossUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		boss, err := options.Store.ActivateBoss(ctx, body)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_ACTIVATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
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

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
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
			Loot   []vote.BossLootEntry `json:"loot"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SetBossLoot(ctx, body.BossID, body.Loot); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_LOOT_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeBossChanged,
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

		var body vote.BossTemplateUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SaveBossTemplate(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/pool/:templateId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.BossTemplateUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.ID = c.Param("templateId")

		if err := options.Store.SaveBossTemplate(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
			return
		}

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

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/pool/:templateId/loot", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			Loot []vote.BossLootEntry `json:"loot"`
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

	router.PUT("/api/admin/boss/pool/:templateId/hero-loot", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			Loot []vote.BossHeroLootEntry `json:"loot"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SetBossTemplateHeroLoot(ctx, c.Param("templateId"), body.Loot); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_HERO_LOOT_FAILED"})
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
			if errors.Is(err, vote.ErrBossPoolEmpty) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_POOL_EMPTY",
					"message": "Boss 池还是空的，先加模板再开启循环。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_ENABLE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, boss)
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

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
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

		history, err := options.Store.ListAdminBossHistoryPage(ctx, page, pageSize)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, history)
	})
}
