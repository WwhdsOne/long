package httpapi

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerAdminRoutes(router route.IRouter, options Options) {
	if options.AdminAuthenticator == nil {
		return
	}

	registerAdminSessionRoutes(router, options)
	registerAdminOverviewRoutes(router, options)
	registerAdminBossRoutes(router, options)
	registerAdminContentRoutes(router, options)
	registerAdminPlayerAuthRoutes(router, options)
	registerAdminResourceRoutes(router, options)
}

func registerAdminSessionRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/login", func(_ context.Context, c *app.RequestContext) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		token, ok := options.AdminAuthenticator.Login(body.Username, body.Password)
		if !ok {
			writeAdminAudit(context.Background(), options.AdminAuditWriter, vote.AdminAuditLog{
				Operator:    strings.TrimSpace(body.Username),
				Action:      "admin.login",
				RequestPath: requestPath(c),
				RequestIP:   requestIP(c),
				Result:      "failed",
				ErrorCode:   "INVALID_CREDENTIALS",
			})
			writeJSON(c, consts.StatusUnauthorized, map[string]string{
				"error":   "INVALID_CREDENTIALS",
				"message": "账号或口令不对。",
			})
			return
		}

		c.SetCookie(adminSessionCookieName, token, 0, "/", "", protocol.CookieSameSiteLaxMode, false, true)
		writeAdminAudit(context.Background(), options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "admin.login",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/logout", func(_ context.Context, c *app.RequestContext) {
		c.SetCookie(adminSessionCookieName, "", -1, "/", "", protocol.CookieSameSiteLaxMode, false, true)
		writeAdminAudit(context.Background(), options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "admin.logout",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.GET("/api/admin/session", func(_ context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		writeJSON(c, consts.StatusOK, map[string]bool{"authenticated": true})
	})
}

func registerAdminOverviewRoutes(router route.IRouter, options Options) {
	router.GET("/api/admin/state", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		state, err := options.Store.GetAdminState(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ADMIN_STATE_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, state)
	})

	router.GET("/api/admin/equipment", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		page, pageSize, ok := parseAdminPageParams(c)
		if !ok {
			return
		}

		result, err := options.Store.ListAdminEquipmentPage(ctx, page, pageSize)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ADMIN_EQUIPMENT_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, result)
	})

	router.GET("/api/admin/players", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		limit := int64(50)
		if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
			parsedLimit, err := strconv.ParseInt(rawLimit, 10, 64)
			if err != nil {
				writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "INVALID_LIMIT"})
				return
			}
			limit = parsedLimit
		}

		page, err := options.Store.ListAdminPlayers(ctx, c.Query("cursor"), limit)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ADMIN_PLAYERS_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, page)
	})

	router.GET("/api/admin/players/:nickname", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		player, err := options.Store.GetAdminPlayer(ctx, c.Param("nickname"))
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ADMIN_PLAYER_FAILED"})
			return
		}
		if player == nil {
			writeJSON(c, consts.StatusNotFound, map[string]string{"error": "PLAYER_NOT_FOUND"})
			return
		}

		writeJSON(c, consts.StatusOK, player)
	})
}
