package vote

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const bossCycleQueueField = "queue"

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
		var layout []BossPart
		if layoutRaw, ok := values["layout"]; ok && layoutRaw != "" {
			_ = sonic.Unmarshal([]byte(layoutRaw), &layout)
		}

		templates = append(templates, BossTemplate{
			ID:                 templateID,
			Name:               firstNonEmpty(strings.TrimSpace(values["name"]), templateID),
			MaxHP:              maxInt64(1, int64FromString(values["max_hp"])),
			GoldOnKill:         maxInt64(0, int64FromString(values["gold_on_kill"])),
			StoneOnKill:        maxInt64(0, int64FromString(values["stone_on_kill"])),
			TalentPointsOnKill: maxInt64(0, int64FromString(values["talent_points_on_kill"])),
			Loot:               loot,
			Layout:             layout,
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
	layout := normalizeBossPartLayout(template.Layout)
	if len(layout) == 0 {
		return ErrBossPartsRequired
	}
	maxHP := maxInt64(1, template.MaxHP)
	maxHP = sumBossPartMaxHP(layout)

	values := map[string]any{
		"name":                  firstNonEmpty(strings.TrimSpace(template.Name), templateID),
		"max_hp":                strconv.FormatInt(maxHP, 10),
		"gold_on_kill":          strconv.FormatInt(maxInt64(0, template.GoldOnKill), 10),
		"stone_on_kill":         strconv.FormatInt(maxInt64(0, template.StoneOnKill), 10),
		"talent_points_on_kill": strconv.FormatInt(maxInt64(0, template.TalentPointsOnKill), 10),
	}
	if len(layout) > 0 {
		layoutRaw, _ := sonic.Marshal(layout)
		values["layout"] = string(layoutRaw)
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
	if err != nil {
		return err
	}

	queue, queueErr := s.GetBossCycleQueue(ctx)
	if queueErr != nil {
		return queueErr
	}
	nextQueue := make([]string, 0, len(queue))
	for _, id := range queue {
		if id == templateID {
			continue
		}
		nextQueue = append(nextQueue, id)
	}
	if len(nextQueue) == len(queue) {
		return nil
	}
	_, err = s.SetBossCycleQueue(ctx, nextQueue)
	return err
}

// SetBossTemplateLoot 保存 Boss 模板掉落池。
func (s *Store) SetBossTemplateLoot(ctx context.Context, templateID string, loot []BossLootEntry) error {
	return s.setLootEntries(ctx, s.bossTemplateLootKey(templateID), loot)
}

// SetBossCycleEnabled 设置 Boss 循环开关；开启时如果当前没有活动 Boss 会立即补位。
func (s *Store) SetBossCycleEnabled(ctx context.Context, enabled bool) (*Boss, error) {
	if enabled {
		if _, err := s.loadBossTemplateQueue(ctx); err != nil {
			return nil, err
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

	return s.activateNextBossFromCycle(ctx, "")
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

// SetBossCycleQueue 保存后台配置的 Boss 循环队列。
func (s *Store) SetBossCycleQueue(ctx context.Context, templateIDs []string) ([]string, error) {
	queue, err := s.normalizeBossCycleQueue(ctx, templateIDs)
	if err != nil {
		return nil, err
	}

	queueRaw, _ := sonic.Marshal(queue)
	if err := s.client.HSet(ctx, s.bossCycleKey, bossCycleQueueField, string(queueRaw)).Err(); err != nil {
		return nil, err
	}

	return queue, nil
}

// GetBossCycleQueue 返回后台配置的 Boss 循环队列。
func (s *Store) GetBossCycleQueue(ctx context.Context) ([]string, error) {
	queueRaw, err := s.client.HGet(ctx, s.bossCycleKey, bossCycleQueueField).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(queueRaw) == "" {
		return []string{}, nil
	}

	var queue []string
	if err := sonic.Unmarshal([]byte(queueRaw), &queue); err != nil {
		return []string{}, nil
	}
	cleaned := make([]string, 0, len(queue))
	seen := make(map[string]struct{}, len(queue))
	for _, templateID := range queue {
		templateID = strings.TrimSpace(templateID)
		if templateID == "" {
			continue
		}
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}
		cleaned = append(cleaned, templateID)
	}
	return cleaned, nil
}

func (s *Store) normalizeBossCycleQueue(ctx context.Context, templateIDs []string) ([]string, error) {
	queue := make([]string, 0, len(templateIDs))
	seen := make(map[string]struct{}, len(templateIDs))
	for _, templateID := range templateIDs {
		templateID = strings.TrimSpace(templateID)
		if templateID == "" {
			continue
		}
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}
		queue = append(queue, templateID)
	}

	templates, err := s.ListBossTemplates(ctx)
	if err != nil {
		return nil, err
	}
	templateSet := make(map[string]struct{}, len(templates))
	for _, template := range templates {
		templateSet[template.ID] = struct{}{}
	}
	for _, templateID := range queue {
		if _, ok := templateSet[templateID]; !ok {
			return nil, ErrBossTemplateNotFound
		}
	}

	return queue, nil
}

func (s *Store) loadBossTemplateQueue(ctx context.Context) ([]BossTemplate, error) {
	queueIDs, err := s.GetBossCycleQueue(ctx)
	if err != nil {
		return nil, err
	}
	if len(queueIDs) == 0 {
		return nil, ErrBossCycleQueueEmpty
	}

	templates, err := s.ListBossTemplates(ctx)
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, ErrBossPoolEmpty
	}

	templateByID := make(map[string]BossTemplate, len(templates))
	for _, template := range templates {
		templateByID[template.ID] = template
	}

	queue := make([]BossTemplate, 0, len(queueIDs))
	for _, templateID := range queueIDs {
		template, ok := templateByID[templateID]
		if !ok {
			continue
		}
		queue = append(queue, template)
	}
	if len(queue) == 0 {
		return nil, ErrBossCycleQueueEmpty
	}

	return queue, nil
}

func (s *Store) activateNextBossFromCycle(ctx context.Context, defeatedTemplateID string) (*Boss, error) {
	queue, err := s.loadBossTemplateQueue(ctx)
	if err != nil {
		return nil, err
	}

	nextIndex := 0
	defeatedTemplateID = strings.TrimSpace(defeatedTemplateID)
	if defeatedTemplateID != "" {
		for idx, template := range queue {
			if template.ID != defeatedTemplateID {
				continue
			}
			nextIndex = (idx + 1) % len(queue)
			break
		}
	}

	return s.activateBossTemplateInstance(ctx, queue[nextIndex])
}

func (s *Store) activateBossTemplateInstance(ctx context.Context, template BossTemplate) (*Boss, error) {
	instanceID, err := s.nextBossInstanceID(ctx, template.ID)
	if err != nil {
		return nil, err
	}
	parts := normalizeBossPartLayout(template.Layout)
	if len(parts) == 0 {
		return nil, ErrBossPartsRequired
	}
	maxHP := maxInt64(1, template.MaxHP)
	maxHP = sumBossPartMaxHP(parts)

	current := &Boss{
		ID:                 instanceID,
		TemplateID:         template.ID,
		Name:               firstNonEmpty(strings.TrimSpace(template.Name), template.ID),
		Status:             bossStatusActive,
		MaxHP:              maxHP,
		CurrentHP:          maxHP,
		GoldOnKill:         maxInt64(0, template.GoldOnKill),
		StoneOnKill:        maxInt64(0, template.StoneOnKill),
		TalentPointsOnKill: maxInt64(0, template.TalentPointsOnKill),
		Parts:              parts,
		StartedAt:          time.Now().Unix(),
	}

	if err := s.setCurrentBoss(ctx, current, template.Loot); err != nil {
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

func (s *Store) setCurrentBoss(ctx context.Context, boss *Boss, loot []BossLootEntry) error {
	if boss == nil || strings.TrimSpace(boss.ID) == "" {
		return nil
	}
	if len(boss.Parts) == 0 {
		return ErrBossPartsRequired
	}

	values := map[string]any{
		"id":                    boss.ID,
		"name":                  boss.Name,
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
	partsRaw, _ := sonic.Marshal(boss.Parts)
	values["parts"] = string(partsRaw)

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.bossCurrentKey)
	pipe.HSet(ctx, s.bossCurrentKey, values)
	pipe.Del(ctx, s.bossLootKey(boss.ID))

	entries := make([]redis.Z, 0, len(loot))
	for _, item := range loot {
		dropRatePercent := normalizeLootDropRate(item)
		if strings.TrimSpace(item.ItemID) == "" || dropRatePercent <= 0 {
			continue
		}
		entries = append(entries, redis.Z{
			Score:  dropRatePercent,
			Member: strings.TrimSpace(item.ItemID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, s.bossLootKey(boss.ID), entries...)
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
				ItemID:          itemID,
				DropRatePercent: clampFloat(entry.Score, 0, 100),
			})
			continue
		}

		loot = append(loot, BossLootEntry{
			ItemID:               itemID,
			ItemName:             definition.Name,
			Slot:                 definition.Slot,
			Rarity:               normalizeEquipmentRarity(definition.Rarity),
			DropRatePercent:      clampFloat(entry.Score, 0, 100),
			AttackPower:          definition.AttackPower,
			ArmorPenPercent:      definition.ArmorPenPercent,
			CritRate:             definition.CritRate,
			CritDamageMultiplier: definition.CritDamageMultiplier,
			PartTypeDamageSoft:   definition.PartTypeDamageSoft,
			PartTypeDamageHeavy:  definition.PartTypeDamageHeavy,
			PartTypeDamageWeak:   definition.PartTypeDamageWeak,
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
		dropRatePercent := normalizeLootDropRate(item)
		if strings.TrimSpace(item.ItemID) == "" || dropRatePercent <= 0 {
			continue
		}
		entries = append(entries, redis.Z{
			Score:  dropRatePercent,
			Member: strings.TrimSpace(item.ItemID),
		})
	}

	if len(entries) == 0 {
		return nil
	}

	return s.client.ZAdd(ctx, key, entries...).Err()
}
