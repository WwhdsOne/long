package httpapi

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	adminauth "long/internal/admin"
	"long/internal/vote"
)

func writeJSON(c *app.RequestContext, status int, payload any) {
	c.JSON(status, payload)
}

func bindJSON(c *app.RequestContext, body any, invalidPayload map[string]string) bool {
	if err := c.BindJSON(body); err != nil {
		writeJSON(c, consts.StatusBadRequest, invalidPayload)
		return false
	}
	return true
}

func writeNicknameError(c *app.RequestContext, err error) bool {
	if errors.Is(err, vote.ErrInvalidNickname) {
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "INVALID_NICKNAME",
			"message": "昵称还没填好，先起个名字再点。",
		})
		return true
	}

	if errors.Is(err, vote.ErrSensitiveNickname) {
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "SENSITIVE_NICKNAME",
			"message": "昵称包含敏感词，请换一个试试。",
		})
		return true
	}

	return false
}

func writeContentError(c *app.RequestContext, err error) bool {
	switch {
	case errors.Is(err, vote.ErrSensitiveContent):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "SENSITIVE_CONTENT",
			"message": "内容包含敏感词，请改一下再发。",
		})
		return true
	case errors.Is(err, vote.ErrMessageEmpty):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "EMPTY_CONTENT",
			"message": "内容不能为空。",
		})
		return true
	case errors.Is(err, vote.ErrMessageTooLong):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   "CONTENT_TOO_LONG",
			"message": "内容最多 200 个字。",
		})
		return true
	}

	return false
}

func clientIdentifier(c *app.RequestContext) string {
	return c.ClientIP()
}

func isAdminAuthenticated(c *app.RequestContext, authenticator *adminauth.Authenticator) bool {
	if authenticator == nil {
		return false
	}

	token := strings.TrimSpace(string(c.Cookie(adminSessionCookieName)))
	if token == "" {
		return false
	}

	return authenticator.Verify(token)
}

func publishChange(ctx context.Context, publisher ChangePublisher, change vote.StateChange) {
	if publisher == nil {
		return
	}
	_ = publisher.PublishChange(ctx, change)
}

func parseAdminPageParams(c *app.RequestContext) (int64, int64, bool) {
	page := int64(1)
	pageSize := int64(20)

	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		parsedPage, err := strconv.ParseInt(rawPage, 10, 64)
		if err != nil {
			writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "INVALID_PAGE"})
			return 0, 0, false
		}
		page = parsedPage
	}

	if rawPageSize := strings.TrimSpace(c.Query("pageSize")); rawPageSize != "" {
		parsedPageSize, err := strconv.ParseInt(rawPageSize, 10, 64)
		if err != nil {
			writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "INVALID_PAGE_SIZE"})
			return 0, 0, false
		}
		pageSize = parsedPageSize
	}

	return page, pageSize, true
}
