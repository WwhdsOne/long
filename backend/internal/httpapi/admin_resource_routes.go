package httpapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerAdminResourceRoutes(router route.IRouter, options Options) {
	registerAdminButtonRoutes(router, options)
	registerAdminEquipmentRoutes(router, options)
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
	router.POST("/api/admin/equipment/generate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		if options.EquipmentDraftGenerator == nil {
			writeJSON(c, consts.StatusServiceUnavailable, map[string]string{"error": "EQUIPMENT_GENERATOR_DISABLED"})
			return
		}

		var body struct {
			Prompt string `json:"prompt"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		draft, err := options.EquipmentDraftGenerator.GenerateEquipmentDraft(ctx, body.Prompt)
		if err != nil {
			if errors.Is(err, ErrInvalidEquipmentDraft) {
				fmt.Printf("invalid equipment draft: %s\n", err)
				writeJSON(c, consts.StatusUnprocessableEntity, map[string]string{"error": "INVALID_EQUIPMENT_DRAFT"})
				return
			}
			writeJSON(c, consts.StatusBadGateway, map[string]string{"error": "EQUIPMENT_GENERATE_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]vote.EquipmentDefinition{"draft": draft})
	})

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
