package vote

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
)

// GetAdminState 返回后台页面所需的聚合数据。
func (s *Store) GetAdminState(ctx context.Context) (AdminState, error) {
	boss, err := s.currentBoss(ctx)
	if err != nil {
		return AdminState{}, err
	}

	bossCycleEnabled, err := s.bossCycleEnabled(ctx)
	if err != nil {
		return AdminState{}, err
	}
	bossPool, err := s.ListBossTemplates(ctx)
	if err != nil {
		return AdminState{}, err
	}
	bossCycleQueue, err := s.GetBossCycleQueue(ctx)
	if err != nil {
		return AdminState{}, err
	}

	bossLeaderboard := []BossLeaderboardEntry{}
	loot := []BossLootEntry{}
	if boss != nil {
		bossLeaderboard, err = s.ListBossLeaderboard(ctx, boss.ID, 20)
		if err != nil {
			return AdminState{}, err
		}

		loot, err = s.loadBossLoot(ctx, boss.ID)
		if err != nil {
			return AdminState{}, err
		}
	}

	playerCount, recentPlayerCount, err := s.playerCounts(ctx)
	if err != nil {
		return AdminState{}, err
	}

	return AdminState{
		Boss:              boss,
		BossLeaderboard:   bossLeaderboard,
		Equipment:         []EquipmentDefinition{},
		Loot:              loot,
		BossCycleEnabled:  bossCycleEnabled,
		BossPool:          bossPool,
		BossCycleQueue:    bossCycleQueue,
		PlayerCount:       playerCount,
		RecentPlayerCount: recentPlayerCount,
		Players:           []AdminPlayerOverview{},
	}, nil
}

// ListEquipmentDefinitions 列出全部装备模板。
func (s *Store) ListEquipmentDefinitions(ctx context.Context) ([]EquipmentDefinition, error) {
	itemIDs, err := s.listEquipmentIDs(ctx)
	if err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return []EquipmentDefinition{}, nil
	}

	equipment := make([]EquipmentDefinition, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		definition, err := s.getEquipmentDefinition(ctx, itemID)
		if err != nil {
			continue
		}
		equipment = append(equipment, definition)
	}

	slices.SortFunc(equipment, func(left, right EquipmentDefinition) int {
		if left.Slot == right.Slot {
			return strings.Compare(left.Name, right.Name)
		}
		return strings.Compare(left.Slot, right.Slot)
	})

	return equipment, nil
}

// SaveEquipmentDefinition 保存或更新装备模板。
func (s *Store) SaveEquipmentDefinition(ctx context.Context, definition EquipmentDefinition) error {
	itemID := strings.TrimSpace(definition.ItemID)
	if itemID == "" {
		return ErrEquipmentNotFound
	}

	values := map[string]any{
		"name":                   firstNonEmpty(strings.TrimSpace(definition.Name), itemID),
		"slot":                   normalizeEquipmentSlot(definition.Slot),
		"rarity":                 normalizeEquipmentRarity(definition.Rarity),
		"image_path":             strings.TrimSpace(definition.ImagePath),
		"image_alt":              strings.TrimSpace(definition.ImageAlt),
		"attack_power":           strconv.FormatInt(definition.AttackPower, 10),
		"armor_pen_percent":      strconv.FormatFloat(definition.ArmorPenPercent, 'f', -1, 64),
		"crit_rate":              strconv.FormatFloat(definition.CritRate, 'f', -1, 64),
		"crit_damage_multiplier": strconv.FormatFloat(definition.CritDamageMultiplier, 'f', -1, 64),
		"part_type_damage_soft":  strconv.FormatFloat(definition.PartTypeDamageSoft, 'f', -1, 64),
		"part_type_damage_heavy": strconv.FormatFloat(definition.PartTypeDamageHeavy, 'f', -1, 64),
		"part_type_damage_weak":  strconv.FormatFloat(definition.PartTypeDamageWeak, 'f', -1, 64),
		"talent_affinity":        strings.TrimSpace(definition.TalentAffinity),
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.equipmentKey(itemID), values)
	pipe.SAdd(ctx, s.equipmentIndexKey, itemID)
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteEquipmentDefinition 删除装备模板。
func (s *Store) DeleteEquipmentDefinition(ctx context.Context, itemID string) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return nil
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.equipmentKey(itemID))
	pipe.SRem(ctx, s.equipmentIndexKey, itemID)
	_, err := pipe.Exec(ctx)
	return err
}

// ActivateBoss 覆盖当前活动 Boss。
func (s *Store) ActivateBoss(ctx context.Context, boss BossUpsert) (*Boss, error) {
	// 保存旧 Boss 到历史
	if old, err := s.currentBoss(ctx); err == nil && old != nil {
		_ = s.SaveBossToHistory(ctx, old)
	}

	bossID := strings.TrimSpace(boss.ID)
	if bossID == "" {
		bossID = "boss-" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	parts := normalizeBossPartLayout(boss.Parts)
	if len(parts) == 0 {
		return nil, ErrBossPartsRequired
	}
	maxHP := maxInt64(1, boss.MaxHP)
	maxHP = sumBossPartMaxHP(parts)

	current := &Boss{
		ID:          bossID,
		Name:        firstNonEmpty(strings.TrimSpace(boss.Name), bossID),
		Status:      bossStatusActive,
		MaxHP:       maxHP,
		CurrentHP:   maxHP,
		GoldOnKill:  maxInt64(0, boss.GoldOnKill),
		StoneOnKill: maxInt64(0, boss.StoneOnKill),
		Parts:       parts,
		StartedAt:   time.Now().Unix(),
	}

	if err := s.setCurrentBoss(ctx, current, nil); err != nil {
		return nil, err
	}

	return current, nil
}

// DeactivateBoss 清空当前 Boss。
func (s *Store) DeactivateBoss(ctx context.Context) error {
	old, err := s.currentBoss(ctx)
	if err == nil && old != nil {
		_ = s.SaveBossToHistory(ctx, old)
	}
	enabled, cycleErr := s.bossCycleEnabled(ctx)
	if cycleErr != nil {
		return cycleErr
	}
	if enabled {
		nextTemplateID := ""
		if old != nil {
			nextTemplateID = old.TemplateID
		}
		if _, err := s.activateNextBossFromCycle(ctx, nextTemplateID); err != nil {
			if !errors.Is(err, ErrBossPoolEmpty) && !errors.Is(err, ErrBossCycleQueueEmpty) {
				return err
			}
			return s.client.Del(ctx, s.bossCurrentKey).Err()
		}
		return nil
	}
	return s.client.Del(ctx, s.bossCurrentKey).Err()
}

// SetBossLoot 覆盖指定 Boss 的掉落池。
func (s *Store) SetBossLoot(ctx context.Context, bossID string, loot []BossLootEntry) error {
	bossID = strings.TrimSpace(bossID)
	if bossID == "" {
		return nil
	}

	return s.setLootEntries(ctx, s.bossLootKey(bossID), loot)
}

// ListPlayerOverviews 列出后台玩家背包与穿戴概览。
func (s *Store) ListPlayerOverviews(ctx context.Context) ([]AdminPlayerOverview, error) {
	nicknames, err := s.listPlayerNicknames(ctx)
	if err != nil {
		return nil, err
	}
	if len(nicknames) == 0 {
		return []AdminPlayerOverview{}, nil
	}

	players := make([]AdminPlayerOverview, 0, len(nicknames))
	for _, nickname := range nicknames {
		if nickname == "" {
			continue
		}

		userStats, err := s.GetUserStats(ctx, nickname)
		if err != nil {
			if errors.Is(err, ErrInvalidNickname) || errors.Is(err, ErrSensitiveNickname) {
				continue
			}
			return nil, err
		}

		loadout, equipped, err := s.loadoutForNickname(ctx, nickname)
		if err != nil {
			return nil, err
		}
		inventory, err := s.inventoryForNickname(ctx, nickname, equipped)
		if err != nil {
			return nil, err
		}

		players = append(players, AdminPlayerOverview{
			Nickname:   nickname,
			ClickCount: userStats.ClickCount,
			Inventory:  inventory,
			Loadout:    loadout,
		})
	}

	slices.SortFunc(players, func(left, right AdminPlayerOverview) int {
		if left.ClickCount == right.ClickCount {
			return strings.Compare(left.Nickname, right.Nickname)
		}
		if left.ClickCount > right.ClickCount {
			return -1
		}
		return 1
	})

	return players, nil
}

// ListAdminPlayers 返回后台玩家分页概览。
func (s *Store) ListAdminPlayers(ctx context.Context, cursor string, limit int64) (AdminPlayerPage, error) {
	if limit <= 0 {
		limit = 50
	}

	total, _, err := s.playerCounts(ctx)
	if err != nil {
		return AdminPlayerPage{}, err
	}

	offset := int64(0)
	if trimmed := strings.TrimSpace(cursor); trimmed != "" {
		parsed, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return AdminPlayerPage{}, err
		}
		if parsed > 0 {
			offset = parsed
		}
	}

	nicknames, err := s.client.ZRevRange(ctx, s.playerIndexKey, offset, offset+limit-1).Result()
	if err != nil {
		return AdminPlayerPage{}, err
	}
	if len(nicknames) == 0 {
		return AdminPlayerPage{
			Items: []AdminPlayerOverview{},
			Total: total,
		}, nil
	}

	items := make([]AdminPlayerOverview, 0, len(nicknames))
	for _, nickname := range nicknames {
		player, err := s.adminPlayerOverview(ctx, nickname)
		if err != nil {
			return AdminPlayerPage{}, err
		}
		if player == nil {
			continue
		}
		items = append(items, *player)
	}

	page := AdminPlayerPage{
		Items: items,
		Total: total,
	}
	if offset+int64(len(nicknames)) < total {
		page.NextCursor = strconv.FormatInt(offset+int64(len(nicknames)), 10)
	}

	return page, nil
}

// GetAdminPlayer 返回单个玩家的后台详情。
func (s *Store) GetAdminPlayer(ctx context.Context, nickname string) (*AdminPlayerOverview, error) {
	return s.adminPlayerOverview(ctx, strings.TrimSpace(nickname))
}

func (s *Store) scanByPrefix(ctx context.Context, prefix string) ([]string, error) {
	var (
		cursor uint64
		keys   []string
	)

	for {
		foundKeys, nextCursor, err := s.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, foundKeys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (s *Store) adminPlayerOverview(ctx context.Context, nickname string) (*AdminPlayerOverview, error) {
	if strings.TrimSpace(nickname) == "" {
		return nil, nil
	}

	userStats, err := s.GetUserStats(ctx, nickname)
	if err != nil {
		if errors.Is(err, ErrInvalidNickname) || errors.Is(err, ErrSensitiveNickname) {
			return nil, nil
		}
		return nil, err
	}

	loadout, equipped, err := s.loadoutForNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}
	inventory, err := s.inventoryForNickname(ctx, nickname, equipped)
	if err != nil {
		return nil, err
	}

	return &AdminPlayerOverview{
		Nickname:   nickname,
		ClickCount: userStats.ClickCount,
		Inventory:  inventory,
		Loadout:    loadout,
	}, nil
}

func (s *Store) playerCounts(ctx context.Context) (int64, int64, error) {
	total, err := s.client.ZCard(ctx, s.playerIndexKey).Result()
	if err != nil {
		return 0, 0, err
	}

	now := time.Now().Unix()
	recent, err := s.client.ZCount(ctx, s.playerIndexKey, strconv.FormatInt(now-24*60*60, 10), "+inf").Result()
	if err != nil {
		return 0, 0, err
	}

	return total, recent, nil
}

func boolToRedis(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func maxInt64(left int64, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
