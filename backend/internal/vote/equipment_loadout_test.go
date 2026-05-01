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

	hammerInst := seedOwnedInstance(t, store, ctx, "阿明", "star-hammer")
	helmInst := seedOwnedInstance(t, store, ctx, "阿明", "star-helm")
	chestInst := seedOwnedInstance(t, store, ctx, "阿明", "star-chest")
	glovesInst := seedOwnedInstance(t, store, ctx, "阿明", "star-gloves")
	legsInst := seedOwnedInstance(t, store, ctx, "阿明", "star-legs")
	badgeInst := seedOwnedInstance(t, store, ctx, "阿明", "star-badge")
	oldArmorInst := seedOwnedInstance(t, store, ctx, "阿明", "old-armor")
	if err := store.client.HSet(ctx, store.loadoutKey("阿明"), map[string]any{
		"weapon":    hammerInst,
		"helmet":    helmInst,
		"chest":     chestInst,
		"gloves":    glovesInst,
		"legs":      legsInst,
		"accessory": badgeInst,
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}

	loadout, equipped, err := store.loadoutForNickname(ctx, "阿明")
	if err != nil {
		t.Fatalf("loadout: %v", err)
	}

	if loadout.Weapon == nil || loadout.Helmet == nil || loadout.Chest == nil || loadout.Gloves == nil || loadout.Legs == nil || loadout.Accessory == nil {
		t.Fatalf("expected six-slot loadout, got %+v", loadout)
	}
	if loadout.Weapon.Slot != "weapon" {
		t.Fatalf("expected Chinese slot name to normalize into weapon, got %q", loadout.Weapon.Slot)
	}
	if equipped[chestInst] != "chest" {
		t.Fatalf("expected equipped chest marker, got %+v", equipped)
	}

	attackPower, _, _, _, _, _, _ := loadoutBonuses(loadout)
	if attackPower != 210 {
		t.Fatalf("expected all six slots to contribute attack power 210, got %d", attackPower)
	}

	if err := store.client.HSet(ctx, store.loadoutKey("阿明"), map[string]any{
		"chest": oldArmorInst,
	}).Err(); err != nil {
		t.Fatalf("seed legacy armor loadout: %v", err)
	}
	loadout, equipped, err = store.loadoutForNickname(ctx, "阿明")
	if err != nil {
		t.Fatalf("legacy loadout: %v", err)
	}
	if loadout.Chest == nil || loadout.Chest.ItemID != "old-armor" || loadout.Chest.Slot != "chest" {
		t.Fatalf("expected legacy armor definition to normalize into chest slot, got %+v", loadout.Chest)
	}
	if equipped[oldArmorInst] != "chest" {
		t.Fatalf("expected legacy armor equipped marker to be chest, got %+v", equipped)
	}
}

func seedOwnedInstance(t *testing.T, store *Store, ctx context.Context, nickname string, itemID string) string {
	t.Helper()

	instanceID, err := store.newEquipmentInstanceID(ctx)
	if err != nil {
		t.Fatalf("alloc instance id: %v", err)
	}
	if err := store.client.HSet(ctx, store.equipmentInstanceKey(instanceID), map[string]any{
		"item_id":       itemID,
		"enhance_level": "0",
		"spent_stones":  "0",
		"bound":         "0",
		"locked":        "0",
		"created_at":    "0",
	}).Err(); err != nil {
		t.Fatalf("seed instance %s: %v", instanceID, err)
	}
	if err := store.client.SAdd(ctx, store.playerInstancesKey(nickname), instanceID).Err(); err != nil {
		t.Fatalf("bind instance %s to player: %v", instanceID, err)
	}

	return instanceID
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
