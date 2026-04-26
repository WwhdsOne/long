package httpapi

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
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
			"loadout":             state.Loadout,
			"combatStats":         state.CombatStats,
			"gold":                state.Gold,
			"stones":              state.Stones,
			"talentPoints":        state.TalentPoints,
			"recentRewards":       state.RecentRewards,
		}
		if resources, err := options.Store.GetBossResources(ctx); err == nil {
			payload["bossLoot"] = resources.BossLoot
			payload["bossGoldRange"] = resources.GoldRange
			payload["bossStoneRange"] = resources.StoneRange
			payload["bossTalentPointsOnKill"] = resources.TalentPointsOnKill
		}
		writeJSON(c, consts.StatusOK, payload)
	})

}
