package vote

import (
	"context"
	"testing"
)

func TestCombatStatsRefreshAfterEquipItem(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "缓存穿戴测试"

	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID:      "cache-sword",
		Name:        "缓存之剑",
		Slot:        "weapon",
		Rarity:      "普通",
		AttackPower: 7,
		CritRate:    0.05,
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}

	before, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state before equip: %v", err)
	}
	if before.CombatStats.AttackPower != 5 {
		t.Fatalf("expected base attack 5 before equip, got %+v", before.CombatStats)
	}

	instanceID := seedOwnedInstance(t, store, ctx, nickname, "cache-sword")
	if _, err := store.EquipItem(ctx, nickname, instanceID); err != nil {
		t.Fatalf("equip item: %v", err)
	}

	after, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after equip: %v", err)
	}
	if after.CombatStats.AttackPower != 12 {
		t.Fatalf("expected attack 12 after equip, got %+v", after.CombatStats)
	}
	if after.CombatStats.CriticalChancePercent != before.CombatStats.CriticalChancePercent+5 {
		t.Fatalf("expected crit rate to increase by 5 after equip, before=%+v after=%+v", before.CombatStats, after.CombatStats)
	}
}

func TestCombatStatsRefreshAfterTalentUpgradeAndReset(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "缓存天赋测试"

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed talent points: %v", err)
	}

	before, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state before upgrade: %v", err)
	}
	if before.CombatStats.AttackPower != 5 {
		t.Fatalf("expected base attack 5 before talent upgrade, got %+v", before.CombatStats)
	}

	if err := store.UpgradeTalent(ctx, nickname, "normal_core", 1); err != nil {
		t.Fatalf("upgrade normal_core: %v", err)
	}
	if err := store.UpgradeTalent(ctx, nickname, "normal_atk_up", 1); err != nil {
		t.Fatalf("upgrade normal_atk_up: %v", err)
	}

	upgraded, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after upgrade: %v", err)
	}
	if upgraded.CombatStats.AttackPower <= before.CombatStats.AttackPower {
		t.Fatalf("expected attack to increase after talent upgrade, before=%+v after=%+v", before.CombatStats, upgraded.CombatStats)
	}

	if err := store.ResetTalents(ctx, nickname); err != nil {
		t.Fatalf("reset talents: %v", err)
	}

	reset, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after reset: %v", err)
	}
	if reset.CombatStats.AttackPower != before.CombatStats.AttackPower {
		t.Fatalf("expected attack to return to base after reset, before=%+v after=%+v", before.CombatStats, reset.CombatStats)
	}
}

func TestCompileTalentSetBuildsNormalThresholdsAndTierFlags(t *testing.T) {
	compiled := compileTalentSet(&TalentState{
		Talents: map[string]int{
			"normal_core":       5,
			"normal_atk_up":     1,
			"normal_dmg_amp":    1,
			"normal_soft_atk":   1,
			"normal_charge":     4,
			"normal_chase_up":   3,
			"normal_combo_ext":  2,
			"normal_filler_t2a": 1,
			"normal_filler_t2b": 1,
			"normal_ultimate":   5,
		},
	})

	if compiled == nil {
		t.Fatal("expected compiled talent set")
	}
	if !compiled.Has("normal_core") || !compiled.Has("normal_ultimate") {
		t.Fatalf("expected compiled set to record learned talents, got %+v", compiled)
	}
	if !compiled.IsTierFull(TalentTreeNormal, 2) {
		t.Fatalf("expected normal tier 2 to be full, got %+v", compiled.tierFull)
	}
	if compiled.Normal.TriggerCount != int64(normalCoreTriggerCountForLevel(5)-20) {
		t.Fatalf("expected compiled normal trigger count %d, got %d", normalCoreTriggerCountForLevel(5)-20, compiled.Normal.TriggerCount)
	}
	expectedHits := int64(normalCoreExtraHitsForLevel(5) + normalComboExtendHitsForLevel(2) + 5)
	if compiled.Normal.ExtraHits != expectedHits {
		t.Fatalf("expected compiled normal extra hits %d, got %d", expectedHits, compiled.Normal.ExtraHits)
	}
	expectedRatio := normalChaseUpgradeRatioForLevel(3) + 0.15
	if compiled.Normal.ChaseRatio != expectedRatio {
		t.Fatalf("expected compiled normal chase ratio %.2f, got %.2f", expectedRatio, compiled.Normal.ChaseRatio)
	}
	if compiled.Normal.RetainPercent != normalChargeRetainPercentForLevel(4) {
		t.Fatalf("expected retain percent %.2f, got %.2f", normalChargeRetainPercentForLevel(4), compiled.Normal.RetainPercent)
	}
	if compiled.Normal.SilverStormDuration != int64(normalSilverStormDurationForLevel(5)) {
		t.Fatalf("expected silver storm duration %d, got %d", normalSilverStormDurationForLevel(5), compiled.Normal.SilverStormDuration)
	}
}

func TestCompileTalentSetBuildsArmorAndCritThresholds(t *testing.T) {
	compiled := compileTalentSet(&TalentState{
		Talents: map[string]int{
			"armor_core":         4,
			"armor_pen_up":       1,
			"armor_boss_hunter":  1,
			"armor_heavy_scale":  1,
			"armor_filler_t1a":   1,
			"armor_filler_t1b":   1,
			"armor_auto_strike":  2,
			"armor_collapse_ext": 3,
			"armor_ruin":         5,
			"armor_ultimate":     4,
			"crit_core":          5,
			"crit_omen_resonate": 3,
			"crit_cruel":         2,
			"crit_skinner":       4,
			"crit_doom_judgment": 2,
			"crit_final_cut":     1,
			"crit_death_ecstasy": 5,
		},
	})

	if compiled == nil {
		t.Fatal("expected compiled talent set")
	}
	if !compiled.IsTierFull(TalentTreeArmor, 1) {
		t.Fatalf("expected armor tier 1 to be full, got %+v", compiled.tierFull)
	}
	expectedCollapseTrigger := int64(max(armorCoreCollapseTriggerForLevel(4)-30, 1))
	if compiled.Armor.CollapseTrigger != expectedCollapseTrigger {
		t.Fatalf("expected compiled collapse trigger %d, got %d", expectedCollapseTrigger, compiled.Armor.CollapseTrigger)
	}
	if compiled.Armor.AutoStrikeInterval != int64(armorAutoStrikeIntervalForLevel(2)) {
		t.Fatalf("expected auto strike interval %d, got %d", armorAutoStrikeIntervalForLevel(2), compiled.Armor.AutoStrikeInterval)
	}
	if compiled.Armor.AutoStrikeRatio != armorAutoStrikeRatioForLevel(2) {
		t.Fatalf("expected auto strike ratio %.2f, got %.2f", armorAutoStrikeRatioForLevel(2), compiled.Armor.AutoStrikeRatio)
	}
	if compiled.Armor.CollapseDuration != int64(armorCollapseExtendForLevel(3)) {
		t.Fatalf("expected collapse duration %d, got %d", armorCollapseExtendForLevel(3), compiled.Armor.CollapseDuration)
	}
	if compiled.Armor.RuinAmp != armorRuinAmpForLevel(5) {
		t.Fatalf("expected ruin amp %.2f, got %.2f", armorRuinAmpForLevel(5), compiled.Armor.RuinAmp)
	}
	if compiled.Crit.SkinnerChance != critSkinnerChanceForLevel(4) {
		t.Fatalf("expected skinner chance %.2f, got %.2f", critSkinnerChanceForLevel(4), compiled.Crit.SkinnerChance)
	}
	if compiled.Crit.FinalCutTrigger != int64(critFinalCutCountForLevel(1)) {
		t.Fatalf("expected final cut trigger %d, got %d", critFinalCutCountForLevel(1), compiled.Crit.FinalCutTrigger)
	}
	if compiled.Crit.DeathEcstasyMult != critDeathEcstasyMultForLevel(5)+2 {
		t.Fatalf("expected death ecstasy mult %.2f, got %.2f", critDeathEcstasyMultForLevel(5)+2, compiled.Crit.DeathEcstasyMult)
	}
	if compiled.Crit.DoomMarkCount != critDoomMarkCountForLevel(2) {
		t.Fatalf("expected doom mark count %d, got %d", critDoomMarkCountForLevel(2), compiled.Crit.DoomMarkCount)
	}
}

func TestApplyBossPartDamageDeltaUpdatesBossCurrentHPIncrementally(t *testing.T) {
	boss := &Boss{
		CurrentHP: 150,
		Parts: []BossPart{
			{X: 0, Y: 0, CurrentHP: 100, MaxHP: 100, Alive: true},
			{X: 1, Y: 0, CurrentHP: 50, MaxHP: 50, Alive: true},
		},
	}
	part := &boss.Parts[0]

	beforeHP, actualDamage, partJustDied := applyBossPartDamageDelta(boss, part, 30)
	if beforeHP != 100 || actualDamage != 30 || partJustDied {
		t.Fatalf("expected before=100 damage=30 alive, got before=%d damage=%d died=%v", beforeHP, actualDamage, partJustDied)
	}
	if boss.CurrentHP != 120 || part.CurrentHP != 70 {
		t.Fatalf("expected incremental hp update to 120/70, got boss=%d part=%d", boss.CurrentHP, part.CurrentHP)
	}
}

func TestApplyBossPartDamageDeltaCapsOverflowDamage(t *testing.T) {
	boss := &Boss{
		CurrentHP: 25,
		Parts: []BossPart{
			{X: 0, Y: 0, CurrentHP: 25, MaxHP: 25, Alive: true},
		},
	}
	part := &boss.Parts[0]

	beforeHP, actualDamage, partJustDied := applyBossPartDamageDelta(boss, part, 99)
	if beforeHP != 25 || actualDamage != 25 || !partJustDied {
		t.Fatalf("expected before=25 damage=25 died=true, got before=%d damage=%d died=%v", beforeHP, actualDamage, partJustDied)
	}
	if boss.CurrentHP != 0 || part.CurrentHP != 0 || part.Alive {
		t.Fatalf("expected boss/part hp to reach zero, got boss=%d part=%d alive=%v", boss.CurrentHP, part.CurrentHP, part.Alive)
	}
}

func TestCompileTalentSetCollectsEnabledTriggerHandlersInStableOrder(t *testing.T) {
	compiled := compileTalentSet(&TalentState{
		Talents: map[string]int{
			"normal_core":        1,
			"armor_auto_strike":  1,
			"crit_bleed":         1,
			"crit_doom_judgment": 1,
		},
	})

	if len(compiled.triggers) != 4 {
		t.Fatalf("expected 4 compiled triggers, got %d", len(compiled.triggers))
	}
	if compiled.triggerNames[0] != "normal_core" {
		t.Fatalf("expected first trigger normal_core, got %+v", compiled.triggerNames)
	}
	if compiled.triggerNames[1] != "armor_auto_strike" {
		t.Fatalf("expected second trigger armor_auto_strike, got %+v", compiled.triggerNames)
	}
	if compiled.triggerNames[2] != "crit_bleed" || compiled.triggerNames[3] != "crit_doom_judgment" {
		t.Fatalf("expected crit triggers to keep declaration order, got %+v", compiled.triggerNames)
	}
}
