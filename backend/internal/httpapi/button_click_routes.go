package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"long/internal/vote"
)

func registerButtonClickRoutes(router route.IRouter, options Options) {
	router.POST("/api/buttons/:slug/click", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname          string `json:"nickname"`
			RealtimeConnected bool   `json:"realtimeConnected"`
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
		})
		if apiErr != nil {
			apiErr.writeTo(c)
			return
		}

		change := vote.StateChange{
			Type:      vote.StateChangeButtonClicked,
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
