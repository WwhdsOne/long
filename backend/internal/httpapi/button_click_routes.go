package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"long/internal/vote"
)

func registerButtonClickRoutes(router route.IRouter, options Options) {
	router.POST("/api/click-tickets", func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		if !ok {
			return
		}
		if options.ManualClick == nil {
			writeJSON(c, 503, map[string]string{
				"error":   "CLICK_TICKET_UNAVAILABLE",
				"message": "点击票据服务暂不可用，请稍后重试。",
			})
			return
		}

		var body struct {
			Slug            string `json:"slug"`
			FingerprintHash string `json:"fingerprintHash"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "按钮标识不能为空。",
		}) {
			return
		}

		ticket, err := options.ManualClick.IssueTicket(ctx, TicketIssueRequest{
			Nickname:        nickname,
			Slug:            body.Slug,
			ClientID:        clientIdentifier(c),
			FingerprintHash: body.FingerprintHash,
		})
		if err != nil {
			if apiErr := manualClickRequestError(err); apiErr != nil {
				apiErr.writeTo(c)
				return
			}
			writeJSON(c, 500, map[string]string{
				"error":   "CLICK_TICKET_FAILED",
				"message": "点击票据签发失败，请稍后重试。",
			})
			return
		}

		writeJSON(c, 200, ticket)
	})

	router.POST("/api/buttons/:slug/click", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname          string `json:"nickname"`
			RealtimeConnected bool   `json:"realtimeConnected"`
			Ticket            string `json:"ticket"`
			PointerType       string `json:"pointerType"`
			PressDurationMS   int64  `json:"pressDurationMs"`
			FingerprintHash   string `json:"fingerprintHash"`
			FingerprintProof  string `json:"fingerprintProof"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再开点。",
		}) {
			return
		}
		nickname := authenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		resolvedNickname, result, apiErr := executeButtonClick(ctx, options, clickRequestContext{
			Slug:                  c.Param("slug"),
			NicknameHint:          body.Nickname,
			AuthenticatedNickname: nickname,
			AuthenticatorEnabled:  options.PlayerAuthenticator != nil,
			ClientID:              clientIdentifier(c),
			Ticket:                body.Ticket,
			EntryType:             clickEntryHTTP,
			FingerprintHash:       body.FingerprintHash,
			FingerprintProof:      body.FingerprintProof,
			Behavior: ClickBehavior{
				PointerType:     body.PointerType,
				PressDurationMS: body.PressDurationMS,
			},
		})
		if apiErr != nil {
			apiErr.writeTo(c)
			return
		}

		changeType := vote.StateChangeButtonClicked
		if result.BroadcastUserAll {
			changeType = vote.StateChangeBossChanged
		}
		change := vote.StateChange{
			Type:      changeType,
			Nickname:  resolvedNickname,
			Timestamp: time.Now().Unix(),
		}
		if result.BroadcastUserAll {
			change.BroadcastUserAll = true
		}
		publishChange(ctx, options.ChangePublisher, change)
		payload := map[string]any{
			"button":   result.Button,
			"delta":    result.Delta,
			"critical": result.Critical,
		}
		if !body.RealtimeConnected {
			payload["userStats"] = result.UserStats
			payload["boss"] = result.Boss
			payload["bossLeaderboard"] = result.BossLeaderboard
			payload["myBossStats"] = result.MyBossStats
			payload["recentRewards"] = result.RecentRewards
			payload["lastReward"] = result.LastReward
		}

		writeJSON(c, 200, payload)
	})
}
