package core

import "testing"

func TestTalentLevelCostUsesFlattenedExponentCurve(t *testing.T) {
	if got := TalentCostTier0Main; got != 20 {
		t.Fatalf("expected tier0 main base cost 20, got %d", got)
	}
	if got := TalentCostTier2Filler; got != 35 {
		t.Fatalf("expected tier2 filler base cost 35, got %d", got)
	}
	if got := TalentCostTier3Filler; got != 60 {
		t.Fatalf("expected tier3 filler base cost 60, got %d", got)
	}

	expectedByLevel := map[int]int64{
		1: 108,
		2: 195,
		3: 275,
		4: 351,
		5: 424,
	}
	for level, want := range expectedByLevel {
		if got := TalentLevelCost(60, level); got != want {
			t.Fatalf("expected level %d cost %d, got %d", level, want, got)
		}
	}
}

func TestTalentLevelCostDiffAndCumulativeUsePerLevelCharging(t *testing.T) {
	if got := TalentLevelCost(20, 1); got != 36 {
		t.Fatalf("expected tier0 lv1 single cost 36, got %d", got)
	}
	if got := TalentLevelCost(20, 2); got != 108 {
		t.Fatalf("expected tier0 lv2 single cost 108, got %d", got)
	}
	if got := TalentLevelCost(20, 3); got != 324 {
		t.Fatalf("expected tier0 lv3 single cost 324, got %d", got)
	}
	if got := TalentLevelCost(20, 4); got != 972 {
		t.Fatalf("expected tier0 lv4 single cost 972, got %d", got)
	}
	if got := TalentLevelCost(20, 5); got != 2916 {
		t.Fatalf("expected tier0 lv5 single cost 2916, got %d", got)
	}
	if got := TalentLevelCostDiff(20, 0, 5); got != 4356 {
		t.Fatalf("expected 0->5 cumulative cost 4356, got %d", got)
	}
	if got := TalentLevelCostDiff(20, 2, 5); got != 4212 {
		t.Fatalf("expected 2->5 diff cost 4212, got %d", got)
	}
	if got := TalentCumulativeCost(20, 5); got != 4356 {
		t.Fatalf("expected cumulative cost 4356, got %d", got)
	}
}
