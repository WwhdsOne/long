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

		state, err := options.Store.EquipItem(ctx, body.Nickname, c.Param("itemId"))
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

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
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

		state, err := options.Store.UnequipItem(ctx, body.Nickname, c.Param("itemId"))
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

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
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

		state, err := options.Store.SynthesizeItem(ctx, body.Nickname, c.Param("itemId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrEquipmentNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
			case errors.Is(err, vote.ErrEquipmentNotOwned), errors.Is(err, vote.ErrEquipmentNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_ENOUGH",
					"message": "至少要有 3 件同名装备才能升星。",
				})
			case errors.Is(err, vote.ErrEquipmentMaxStar):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_MAX_STAR",
					"message": "这件装备已经满星了。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SYNTHESIZE_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
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

		state, err := options.Store.SalvageEquipment(ctx, body.Nickname, c.Param("itemId"), body.Quantity)
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

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/equipment/:itemId/reforge", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再强化。",
		}) {
			return
		}

		state, err := options.Store.ReforgeEquipment(ctx, body.Nickname, c.Param("itemId"))
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
			case errors.Is(err, vote.ErrGemsNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "GEMS_NOT_ENOUGH",
					"message": "原石不够，先去分解点重复装备吧。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_REFORGE_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
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
