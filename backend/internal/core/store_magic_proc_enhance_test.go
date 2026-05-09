package core

import (
	"math"
	"testing"
)

func TestBuildInventoryItemMagicProcRateBonusUsesFixedStepPerEnhanceLevel(t *testing.T) {
	definition := EquipmentDefinition{
		ItemID:             "magic-ring",
		Name:               "奥术指环",
		Slot:               "accessory",
		Rarity:             "史诗",
		MagicProcRateBonus: 0.015,
	}

	item := buildInventoryItem(definition, 1, false, 3, "inst-1", false, false)
	want := 0.015 + 3*0.001
	if math.Abs(item.MagicProcRateBonus-want) > 1e-9 {
		t.Fatalf("expected magic proc rate bonus %.6f at +3, got %.6f", want, item.MagicProcRateBonus)
	}
}

