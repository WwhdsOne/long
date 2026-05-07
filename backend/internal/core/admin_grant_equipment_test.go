package core

import (
	"context"
	"testing"
)

func TestGrantEquipmentToPlayerCreatesMultipleInstances(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID: "wood-sword",
		Name:   "木剑",
		Slot:   "weapon",
		Rarity: "普通",
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}

	state, err := store.GrantEquipmentToPlayer(ctx, "阿明", "wood-sword", 3)
	if err != nil {
		t.Fatalf("grant equipment: %v", err)
	}

	if len(state.Inventory) != 3 {
		t.Fatalf("expected 3 inventory items, got %+v", state.Inventory)
	}
	for _, item := range state.Inventory {
		if item.ItemID != "wood-sword" {
			t.Fatalf("expected item wood-sword, got %+v", item)
		}
		if item.InstanceID == "" {
			t.Fatalf("expected instance id, got %+v", item)
		}
	}

	instances, err := store.itemInstancesByIDForNickname(ctx, "阿明")
	if err != nil {
		t.Fatalf("list item instances: %v", err)
	}
	if len(instances) != 3 {
		t.Fatalf("expected 3 instances, got %d", len(instances))
	}
}

func TestGrantEquipmentToPlayerRejectsInvalidQuantity(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID: "wood-sword",
		Name:   "木剑",
		Slot:   "weapon",
		Rarity: "普通",
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}

	if _, err := store.GrantEquipmentToPlayer(ctx, "阿明", "wood-sword", 0); err == nil {
		t.Fatal("expected invalid quantity error")
	}
}
