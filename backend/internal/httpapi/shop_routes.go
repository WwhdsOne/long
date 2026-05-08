package httpapi

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

func registerShopRoutes(router route.IRouter, options Options) {
	router.GET("/api/shop/items", func(ctx context.Context, c *app.RequestContext) {
		nickname := resolvedPlayerNicknameForRead(ctx, c, options.PlayerAuthenticator)
		items, err := options.Store.ListShopCatalogItemsForPlayer(ctx, nickname)
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SHOP_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, items)
	})

	router.POST("/api/shop/items/:itemId/purchase", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		state, err := options.Store.PurchaseShopItem(ctx, nickname, c.Param("itemId"))
		if err != nil {
			if writeShopError(c, err) || writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SHOP_PURCHASE_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "shop.item_purchased",
			Nickname:  nickname,
			ItemID:    c.Param("itemId"),
		})
		writeJSON(c, consts.StatusOK, core.ShopActionResult{
			ItemID:    c.Param("itemId"),
			UserState: state,
		})
	})

	router.POST("/api/shop/items/unequip", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		state, err := options.Store.UnequipShopItem(ctx, nickname)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SHOP_UNEQUIP_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "shop.item_unequipped",
			Nickname:  nickname,
		})
		writeJSON(c, consts.StatusOK, map[string]any{
			"userState": state,
		})
	})

	router.POST("/api/shop/items/:itemId/equip", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		state, err := options.Store.EquipShopItem(ctx, nickname, c.Param("itemId"))
		if err != nil {
			if writeShopError(c, err) || writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "SHOP_EQUIP_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "shop.item_equipped",
			Nickname:  nickname,
			ItemID:    c.Param("itemId"),
		})
		writeJSON(c, consts.StatusOK, core.ShopActionResult{
			ItemID:    c.Param("itemId"),
			UserState: state,
		})
	})

	router.POST("/api/shop/stamina/full/purchase", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			TurnstileToken string `json:"turnstileToken"`
		}
		if rawBody := c.Request.Body(); len(rawBody) > 0 && !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		if options.StaminaPurchaseTurnstile != nil {
			turnstileResult, err := options.StaminaPurchaseTurnstile.CheckPurchaseStamina(ctx, StaminaPurchaseTurnstileRequest{
				Nickname: nickname,
				Token:    strings.TrimSpace(body.TurnstileToken),
				RemoteIP: requestIP(c),
			})
			if err != nil {
				writeJSON(c, consts.StatusServiceUnavailable, map[string]string{
					"error":   "CAPTCHA_VERIFY_UNAVAILABLE",
					"message": "验证服务暂时不可用，请稍后再试",
				})
				return
			}
			switch turnstileResult.Decision {
			case StaminaPurchaseTurnstileAllow:
			case StaminaPurchaseTurnstileRequire:
				writeJSON(c, consts.StatusBadRequest, map[string]any{
					"error":   "CAPTCHA_REQUIRED",
					"message": "本次购买需要完成人机验证",
					"siteKey": turnstileResult.SiteKey,
				})
				return
			case StaminaPurchaseTurnstileInvalid:
				if options.AccountRisk != nil {
					if _, err := options.AccountRisk.RecordAccountRiskEvent(ctx, nickname, core.AccountRiskEventStaminaTurnstileInvalid); err != nil {
						writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ACCOUNT_RISK_FAILED"})
						return
					}
				}
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "CAPTCHA_INVALID",
					"message": "验证失败，请重试",
				})
				return
			default:
				writeJSON(c, consts.StatusServiceUnavailable, map[string]string{
					"error":   "CAPTCHA_VERIFY_UNAVAILABLE",
					"message": "验证服务暂时不可用，请稍后再试",
				})
				return
			}
		}
		state, err := options.Store.PurchaseStaminaFull(ctx, nickname)
		if err != nil {
			if writeShopError(c, err) || writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "STAMINA_PURCHASE_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, map[string]any{
			"userState": state,
		})
	})

	router.POST("/api/shop/stamina/cap/upgrade", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		state, err := options.Store.UpgradeStaminaCap(ctx, nickname)
		if err != nil {
			if writeShopError(c, err) || writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "STAMINA_UPGRADE_FAILED"})
			return
		}

		publishEquipmentChange(ctx, nickname, options.ChangePublisher)
		writeJSON(c, consts.StatusOK, map[string]any{
			"userState": state,
		})
	})
}

func writeShopError(c *app.RequestContext, err error) bool {
	switch {
	case errors.Is(err, core.ErrShopItemNotFound):
		writeJSON(c, consts.StatusNotFound, map[string]string{
			"error":   "SHOP_ITEM_NOT_FOUND",
			"message": "商品不存在。",
		})
		return true
	case errors.Is(err, core.ErrShopItemNotPurchasable), errors.Is(err, core.ErrShopUnsupportedItemType):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "SHOP_ITEM_NOT_PURCHASABLE",
			"message": "商品当前不可购买或类型暂不支持。",
		})
		return true
	case errors.Is(err, core.ErrShopItemAlreadyOwned):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "SHOP_ITEM_ALREADY_OWNED",
			"message": "这个点击图标你已经拥有了。",
		})
		return true
	case errors.Is(err, core.ErrShopItemNotOwned):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "SHOP_ITEM_NOT_OWNED",
			"message": "还没有拥有这个点击图标，不能直接使用。",
		})
		return true
	case errors.Is(err, core.ErrShopInsufficientGold):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "SHOP_GOLD_NOT_ENOUGH",
			"message": "金币不足，无法购买。",
		})
		return true
	case errors.Is(err, core.ErrStaminaMaxLevelReached):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "STAMINA_MAX_LEVEL",
			"message": "体力上限已升到最高。",
		})
		return true
	case errors.Is(err, core.ErrStaminaAlreadyFull):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "STAMINA_ALREADY_FULL",
			"message": "当前体力已经是满的。",
		})
		return true
	case errors.Is(err, core.ErrAccountRiskBanned):
		writeJSON(c, consts.StatusLocked, map[string]string{
			"error":   "ACCOUNT_RISK_BANNED",
			"message": "账号风险过高，当前不可手点/挂机/购买体力。",
		})
		return true
	default:
		return false
	}
}
