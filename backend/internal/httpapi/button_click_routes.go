package httpapi

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/ratelimit"
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
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			if options.PlayerAuthenticator != nil {
				return
			}
			writeJSON(c, consts.StatusBadRequest, map[string]string{
				"error":   "INVALID_NICKNAME",
				"message": "昵称还没填好，先起个名字再点。",
			})
			return
		}

		if err := enforceClickRateLimit(c, options.ClickGuard, nickname); err != nil {
			return
		}

		result, err := options.Store.ClickButton(ctx, c.Param("slug"), nickname)
		if err != nil {
			if errors.Is(err, vote.ErrButtonNotFound) {
				writeJSON(c, consts.StatusNotFound, map[string]string{"error": "BUTTON_NOT_FOUND"})
				return
			}
			if writeNicknameError(c, err) {
				return
			}

			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "INCREMENT_FAILED"})
			return
		}

		change := vote.StateChange{
			Type:      vote.StateChangeButtonClicked,
			Nickname:  nickname,
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

		writeJSON(c, consts.StatusOK, payload)
	})
}

func enforceClickRateLimit(c *app.RequestContext, guard ClickGuard, nickname string) error {
	if guard == nil {
		return nil
	}

	for _, key := range []string{
		"ip:" + clientIdentifier(c),
		"nickname:" + nickname,
	} {
		retryAfter, err := guard.Allow(key)
		if err == nil {
			continue
		}
		if errors.Is(err, ratelimit.ErrTooManyRequests) {
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter/time.Second), 10))
			writeJSON(c, consts.StatusTooManyRequests, map[string]string{
				"error":   "TOO_MANY_REQUESTS",
				"message": "点得太快了，先歇 10 分钟再来。",
			})
			return err
		}

		writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "RATE_LIMIT_FAILED"})
		return err
	}

	return nil
}
