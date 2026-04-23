package vote

import (
	"context"
	"testing"
)

func TestEquipmentDefinitionDefaultsRarityToCommon(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:wood-sword", map[string]any{
		"name":         "木剑",
		"slot":         "weapon",
		"bonus_clicks": "2",
		"enhance_cap":  "3",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.client.SAdd(ctx, "vote:equipment:index", "wood-sword").Err(); err != nil {
		t.Fatalf("seed equipment index: %v", err)
	}

	items, err := store.ListEquipmentDefinitions(ctx)
	if err != nil {
		t.Fatalf("list definitions: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(items))
	}
	if items[0].Rarity != "普通" {
		t.Fatalf("expected default rarity 普通, got %q", items[0].Rarity)
	}
}

func TestGetStateAndSnapshotExposeRarityAndGrowthCaps(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID:                     "fire-ring",
		Name:                       "🔥 炽焰戒",
		Slot:                       "accessory",
		Rarity:                     "至臻",
		BonusClicks:                2,
		BonusCriticalChancePercent: 6,
		BonusCriticalCount:         4,
		EnhanceCap:                 7,
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}
	if err := store.SaveHeroDefinition(ctx, HeroDefinition{
		HeroID:                     "spark-cat",
		Name:                       "星火猫",
		BonusClicks:                1,
		BonusCriticalChancePercent: 2,
		BonusCriticalCount:         3,
		AwakenCap:                  5,
	}); err != nil {
		t.Fatalf("save hero definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-inventory:阿明", map[string]any{
		"fire-ring": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.SetBossTemplateLoot(ctx, "dragon", []BossLootEntry{
		{ItemID: "fire-ring", Weight: 4},
	}); err != nil {
		t.Fatalf("set boss template loot: %v", err)
	}
	if err := store.SetBossTemplateHeroLoot(ctx, "dragon", []BossHeroLootEntry{
		{HeroID: "spark-cat", Weight: 2},
	}); err != nil {
		t.Fatalf("set boss template hero loot: %v", err)
	}
	if err := store.setCurrentBoss(ctx, &Boss{
		ID:        "dragon-1",
		Name:      "火龙",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
	}, []BossLootEntry{
		{ItemID: "fire-ring", Weight: 4},
	}, []BossHeroLootEntry{
		{HeroID: "spark-cat", Weight: 2},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if len(state.Inventory) != 1 {
		t.Fatalf("expected 1 inventory item, got %d", len(state.Inventory))
	}
	if state.Inventory[0].Rarity != "至臻" {
		t.Fatalf("expected inventory rarity 至臻, got %q", state.Inventory[0].Rarity)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 1 {
		t.Fatalf("expected 1 boss loot item, got %d", len(resources.BossLoot))
	}
	if resources.BossLoot[0].Rarity != "至臻" {
		t.Fatalf("expected boss loot rarity 至臻, got %q", resources.BossLoot[0].Rarity)
	}
	if resources.BossLoot[0].EnhanceCap != 7 {
		t.Fatalf("expected enhance cap 7, got %d", resources.BossLoot[0].EnhanceCap)
	}
	if len(resources.BossHeroLoot) != 1 {
		t.Fatalf("expected 1 boss hero loot item, got %d", len(resources.BossHeroLoot))
	}
	if resources.BossHeroLoot[0].AwakenCap != 5 {
		t.Fatalf("expected awaken cap 5, got %d", resources.BossHeroLoot[0].AwakenCap)
	}
}
