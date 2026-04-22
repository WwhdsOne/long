package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerHeroRoutes(router route.IRouter, options Options) {
	router.POST("/api/heroes/:heroId/equip", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再派出英雄。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.EquipHero(ctx, nickname, c.Param("heroId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrHeroNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "HERO_NOT_FOUND"})
			case errors.Is(err, vote.ErrHeroNotOwned):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "HERO_NOT_OWNED",
					"message": "这位小小英雄还没加入你的队伍。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_EQUIP_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/heroes/:heroId/unequip", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再收回英雄。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.UnequipHero(ctx, nickname, c.Param("heroId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrHeroNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "HERO_NOT_FOUND"})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_UNEQUIP_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/heroes/:heroId/salvage", func(ctx context.Context, c *app.RequestContext) {
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

		state, err := options.Store.SalvageHero(ctx, nickname, c.Param("heroId"), body.Quantity)
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrHeroNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "HERO_NOT_FOUND"})
			case errors.Is(err, vote.ErrInvalidQuantity):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "INVALID_QUANTITY",
					"message": "分解数量至少要填 1。",
				})
			case errors.Is(err, vote.ErrHeroNotOwned), errors.Is(err, vote.ErrHeroNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "HERO_NOT_ENOUGH",
					"message": "当前只能分解重复英雄，出战中的最后一位必须保留。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_SALVAGE_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})

	router.POST("/api/heroes/:heroId/awaken", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再觉醒。",
		}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		state, err := options.Store.AwakenHero(ctx, nickname, c.Param("heroId"))
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrHeroNotFound):
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "HERO_NOT_FOUND"})
			case errors.Is(err, vote.ErrHeroNotOwned):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "HERO_NOT_OWNED",
					"message": "这位小小英雄还没加入你的队伍。",
				})
			case errors.Is(err, vote.ErrHeroMaxAwaken):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "HERO_MAX_AWAKEN",
					"message": "这位小小英雄已经达到觉醒上限。",
				})
			case errors.Is(err, vote.ErrGemsNotEnough):
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "GEMS_NOT_ENOUGH",
					"message": "原石不够，先去分解点重复英雄吧。",
				})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "HERO_AWAKEN_FAILED"})
			}
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, state)
	})
}
