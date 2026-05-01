package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

func writeTaskAdminError(c *app.RequestContext, err error, fallbackCode string) {
	switch {
	case errors.Is(err, vote.ErrTaskNotFound):
		writeJSON(c, consts.StatusNotFound, map[string]string{
			"error":   "TASK_NOT_FOUND",
			"message": "任务不存在。",
		})
	case errors.Is(err, vote.ErrTaskImmutable):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   fallbackCode,
			"message": "生效中的任务核心规则不能直接修改，建议复制新任务后再上线。",
		})
	case errors.Is(err, vote.ErrTaskNotClaimable):
		writeJSON(c, consts.StatusBadRequest, map[string]string{
			"error":   fallbackCode,
			"message": "任务定义不合法，请检查目标值、奖励配置和限时任务的时间窗口。",
		})
	default:
		writeJSON(c, consts.StatusBadRequest, map[string]string{"error": fallbackCode})
	}
}

func registerAdminTaskRoutes(router route.IRouter, options Options) {
	router.GET("/api/admin/tasks", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		items, err := options.Store.ListTaskDefinitions(ctx)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "TASK_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, items)
	})

	router.POST("/api/admin/tasks", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		var body vote.TaskDefinition
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		if err := options.Store.SaveTaskDefinition(ctx, body); err != nil {
			writeTaskAdminError(c, err, "TASK_SAVE_FAILED")
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "task.save",
			TargetType:  "task",
			TargetID:    body.TaskID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "task.saved",
			Payload: map[string]any{
				"task_id":  body.TaskID,
				"operator": options.AdminAuthenticator.Username(),
			},
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/tasks/:taskId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		var body vote.TaskDefinition
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		body.TaskID = c.Param("taskId")
		if err := options.Store.SaveTaskDefinition(ctx, body); err != nil {
			writeTaskAdminError(c, err, "TASK_SAVE_FAILED")
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "task.save",
			TargetType:  "task",
			TargetID:    body.TaskID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, vote.DomainEvent{
			EventType: "task.saved",
			Payload: map[string]any{
				"task_id":  body.TaskID,
				"operator": options.AdminAuthenticator.Username(),
			},
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/tasks/:taskId/activate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		if err := options.Store.ActivateTaskDefinition(ctx, c.Param("taskId")); err != nil {
			writeTaskAdminError(c, err, "TASK_ACTIVATE_FAILED")
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "task.activate",
			TargetType:  "task",
			TargetID:    c.Param("taskId"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/tasks/:taskId/deactivate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		if err := options.Store.DeactivateTaskDefinition(ctx, c.Param("taskId")); err != nil {
			writeTaskAdminError(c, err, "TASK_DEACTIVATE_FAILED")
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "task.deactivate",
			TargetType:  "task",
			TargetID:    c.Param("taskId"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/tasks/:taskId/duplicate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		var body struct {
			TaskID string `json:"taskId"`
		}
		_ = bindJSON(c, &body, nil)
		item, err := options.Store.DuplicateTaskDefinition(ctx, c.Param("taskId"), body.TaskID)
		if err != nil {
			writeTaskAdminError(c, err, "TASK_DUPLICATE_FAILED")
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "task.duplicate",
			TargetType:  "task",
			TargetID:    c.Param("taskId"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, item)
	})

	router.POST("/api/admin/tasks/archive-expired", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		items, err := options.Store.ArchiveExpiredTaskCycles(ctx, time.Now())
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "TASK_ARCHIVE_FAILED"})
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, vote.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "task.archive_expired",
			TargetType:  "task_cycle",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, items)
	})

	router.GET("/api/admin/tasks/:taskId/cycles", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		items, err := options.Store.ListTaskCycleArchives(ctx, c.Param("taskId"))
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "TASK_CYCLE_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, items)
	})

	router.GET("/api/admin/tasks/:taskId/cycles/:cycleKey/results", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		item, err := options.Store.GetTaskCycleResults(ctx, c.Param("taskId"), c.Param("cycleKey"))
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "TASK_CYCLE_RESULTS_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, item)
	})
}
