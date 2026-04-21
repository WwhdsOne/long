package vote

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// ListHeroDefinitions 返回全部小小英雄模板。
func (s *Store) ListHeroDefinitions(ctx context.Context) ([]HeroDefinition, error) {
	heroIDs, err := s.client.SMembers(ctx, s.heroIndexKey).Result()
	if err != nil {
		return nil, err
	}
	if len(heroIDs) == 0 {
		return []HeroDefinition{}, nil
	}

	items := make([]HeroDefinition, 0, len(heroIDs))
	for _, heroID := range heroIDs {
		definition, err := s.getHeroDefinition(ctx, heroID)
		if err != nil {
			if errors.Is(err, ErrHeroNotFound) {
				continue
			}
			return nil, err
		}
		items = append(items, definition)
	}

	slices.SortFunc(items, func(left, right HeroDefinition) int {
		if left.Name == right.Name {
			return strings.Compare(left.HeroID, right.HeroID)
		}
		return strings.Compare(left.Name, right.Name)
	})

	return items, nil
}

// SaveHeroDefinition 保存或更新小小英雄模板。
func (s *Store) SaveHeroDefinition(ctx context.Context, definition HeroDefinition) error {
	heroID := strings.TrimSpace(definition.HeroID)
	if heroID == "" {
		return ErrHeroNotFound
	}

	values := map[string]any{
		"name":                          firstNonEmpty(strings.TrimSpace(definition.Name), heroID),
		"image_path":                    strings.TrimSpace(definition.ImagePath),
		"image_alt":                     strings.TrimSpace(definition.ImageAlt),
		"bonus_clicks":                  strconv.FormatInt(definition.BonusClicks, 10),
		"bonus_critical_chance_percent": strconv.Itoa(definition.BonusCriticalChancePercent),
		"bonus_critical_count":          strconv.FormatInt(definition.BonusCriticalCount, 10),
		"trait_type":                    strings.TrimSpace(string(definition.TraitType)),
		"trait_value":                   strconv.FormatInt(definition.TraitValue, 10),
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.heroKey(heroID), values)
	pipe.SAdd(ctx, s.heroIndexKey, heroID)
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteHeroDefinition 删除小小英雄模板。
func (s *Store) DeleteHeroDefinition(ctx context.Context, heroID string) error {
	heroID = strings.TrimSpace(heroID)
	if heroID == "" {
		return nil
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.heroKey(heroID))
	pipe.SRem(ctx, s.heroIndexKey, heroID)
	_, err := pipe.Exec(ctx)
	return err
}

// EquipHero 设置当前出战英雄。
func (s *Store) EquipHero(ctx context.Context, nickname string, heroID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	heroID = strings.TrimSpace(heroID)
	if heroID == "" {
		return State{}, ErrHeroNotFound
	}
	if _, err := s.getHeroDefinition(ctx, heroID); err != nil {
		return State{}, err
	}

	quantity, err := s.client.HGet(ctx, s.heroInventoryKey(normalizedNickname), heroID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrHeroNotOwned
		}
		return State{}, err
	}
	if quantity <= 0 {
		return State{}, ErrHeroNotOwned
	}

	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.activeHeroKey(normalizedNickname), heroID, 0)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(s.now().Unix()),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

// UnequipHero 卸下当前出战英雄。
func (s *Store) UnequipHero(ctx context.Context, nickname string, heroID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	heroID = strings.TrimSpace(heroID)
	if heroID == "" {
		return State{}, ErrHeroNotFound
	}

	currentHeroID, err := s.client.Get(ctx, s.activeHeroKey(normalizedNickname)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return State{}, err
	}
	if currentHeroID != "" && currentHeroID != heroID {
		return State{}, ErrHeroNotFound
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.activeHeroKey(normalizedNickname))
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(s.now().Unix()),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

func (s *Store) getHeroDefinition(ctx context.Context, heroID string) (HeroDefinition, error) {
	heroID = strings.TrimSpace(heroID)
	if heroID == "" {
		return HeroDefinition{}, ErrHeroNotFound
	}

	values, err := s.client.HGetAll(ctx, s.heroKey(heroID)).Result()
	if err != nil {
		return HeroDefinition{}, err
	}
	if len(values) == 0 {
		return HeroDefinition{}, ErrHeroNotFound
	}

	return HeroDefinition{
		HeroID:                     heroID,
		Name:                       firstNonEmpty(strings.TrimSpace(values["name"]), heroID),
		ImagePath:                  strings.TrimSpace(values["image_path"]),
		ImageAlt:                   strings.TrimSpace(values["image_alt"]),
		BonusClicks:                int64FromString(values["bonus_clicks"]),
		BonusCriticalChancePercent: int(int64FromString(values["bonus_critical_chance_percent"])),
		BonusCriticalCount:         int64FromString(values["bonus_critical_count"]),
		TraitType:                  HeroTraitType(strings.TrimSpace(values["trait_type"])),
		TraitValue:                 int64FromString(values["trait_value"]),
	}, nil
}

func (s *Store) heroInventoryQuantities(ctx context.Context, nickname string) (map[string]int64, error) {
	values, err := s.client.HGetAll(ctx, s.heroInventoryKey(nickname)).Result()
	if err != nil {
		return nil, err
	}

	quantities := make(map[string]int64, len(values))
	for heroID, rawQuantity := range values {
		quantity := int64FromString(rawQuantity)
		if quantity <= 0 {
			continue
		}
		quantities[heroID] = quantity
	}

	return quantities, nil
}

func (s *Store) activeHeroForNickname(ctx context.Context, nickname string, quantities map[string]int64) (*HeroInventoryItem, error) {
	heroID, err := s.client.Get(ctx, s.activeHeroKey(nickname)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if strings.TrimSpace(heroID) == "" || quantities[heroID] <= 0 {
		return nil, nil
	}

	definition, err := s.getHeroDefinition(ctx, heroID)
	if err != nil {
		if errors.Is(err, ErrHeroNotFound) {
			return nil, nil
		}
		return nil, err
	}

	item := buildHeroInventoryItem(definition, quantities[heroID], true)
	return &item, nil
}

func (s *Store) heroInventoryForNickname(ctx context.Context, _ string, quantities map[string]int64, activeHero *HeroInventoryItem) ([]HeroInventoryItem, error) {
	if len(quantities) == 0 {
		return []HeroInventoryItem{}, nil
	}

	items := make([]HeroInventoryItem, 0, len(quantities))
	for heroID, quantity := range quantities {
		definition, err := s.getHeroDefinition(ctx, heroID)
		if err != nil {
			if errors.Is(err, ErrHeroNotFound) {
				continue
			}
			return nil, err
		}
		items = append(items, buildHeroInventoryItem(definition, quantity, activeHero != nil && activeHero.HeroID == heroID))
	}

	slices.SortFunc(items, func(left, right HeroInventoryItem) int {
		if left.Name == right.Name {
			return strings.Compare(left.HeroID, right.HeroID)
		}
		return strings.Compare(left.Name, right.Name)
	})

	return items, nil
}

func buildHeroInventoryItem(definition HeroDefinition, quantity int64, active bool) HeroInventoryItem {
	return HeroInventoryItem{
		HeroID:                     definition.HeroID,
		Name:                       definition.Name,
		ImagePath:                  definition.ImagePath,
		ImageAlt:                   definition.ImageAlt,
		Quantity:                   quantity,
		Active:                     active,
		BonusClicks:                definition.BonusClicks,
		BonusCriticalChancePercent: definition.BonusCriticalChancePercent,
		BonusCriticalCount:         definition.BonusCriticalCount,
		TraitType:                  definition.TraitType,
		TraitValue:                 definition.TraitValue,
	}
}

func (s *Store) SetBossTemplateHeroLoot(ctx context.Context, templateID string, loot []BossHeroLootEntry) error {
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return nil
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.bossTemplateHeroLootKey(templateID))
	entries := make([]redis.Z, 0, len(loot))
	for _, item := range loot {
		if strings.TrimSpace(item.HeroID) == "" || item.Weight <= 0 {
			continue
		}
		entries = append(entries, redis.Z{
			Score:  float64(item.Weight),
			Member: strings.TrimSpace(item.HeroID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, s.bossTemplateHeroLootKey(templateID), entries...)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) loadBossTemplateHeroLoot(ctx context.Context, templateID string) ([]BossHeroLootEntry, error) {
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return []BossHeroLootEntry{}, nil
	}

	entries, err := s.client.ZRangeWithScores(ctx, s.bossTemplateHeroLootKey(templateID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return s.normalizeHeroLootEntries(ctx, entries)
}

func (s *Store) loadBossHeroLoot(ctx context.Context, bossID string) ([]BossHeroLootEntry, error) {
	bossID = strings.TrimSpace(bossID)
	if bossID == "" {
		return []BossHeroLootEntry{}, nil
	}

	entries, err := s.client.ZRangeWithScores(ctx, s.bossHeroLootKey(bossID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return s.normalizeHeroLootEntries(ctx, entries)
}

func (s *Store) normalizeHeroLootEntries(ctx context.Context, entries []redis.Z) ([]BossHeroLootEntry, error) {
	totalWeight := int64(0)
	for _, entry := range entries {
		if entry.Score > 0 {
			totalWeight += int64(entry.Score)
		}
	}

	loot := make([]BossHeroLootEntry, 0, len(entries))
	for _, entry := range entries {
		heroID, ok := entry.Member.(string)
		if !ok || strings.TrimSpace(heroID) == "" {
			continue
		}

		dropRatePercent := percentageFromWeight(int64(entry.Score), totalWeight)
		definition, err := s.getHeroDefinition(ctx, heroID)
		if err != nil {
			if errors.Is(err, ErrHeroNotFound) {
				loot = append(loot, BossHeroLootEntry{
					HeroID:          heroID,
					Weight:          int64(entry.Score),
					DropRatePercent: dropRatePercent,
				})
				continue
			}
			return nil, err
		}

		loot = append(loot, BossHeroLootEntry{
			HeroID:                     heroID,
			HeroName:                   definition.Name,
			ImagePath:                  definition.ImagePath,
			ImageAlt:                   definition.ImageAlt,
			Weight:                     int64(entry.Score),
			DropRatePercent:            dropRatePercent,
			BonusClicks:                definition.BonusClicks,
			BonusCriticalChancePercent: definition.BonusCriticalChancePercent,
			BonusCriticalCount:         definition.BonusCriticalCount,
			TraitType:                  definition.TraitType,
			TraitValue:                 definition.TraitValue,
		})
	}

	return loot, nil
}

func (s *Store) chooseHeroLoot(entries []BossHeroLootEntry) *BossHeroLootEntry {
	if len(entries) == 0 {
		return nil
	}

	totalWeight := 0
	for _, entry := range entries {
		if entry.Weight > 0 {
			totalWeight += int(entry.Weight)
		}
	}
	if totalWeight <= 0 {
		return nil
	}

	cursor := s.roll(totalWeight)
	running := 0
	for _, entry := range entries {
		if entry.Weight <= 0 {
			continue
		}
		running += int(entry.Weight)
		if cursor < running {
			selected := entry
			return &selected
		}
	}

	selected := entries[len(entries)-1]
	return &selected
}

func heroBonuses(hero *HeroInventoryItem) (int64, int, int64, int64) {
	if hero == nil {
		return 0, 0, 0, 0
	}

	bonusClicks := hero.BonusClicks
	bonusChance := hero.BonusCriticalChancePercent
	bonusCount := hero.BonusCriticalCount
	finalDamagePercent := int64(0)

	switch hero.TraitType {
	case HeroTraitBonusClicks:
		bonusClicks += hero.TraitValue
	case HeroTraitCriticalChancePercent:
		bonusChance += int(hero.TraitValue)
	case HeroTraitCriticalCountBonus:
		bonusCount += hero.TraitValue
	case HeroTraitFinalDamagePercent:
		finalDamagePercent += hero.TraitValue
	}

	return bonusClicks, bonusChance, bonusCount, finalDamagePercent
}

func applyFinalDamagePercent(stats CombatStats, percent int64) CombatStats {
	if percent == 0 {
		return stats
	}

	multiplier := 100 + percent
	if multiplier <= 0 {
		multiplier = 1
	}

	stats.NormalDamage = max(scaleDamage(stats.NormalDamage, multiplier), 1)
	stats.CriticalDamage = max(scaleDamage(stats.CriticalDamage, multiplier), stats.NormalDamage)
	return stats
}

func scaleDamage(value int64, multiplier int64) int64 {
	if value <= 0 {
		return 0
	}
	if multiplier <= 0 {
		return value
	}
	return (value*multiplier + 99) / 100
}
