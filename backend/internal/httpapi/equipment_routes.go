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
	router.POST("/api/equipment/:itemId/equip", func(ctx context.Context, c *app.RequestContext) {
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

		state, err := options.Store.EquipItem(ctx, nickname, c.Param("itemId"))
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

	router.POST("/api/equipment/:itemId/unequip", func(ctx context.Context, c *app.RequestContext) {
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

		state, err := options.Store.UnequipItem(ctx, nickname, c.Param("itemId"))
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

	router.POST("/api/equipment/:itemId/synthesize", func(ctx context.Context, c *app.RequestContext) {
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

	router.POST("/api/equipment/:itemId/salvage", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
			Quantity int64  `json:"quantity"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称和分解数量都得带上。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.SalvageEquipment(ctx, nickname, c.Param("itemId"), body.Quantity)
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrEquipmentNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
			case errors.Is(err, vote.ErrInvalidQuantity):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "INVALID_QUANTITY",
					"message": "分解数量至少要填 1。",
				})
			case errors.Is(err, vote.ErrEquipmentNotOwned), errors.Is(err, vote.ErrEquipmentNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_ENOUGH",
					"message": "当前只能分解多出来的装备，穿戴中的那一件必须留着。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_SALVAGE_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/equipment/:itemId/enhance", func(ctx context.Context, c *app.RequestContext) {
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

		state, err := options.Store.EnhanceEquipment(ctx, nickname, c.Param("itemId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrEquipmentNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
			case errors.Is(err, vote.ErrEquipmentNotOwned):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_OWNED",
					"message": "这件装备还不在你的背包里。",
				})
			case errors.Is(err, vote.ErrEquipmentMaxEnhance):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_MAX_ENHANCE",
					"message": "这件装备已经达到强化上限。",
				})
			case errors.Is(err, vote.ErrGemsNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "GEMS_NOT_ENOUGH",
					"message": "原石不够，先去分解点重复装备吧。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_ENHANCE_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})
}

func publishEquipmentChange(ctx context.Context, nickname string, publisher ChangePublisher) {
	publishChange(ctx, publisher, vote.StateChange{
		Type:      vote.StateChangeEquipmentChanged,
		Nickname:  strings.TrimSpace(nickname),
		Timestamp: time.Now().Unix(),
	})
}
