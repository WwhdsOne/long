package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerAdminResourceRoutes(router route.IRouter, options Options) {
	registerAdminButtonRoutes(router, options)
	registerAdminEquipmentRoutes(router, options)
	registerAdminHeroRoutes(router, options)
}

func registerAdminButtonRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/buttons", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.ButtonUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SaveButton(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BUTTON_SAVE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeButtonMetaChanged,
			Timestamp: time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/buttons/:slug", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.ButtonUpsert
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.Slug = c.Param("slug")

		if err := options.Store.SaveButton(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BUTTON_SAVE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeButtonMetaChanged,
			Timestamp: time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}

func registerAdminEquipmentRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/equipment", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.EquipmentDefinition
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SaveEquipmentDefinition(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_SAVE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/equipment/:itemId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.EquipmentDefinition
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.ItemID = c.Param("itemId")

		if err := options.Store.SaveEquipmentDefinition(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_SAVE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.DELETE("/api/admin/equipment/:itemId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := options.Store.DeleteEquipmentDefinition(ctx, c.Param("itemId")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_DELETE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}

func registerAdminHeroRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/heroes", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.HeroDefinition
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SaveHeroDefinition(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_SAVE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/heroes/:heroId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body vote.HeroDefinition
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.HeroID = c.Param("heroId")

		if err := options.Store.SaveHeroDefinition(ctx, body); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_SAVE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.DELETE("/api/admin/heroes/:heroId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := options.Store.DeleteHeroDefinition(ctx, c.Param("heroId")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_DELETE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}
