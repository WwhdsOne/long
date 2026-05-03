package core

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"

	"long/internal/xlog"
)

// SaveBossToHistory 将 Boss 快照存入历史列表。
func (s *Store) SaveBossToHistory(ctx context.Context, boss *Boss) error {
	if boss == nil || strings.TrimSpace(boss.ID) == "" {
		return nil
	}

	values := map[string]any{
		"id":                    boss.ID,
		"name":                  boss.Name,
		"room_id":               boss.RoomID,
		"queue_id":              boss.QueueID,
		"status":                boss.Status,
		"max_hp":                strconv.FormatInt(boss.MaxHP, 10),
		"current_hp":            strconv.FormatInt(boss.CurrentHP, 10),
		"gold_on_kill":          strconv.FormatInt(maxInt64(0, boss.GoldOnKill), 10),
		"stone_on_kill":         strconv.FormatInt(maxInt64(0, boss.StoneOnKill), 10),
		"talent_points_on_kill": strconv.FormatInt(maxInt64(0, boss.TalentPointsOnKill), 10),
		"started_at":            strconv.FormatInt(boss.StartedAt, 10),
	}
	if strings.TrimSpace(boss.TemplateID) != "" {
		values["template_id"] = boss.TemplateID
	}
	if boss.DefeatedAt != 0 {
		values["defeated_at"] = strconv.FormatInt(boss.DefeatedAt, 10)
	}
	if len(boss.Parts) > 0 {
		partsRaw, _ := sonic.Marshal(boss.Parts)
		values["parts"] = string(partsRaw)
	}

	entry := BossHistoryEntry{
		Boss: Boss{
			ID:                 boss.ID,
			TemplateID:         boss.TemplateID,
			RoomID:             boss.RoomID,
			QueueID:            boss.QueueID,
			Name:               boss.Name,
			Status:             boss.Status,
			MaxHP:              boss.MaxHP,
			CurrentHP:          boss.CurrentHP,
			GoldOnKill:         maxInt64(0, boss.GoldOnKill),
			StoneOnKill:        maxInt64(0, boss.StoneOnKill),
			TalentPointsOnKill: maxInt64(0, boss.TalentPointsOnKill),
			Parts:              boss.Parts,
			StartedAt:          boss.StartedAt,
			DefeatedAt:         boss.DefeatedAt,
		},
	}
	loot, err := s.loadBossLoot(ctx, boss.ID)
	if err == nil {
		entry.Loot = loot
	}
	damage, err := s.ListBossLeaderboard(ctx, boss.ID, 20)
	if err == nil {
		entry.Damage = damage
	}

	if s.bossHistoryArchiver != nil {
		s.enqueueBossHistoryArchive(ctx, entry)
		return nil
	}

	key := s.bossHistoryPrefix + boss.ID
	if err := s.client.HSet(ctx, key, values).Err(); err != nil {
		return err
	}

	score := float64(boss.StartedAt)
	if score == 0 {
		score = float64(time.Now().Unix())
	}
	if err := s.client.ZAdd(ctx, s.bossHistoryKey, redis.Z{
		Score:  score,
		Member: boss.ID,
	}).Err(); err != nil {
		return err
	}

	return nil
}

// ListBossHistory 返回历史 Boss 列表（按时间倒序）。
func (s *Store) ListBossHistory(ctx context.Context) ([]BossHistoryEntry, error) {
	if s.bossHistoryStore != nil {
		return s.bossHistoryStore.ListBossHistory(ctx)
	}

	ids, err := s.client.ZRevRange(ctx, s.bossHistoryKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]BossHistoryEntry, 0, len(ids))
	for _, id := range ids {
		values, err := s.client.HGetAll(ctx, s.bossHistoryPrefix+id).Result()
		if err != nil || len(values) == 0 {
			continue
		}

		boss := normalizeBoss(values)
		if boss == nil {
			continue
		}

		loot, _ := s.loadBossLoot(ctx, id)
		damage, _ := s.ListBossLeaderboard(ctx, id, 20)

		entries = append(entries, BossHistoryEntry{
			Boss:   *boss,
			Loot:   loot,
			Damage: damage,
		})
	}

	return entries, nil
}

func (s *Store) enqueueBossHistoryArchive(ctx context.Context, entry BossHistoryEntry) {
	if s.bossHistoryArchiver == nil {
		return
	}
	_ = ctx
	if ok := s.bossHistoryArchiver.Enqueue(entry); !ok {
		xlog.L().Error("enqueue boss history archive failed")
	}
}
