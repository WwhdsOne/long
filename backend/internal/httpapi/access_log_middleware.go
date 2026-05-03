package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"long/internal/archive"
	"long/internal/xlog"
)

// AccessLogMiddleware 记录每个 HTTP 请求的 method、path、用户昵称、状态码、耗时，
// 输出到控制台，同时通过异步队列写入 MongoDB（含完整请求体用于事后排查）。
func AccessLogMiddleware(queue *archive.AsyncQueue[xlog.AccessLogEntry], authenticator PlayerAuthenticator) app.HandlerFunc {
	if queue == nil {
		return func(ctx context.Context, c *app.RequestContext) {
			start := time.Now()
			c.Next(ctx)
			nickname := authenticatedPlayerNickname(ctx, c, authenticator)
			hlog.CtxInfof(ctx, "%s %s %s %d %dms",
				nickname, string(c.Method()), string(c.Path()),
				c.Response.StatusCode(), time.Since(start).Milliseconds())
		}
	}
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		body := peekRequestBody(c)
		c.Next(ctx)
		latencyMs := time.Since(start).Milliseconds()
		statusCode := c.Response.StatusCode()
		nickname := authenticatedPlayerNickname(ctx, c, authenticator)
		hlog.CtxInfof(ctx, "%s %s %s %d %dms",
			nickname, string(c.Method()), string(c.Path()), statusCode, latencyMs)
		queue.Enqueue(xlog.AccessLogEntry{
			Method:     string(c.Method()),
			Path:       string(c.Path()),
			Nickname:   nickname,
			Body:       body,
			StatusCode: statusCode,
			LatencyMs:  latencyMs,
			ClientIP:   requestIP(c),
			UserAgent:  string(c.Request.Header.UserAgent()),
			CreatedAt:  time.Now().Unix(),
		})
	}
}

// peekRequestBody 读取请求体用于 MongoDB 日志落库。
func peekRequestBody(c *app.RequestContext) string {
	data := c.Request.Body()
	if len(data) == 0 {
		return ""
	}
	if len(data) > 1024 {
		return string(data[:1024])
	}
	return string(data)
}
