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
)

func registerEquipmentRoutes(router route.IRouter, options Options) {
	router.POST("/api/equipment/:instanceId/equip", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再穿装备。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.EquipItem(ctx, nickname, c.Param("instanceId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotFound) {
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotOwned) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_OWNED",
					"message": "这件装备还不在你的背包里。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIP_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/equipment/:instanceId/unequip", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再卸装备。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.UnequipItem(ctx, nickname, c.Param("instanceId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotFound) {
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "UNEQUIP_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/equipment/:instanceId/synthesize", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再升星。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}
		_ = nickname
		writeJSON(c, consts.StatusGone, map[string]string{
			"error":   "EQUIPMENT_SYNTHESIZE_DEPRECATED",
			"message": "3 合 1 升星已废弃，请改用装备强化。",
		})
	})

	router.POST("/api/equipment/:instanceId/enhance", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再强化。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.EnhanceItem(ctx, nickname, c.Param("instanceId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotFound) {
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotOwned) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_OWNED",
					"message": "这件装备还不在你的背包里。",
				})
				return
			}
			if errors.Is(err, vote.ErrEquipmentEnhanceMaxLevel) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_ENHANCE_MAX_LEVEL",
					"message": "这件装备已达到强化上限。",
				})
				return
			}
			if errors.Is(err, vote.ErrEquipmentEnhanceInsufficientGold) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_ENHANCE_GOLD_NOT_ENOUGH",
					"message": "金币不足，无法强化。",
				})
				return
			}
			if errors.Is(err, vote.ErrEquipmentEnhanceInsufficientStones) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_ENHANCE_STONE_NOT_ENOUGH",
					"message": "强化石不足，无法强化。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ENHANCE_FAILED"})
			return
		}
		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/equipment/:instanceId/salvage", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再分解。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		result, err := options.Store.SalvageItem(ctx, nickname, c.Param("instanceId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotFound) {
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotOwned) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_OWNED",
					"message": "这件装备还不在你的背包里。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SALVAGE_FAILED"})
			return
		}
		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, result)
	})
}

func publishEquipmentChange(ctx context.Context, nickname string, publisher ChangePublisher) {
	publishChange(ctx, publisher, vote.StateChange{
		Type:      vote.StateChangeEquipmentChanged,
		Nickname:  strings.TrimSpace(nickname),
		Timestamp: time.Now().Unix(),
	})
}
