package vote

import (
	"context"
	"testing"
)

func TestCombatStatsRefreshAfterEquipItem(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "缓存穿戴测试"

	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID:      "cache-sword",
		Name:        "缓存之剑",
		Slot:        "weapon",
		Rarity:      "普通",
		AttackPower: 7,
		CritRate:    0.05,
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}

	before, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state before equip: %v", err)
	}
	if before.CombatStats.AttackPower != 5 {
		t.Fatalf("expected base attack 5 before equip, got %+v", before.CombatStats)
	}

	instanceID := seedOwnedInstance(t, store, ctx, nickname, "cache-sword")
	if _, err := store.EquipItem(ctx, nickname, instanceID); err != nil {
		t.Fatalf("equip item: %v", err)
	}

	after, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after equip: %v", err)
	}
	if after.CombatStats.AttackPower != 12 {
		t.Fatalf("expected attack 12 after equip, got %+v", after.CombatStats)
	}
	if after.CombatStats.CriticalChancePercent != before.CombatStats.CriticalChancePercent+5 {
		t.Fatalf("expected crit rate to increase by 5 after equip, before=%+v after=%+v", before.CombatStats, after.CombatStats)
	}
}

func TestCombatStatsRefreshAfterTalentUpgradeAndReset(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "缓存天赋测试"

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed talent points: %v", err)
	}

	before, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state before upgrade: %v", err)
	}
	if before.CombatStats.AttackPower != 5 {
		t.Fatalf("expected base attack 5 before talent upgrade, got %+v", before.CombatStats)
	}

	if err := store.UpgradeTalent(ctx, nickname, "normal_core", 1); err != nil {
		t.Fatalf("upgrade normal_core: %v", err)
	}
	if err := store.UpgradeTalent(ctx, nickname, "normal_atk_up", 1); err != nil {
		t.Fatalf("upgrade normal_atk_up: %v", err)
	}

	upgraded, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after upgrade: %v", err)
	}
	if upgraded.CombatStats.AttackPower <= before.CombatStats.AttackPower {
		t.Fatalf("expected attack to increase after talent upgrade, before=%+v after=%+v", before.CombatStats, upgraded.CombatStats)
	}

	if err := store.ResetTalents(ctx, nickname); err != nil {
		t.Fatalf("reset talents: %v", err)
	}

	reset, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after reset: %v", err)
	}
	if reset.CombatStats.AttackPower != before.CombatStats.AttackPower {
		t.Fatalf("expected attack to return to base after reset, before=%+v after=%+v", before.CombatStats, reset.CombatStats)
	}
}
