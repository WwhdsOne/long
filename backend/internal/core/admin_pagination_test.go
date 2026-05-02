package core

import (
	"context"
	"testing"
)

func TestGetAdminStateOmitsHeavyCollections(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "1",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:wood-sword", map[string]any{
		"name":         "木剑",
		"slot":         "weapon",
		"bonus_clicks": "2",
	}).Err(); err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
	if err := store.client.SAdd(ctx, "vote:equipment:index", "wood-sword").Err(); err != nil {
		t.Fatalf("seed equipment index: %v", err)
	}

	state, err := store.GetAdminState(ctx)
	if err != nil {
		t.Fatalf("get admin state: %v", err)
	}

	if len(state.Equipment) != 0 {
		t.Fatalf("expected equipment omitted from admin summary, got %+v", state.Equipment)
	}
}

func TestListAdminEquipmentPageReturnsStablePagination(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedEquipmentDefinitionForPage(t, store, ctx, "wood-sword", "木剑", "weapon")
	seedEquipmentDefinitionForPage(t, store, ctx, "cloth-armor", "布甲", "armor")
	seedEquipmentDefinitionForPage(t, store, ctx, "fire-ring", "火戒", "accessory")

	page, err := store.ListAdminEquipmentPage(ctx, 2, 1)
	if err != nil {
		t.Fatalf("list admin equipment page: %v", err)
	}

	if page.Page != 2 || page.PageSize != 1 || page.Total != 3 || page.TotalPages != 3 {
		t.Fatalf("unexpected page meta: %+v", page)
	}
	if len(page.Items) != 1 || page.Items[0].ItemID != "cloth-armor" {
		t.Fatalf("unexpected equipment page items: %+v", page.Items)
	}
}

func TestListAdminEquipmentPageReadsMediaAndTalentFields(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID:               "soft-blade",
		Name:                 "软组织切割刃",
		Slot:                 "weapon",
		Rarity:               "史诗",
		ImagePath:            "https://cdn.example.com/soft-blade.png",
		ImageAlt:             "软组织切割刃",
		AttackPower:          12,
		ArmorPenPercent:      0.2,
		CritRate:             0.22,
		CritDamageMultiplier: 1.5,
		PartTypeDamageSoft:   0.35,
		PartTypeDamageHeavy:  0.05,
		PartTypeDamageWeak:   0.15,
		TalentAffinity:       "normal",
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}

	page, err := store.ListAdminEquipmentPage(ctx, 1, 20)
	if err != nil {
		t.Fatalf("list admin equipment page: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one equipment item, got %+v", page.Items)
	}

	item := page.Items[0]
	if item.ImagePath != "https://cdn.example.com/soft-blade.png" {
		t.Fatalf("expected image path to round trip, got %q", item.ImagePath)
	}
	if item.ImageAlt != "软组织切割刃" {
		t.Fatalf("expected image alt to round trip, got %q", item.ImageAlt)
	}
	if item.TalentAffinity != "normal" {
		t.Fatalf("expected talent affinity normal, got %q", item.TalentAffinity)
	}
	if item.CritRate != 0.22 {
		t.Fatalf("expected crit rate to round trip, got %v", item.CritRate)
	}
}

func TestListAdminBossHistoryPageReturnsStablePagination(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	seedBossHistoryForPage(t, store, ctx, &Boss{ID: "boss-1", Name: "一号", Status: bossStatusDefeated, StartedAt: 1})
	seedBossHistoryForPage(t, store, ctx, &Boss{ID: "boss-2", Name: "二号", Status: bossStatusDefeated, StartedAt: 2})
	seedBossHistoryForPage(t, store, ctx, &Boss{ID: "boss-3", Name: "三号", Status: bossStatusDefeated, StartedAt: 3})

	page, err := store.ListAdminBossHistoryPage(ctx, 2, 1)
	if err != nil {
		t.Fatalf("list admin boss history page: %v", err)
	}

	if page.Page != 2 || page.PageSize != 1 || page.Total != 3 || page.TotalPages != 3 {
		t.Fatalf("unexpected page meta: %+v", page)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "boss-2" {
		t.Fatalf("unexpected boss history page items: %+v", page.Items)
	}
}

func seedEquipmentDefinitionForPage(t *testing.T, store *Store, ctx context.Context, itemID string, name string, slot string) {
	t.Helper()

	if err := store.client.HSet(ctx, "vote:equip:def:"+itemID, map[string]any{
		"name":         name,
		"slot":         slot,
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed equipment %s: %v", itemID, err)
	}
	if err := store.client.SAdd(ctx, "vote:equipment:index", itemID).Err(); err != nil {
		t.Fatalf("seed equipment index %s: %v", itemID, err)
	}
}

func seedBossHistoryForPage(t *testing.T, store *Store, ctx context.Context, boss *Boss) {
	t.Helper()

	if err := store.SaveBossToHistory(ctx, boss); err != nil {
		t.Fatalf("seed boss history %s: %v", boss.ID, err)
	}
}
