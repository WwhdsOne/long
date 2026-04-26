package vote

import (
	"context"
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

	if result.SalvagedCount != 1 || result.ExcludedEquipped != 1 || result.ExcludedLocked != 1 || result.ExcludedTopRarity != 1 {
		t.Fatalf("unexpected bulk salvage summary: %+v", result)
	}
	if result.GoldReward != 500 || result.StoneReward != 1 || result.RefundedStones != 2 {
		t.Fatalf("unexpected bulk salvage rewards: %+v", result)
	}
	if result.Gold != 500 || result.Stones != 3 {
		t.Fatalf("expected resources after bulk salvage gold=500 stones=3, got %+v", result)
	}

	remainingIDs, err := store.client.SMembers(ctx, store.playerInstancesKey(nickname)).Result()
	if err != nil {
		t.Fatalf("list remaining instances: %v", err)
	}
	if len(remainingIDs) != 3 {
		t.Fatalf("expected 3 remaining instances, got %v", remainingIDs)
	}
	if !containsString(remainingIDs, equippedID) || !containsString(remainingIDs, lockedID) || !containsString(remainingIDs, perfectID) {
		t.Fatalf("expected equipped/locked/top rarity instances kept, got %v", remainingIDs)
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
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
