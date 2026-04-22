package vote

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ListBossTemplates 返回后台 Boss 池模板。
func (s *Store) ListBossTemplates(ctx context.Context) ([]BossTemplate, error) {
	templateIDs, err := s.client.SMembers(ctx, s.bossTemplateIndexKey).Result()
	if err != nil {
		return nil, err
	}
	if len(templateIDs) == 0 {
		return []BossTemplate{}, nil
	}

	templates := make([]BossTemplate, 0, len(templateIDs))
	for _, templateID := range templateIDs {
		templateID = strings.TrimSpace(templateID)
		if templateID == "" {
			continue
		}

		values, err := s.client.HGetAll(ctx, s.bossTemplateKey(templateID)).Result()
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			continue
		}

		loot, err := s.loadBossTemplateLoot(ctx, templateID)
		if err != nil {
			return nil, err
		}
		heroLoot, err := s.loadBossTemplateHeroLoot(ctx, templateID)
		if err != nil {
			return nil, err
		}

		templates = append(templates, BossTemplate{
			ID:       templateID,
			Name:     firstNonEmpty(strings.TrimSpace(values["name"]), templateID),
			MaxHP:    maxInt64(1, int64FromString(values["max_hp"])),
			Loot:     loot,
			HeroLoot: heroLoot,
		})
	}

	slices.SortFunc(templates, func(left, right BossTemplate) int {
		if left.Name == right.Name {
			return strings.Compare(left.ID, right.ID)
		}
		return strings.Compare(left.Name, right.Name)
	})

	return templates, nil
}

// SaveBossTemplate 保存或更新 Boss 池模板。
func (s *Store) SaveBossTemplate(ctx context.Context, template BossTemplateUpsert) error {
	templateID := strings.TrimSpace(template.ID)
	if templateID == "" {
		return ErrBossTemplateNotFound
	}

	values := map[string]any{
		"name":   firstNonEmpty(strings.TrimSpace(template.Name), templateID),
		"max_hp": strconv.FormatInt(maxInt64(1, template.MaxHP), 10),
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.bossTemplateKey(templateID), values)
	pipe.SAdd(ctx, s.bossTemplateIndexKey, templateID)
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteBossTemplate 删除 Boss 池模板。
func (s *Store) DeleteBossTemplate(ctx context.Context, templateID string) error {
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return nil
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.bossTemplateKey(templateID))
	pipe.Del(ctx, s.bossTemplateLootKey(templateID))
	pipe.SRem(ctx, s.bossTemplateIndexKey, templateID)
	_, err := pipe.Exec(ctx)
	return err
}

// SetBossTemplateLoot 保存 Boss 模板掉落池。
func (s *Store) SetBossTemplateLoot(ctx context.Context, templateID string, loot []BossLootEntry) error {
	return s.setLootEntries(ctx, s.bossTemplateLootKey(templateID), loot)
}

// SetBossCycleEnabled 设置 Boss 循环开关；开启时如果当前没有活动 Boss 会立即补位。
func (s *Store) SetBossCycleEnabled(ctx context.Context, enabled bool) (*Boss, error) {
	if enabled {
		templates, err := s.ListBossTemplates(ctx)
		if err != nil {
			return nil, err
		}
		if len(templates) == 0 {
			return nil, ErrBossPoolEmpty
		}
	}

	if err := s.client.HSet(ctx, s.bossCycleKey, "enabled", boolToRedis(enabled)).Err(); err != nil {
		return nil, err
	}

	current, err := s.currentBoss(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return current, nil
	}
	if current != nil && current.Status == bossStatusActive {
		return current, nil
	}
	if current != nil {
		_ = s.SaveBossToHistory(ctx, current)
	}

	return s.activateRandomBossFromPool(ctx)
}

func (s *Store) bossCycleEnabled(ctx context.Context) (bool, error) {
	value, err := s.client.HGet(ctx, s.bossCycleKey, "enabled").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return strings.TrimSpace(value) == "1", nil
}

func (s *Store) activateRandomBossFromPool(ctx context.Context) (*Boss, error) {
	templates, err := s.ListBossTemplates(ctx)
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, ErrBossPoolEmpty
	}

	index := 0
	if len(templates) > 1 {
		index = max(s.roll(len(templates)), 0)
		if index >= len(templates) {
			index = len(templates) - 1
		}
	}

	return s.activateBossTemplateInstance(ctx, templates[index])
}

func (s *Store) activateBossTemplateInstance(ctx context.Context, template BossTemplate) (*Boss, error) {
	instanceID, err := s.nextBossInstanceID(ctx, template.ID)
	if err != nil {
		return nil, err
	}

	current := &Boss{
		ID:         instanceID,
		TemplateID: template.ID,
		Name:       firstNonEmpty(strings.TrimSpace(template.Name), template.ID),
		Status:     bossStatusActive,
		MaxHP:      maxInt64(1, template.MaxHP),
		CurrentHP:  maxInt64(1, template.MaxHP),
		StartedAt:  time.Now().Unix(),
	}

	if err := s.setCurrentBoss(ctx, current, template.Loot, template.HeroLoot); err != nil {
		return nil, err
	}

	return current, nil
}

func (s *Store) nextBossInstanceID(ctx context.Context, templateID string) (string, error) {
	seq, err := s.client.Incr(ctx, s.bossInstanceSeqKey).Result()
	if err != nil {
		return "", err
	}

	baseID := strings.TrimSpace(templateID)
	if baseID == "" {
		baseID = "boss"
	}

	return baseID + "-" + strconv.FormatInt(seq, 10), nil
}

func (s *Store) setCurrentBoss(ctx context.Context, boss *Boss, loot []BossLootEntry, heroLoot []BossHeroLootEntry) error {
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

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.bossCurrentKey)
	pipe.HSet(ctx, s.bossCurrentKey, values)
	pipe.Del(ctx, s.bossLootKey(boss.ID))
	pipe.Del(ctx, s.bossHeroLootKey(boss.ID))

	entries := make([]redis.Z, 0, len(loot))
	for _, item := range loot {
		if strings.TrimSpace(item.ItemID) == "" || item.Weight <= 0 {
			continue
		}
		entries = append(entries, redis.Z{
			Score:  float64(item.Weight),
			Member: strings.TrimSpace(item.ItemID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, s.bossLootKey(boss.ID), entries...)
	}
	heroEntries := make([]redis.Z, 0, len(heroLoot))
	for _, item := range heroLoot {
		if strings.TrimSpace(item.HeroID) == "" || item.Weight <= 0 {
			continue
		}
		heroEntries = append(heroEntries, redis.Z{
			Score:  float64(item.Weight),
			Member: strings.TrimSpace(item.HeroID),
		})
	}
	if len(heroEntries) > 0 {
		pipe.ZAdd(ctx, s.bossHeroLootKey(boss.ID), heroEntries...)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) loadBossTemplateLoot(ctx context.Context, templateID string) ([]BossLootEntry, error) {
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return []BossLootEntry{}, nil
	}

	entries, err := s.client.ZRangeWithScores(ctx, s.bossTemplateLootKey(templateID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	loot := make([]BossLootEntry, 0, len(entries))
	for _, entry := range entries {
		itemID, ok := entry.Member.(string)
		if !ok || strings.TrimSpace(itemID) == "" {
			continue
		}

		definition, defErr := s.getEquipmentDefinition(ctx, itemID)
		if defErr != nil {
			loot = append(loot, BossLootEntry{
				ItemID: itemID,
				Weight: int64(entry.Score),
			})
			continue
		}

		loot = append(loot, BossLootEntry{
			ItemID:                     itemID,
			ItemName:                   definition.Name,
			Slot:                       definition.Slot,
			Rarity:                     normalizeEquipmentRarity(definition.Rarity),
			Weight:                     int64(entry.Score),
			EnhanceCap:                 definition.EnhanceCap,
			BonusClicks:                definition.BonusClicks,
			BonusCriticalChancePercent: definition.BonusCriticalChancePercent,
			BonusCriticalCount:         definition.BonusCriticalCount,
		})
	}

	return loot, nil
}

func (s *Store) setLootEntries(ctx context.Context, key string, loot []BossLootEntry) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	if len(loot) == 0 {
		return nil
	}

	entries := make([]redis.Z, 0, len(loot))
	for _, item := range loot {
		if strings.TrimSpace(item.ItemID) == "" || item.Weight <= 0 {
			continue
		}
		entries = append(entries, redis.Z{
			Score:  float64(item.Weight),
			Member: strings.TrimSpace(item.ItemID),
		})
	}

	if len(entries) == 0 {
		return nil
	}

	return s.client.ZAdd(ctx, key, entries...).Err()
}
