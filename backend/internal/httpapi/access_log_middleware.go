package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"long/internal/archive"
	"long/internal/xlog"
)

// AccessLogMiddleware 记录每个 HTTP 请求的 method、path、状态码、耗时等信息，
// 通过异步队列写入 MongoDB。
func AccessLogMiddleware(queue *archive.AsyncQueue[xlog.AccessLogEntry]) app.HandlerFunc {
	if queue == nil {
		return func(ctx context.Context, c *app.RequestContext) {
			c.Next(ctx)
		}
	}
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		queue.Enqueue(xlog.AccessLogEntry{
			Method:     string(c.Method()),
			Path:       string(c.Path()),
			StatusCode: c.Response.StatusCode(),
			LatencyMs:  time.Since(start).Milliseconds(),
			ClientIP:   requestIP(c),
			UserAgent:  string(c.Request.Header.UserAgent()),
			CreatedAt:  time.Now().Unix(),
		})
	}
}
