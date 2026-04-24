package vote

import (
	"context"
	"errors"
	"testing"
)

func seedEquipmentDefinition(t *testing.T, store *Store, definition EquipmentDefinition) {
	t.Helper()

	if err := store.SaveEquipmentDefinition(context.Background(), definition); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}
}

func seedHeroDefinition(t *testing.T, store *Store, definition HeroDefinition) {
	t.Helper()

	if err := store.SaveHeroDefinition(context.Background(), definition); err != nil {
		t.Fatalf("save hero definition: %v", err)
	}
}

func TestSalvageEquipmentAddsGemsAndKeepsEquippedCopy(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedEquipmentDefinition(t, store, EquipmentDefinition{
		ItemID:      "wood-sword",
		Name:        "木剑",
		Slot:        "weapon",
		BonusClicks: 2,
	})
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), map[string]any{
		"wood-sword": "3",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, store.loadoutKey("阿明"), map[string]any{
		"weapon": "wood-sword",
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}

	state, err := store.SalvageEquipment(ctx, "阿明", "wood-sword", 2)
	if err != nil {
		t.Fatalf("salvage equipment: %v", err)
	}

	if state.Gems != 2 {
		t.Fatalf("expected gems to become 2, got %d", state.Gems)
	}
	if len(state.Inventory) != 1 || state.Inventory[0].Quantity != 1 {
		t.Fatalf("expected to keep 1 equipped copy, got %+v", state.Inventory)
	}
	if state.Loadout.Weapon == nil || state.Loadout.Weapon.Quantity != 1 {
		t.Fatalf("expected equipped weapon to remain, got %+v", state.Loadout.Weapon)
	}
	if state.LastForgeResult == nil || state.LastForgeResult.Kind != "equipment_salvage" {
		t.Fatalf("expected salvage result payload, got %+v", state.LastForgeResult)
	}
	if state.LastForgeResult.GemsDelta != 2 || state.LastForgeResult.RemainingGems != 2 {
		t.Fatalf("unexpected salvage result payload: %+v", state.LastForgeResult)
	}
}

func TestSalvageHeroAddsGemsAndKeepsActiveCopy(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedHeroDefinition(t, store, HeroDefinition{
		HeroID:      "spark-cat",
		Name:        "星火猫",
		BonusClicks: 2,
	})
	if err := store.client.HSet(ctx, store.heroInventoryKey("阿明"), map[string]any{
		"spark-cat": "3",
	}).Err(); err != nil {
		t.Fatalf("seed hero inventory: %v", err)
	}
	if err := store.client.Set(ctx, store.activeHeroKey("阿明"), "spark-cat", 0).Err(); err != nil {
		t.Fatalf("seed active hero: %v", err)
	}

	state, err := store.SalvageHero(ctx, "阿明", "spark-cat", 2)
	if err != nil {
		t.Fatalf("salvage hero: %v", err)
	}

	if state.Gems != 2 {
		t.Fatalf("expected gems to become 2, got %d", state.Gems)
	}
	if len(state.Heroes) != 1 || state.Heroes[0].Quantity != 1 || !state.Heroes[0].Active {
		t.Fatalf("expected to keep 1 active hero copy, got %+v", state.Heroes)
	}
	if state.ActiveHero == nil || state.ActiveHero.Quantity != 1 {
		t.Fatalf("expected active hero to remain, got %+v", state.ActiveHero)
	}
	if state.LastForgeResult == nil || state.LastForgeResult.Kind != "hero_salvage" {
		t.Fatalf("expected hero salvage result payload, got %+v", state.LastForgeResult)
	}
}

func TestEnhanceEquipmentConsumesGemsAndAppliesSingleStatGrowth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	seedEquipmentDefinition(t, store, EquipmentDefinition{
		ItemID:      "wood-sword",
		Name:        "木剑",
		Slot:        "weapon",
		BonusClicks: 2,
		EnhanceCap:  3,
	})
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), map[string]any{
		"wood-sword": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.setGems(ctx, "阿明", equipmentEnhanceCost); err != nil {
		t.Fatalf("seed gems: %v", err)
	}

	state, err := store.EnhanceEquipment(ctx, "阿明", "wood-sword")
	if err != nil {
		t.Fatalf("enhance equipment: %v", err)
	}

	if state.Gems != 0 {
		t.Fatalf("expected gems to become 0, got %d", state.Gems)
	}
	if len(state.Inventory) != 1 || state.Inventory[0].BonusClicks != 3 {
		t.Fatalf("expected click stat to grow by 1, got %+v", state.Inventory)
	}
	if state.Inventory[0].EnhanceLevel != 1 || state.Inventory[0].BonusClicksDelta != 1 {
		t.Fatalf("expected enhance level and delta to update, got %+v", state.Inventory[0])
	}
	if state.LastForgeResult == nil || state.LastForgeResult.Kind != "equipment_enhance" || state.LastForgeResult.Jackpot {
		t.Fatalf("expected normal enhance result, got %+v", state.LastForgeResult)
	}
	upgrade, err := store.getEquipmentUpgrade(ctx, "阿明", "wood-sword")
	if err != nil {
		t.Fatalf("load equipment upgrade: %v", err)
	}
	if upgrade.EnhanceLevel != 1 || upgrade.BonusClicks != 1 {
		t.Fatalf("expected stored enhance delta, got %+v", upgrade)
	}
}

func TestEnhanceEquipmentAppliesFixedCriticalChanceGrowth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 2 }

	ctx := context.Background()
	seedEquipmentDefinition(t, store, EquipmentDefinition{
		ItemID:      "wood-sword",
		Name:        "木剑",
		Slot:        "weapon",
		BonusClicks: 2,
		EnhanceCap:  3,
	})
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), map[string]any{
		"wood-sword": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.setGems(ctx, "阿明", equipmentEnhanceCost); err != nil {
		t.Fatalf("seed gems: %v", err)
	}

	state, err := store.EnhanceEquipment(ctx, "阿明", "wood-sword")
	if err != nil {
		t.Fatalf("enhance equipment: %v", err)
	}

	item := state.Inventory[0]
	if item.BonusCriticalChancePercent != 0.2 || item.BonusCriticalChancePercentDelta != 0.2 {
		t.Fatalf("expected fixed crit chance growth, got %+v", item)
	}
	if state.LastForgeResult == nil || state.LastForgeResult.RewardSummary != "暴击率 +0.20%" {
		t.Fatalf("expected fixed crit chance summary, got %+v", state.LastForgeResult)
	}
}

func TestEnhanceEquipmentRejectsAtTemplateCap(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedEquipmentDefinition(t, store, EquipmentDefinition{
		ItemID:      "wood-sword",
		Name:        "木剑",
		Slot:        "weapon",
		BonusClicks: 2,
		EnhanceCap:  1,
	})
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), map[string]any{
		"wood-sword": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, store.upgradeKey("阿明", "wood-sword"), map[string]any{
		"enhance_level": "1",
		"bonus_clicks":  "1",
	}).Err(); err != nil {
		t.Fatalf("seed upgrade: %v", err)
	}
	if err := store.setGems(ctx, "阿明", equipmentEnhanceCost); err != nil {
		t.Fatalf("seed gems: %v", err)
	}

	if _, err := store.EnhanceEquipment(ctx, "阿明", "wood-sword"); !errors.Is(err, ErrEquipmentMaxEnhance) {
		t.Fatalf("expected enhance cap error, got %v", err)
	}
}

func TestAwakenHeroConsumesGemsAndFeedsCombatStats(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	rolls := []int{0, 99}
	store.roll = func(limit int) int {
		next := rolls[0]
		rolls = rolls[1:]
		if next >= limit {
			return limit - 1
		}
		return next
	}

	ctx := context.Background()
	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "feel",
		Label:   "有感觉吗",
		Sort:    10,
		Enabled: true,
	}); err != nil {
		t.Fatalf("save button: %v", err)
	}
	seedHeroDefinition(t, store, HeroDefinition{
		HeroID:      "spark-cat",
		Name:        "星火猫",
		BonusClicks: 2,
		AwakenCap:   3,
	})
	if err := store.client.HSet(ctx, store.heroInventoryKey("阿明"), map[string]any{
		"spark-cat": "1",
	}).Err(); err != nil {
		t.Fatalf("seed hero inventory: %v", err)
	}
	if err := store.client.Set(ctx, store.activeHeroKey("阿明"), "spark-cat", 0).Err(); err != nil {
		t.Fatalf("seed active hero: %v", err)
	}
	if err := store.setGems(ctx, "阿明", heroAwakenCost); err != nil {
		t.Fatalf("seed gems: %v", err)
	}
	if err := store.client.HSet(ctx, store.bossCurrentKey, map[string]any{
		"id":         "slime-king",
		"name":       "史莱姆王",
		"status":     bossStatusActive,
		"max_hp":     "10",
		"current_hp": "10",
	}).Err(); err != nil {
		t.Fatalf("seed boss: %v", err)
	}

	state, err := store.AwakenHero(ctx, "阿明", "spark-cat")
	if err != nil {
		t.Fatalf("awaken hero: %v", err)
	}

	if state.Gems != 0 {
		t.Fatalf("expected gems to become 0, got %d", state.Gems)
	}
	if state.ActiveHero == nil || state.ActiveHero.BonusClicks != 3 || state.ActiveHero.AwakenLevel != 1 {
		t.Fatalf("expected awakened hero stats, got %+v", state.ActiveHero)
	}
	if len(state.ActiveHero.Effects) != 0 {
		t.Fatalf("expected awaken to affect only base stats, got %+v", state.ActiveHero.Effects)
	}
	if state.CombatStats.NormalDamage != 4 {
		t.Fatalf("expected awakened hero to raise normal damage to 4, got %+v", state.CombatStats)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click after awaken: %v", err)
	}
	if result.Delta != 4 {
		t.Fatalf("expected awakened hero bonus to affect boss damage, got %+v", result)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 6 {
		t.Fatalf("expected boss hp to drop to 6, got %+v", result.Boss)
	}
}

func TestAwakenHeroKeepsEffectsAndIgnoresLegacyUpgradeFields(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	seedHeroDefinition(t, store, HeroDefinition{
		HeroID:      "spark-cat",
		Name:        "星火猫",
		BonusClicks: 2,
		AwakenCap:   3,
		Effects: []HeroEffect{
			{
				Type:        HeroEffectFinalDamagePercent,
				Value:       5,
				DisplayName: "终幕打击",
				Description: "最终伤害 +5%",
			},
		},
	})
	if err := store.client.HSet(ctx, store.heroInventoryKey("阿明"), map[string]any{
		"spark-cat": "1",
	}).Err(); err != nil {
		t.Fatalf("seed hero inventory: %v", err)
	}
	if err := store.client.HSet(ctx, store.heroUpgradeKey("阿明", "spark-cat"), map[string]any{
		"pity_counter":                  "30",
		"bonus_clicks":                  "99",
		"bonus_critical_chance_percent": "50",
		"bonus_critical_count":          "99",
		"trait_value":                   "99",
	}).Err(); err != nil {
		t.Fatalf("seed legacy hero upgrade: %v", err)
	}
	if err := store.setGems(ctx, "阿明", 50); err != nil {
		t.Fatalf("seed gems: %v", err)
	}

	state, err := store.AwakenHero(ctx, "阿明", "spark-cat")
	if err != nil {
		t.Fatalf("awaken hero: %v", err)
	}

	hero := state.Heroes[0]
	if hero.BonusClicks != 3 || hero.BonusClicksDelta != 1 {
		t.Fatalf("expected awaken to apply fresh click growth only, got %+v", hero)
	}
	if len(hero.Effects) != 1 || hero.Effects[0].Type != HeroEffectFinalDamagePercent || hero.Effects[0].Value != 5 {
		t.Fatalf("expected effects to remain unchanged, got %+v", hero.Effects)
	}
	if state.LastForgeResult == nil || state.LastForgeResult.Jackpot {
		t.Fatalf("expected non-jackpot awaken payload, got %+v", state.LastForgeResult)
	}
}

func TestAwakenHeroAppliesFixedCriticalChanceGrowthAndRejectsAtCap(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 2 }

	ctx := context.Background()
	seedHeroDefinition(t, store, HeroDefinition{
		HeroID:      "spark-cat",
		Name:        "星火猫",
		BonusClicks: 2,
		AwakenCap:   1,
	})
	if err := store.client.HSet(ctx, store.heroInventoryKey("阿明"), map[string]any{
		"spark-cat": "1",
	}).Err(); err != nil {
		t.Fatalf("seed hero inventory: %v", err)
	}
	if err := store.setGems(ctx, "阿明", 50); err != nil {
		t.Fatalf("seed gems: %v", err)
	}

	state, err := store.AwakenHero(ctx, "阿明", "spark-cat")
	if err != nil {
		t.Fatalf("awaken hero: %v", err)
	}

	hero := state.Heroes[0]
	if hero.BonusCriticalChancePercent != 0.2 || hero.BonusCriticalChancePercentDelta != 0.2 {
		t.Fatalf("expected fixed crit chance growth, got %+v", hero)
	}
	if state.LastForgeResult == nil || state.LastForgeResult.RewardSummary != "暴击率 +0.20%" {
		t.Fatalf("expected crit chance awaken summary, got %+v", state.LastForgeResult)
	}
	if _, err := store.AwakenHero(ctx, "阿明", "spark-cat"); !errors.Is(err, ErrHeroMaxAwaken) {
		t.Fatalf("expected awaken cap error, got %v", err)
	}
}

func TestProgressionActionsRejectInsufficientGems(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedEquipmentDefinition(t, store, EquipmentDefinition{
		ItemID: "wood-sword",
		Name:   "木剑",
		Slot:   "weapon",
	})
	seedHeroDefinition(t, store, HeroDefinition{
		HeroID: "spark-cat",
		Name:   "星火猫",
	})
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), map[string]any{
		"wood-sword": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, store.heroInventoryKey("阿明"), map[string]any{
		"spark-cat": "1",
	}).Err(); err != nil {
		t.Fatalf("seed hero inventory: %v", err)
	}

	if _, err := store.EnhanceEquipment(ctx, "阿明", "wood-sword"); !errors.Is(err, ErrGemsNotEnough) {
		t.Fatalf("expected enhance to reject insufficient gems, got %v", err)
	}
	if _, err := store.AwakenHero(ctx, "阿明", "spark-cat"); !errors.Is(err, ErrGemsNotEnough) {
		t.Fatalf("expected awaken to reject insufficient gems, got %v", err)
	}
}
