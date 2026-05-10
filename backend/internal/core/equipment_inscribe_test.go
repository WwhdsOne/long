package core

import (
	"context"
	"testing"
)

func TestMaxAffixCountByRarity(t *testing.T) {
	cases := map[string]int{
		"普通": 2,
		"优秀": 2,
		"稀有": 3,
		"史诗": 5,
		"传说": 6,
		"神话": 7,
		"至臻": 8,
		"未知": 2,
	}

	for rarity, want := range cases {
		if got := maxAffixCount(rarity); got != want {
			t.Fatalf("rarity %q max affix count = %d, want %d", rarity, got, want)
		}
	}
}

func TestInscribeItemConsumesInscriptionStoneAndAddsAffix(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "legend-sword", "weapon", "传说", 20)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "legend-sword")
	if err := store.client.HSet(ctx, store.resourceKey(nickname), map[string]any{
		"inscription_stones": "2",
	}).Err(); err != nil {
		t.Fatalf("seed inscription stones: %v", err)
	}

	state, err := store.InscribeItem(ctx, nickname, instanceID)
	if err != nil {
		t.Fatalf("inscribe item: %v", err)
	}

	if state.InscriptionStones != 1 {
		t.Fatalf("expected inscription stones to be consumed to 1, got %d", state.InscriptionStones)
	}
	if len(state.Inventory) != 1 {
		t.Fatalf("expected one inventory item, got %+v", state.Inventory)
	}
	item := state.Inventory[0]
	if item.AffixCount != 1 {
		t.Fatalf("expected one affix after inscribe, got %+v", item)
	}
	if item.AffixLimit != 6 {
		t.Fatalf("expected legend affix limit 6, got %+v", item)
	}
	if len(item.Affixes) != 1 {
		t.Fatalf("expected affix list length 1, got %+v", item)
	}

	instance, err := store.getOwnedInstance(ctx, nickname, instanceID)
	if err != nil {
		t.Fatalf("reload instance: %v", err)
	}
	if len(instance.Affixes) != 1 {
		t.Fatalf("expected stored affixes on instance, got %+v", instance)
	}
	if instance.Affixes[0].ID == "" || instance.Affixes[0].Name == "" || instance.Affixes[0].ValueText == "" {
		t.Fatalf("expected generated affix fields populated, got %+v", instance.Affixes[0])
	}
}

func TestInscribeItemRejectsWhenInscriptionStonesInsufficient(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "rare-sword", "weapon", "稀有", 20)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "rare-sword")

	if _, err := store.InscribeItem(ctx, nickname, instanceID); err != ErrInscriptionStoneInsufficient {
		t.Fatalf("expected insufficient inscription stones, got %v", err)
	}
}

func TestInscribeItemRejectsWhenAffixLimitReached(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	seedEquipmentDefinitionWithRarity(t, store, ctx, "common-sword", "weapon", "普通", 10)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "common-sword")
	if err := store.client.HSet(ctx, store.resourceKey(nickname), map[string]any{
		"inscription_stones": "5",
	}).Err(); err != nil {
		t.Fatalf("seed inscription stones: %v", err)
	}

	instance, err := store.getOwnedInstance(ctx, nickname, instanceID)
	if err != nil {
		t.Fatalf("load instance: %v", err)
	}
	instance.Affixes = []ItemAffix{
		{ID: "a1", Name: "猛攻", Type: "attack_power", Value: 12, ValueText: "+12", Tier: "normal"},
		{ID: "a2", Name: "猛攻", Type: "attack_power", Value: 15, ValueText: "+15", Tier: "normal"},
	}
	if err := store.saveItemInstanceAffixes(ctx, instanceID, instance.Affixes); err != nil {
		t.Fatalf("seed affixes: %v", err)
	}

	if _, err := store.InscribeItem(ctx, nickname, instanceID); err != ErrEquipmentAffixLimitReached {
		t.Fatalf("expected affix limit reached, got %v", err)
	}
}

func TestGenerateInscriptionAffixSkipsSensitiveTypeAfterCap(t *testing.T) {
	instance := &ItemInstance{
		InstanceID: "inst-1",
		ItemID:     "legend-sword",
		Affixes: []ItemAffix{
			{ID: "a1", Name: "精准", Type: "crit_rate", Value: 0.03, ValueText: "+3%", Tier: "sensitive"},
			{ID: "a2", Name: "精准", Type: "crit_rate", Value: 0.03, ValueText: "+3%", Tier: "sensitive"},
		},
	}

	for range 20 {
		affix := generateInscriptionAffix(instance)
		if affix.Type == "crit_rate" {
			t.Fatalf("expected crit_rate affix blocked after reaching cap, got %+v", affix)
		}
	}
}
