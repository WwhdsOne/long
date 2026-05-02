package httpapi

import (
	"context"
	"errors"

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
	default:
		return false
	}
}
