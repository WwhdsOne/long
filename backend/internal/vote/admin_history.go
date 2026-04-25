package vote

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

// SaveBossToHistory 将 Boss 快照存入历史列表。
func (s *Store) SaveBossToHistory(ctx context.Context, boss *Boss) error {
	if boss == nil || strings.TrimSpace(boss.ID) == "" {
		return nil
	}

	values := map[string]any{
		"id":         boss.ID,
		"name":       boss.Name,
		"status":     boss.Status,
		"max_hp":     strconv.FormatInt(boss.MaxHP, 10),
		"current_hp": strconv.FormatInt(boss.CurrentHP, 10),
		"started_at": strconv.FormatInt(boss.StartedAt, 10),
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

	key := s.bossHistoryPrefix + boss.ID
	if err := s.client.HSet(ctx, key, values).Err(); err != nil {
		return err
	}

	score := float64(boss.StartedAt)
	if score == 0 {
		score = float64(time.Now().Unix())
	}
	return s.client.ZAdd(ctx, s.bossHistoryKey, redis.Z{
		Score:  score,
		Member: boss.ID,
	}).Err()
}

// ListBossHistory 返回历史 Boss 列表（按时间倒序）。
func (s *Store) ListBossHistory(ctx context.Context) ([]BossHistoryEntry, error) {
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
