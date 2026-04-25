package vote

import (
	"context"
	"testing"
)

func TestLoadoutSupportsDesignEquipmentSlotsAndLegacyArmor(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedEquipmentDefinition(t, store, ctx, "star-hammer", "武器", 10)
	seedEquipmentDefinition(t, store, ctx, "star-helm", "helmet", 20)
	seedEquipmentDefinition(t, store, ctx, "star-chest", "chest", 30)
	seedEquipmentDefinition(t, store, ctx, "star-gloves", "gloves", 40)
	seedEquipmentDefinition(t, store, ctx, "star-legs", "legs", 50)
	seedEquipmentDefinition(t, store, ctx, "star-badge", "accessory", 60)
	seedEquipmentDefinition(t, store, ctx, "old-armor", "armor", 70)

	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), map[string]any{
		"star-hammer": "1",
		"star-helm":   "1",
		"star-chest":  "1",
		"star-gloves": "1",
		"star-legs":   "1",
		"star-badge":  "1",
		"old-armor":   "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, store.loadoutKey("阿明"), map[string]any{
		"weapon":    "star-hammer",
		"helmet":    "star-helm",
		"chest":     "star-chest",
		"gloves":    "star-gloves",
		"legs":      "star-legs",
		"accessory": "star-badge",
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}

	quantities, err := store.inventoryQuantities(ctx, "阿明")
	if err != nil {
		t.Fatalf("inventory quantities: %v", err)
	}
	loadout, equipped, err := store.loadoutForNickname(ctx, "阿明", quantities)
	if err != nil {
		t.Fatalf("loadout: %v", err)
	}

	if loadout.Weapon == nil || loadout.Helmet == nil || loadout.Chest == nil || loadout.Gloves == nil || loadout.Legs == nil || loadout.Accessory == nil {
		t.Fatalf("expected six-slot loadout, got %+v", loadout)
	}
	if loadout.Weapon.Slot != "weapon" {
		t.Fatalf("expected Chinese slot name to normalize into weapon, got %q", loadout.Weapon.Slot)
	}
	if equipped["star-chest"] != "chest" {
		t.Fatalf("expected equipped chest marker, got %+v", equipped)
	}

	attackPower, _, _ := loadoutBonuses(loadout)
	if attackPower != 210 {
		t.Fatalf("expected all six slots to contribute attack power 210, got %d", attackPower)
	}

	if err := store.client.HSet(ctx, store.loadoutKey("阿明"), map[string]any{
		"chest": "old-armor",
	}).Err(); err != nil {
		t.Fatalf("seed legacy armor loadout: %v", err)
	}
	loadout, equipped, err = store.loadoutForNickname(ctx, "阿明", quantities)
	if err != nil {
		t.Fatalf("legacy loadout: %v", err)
	}
	if loadout.Chest == nil || loadout.Chest.ItemID != "old-armor" || loadout.Chest.Slot != "chest" {
		t.Fatalf("expected legacy armor definition to normalize into chest slot, got %+v", loadout.Chest)
	}
	if equipped["old-armor"] != "chest" {
		t.Fatalf("expected legacy armor equipped marker to be chest, got %+v", equipped)
	}
}

func seedEquipmentDefinition(t *testing.T, store *Store, ctx context.Context, itemID string, slot string, attackPower int64) {
	t.Helper()

	if err := store.client.HSet(ctx, store.equipmentKey(itemID), map[string]any{
		"name":         itemID,
		"slot":         slot,
		"rarity":       "普通",
		"attack_power": attackPower,
	}).Err(); err != nil {
		t.Fatalf("seed equipment %s: %v", itemID, err)
	}
	if err := store.client.SAdd(ctx, store.equipmentIndexKey, itemID).Err(); err != nil {
		t.Fatalf("seed equipment index %s: %v", itemID, err)
	}
}
