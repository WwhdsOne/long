package httpapi

import (
	"context"
	"errors"
	"strconv"
	"strings"
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
			Nickname string `json:"nickname"`
		}
		if !bindJSON(c, &body, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "昵称没有带上，先报个名再开点。",
		}) {
			return
		}
		if strings.TrimSpace(body.Nickname) == "" {
			writeJSON(c, consts.StatusBadRequest, map[string]string{
				"error":   "INVALID_NICKNAME",
				"message": "昵称还没填好，先起个名字再点。",
			})
			return
		}

		if options.ClickGuard != nil {
			retryAfter, err := options.ClickGuard.Allow(clientIdentifier(c))
			if err != nil {
				if errors.Is(err, ratelimit.ErrTooManyRequests) {
					c.Header("Retry-After", strconv.FormatInt(int64(retryAfter/time.Second), 10))
					writeJSON(c, consts.StatusTooManyRequests, map[string]string{
						"error":   "TOO_MANY_REQUESTS",
						"message": "点得太快了，先歇 10 分钟再来。",
					})
					return
				}

				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "RATE_LIMIT_FAILED"})
				return
			}
		}

		result, err := options.Store.ClickButton(ctx, c.Param("slug"), body.Nickname)
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
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		}
		if result.BroadcastUserAll {
			change.BroadcastUserAll = true
		}
		publishChange(ctx, options.ChangePublisher, change)
		writeJSON(c, consts.StatusOK, map[string]any{
			"button":          result.Button,
			"userStats":       result.UserStats,
			"delta":           result.Delta,
			"critical":        result.Critical,
			"boss":            result.Boss,
			"bossLeaderboard": result.BossLeaderboard,
			"myBossStats":     result.MyBossStats,
			"recentRewards":   result.RecentRewards,
			"lastReward":      result.LastReward,
		})
	})
}
