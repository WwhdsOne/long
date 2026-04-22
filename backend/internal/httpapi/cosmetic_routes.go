package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerCosmeticRoutes(router route.IRouter, options Options) {
	router.POST("/api/shop/cosmetics/:cosmeticId/purchase", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再买外观。",
		}) {
			return
		}

		state, err := options.Store.PurchaseCosmetic(ctx, body.Nickname, c.Param("cosmeticId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrCosmeticNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "COSMETIC_NOT_FOUND"})
			case errors.Is(err, vote.ErrCosmeticAlreadyOwned):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "COSMETIC_ALREADY_OWNED",
					"message": "这件外观已经在你的衣柜里了。",
				})
			case errors.Is(err, vote.ErrGemsNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "GEMS_NOT_ENOUGH",
					"message": "原石不够，先去分解重复资源吧。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "COSMETIC_PURCHASE_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/shop/cosmetics/equip", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
			TrailID  string `json:"trailId"`
			ImpactID string `json:"impactId"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称和外观槽位都得带上。",
		}) {
			return
		}

		state, err := options.Store.EquipCosmetics(ctx, body.Nickname, body.TrailID, body.ImpactID)
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrCosmeticNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "COSMETIC_NOT_FOUND"})
			case errors.Is(err, vote.ErrCosmeticNotOwned), errors.Is(err, vote.ErrInvalidCosmeticLoadout):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "COSMETIC_EQUIP_FAILED",
					"message": "只能装备已经拥有、且槽位匹配的外观。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "COSMETIC_EQUIP_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, body.Nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})
}
