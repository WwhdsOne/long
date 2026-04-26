package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"long/internal/ratelimit"
	"long/internal/vote"
)

type apiResponseError struct {
	Status     int
	Code       string
	Message    string
	RetryAfter time.Duration
}

type clickRequestContext struct {
	Slug                  string
	NicknameHint          string
	AuthenticatedNickname string
	AuthenticatorEnabled  bool
	ClientID              string
}

func (e *apiResponseError) writeTo(c *app.RequestContext) {
	if e == nil {
		return
	}
	if e.RetryAfter > 0 {
		c.Header("Retry-After", strconv.FormatInt(int64(e.RetryAfter/time.Second), 10))
	}
	payload := map[string]string{"error": e.Code}
	if strings.TrimSpace(e.Message) != "" {
		payload["message"] = e.Message
	}
	writeJSON(c, e.Status, payload)
}

func resolveRealtimeReadNickname(authenticatorEnabled bool, authenticatedNickname string, requestedNickname string) string {
	if authenticatorEnabled {
		return strings.TrimSpace(authenticatedNickname)
	}
	return strings.TrimSpace(requestedNickname)
}

func resolveClickNickname(request clickRequestContext) (string, *apiResponseError) {
	if request.AuthenticatorEnabled {
		nickname := strings.TrimSpace(request.AuthenticatedNickname)
		if nickname == "" {
			return "", &apiResponseError{
				Status:  http.StatusUnauthorized,
				Code:    "UNAUTHORIZED",
				Message: "请先登录账号再点。",
			}
		}
		return nickname, nil
	}

	nickname := strings.TrimSpace(request.NicknameHint)
	if nickname == "" {
		return "", &apiResponseError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_NICKNAME",
			Message: "昵称还没填好，先起个名字再点。",
		}
	}
	return nickname, nil
}

func enforceClickRateLimitForClient(guard ClickGuard, clientID string, nickname string) *apiResponseError {
	if guard == nil {
		return nil
	}

	keys := []string{
		"ip:" + strings.TrimSpace(clientID),
		"nickname:" + strings.TrimSpace(nickname),
	}
	for _, key := range keys {
		retryAfter, err := guard.Allow(key)
		if err == nil {
			continue
		}
		if errors.Is(err, ratelimit.ErrTooManyRequests) {
			return &apiResponseError{
				Status:     consts.StatusTooManyRequests,
				Code:       "TOO_MANY_REQUESTS",
				Message:    "点得太快了，先歇 10 分钟再来。",
				RetryAfter: retryAfter,
			}
		}
		return &apiResponseError{
			Status:  consts.StatusInternalServerError,
			Code:    "RATE_LIMIT_FAILED",
			Message: "限流检查失败，请稍后重试。",
		}
	}

	return nil
}

func clickRequestError(err error) *apiResponseError {
	switch {
	case errors.Is(err, vote.ErrBossPartNotFound):
		return &apiResponseError{
			Status:  consts.StatusNotFound,
			Code:    "BOSS_PART_NOT_FOUND",
			Message: "Boss 部位不存在或当前不可攻击。",
		}
	case errors.Is(err, vote.ErrBossPartAlreadyDead):
		return &apiResponseError{
			Status:  consts.StatusConflict,
			Code:    "BOSS_PART_ALREADY_DEAD",
			Message: "Boss 部位已被击碎，请选择其他部位。",
		}
	case errors.Is(err, vote.ErrInvalidNickname):
		return &apiResponseError{
			Status:  consts.StatusBadRequest,
			Code:    "INVALID_NICKNAME",
			Message: "昵称还没填好，先起个名字再点。",
		}
	case errors.Is(err, vote.ErrSensitiveNickname):
		return &apiResponseError{
			Status:  consts.StatusBadRequest,
			Code:    "SENSITIVE_NICKNAME",
			Message: "昵称包含敏感词，请换一个试试。",
		}
	default:
		return &apiResponseError{
			Status:  consts.StatusInternalServerError,
			Code:    "INCREMENT_FAILED",
			Message: "点击失败，请稍后重试。",
		}
	}
}

func executeButtonClick(ctx context.Context, options Options, request clickRequestContext) (string, vote.ClickResult, *apiResponseError) {
	nickname, apiErr := resolveClickNickname(request)
	if apiErr != nil {
		return "", vote.ClickResult{}, apiErr
	}

	if apiErr := enforceClickRateLimitForClient(options.ClickGuard, request.ClientID, nickname); apiErr != nil {
		return "", vote.ClickResult{}, apiErr
	}

	result, err := options.Store.ClickButton(ctx, request.Slug, nickname)
	if err != nil {
		return "", vote.ClickResult{}, clickRequestError(err)
	}

	return nickname, result, nil
}
