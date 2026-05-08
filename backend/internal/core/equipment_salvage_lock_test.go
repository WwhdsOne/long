package core

import (
	"context"
	"slices"
	"testing"
)

func TestSalvageItemReturnsRarityRewardsAndRefund(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "rare-sword", "weapon", "稀有", 20)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(instanceID), map[string]any{
		"spent_stones": "9",
	}).Err(); err != nil {
		t.Fatalf("seed spent stones: %v", err)
	}
	if err := store.client.HSet(ctx, store.resourceKey(nickname), map[string]any{
		"gold":   "100",
		"stones": "10",
	}).Err(); err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	result, err := store.SalvageItem(ctx, nickname, instanceID)
	if err != nil {
		t.Fatalf("salvage item: %v", err)
	}

	if result.GoldReward != 500 || result.StoneReward != 1 || result.RefundedStones != 5 {
		t.Fatalf("unexpected salvage rewards: %+v", result)
	}
	if result.Gold != 600 || result.Stones != 16 {
		t.Fatalf("expected resources after salvage gold=600 stones=16, got %+v", result)
	}

	exists, err := store.client.Exists(ctx, store.equipmentInstanceKey(instanceID)).Result()
	if err != nil {
		t.Fatalf("check instance exists: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected salvaged instance removed, exists=%d", exists)
	}
}

func TestBulkSalvageUnequippedSkipsProtectedItems(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "rare-sword", "weapon", "稀有", 20)
	seedEquipmentDefinitionWithRarity(t, store, ctx, "epic-armor", "chest", "史诗", 30)
	seedEquipmentDefinitionWithRarity(t, store, ctx, "perfect-ring", "accessory", "至臻", 50)

	equippedID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")
	lockedID := seedOwnedInstance(t, store, ctx, nickname, "epic-armor")
	perfectID := seedOwnedInstance(t, store, ctx, nickname, "perfect-ring")
	salvageID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")

	if err := store.client.HSet(ctx, store.loadoutKey(nickname), "weapon", equippedID).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(lockedID), map[string]any{
		"locked": "1",
	}).Err(); err != nil {
		t.Fatalf("seed locked instance: %v", err)
	}
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(salvageID), map[string]any{
		"spent_stones": "4",
	}).Err(); err != nil {
		t.Fatalf("seed salvage spent stones: %v", err)
	}

	result, err := store.BulkSalvageUnequipped(ctx, nickname)
	if err != nil {
		t.Fatalf("bulk salvage: %v", err)
	}

	if result.SalvagedCount != 0 || result.ExcludedEquipped != 1 || result.ExcludedLocked != 1 || result.ExcludedTopRarity != 1 {
		t.Fatalf("unexpected bulk salvage summary: %+v", result)
	}
	if result.GoldReward != 0 || result.StoneReward != 0 || result.RefundedStones != 0 {
		t.Fatalf("unexpected bulk salvage rewards: %+v", result)
	}
	if result.Gold != 0 || result.Stones != 0 {
		t.Fatalf("expected resources unchanged after bulk salvage, got %+v", result)
	}

	remainingIDs, err := store.client.SMembers(ctx, store.playerInstancesKey(nickname)).Result()
	if err != nil {
		t.Fatalf("list remaining instances: %v", err)
	}
	if len(remainingIDs) != 4 {
		t.Fatalf("expected 4 remaining instances, got %v", remainingIDs)
	}
	if !containsString(remainingIDs, equippedID) || !containsString(remainingIDs, lockedID) || !containsString(remainingIDs, perfectID) || !containsString(remainingIDs, salvageID) {
		t.Fatalf("expected equipped/locked/top rarity/single candidate instances kept, got %v", remainingIDs)
	}
}

func TestBulkSalvageUnequippedKeepsHighestEnhancedPerItem(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "rare-sword", "weapon", "稀有", 20)
	seedEquipmentDefinitionWithRarity(t, store, ctx, "epic-armor", "chest", "史诗", 30)

	lowID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")
	highAID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")
	highBID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")
	otherSingleID := seedOwnedInstance(t, store, ctx, nickname, "epic-armor")

	if err := store.client.HSet(ctx, store.equipmentInstanceKey(lowID), map[string]any{
		"enhance_level": "1",
		"spent_stones":  "3",
	}).Err(); err != nil {
		t.Fatalf("seed low enhance instance: %v", err)
	}
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(highAID), map[string]any{
		"enhance_level": "4",
		"spent_stones":  "12",
	}).Err(); err != nil {
		t.Fatalf("seed high enhance instance a: %v", err)
	}
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(highBID), map[string]any{
		"enhance_level": "4",
		"spent_stones":  "12",
	}).Err(); err != nil {
		t.Fatalf("seed high enhance instance b: %v", err)
	}
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(otherSingleID), map[string]any{
		"enhance_level": "2",
		"spent_stones":  "5",
	}).Err(); err != nil {
		t.Fatalf("seed single instance: %v", err)
	}

	result, err := store.BulkSalvageUnequipped(ctx, nickname)
	if err != nil {
		t.Fatalf("bulk salvage: %v", err)
	}

	if result.SalvagedCount != 2 {
		t.Fatalf("expected two rare swords salvaged, got %+v", result)
	}
	if result.GoldReward != 1000 || result.StoneReward != 2 || result.RefundedStones != 8 {
		t.Fatalf("unexpected bulk salvage rewards: %+v", result)
	}

	remainingIDs, err := store.client.SMembers(ctx, store.playerInstancesKey(nickname)).Result()
	if err != nil {
		t.Fatalf("list remaining instances: %v", err)
	}
	if len(remainingIDs) != 2 {
		t.Fatalf("expected 2 remaining instances, got %v", remainingIDs)
	}
	if containsString(remainingIDs, lowID) {
		t.Fatalf("expected lowest enhanced rare sword salvaged, got %v", remainingIDs)
	}
	if !containsString(remainingIDs, otherSingleID) {
		t.Fatalf("expected single epic armor kept, got %v", remainingIDs)
	}

	keptHighCount := 0
	if containsString(remainingIDs, highAID) {
		keptHighCount++
	}
	if containsString(remainingIDs, highBID) {
		keptHighCount++
	}
	if keptHighCount != 1 {
		t.Fatalf("expected one random highest-enhanced rare sword kept, got %v", remainingIDs)
	}
}

func TestLockAndUnlockItemUpdatesInventoryLockState(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "wood-sword", "weapon", "普通", 10)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "wood-sword")

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state before lock: %v", err)
	}
	if len(state.Inventory) != 1 || state.Inventory[0].Locked {
		t.Fatalf("expected default unlocked inventory item, got %+v", state.Inventory)
	}

	if _, err := store.LockItem(ctx, nickname, instanceID); err != nil {
		t.Fatalf("lock item: %v", err)
	}
	state, err = store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after lock: %v", err)
	}
	if len(state.Inventory) != 1 || !state.Inventory[0].Locked {
		t.Fatalf("expected inventory item locked, got %+v", state.Inventory)
	}

	if _, err := store.UnlockItem(ctx, nickname, instanceID); err != nil {
		t.Fatalf("unlock item: %v", err)
	}
	state, err = store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after unlock: %v", err)
	}
	if len(state.Inventory) != 1 || state.Inventory[0].Locked {
		t.Fatalf("expected inventory item unlocked after unlock action, got %+v", state.Inventory)
	}
}

func seedEquipmentDefinitionWithRarity(t *testing.T, store *Store, ctx context.Context, itemID string, slot string, rarity string, attackPower int64) {
	t.Helper()

	if err := store.client.HSet(ctx, store.equipmentKey(itemID), map[string]any{
		"name":         itemID,
		"slot":         slot,
		"rarity":       rarity,
		"attack_power": attackPower,
	}).Err(); err != nil {
		t.Fatalf("seed equipment %s: %v", itemID, err)
	}
	if err := store.client.SAdd(ctx, store.equipmentIndexKey, itemID).Err(); err != nil {
		t.Fatalf("seed equipment index %s: %v", itemID, err)
	}
}

func containsString(items []string, target string) bool {
	return slices.Contains(items, target)
}
