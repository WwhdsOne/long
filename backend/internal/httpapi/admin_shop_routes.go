package httpapi

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

func registerAdminShopRoutes(router route.IRouter, options Options) {
	router.GET("/api/admin/shop/items", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		items, err := options.Store.ListShopItems(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ADMIN_SHOP_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, items)
	})

	router.POST("/api/admin/shop/items", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		var body core.ShopItem
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		if err := options.Store.SaveShopItem(ctx, body); err != nil {
			if writeShopError(c, err) {
				return
			}
			writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "ADMIN_SHOP_SAVE_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/shop/items/:itemId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		var body core.ShopItem
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.ItemID = c.Param("itemId")
		if err := options.Store.SaveShopItem(ctx, body); err != nil {
			if writeShopError(c, err) {
				return
			}
			writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "ADMIN_SHOP_SAVE_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.DELETE("/api/admin/shop/items/:itemId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		if err := options.Store.DeleteShopItem(ctx, c.Param("itemId")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ADMIN_SHOP_DELETE_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}
