package core

import (
	"context"
	"testing"
)

func TestEnhanceItemBatchConsumesAccumulatedCosts(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "wood-sword", "weapon", "普通", 10)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "wood-sword")

	if err := store.client.HSet(ctx, store.resourceKey(nickname), map[string]any{
		"gold":   "5000",
		"stones": "30",
	}).Err(); err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	state, err := store.EnhanceItemBatch(ctx, nickname, instanceID, 3)
	if err != nil {
		t.Fatalf("enhance item batch: %v", err)
	}

	if len(state.Inventory) != 1 {
		t.Fatalf("expected one inventory item, got %+v", state.Inventory)
	}
	item := state.Inventory[0]
	if item.EnhanceLevel != 3 {
		t.Fatalf("expected enhance level 3, got %+v", item)
	}
	if item.AttackPower != 14 {
		t.Fatalf("expected attack power 14 after batch enhance, got %+v", item)
	}
	if state.Gold != 2625 || state.Stones != 15 {
		t.Fatalf("expected remaining resources gold=2625 stones=15, got gold=%d stones=%d", state.Gold, state.Stones)
	}

	instance, err := store.getOwnedInstance(ctx, nickname, instanceID)
	if err != nil {
		t.Fatalf("reload owned instance: %v", err)
	}
	if instance.EnhanceLevel != 3 || instance.SpentStones != 15 {
		t.Fatalf("expected stored instance level=3 spent_stones=15, got %+v", instance)
	}
}

func TestEnhanceItemBatchIsAtomicWhenResourcesInsufficient(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "wood-sword", "weapon", "普通", 10)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "wood-sword")

	if err := store.client.HSet(ctx, store.resourceKey(nickname), map[string]any{
		"gold":   "5000",
		"stones": "15",
	}).Err(); err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	if _, err := store.EnhanceItemBatch(ctx, nickname, instanceID, 4); err != ErrEquipmentEnhanceInsufficientStones {
		t.Fatalf("expected insufficient stones, got %v", err)
	}

	instance, err := store.getOwnedInstance(ctx, nickname, instanceID)
	if err != nil {
		t.Fatalf("reload owned instance: %v", err)
	}
	if instance.EnhanceLevel != 0 || instance.SpentStones != 0 {
		t.Fatalf("expected instance unchanged after failed batch enhance, got %+v", instance)
	}

	resources, err := store.resourcesForNickname(ctx, nickname)
	if err != nil {
		t.Fatalf("reload resources: %v", err)
	}
	if resources.Gold != 5000 || resources.Stones != 15 {
		t.Fatalf("expected resources unchanged after failed batch enhance, got %+v", resources)
	}
}
