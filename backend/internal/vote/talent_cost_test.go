package vote

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
	if got := TalentLevelCostDiff(20, 0, 5); got != 451 {
		t.Fatalf("expected 0->5 cumulative cost 451, got %d", got)
	}
	if got := TalentLevelCostDiff(20, 2, 5); got != 350 {
		t.Fatalf("expected 2->5 diff cost 350, got %d", got)
	}
	if got := TalentCumulativeCost(20, 5); got != 451 {
		t.Fatalf("expected cumulative cost 451, got %d", got)
	}
}
