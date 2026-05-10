package core

import "testing"

func TestEncodeDecodeBossPartsPreservesDamageAffinity(t *testing.T) {
	encoded, err := encodeBossParts([]BossPart{{
		X:              0,
		Y:              0,
		Type:           PartTypeArcane,
		DamageAffinity: PartDamageAffinityMagicOnly,
		MaxHP:          100,
		CurrentHP:      100,
		Armor:          0,
		Alive:          true,
	}})
	if err != nil {
		t.Fatalf("encode boss parts: %v", err)
	}

	decoded, err := decodeBossParts(encoded)
	if err != nil {
		t.Fatalf("decode boss parts: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 boss part, got %d", len(decoded))
	}
	if decoded[0].DamageAffinity != PartDamageAffinityMagicOnly {
		t.Fatalf("expected damage affinity preserved, got %+v", decoded[0])
	}
}
