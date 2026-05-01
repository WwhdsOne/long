package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
	"long/internal/xlog"
)

func registerAdminResourceRoutes(router route.IRouter, options Options) {
	registerAdminEquipmentRoutes(router, options)
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
			writeEquipmentDraftFailureLog(ctx, options.EquipmentDraftFailureWriter, body.Prompt, err)
			if errors.Is(err, ErrInvalidEquipmentDraft) {
				xlog.L().Warn("invalid equipment draft", xlog.Err(err))
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
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "equipment.create",
			TargetType:  "equipment",
			TargetID:    body.ItemID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "equipment.created",
			ItemID:    body.ItemID,
			Payload: map[string]any{
				"name": body.Name,
				"slot": body.Slot,
			},
		})

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
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "equipment.update",
			TargetType:  "equipment",
			TargetID:    body.ItemID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "equipment.updated",
			ItemID:    body.ItemID,
			Payload: map[string]any{
				"name": body.Name,
				"slot": body.Slot,
			},
		})

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
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "equipment.delete",
			TargetType:  "equipment",
			TargetID:    c.Param("itemId"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "equipment.deleted",
			ItemID:    c.Param("itemId"),
		})

		publishChange(ctx, options.ChangePublisher, vote.StateChange{
			Type:             vote.StateChangeEquipmentMetaChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}

func writeEquipmentDraftFailureLog(ctx context.Context, writer EquipmentDraftFailureWriter, prompt string, err error) {
	if writer == nil || err == nil {
		return
	}

	item := vote.EquipmentDraftFailureLog{
		Prompt:       prompt,
		ErrorMessage: err.Error(),
		CreatedAt:    time.Now().Unix(),
	}

	var generateErr *EquipmentDraftGenerateError
	if errors.As(err, &generateErr) {
		if strings.TrimSpace(generateErr.Prompt) != "" {
			item.Prompt = strings.TrimSpace(generateErr.Prompt)
		}
		item.Draft = generateErr.Draft
		item.RawResponse = generateErr.RawResponse
		if strings.TrimSpace(generateErr.Error()) != "" {
			item.ErrorMessage = generateErr.Error()
		}
	}

	if writeErr := writer.WriteEquipmentDraftFailure(ctx, item); writeErr != nil {
		xlog.L().Warn("write equipment draft failure log failed", xlog.Err(writeErr))
	}
}
