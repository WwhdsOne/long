package httpapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func registerBattleRoutes(router route.IRouter, options Options) {
	router.GET("/api/battle/state", func(ctx context.Context, c *app.RequestContext) {
		nickname := resolvedPlayerNicknameForRead(ctx, c, options.PlayerAuthenticator)
		state, err := options.Store.GetState(ctx, nickname)
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		payload := map[string]any{
			"totalVotes":          state.TotalVotes,
			"leaderboard":         state.Leaderboard,
			"boss":                state.Boss,
			"bossLeaderboard":     state.BossLeaderboard,
			"bossLoot":            state.BossLoot,
			"announcementVersion": state.AnnouncementVersion,
			"latestAnnouncement":  state.LatestAnnouncement,
			"userStats":           state.UserStats,
			"myBossStats":         state.MyBossStats,
			"inventory":           state.Inventory,
			"loadout":             state.Loadout,
			"combatStats":         state.CombatStats,
			"gems":                state.Gems,
			"gold":                state.Gold,
			"stones":              state.Stones,
			"recentRewards":       state.RecentRewards,
			"lastReward":          state.LastReward,
		}
		if resources, err := options.Store.GetBossResources(ctx); err == nil {
			payload["bossLoot"] = resources.BossLoot
		}
		writeJSON(c, consts.StatusOK, payload)
	})

	router.POST("/api/boss/parts/:x/:y/attack", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname          string `json:"nickname"`
			RealtimeConnected bool   `json:"realtimeConnected"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再开打。",
		}) {
			return
		}

		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}

		if apiErr := enforceClickRateLimitForClient(options.ClickGuard, clientIdentifier(c), nickname); apiErr != nil {
			apiErr.writeTo(c)
			return
		}

		x, xErr := strconv.Atoi(strings.TrimSpace(c.Param("x")))
		y, yErr := strconv.Atoi(strings.TrimSpace(c.Param("y")))
		if xErr != nil || yErr != nil {
			writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "INVALID_PART_COORDINATE"})
			return
		}

		result, err := options.Store.ClickBossPart(ctx, fmt.Sprintf("%d-%d", x, y), nickname)
		if err != nil {
			clickRequestError(err).writeTo(c)
			return
		}

		change := vote.StateChange{
			Type:      vote.StateChangeBossChanged,
			Nickname:  nickname,
			Timestamp: time.Now().Unix(),
		}
		if result.BroadcastUserAll {
			change.BroadcastUserAll = true
		}
		publishChange(ctx, options.ChangePublisher, change)

		payload := map[string]any{
			"delta":           result.Delta,
			"bossDamage":      result.BossDamage,
			"critical":        result.Critical,
			"userStats":       result.UserStats,
			"boss":            result.Boss,
			"bossLeaderboard": result.BossLeaderboard,
			"myBossStats":     result.MyBossStats,
			"recentRewards":   result.RecentRewards,
			"lastReward":      result.LastReward,
		}
		if !body.RealtimeConnected {
			payload["userStats"] = result.UserStats
		}

		writeJSON(c, consts.StatusOK, payload)
	})
}
