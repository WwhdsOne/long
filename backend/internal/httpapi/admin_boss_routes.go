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

	"long/internal/core"
)

type flexibleInt64 int64

func (v *flexibleInt64) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*v = 0
		return nil
	}
	if strings.HasPrefix(raw, `"`) {
		unquoted, err := strconv.Unquote(raw)
		if err != nil {
			return err
		}
		raw = strings.TrimSpace(unquoted)
	}
	if raw == "" {
		*v = 0
		return nil
	}
	if strings.ContainsAny(raw, ".eE") {
		return errors.New("invalid int64")
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*v = flexibleInt64(parsed)
	return nil
}

type bossPartPayload struct {
	X           int           `json:"x"`
	Y           int           `json:"y"`
	Type        core.PartType `json:"type"`
	DisplayName string        `json:"displayName,omitempty"`
	ImagePath   string        `json:"imagePath,omitempty"`
	MaxHP       flexibleInt64 `json:"maxHp"`
	CurrentHP   flexibleInt64 `json:"currentHp"`
	Armor       flexibleInt64 `json:"armor"`
	Alive       bool          `json:"alive"`
}

func (p bossPartPayload) toCore() core.BossPart {
	return core.BossPart{
		X:           p.X,
		Y:           p.Y,
		Type:        p.Type,
		DisplayName: p.DisplayName,
		ImagePath:   p.ImagePath,
		MaxHP:       int64(p.MaxHP),
		CurrentHP:   int64(p.CurrentHP),
		Armor:       int64(p.Armor),
		Alive:       p.Alive,
	}
}

type bossUpsertPayload struct {
	ID                 string            `json:"id"`
	RoomID             string            `json:"roomId,omitempty"`
	Name               string            `json:"name"`
	MaxHP              flexibleInt64     `json:"maxHp"`
	GoldOnKill         flexibleInt64     `json:"goldOnKill"`
	StoneOnKill        flexibleInt64     `json:"stoneOnKill"`
	TalentPointsOnKill flexibleInt64     `json:"talentPointsOnKill"`
	Parts              []bossPartPayload `json:"parts,omitempty"`
}

func (p bossUpsertPayload) toCore() core.BossUpsert {
	parts := make([]core.BossPart, 0, len(p.Parts))
	for _, part := range p.Parts {
		parts = append(parts, part.toCore())
	}
	return core.BossUpsert{
		ID:                 p.ID,
		RoomID:             p.RoomID,
		Name:               p.Name,
		MaxHP:              int64(p.MaxHP),
		GoldOnKill:         int64(p.GoldOnKill),
		StoneOnKill:        int64(p.StoneOnKill),
		TalentPointsOnKill: int64(p.TalentPointsOnKill),
		Parts:              parts,
	}
}

type bossTemplateUpsertPayload struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	MaxHP              flexibleInt64     `json:"maxHp"`
	GoldOnKill         flexibleInt64     `json:"goldOnKill"`
	StoneOnKill        flexibleInt64     `json:"stoneOnKill"`
	TalentPointsOnKill flexibleInt64     `json:"talentPointsOnKill"`
	Layout             []bossPartPayload `json:"layout,omitempty"`
}

func (p bossTemplateUpsertPayload) toCore() core.BossTemplateUpsert {
	layout := make([]core.BossPart, 0, len(p.Layout))
	for _, part := range p.Layout {
		layout = append(layout, part.toCore())
	}
	return core.BossTemplateUpsert{
		ID:                 p.ID,
		Name:               p.Name,
		MaxHP:              int64(p.MaxHP),
		GoldOnKill:         int64(p.GoldOnKill),
		StoneOnKill:        int64(p.StoneOnKill),
		TalentPointsOnKill: int64(p.TalentPointsOnKill),
		Layout:             layout,
	}
}

func registerAdminBossRoutes(router route.IRouter, options Options) {
	router.POST("/api/admin/boss/activate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body bossUpsertPayload
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		upsert := body.toCore()

		roomID := strings.TrimSpace(firstNonEmptyString(upsert.RoomID, c.Query("roomId")))
		var boss *core.Boss
		var err error
		if store, ok := options.Store.(interface {
			ActivateBossInRoom(context.Context, string, core.BossUpsert) (*core.Boss, error)
		}); ok {
			boss, err = store.ActivateBossInRoom(ctx, roomID, upsert)
		} else {
			boss, err = options.Store.ActivateBoss(ctx, upsert)
		}
		if err != nil {
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 必须配置至少一个部位。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_ACTIVATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			RoomID:           boss.RoomID,
			QueueID:          boss.QueueID,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.activate",
			RoomID:      boss.RoomID,
			QueueID:     boss.QueueID,
			TargetType:  "boss",
			TargetID:    boss.ID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "boss.activated",
			BossID:    boss.ID,
			RoomID:    boss.RoomID,
			QueueID:   boss.QueueID,
			Payload: map[string]any{
				"name": boss.Name,
			},
		})
		writeJSON(c, consts.StatusOK, boss)
	})

	router.POST("/api/admin/boss/deactivate", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		roomID := strings.TrimSpace(c.Query("roomId"))
		var err error
		if store, ok := options.Store.(interface {
			DeactivateBossInRoom(context.Context, string) error
		}); ok {
			err = store.DeactivateBossInRoom(ctx, roomID)
		} else {
			err = options.Store.DeactivateBoss(ctx)
		}
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_DEACTIVATE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			RoomID:           roomID,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.deactivate",
			RoomID:      roomID,
			TargetType:  "boss",
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeDomainEvent(ctx, options.DomainEventWriter, core.DomainEvent{
			EventType: "boss.deactivated",
			RoomID:    roomID,
			Payload: map[string]any{
				"operator": options.AdminAuthenticator.Username(),
			},
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/loot", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			BossID string               `json:"bossId"`
			Loot   []core.BossLootEntry `json:"loot"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SetBossLoot(ctx, body.BossID, body.Loot); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_LOOT_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/boss/pool", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body bossTemplateUpsertPayload
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		upsert := body.toCore()

		if err := options.Store.SaveBossTemplate(ctx, upsert); err != nil {
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 模板必须配置至少一个部位。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
			return
		}

		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.template.create",
			TargetType:  "boss_template",
			TargetID:    upsert.ID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/pool/:templateId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body bossTemplateUpsertPayload
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		upsert := body.toCore()
		upsert.ID = c.Param("templateId")

		if err := options.Store.SaveBossTemplate(ctx, upsert); err != nil {
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 模板必须配置至少一个部位。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
			return
		}

		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.template.update",
			TargetType:  "boss_template",
			TargetID:    upsert.ID,
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.DELETE("/api/admin/boss/pool/:templateId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		if err := options.Store.DeleteBossTemplate(ctx, c.Param("templateId")); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_DELETE_FAILED"})
			return
		}
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.template.delete",
			TargetType:  "boss_template",
			TargetID:    c.Param("templateId"),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.PUT("/api/admin/boss/pool/:templateId/loot", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			Loot []core.BossLootEntry `json:"loot"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		if err := options.Store.SetBossTemplateLoot(ctx, c.Param("templateId"), body.Loot); err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_LOOT_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]bool{"ok": true})
	})

	router.POST("/api/admin/boss/cycle/enable", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		roomID := strings.TrimSpace(c.Query("roomId"))
		var boss *core.Boss
		var err error
		if store, ok := options.Store.(interface {
			SetBossCycleEnabledForRoom(context.Context, string, bool) (*core.Boss, error)
		}); ok {
			boss, err = store.SetBossCycleEnabledForRoom(ctx, roomID, true)
		} else {
			boss, err = options.Store.SetBossCycleEnabled(ctx, true)
		}
		if err != nil {
			if errors.Is(err, core.ErrBossPoolEmpty) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_POOL_EMPTY",
					"message": "Boss 池还是空的，先加模板再开启循环。",
				})
				return
			}
			if errors.Is(err, core.ErrBossCycleQueueEmpty) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_CYCLE_QUEUE_EMPTY",
					"message": "请先在 Boss 池里配置循环队列。",
				})
				return
			}
			if errors.Is(err, core.ErrBossPartsRequired) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_PARTS_REQUIRED",
					"message": "Boss 模板缺少部位，请先修正模板。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_ENABLE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			RoomID:           roomIDFromBossOrRequest(boss, roomID),
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.cycle.enable",
			RoomID:      roomIDFromBossOrRequest(boss, roomID),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, boss)
	})

	router.PUT("/api/admin/boss/cycle/queue", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			RoomID      string   `json:"roomId"`
			TemplateIDs []string `json:"templateIds"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		roomID := strings.TrimSpace(firstNonEmptyString(body.RoomID, c.Query("roomId")))
		var queue []string
		var err error
		if store, ok := options.Store.(interface {
			SetBossCycleQueueForRoom(context.Context, string, []string) ([]string, error)
		}); ok {
			queue, err = store.SetBossCycleQueueForRoom(ctx, roomID, body.TemplateIDs)
		} else {
			queue, err = options.Store.SetBossCycleQueue(ctx, body.TemplateIDs)
		}
		if err != nil {
			if errors.Is(err, core.ErrBossTemplateNotFound) {
				writeJSON(c, consts.StatusBadRequest, map[string]string{
					"error":   "BOSS_TEMPLATE_NOT_FOUND",
					"message": "循环队列里包含不存在的 Boss 模板。",
				})
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_QUEUE_SAVE_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, map[string]any{
			"ok":          true,
			"roomId":      roomID,
			"templateIds": queue,
		})
	})

	router.POST("/api/admin/boss/cycle/disable", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		roomID := strings.TrimSpace(c.Query("roomId"))
		var boss *core.Boss
		var err error
		if store, ok := options.Store.(interface {
			SetBossCycleEnabledForRoom(context.Context, string, bool) (*core.Boss, error)
		}); ok {
			boss, err = store.SetBossCycleEnabledForRoom(ctx, roomID, false)
		} else {
			boss, err = options.Store.SetBossCycleEnabled(ctx, false)
		}
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_DISABLE_FAILED"})
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			RoomID:           roomIDFromBossOrRequest(boss, roomID),
			BroadcastUserAll: true,
			Timestamp:        time.Now().Unix(),
		})
		writeAdminAudit(ctx, options.AdminAuditWriter, core.AdminAuditLog{
			Operator:    options.AdminAuthenticator.Username(),
			Action:      "boss.cycle.disable",
			RoomID:      roomIDFromBossOrRequest(boss, roomID),
			RequestPath: requestPath(c),
			RequestIP:   requestIP(c),
			Result:      "success",
		})
		writeJSON(c, consts.StatusOK, boss)
	})

	router.GET("/api/admin/boss/history", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		page, pageSize, ok := parseAdminPageParams(c)
		if !ok {
			return
		}

		var historyReader AdminBossHistoryReader = options.Store
		if options.AdminBossHistoryReader != nil {
			historyReader = options.AdminBossHistoryReader
		}

		history, err := historyReader.ListAdminBossHistoryPage(ctx, page, pageSize)
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
			return
		}

		writeJSON(c, consts.StatusOK, history)
	})
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func roomIDFromBossOrRequest(boss *core.Boss, fallback string) string {
	if boss != nil && strings.TrimSpace(boss.RoomID) != "" {
		return strings.TrimSpace(boss.RoomID)
	}
	return strings.TrimSpace(fallback)
}
