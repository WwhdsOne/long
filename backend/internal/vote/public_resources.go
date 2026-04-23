package vote

import (
	"context"
	"strings"
)

// GetBossResources 返回当前 Boss 的低频公共资源。
func (s *Store) GetBossResources(ctx context.Context) (BossResources, error) {
	boss, err := s.currentBoss(ctx)
	if err != nil {
		return BossResources{}, err
	}
	if boss == nil {
		return BossResources{
			BossLoot:     []BossLootEntry{},
			BossHeroLoot: []BossHeroLootEntry{},
		}, nil
	}

	loot, err := s.loadBossLoot(ctx, boss.ID)
	if err != nil {
		return BossResources{}, err
	}
	heroLoot, err := s.loadBossHeroLoot(ctx, boss.ID)
	if err != nil {
		return BossResources{}, err
	}

	return BossResources{
		BossID:       boss.ID,
		TemplateID:   boss.TemplateID,
		Status:       boss.Status,
		BossLoot:     loot,
		BossHeroLoot: heroLoot,
	}, nil
}

// GetLatestAnnouncementVersion 返回最新生效公告的版本标记。
func (s *Store) GetLatestAnnouncementVersion(ctx context.Context) (string, error) {
	ids, err := s.client.ZRevRange(ctx, s.announcementKey, 0, -1).Result()
	if err != nil {
		return "", err
	}

	for _, id := range ids {
		values, err := s.client.HMGet(ctx, s.announcementItemKey(id), "id", "active").Result()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(stringValue(values, 1)) == "0" {
			continue
		}
		return firstNonEmpty(strings.TrimSpace(stringValue(values, 0)), strings.TrimSpace(id)), nil
	}

	return "", nil
}
