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
