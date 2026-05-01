package httpapi

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerAdminPlayerAuthRoutes(router route.IRouter, options Options) {
	if options.PlayerAuthenticator == nil {
		return
	}

	router.POST("/api/admin/players/:nickname/password/reset", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			Password string `json:"password"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.PlayerAuthenticator.ResetPassword(ctx, c.Param("nickname"), body.Password); err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusBadRequest, map[string]string{
				"error":   "PLAYER_PASSWORD_RESET_FAILED",
				"message": "玩家密码重置失败。",
			})
			return
		}

		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "player.password.reset",
			TargetType:  "player",
			TargetID:    c.Param("nickname"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "player.password.reset",
			Nickname:  c.Param("nickname"),
			Payload: map[string]any{
				"operator": options.AdminAuthenticator.Username(),
			},
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})
}
