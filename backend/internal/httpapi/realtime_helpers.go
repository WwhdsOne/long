package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"long/internal/core"
)

type apiResponseError struct {
	Status  int
	Code    string
	Message string
}

type clickRequestContext struct {
	Slug                  string
	NicknameHint          string
	AuthenticatedNickname string
	AuthenticatorEnabled  bool
	ClientID              string
	ComboCount            int64
}

func (e *apiResponseError) writeTo(c *app.RequestContext) {
	if e == nil {
		return
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

func enforceClickRateLimitForClient(ctx context.Context, detector ClickRiskDetector, accountRisk AccountRiskManager, clientID string, nickname string) (bool, *apiResponseError) {
	if detector == nil || accountRisk == nil {
		return false, nil
	}

	type clickRiskCountDetector interface {
		DetectCount(string) (int, error)
	}

	nicknameHit := false
	keys := []string{
		"ip:" + strings.TrimSpace(clientID),
		"nickname:" + strings.TrimSpace(nickname),
	}
	for _, key := range keys {
		hitCount := 0
		var err error
		if countDetector, ok := detector.(clickRiskCountDetector); ok {
			hitCount, err = countDetector.DetectCount(key)
		} else {
			var hit bool
			hit, err = detector.Detect(key)
			if hit {
				hitCount = 1
			}
		}
		if err == nil && hitCount <= 0 {
			continue
		}
		if err == nil && hitCount > 0 {
			if strings.HasPrefix(key, "nickname:") {
				nicknameHit = true
				for range hitCount {
					if _, recordErr := accountRisk.RecordAccountRiskEvent(ctx, nickname, core.AccountRiskEventClickRateLimitHit); recordErr != nil {
						return false, &apiResponseError{
							Status:  consts.StatusInternalServerError,
							Code:    "ACCOUNT_RISK_FAILED",
							Message: "风险积分记录失败，请稍后重试。",
						}
					}
				}
			}
			continue
		}
		return false, &apiResponseError{
			Status:  consts.StatusInternalServerError,
			Code:    "RATE_LIMIT_DETECT_FAILED",
			Message: "点击异常检测失败，请稍后重试。",
		}
	}

	return nicknameHit, nil
}

func shouldSkipClickRateLimit(authenticatorEnabled bool, nickname string, nicknameWhitelist []string) bool {
	if !authenticatorEnabled || len(nicknameWhitelist) == 0 {
		return false
	}

	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return false
	}
	for _, allowedNickname := range nicknameWhitelist {
		if normalizedNickname == strings.TrimSpace(allowedNickname) {
			return true
		}
	}
	return false
}

func clickRequestError(err error) *apiResponseError {
	switch {
	case errors.Is(err, core.ErrBossPartNotFound):
		return &apiResponseError{
			Status:  consts.StatusNotFound,
			Code:    "BOSS_PART_NOT_FOUND",
			Message: "Boss 部位不存在或当前不可攻击。",
		}
	case errors.Is(err, core.ErrBossPartAlreadyDead):
		return &apiResponseError{
			Status:  consts.StatusConflict,
			Code:    "BOSS_PART_ALREADY_DEAD",
			Message: "Boss 部位已被击碎，请选择其他部位。",
		}
	case errors.Is(err, core.ErrInvalidNickname):
		return &apiResponseError{
			Status:  consts.StatusBadRequest,
			Code:    "INVALID_NICKNAME",
			Message: "昵称还没填好，先起个名字再点。",
		}
	case errors.Is(err, core.ErrSensitiveNickname):
		return &apiResponseError{
			Status:  consts.StatusBadRequest,
			Code:    "SENSITIVE_NICKNAME",
			Message: "昵称包含敏感词，请换一个试试。",
		}
	case errors.Is(err, core.ErrAccountRiskBanned):
		return &apiResponseError{
			Status:  consts.StatusLocked,
			Code:    "ACCOUNT_RISK_BANNED",
			Message: "账号风险过高，当前不可手点/挂机/购买体力。",
		}
	default:
		return &apiResponseError{
			Status:  consts.StatusInternalServerError,
			Code:    "INCREMENT_FAILED",
			Message: "点击失败，请稍后重试。",
		}
	}
}

func executeButtonClick(ctx context.Context, options Options, request clickRequestContext) (string, core.ClickResult, *apiResponseError) {
	nickname, apiErr := resolveClickNickname(request)
	if apiErr != nil {
		return "", core.ClickResult{}, apiErr
	}

	if !shouldSkipClickRateLimit(request.AuthenticatorEnabled, nickname, options.RateLimitNicknameWhitelist) {
		_, apiErr := enforceClickRateLimitForClient(ctx, options.ClickRiskDetector, options.AccountRisk, request.ClientID, nickname)
		if apiErr != nil {
			return "", core.ClickResult{}, apiErr
		}
	}

	result, err := options.Store.ClickButton(ctx, request.Slug, nickname, request.ComboCount)
	if err != nil {
		return "", core.ClickResult{}, clickRequestError(err)
	}

	return nickname, result, nil
}
