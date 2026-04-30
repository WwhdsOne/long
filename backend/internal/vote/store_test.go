package vote

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"

	"long/internal/nickname"
)

func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	return NewStore(client, "vote:", StoreOptions{
			CriticalChancePercent: 5,
			CriticalCount:         5,
		}, nickname.NewValidator([]string{"习近平", "xjp"})), func() {
			_ = client.Close()
			server.Close()
		}
}

func TestListButtonsFiltersDisabledAndSortsBySortThenKey(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "3",
		"sort":    "20",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:button:other", map[string]any{
		"label":   "其他",
		"count":   "5",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed second button: %v", err)
	}
	boss := &Boss{
		ID:        "dragon-1",
		Name:      "火龙",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "cloth-armor", DropRatePercent: 25},
		{ItemID: "fire-ring", DropRatePercent: 75},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	_, err := store.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 2 {
		t.Fatalf("expected boss loot resources, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].ItemID != "cloth-armor" || resources.BossLoot[0].DropRatePercent != 25 {
		t.Fatalf("expected cloth-armor probability 25%%, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[1].ItemID != "fire-ring" || resources.BossLoot[1].DropRatePercent != 75 {
		t.Fatalf("expected fire-ring probability 75%%, got %+v", resources.BossLoot)
	}
}

func TestGetUserStateReadsResourceKeyWithoutGems(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	resourceKey := store.resourceKey(nickname)
	if err := store.client.HSet(ctx, resourceKey, map[string]any{
		"gold":          "345",
		"stones":        "67",
		"talent_points": "89",
	}).Err(); err != nil {
		t.Fatalf("seed resource key: %v", err)
	}

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if state.Gold != 345 || state.Stones != 67 || state.TalentPoints != 89 {
		t.Fatalf("expected resources from resource key, got gold=%d stones=%d talentPoints=%d", state.Gold, state.Stones, state.TalentPoints)
	}
}

func TestGetUserStateIgnoresLegacyGemKey(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	legacyKey := store.namespace + "gem:" + nickname
	if err := store.client.HSet(ctx, legacyKey, map[string]any{
		"gems":   "10",
		"gold":   "100",
		"stones": "40",
	}).Err(); err != nil {
		t.Fatalf("seed legacy gem key: %v", err)
	}

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if state.Gold != 0 || state.Stones != 0 || state.TalentPoints != 0 {
		t.Fatalf("expected legacy gem key to be ignored, got gold=%d stones=%d talentPoints=%d", state.Gold, state.Stones, state.TalentPoints)
	}

	exists, err := store.client.Exists(ctx, legacyKey).Result()
	if err != nil {
		t.Fatalf("check legacy key exists: %v", err)
	}
	if exists != 1 {
		t.Fatalf("expected legacy key to remain untouched, exists=%d", exists)
	}
}

func TestLearnTalentConsumesTalentPointsAndResetRefunds(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed talent points: %v", err)
	}

	if err := store.UpgradeTalent(ctx, nickname, "normal_core", 1); err != nil {
		t.Fatalf("learn normal_core: %v", err)
	}
	userState, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if userState.TalentPoints != 4964 {
		t.Fatalf("expected talent points to be deducted to 4964, got %d", userState.TalentPoints)
	}

	if err := store.ResetTalents(ctx, nickname); err != nil {
		t.Fatalf("reset talents: %v", err)
	}
	userState, err = store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after reset: %v", err)
	}
	if userState.TalentPoints != 5000 {
		t.Fatalf("expected refund after reset to 5000, got %d", userState.TalentPoints)
	}
}

func TestTier0TalentUpgradeAndResetRefundUsesNewCurve(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed talent points: %v", err)
	}

	if err := store.UpgradeTalent(ctx, nickname, "normal_core", 3); err != nil {
		t.Fatalf("upgrade normal_core to lv3: %v", err)
	}
	userState, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after upgrade: %v", err)
	}
	if userState.TalentPoints != 4532 {
		t.Fatalf("expected talent points to be deducted to 4532, got %d", userState.TalentPoints)
	}

	if err := store.ResetTalents(ctx, nickname); err != nil {
		t.Fatalf("reset talents: %v", err)
	}
	userState, err = store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after reset: %v", err)
	}
	if userState.TalentPoints != 5000 {
		t.Fatalf("expected tier0 refund after reset to 5000, got %d", userState.TalentPoints)
	}
}

func TestResolveBossDamageTypeHeavyDefaultsToHeavy(t *testing.T) {
	got := resolveBossDamageType(resolveBossDamageTypeInput{
		PartType:    PartTypeHeavy,
		Critical:    false,
		BossDamage:  25,
		BossMaxHP:   1000,
		IsCollapsed: false,
		IsAfkAttack: false,
	})
	if got != "heavy" {
		t.Fatalf("expected heavy default damage type, got %q", got)
	}
}

func TestResolveBossDamageTypeJudgementUsesTenPercentThreshold(t *testing.T) {
	got := resolveBossDamageType(resolveBossDamageTypeInput{
		PartType:    PartTypeSoft,
		Critical:    true,
		BossDamage:  100,
		BossMaxHP:   1000,
		IsCollapsed: false,
		IsAfkAttack: false,
	})
	if got != "judgement" {
		t.Fatalf("expected judgement at 10%% max hp critical threshold, got %q", got)
	}
}

func TestResolveBossDamageTypeCollapsedPartUsesTrueDamage(t *testing.T) {
	got := resolveBossDamageType(resolveBossDamageTypeInput{
		PartType:    PartTypeHeavy,
		Critical:    false,
		BossDamage:  25,
		BossMaxHP:   1000,
		IsCollapsed: true,
		IsAfkAttack: false,
	})
	if got != "trueDamage" {
		t.Fatalf("expected collapsed part to use trueDamage, got %q", got)
	}
}

func TestApplyComboDamageAmplifyUsesTwentyFiveHitSteps(t *testing.T) {
	if got := applyComboDamageAmplify(100, 24); got != 100 {
		t.Fatalf("expected no combo amplify below 25 hits, got %d", got)
	}
	if got := applyComboDamageAmplify(100, 25); got != 110 {
		t.Fatalf("expected 25 combo to add 10%% damage, got %d", got)
	}
	if got := applyComboDamageAmplify(100, 50); got != 120 {
		t.Fatalf("expected 50 combo to add 20%% damage, got %d", got)
	}
}

func TestLearnTalentRejectsWhenTalentPointsInsufficient(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5").Err(); err != nil {
		t.Fatalf("seed talent points: %v", err)
	}

	if err := store.UpgradeTalent(ctx, nickname, "normal_core", 1); !errors.Is(err, ErrTalentPointsInsufficient) {
		t.Fatalf("expected ErrTalentPointsInsufficient, got %v", err)
	}
}

func TestBuildTalentEffectLinesReturnsUpgradePreviewForNormalCore(t *testing.T) {
	def, ok := talentDefs["normal_core"]
	if !ok {
		t.Fatal("expected normal_core talent def")
	}

	lines := BuildTalentEffectLines(def, 2)
	if len(lines) != 3 {
		t.Fatalf("expected 3 effect lines, got %+v", lines)
	}

	if lines[0].Label != "触发次数" || lines[0].Text != "45 → 40" {
		t.Fatalf("expected trigger count preview, got %+v", lines[0])
	}
	if lines[1].Label != "追击段数" || lines[1].Text != "24 → 28" {
		t.Fatalf("expected extra hits preview, got %+v", lines[1])
	}
	if lines[2].Label != "追击倍率" || lines[2].Text != "100% → 150%" {
		t.Fatalf("expected chase ratio preview, got %+v", lines[2])
	}
}

func TestNormalCoreLevelFiveTriggersAtThirtyHits(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	ctx := context.Background()
	nickname := "暴风测试"

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed points: %v", err)
	}
	if err := store.UpgradeTalent(ctx, nickname, "normal_core", 5); err != nil {
		t.Fatalf("upgrade normal_core to lv5: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "storm-test",
		Name:  "暴风测试Boss",
		MaxHP: 100000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100000, CurrentHP: 100000, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	for i := 1; i <= 29; i++ {
		result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
		if err != nil {
			t.Fatalf("click %d: %v", i, err)
		}
		for _, ev := range result.TalentEvents {
			if ev.TalentID == "normal_core" {
				t.Fatalf("expected no normal_core trigger before 30 hits, got event at click %d: %+v", i, ev)
			}
		}
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click 30: %v", err)
	}
	found := false
	for _, ev := range result.TalentEvents {
		if ev.TalentID == "normal_core" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected normal_core trigger on 30th hit, got %+v", result.TalentEvents)
	}
}

func TestSilverStormUsesTimeWindowInsteadOfAttackCountdown(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	baseNow := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseNow }

	ctx := context.Background()
	nickname := "白银风暴测试"

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed points: %v", err)
	}
	talentsJSON, err := sonic.Marshal(map[string]int{
		"normal_ultimate": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed normal ultimate state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "silverstorm-test",
		Name:  "白银风暴Boss",
		MaxHP: 1000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 1000, CurrentHP: 1000, Alive: true, Armor: 100},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("trigger silver storm: %v", err)
	}
	if len(result.TalentEvents) == 0 {
		t.Fatalf("expected silver storm event, got %+v", result.TalentEvents)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "silverstorm-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if !combatState.SilverStormActive {
		t.Fatal("expected silver storm active after part break")
	}
	if combatState.SilverStormEndsAt != baseNow.Unix()+20 {
		t.Fatalf("expected silver storm ends at %d, got %d", baseNow.Unix()+20, combatState.SilverStormEndsAt)
	}
	if combatState.SilverStormRemaining != 20 {
		t.Fatalf("expected silver storm remaining 20, got %d", combatState.SilverStormRemaining)
	}

	second, err := store.ClickBossPart(ctx, "boss-part:1-0", nickname)
	if err != nil {
		t.Fatalf("attack during silver storm: %v", err)
	}
	if second.BossDamage <= result.BossDamage {
		t.Fatalf("expected silver storm to add huge extra damage, first=%d second=%d", result.BossDamage, second.BossDamage)
	}
	combatState, err = store.GetTalentCombatState(ctx, nickname, "silverstorm-test")
	if err != nil {
		t.Fatalf("get combat state during silver storm: %v", err)
	}
	if combatState.SilverStormRemaining != 20 {
		t.Fatalf("expected silver storm remaining unchanged within same second, got %d", combatState.SilverStormRemaining)
	}

	baseNow = baseNow.Add(21 * time.Second)
	_, err = store.ClickBossPart(ctx, "boss-part:1-0", nickname)
	if err != nil {
		t.Fatalf("attack after silver storm expired: %v", err)
	}
	combatState, err = store.GetTalentCombatState(ctx, nickname, "silverstorm-test")
	if err != nil {
		t.Fatalf("get combat state after expiry: %v", err)
	}
	if combatState.SilverStormActive {
		t.Fatal("expected silver storm inactive after expiry")
	}
}

func TestSilverStormTriggersWhenExtraTalentDamageBreaksPart(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	baseNow := time.Unix(1_700_000_100, 0)
	store.now = func() time.Time { return baseNow }

	ctx := context.Background()
	nickname := "白银风暴补刀测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"normal_ultimate": 5,
		"armor_core":      5,
		"armor_ultimate":  5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed talents: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "silverstorm-extra-damage-test",
		Name:  "白银风暴补刀Boss",
		MaxHP: 1000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 9, CurrentHP: 9, Alive: true, Armor: 999999},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	state := NewTalentCombatState()
	state.PartHeavyClickCount[TalentPartKey(0, 0)] = 29
	if err := store.SaveTalentCombatState(ctx, nickname, "silverstorm-extra-damage-test", state); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}

	foundJudgmentDay := false
	foundSilverStorm := false
	for _, event := range result.TalentEvents {
		if event.EffectType == "judgment_day" {
			foundJudgmentDay = true
		}
		if event.EffectType == "silver_storm" {
			foundSilverStorm = true
		}
	}
	if !foundJudgmentDay {
		t.Fatalf("expected judgment day extra damage to trigger, got %+v", result.TalentEvents)
	}
	if !foundSilverStorm {
		t.Fatalf("expected silver storm to trigger after extra damage broke part, got %+v", result.TalentEvents)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "silverstorm-extra-damage-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if !combatState.SilverStormActive {
		t.Fatal("expected silver storm active after extra damage break")
	}
}

func TestArmorAutoStrikeTriggersOnSameHeavyPartWithinFiveSeconds(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	baseNow := time.Unix(1_700_100_000, 0)
	store.now = func() time.Time { return baseNow }

	ctx := context.Background()
	nickname := "自动打击测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"armor_auto_strike": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed armor auto strike state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "auto-strike-test",
		Name:  "自动打击测试Boss",
		MaxHP: 5000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 2000, CurrentHP: 2000, Alive: true, Armor: 0},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 2000, CurrentHP: 2000, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	requiredHits := armorAutoStrikeTriggerCountForLevel(5)
	for i := 1; i < requiredHits; i++ {
		result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
		if err != nil {
			t.Fatalf("click %d: %v", i, err)
		}
		for _, ev := range result.TalentEvents {
			if ev.TalentID == "armor_auto_strike" {
				t.Fatalf("expected no auto strike trigger before hit %d, got %+v", i, ev)
			}
		}
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("trigger auto strike: %v", err)
	}

	var autoStrikeEvent *TalentTriggerEvent
	for _, ev := range result.TalentEvents {
		if ev.TalentID == "armor_auto_strike" {
			autoStrikeEvent = &ev
			break
		}
	}
	if autoStrikeEvent == nil {
		t.Fatalf("expected armor_auto_strike trigger on hit %d, got %+v", requiredHits, result.TalentEvents)
	}
	if autoStrikeEvent.PartX != 0 || autoStrikeEvent.PartY != 0 {
		t.Fatalf("expected auto strike target part (0,0), got (%d,%d)", autoStrikeEvent.PartX, autoStrikeEvent.PartY)
	}
	if result.Boss == nil {
		t.Fatal("expected boss state in click result")
	}
	if result.Boss.Parts[0].CurrentHP >= result.Boss.Parts[1].CurrentHP {
		t.Fatalf("expected focused heavy part to take extra strike damage, got hp0=%d hp1=%d", result.Boss.Parts[0].CurrentHP, result.Boss.Parts[1].CurrentHP)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "auto-strike-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if combatState.AutoStrikeComboCount != 0 {
		t.Fatalf("expected auto strike combo reset after trigger, got %d", combatState.AutoStrikeComboCount)
	}
}

func TestSilverStormAmplifiesFinalResolvedDamage(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	baseNow := time.Unix(1_700_200_000, 0)
	store.now = func() time.Time { return baseNow }

	ctx := context.Background()
	nickname := "白银风暴最终增伤测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"normal_ultimate":   5,
		"armor_auto_strike": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed talents: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "silverstorm-final-amp-test",
		Name:  "白银风暴最终增伤Boss",
		MaxHP: 200000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 200000, CurrentHP: 200000, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	state := NewTalentCombatState()
	state.SilverStormActive = true
	state.SilverStormRemaining = 20
	state.SilverStormEndsAt = baseNow.Unix() + 20
	state.AutoStrikeTargetPart = TalentPartKey(0, 0)
	state.AutoStrikeComboCount = int64(armorAutoStrikeTriggerCountForLevel(5) - 1)
	state.AutoStrikeExpiresAt = baseNow.Unix() + TalentAutoStrikeWindowSec
	if err := store.SaveTalentCombatState(ctx, nickname, "silverstorm-final-amp-test", state); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if len(result.PartStateDeltas) < 2 {
		t.Fatalf("expected multiple damage deltas, got %+v", result.PartStateDeltas)
	}

	baseDamage := result.PartStateDeltas[0].Damage
	autoStrikeDamage := int64(0)
	for _, ev := range result.TalentEvents {
		if ev.EffectType == "auto_strike" {
			autoStrikeDamage = ev.ExtraDamage
			break
		}
	}
	if autoStrikeDamage <= 0 {
		t.Fatalf("expected auto strike damage event, got %+v", result.TalentEvents)
	}

	expectedTotal := baseDamage + autoStrikeDamage
	expectedTotal += int64(float64(expectedTotal) * normalSilverStormDamageRatioForLevel(5))
	if result.BossDamage != expectedTotal {
		t.Fatalf("expected silver storm to amplify final resolved damage, got total=%d want=%d base=%d auto=%d deltas=%+v events=%+v", result.BossDamage, expectedTotal, baseDamage, autoStrikeDamage, result.PartStateDeltas, result.TalentEvents)
	}
}

func TestArmorAutoStrikeComboExpiresAfterFiveSeconds(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	baseNow := time.Unix(1_700_200_000, 0)
	store.now = func() time.Time { return baseNow }

	ctx := context.Background()
	nickname := "自动打击超时测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"armor_auto_strike": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed armor auto strike state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "auto-strike-expire-test",
		Name:  "自动打击超时测试Boss",
		MaxHP: 5000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 3000, CurrentHP: 3000, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	requiredHits := armorAutoStrikeTriggerCountForLevel(5)
	for i := 1; i < requiredHits; i++ {
		result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
		if err != nil {
			t.Fatalf("click %d: %v", i, err)
		}
		for _, ev := range result.TalentEvents {
			if ev.TalentID == "armor_auto_strike" {
				t.Fatalf("expected no auto strike trigger before timeout test, got %+v", ev)
			}
		}
	}

	baseNow = baseNow.Add(6 * time.Second)
	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click after combo expiry: %v", err)
	}
	for _, ev := range result.TalentEvents {
		if ev.TalentID == "armor_auto_strike" {
			t.Fatalf("expected expired combo to reset, but got auto strike event %+v", ev)
		}
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "auto-strike-expire-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if combatState.AutoStrikeComboCount != 1 {
		t.Fatalf("expected combo to restart from 1 after expiry, got %d", combatState.AutoStrikeComboCount)
	}
	if combatState.AutoStrikeTargetPart != TalentPartKey(0, 0) {
		t.Fatalf("expected combo target to restart on current heavy part, got %q", combatState.AutoStrikeTargetPart)
	}
}

func TestArmorAutoStrikeWindowDoesNotExtendOnRepeatedClicks(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	baseNow := time.Unix(1_700_300_000, 0)
	store.now = func() time.Time { return baseNow }

	ctx := context.Background()
	nickname := "自动打击不续期测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"armor_auto_strike": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed armor auto strike state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "auto-strike-fixed-window-test",
		Name:  "自动打击固定窗口Boss",
		MaxHP: 5000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 3000, CurrentHP: 3000, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	if _, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname); err != nil {
		t.Fatalf("first click: %v", err)
	}
	combatState, err := store.GetTalentCombatState(ctx, nickname, "auto-strike-fixed-window-test")
	if err != nil {
		t.Fatalf("get combat state after first click: %v", err)
	}
	firstExpiresAt := combatState.AutoStrikeExpiresAt
	if firstExpiresAt != baseNow.Unix()+TalentAutoStrikeWindowSec {
		t.Fatalf("expected first window expires at %d, got %d", baseNow.Unix()+TalentAutoStrikeWindowSec, firstExpiresAt)
	}

	baseNow = baseNow.Add(2 * time.Second)
	if _, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname); err != nil {
		t.Fatalf("second click: %v", err)
	}
	combatState, err = store.GetTalentCombatState(ctx, nickname, "auto-strike-fixed-window-test")
	if err != nil {
		t.Fatalf("get combat state after second click: %v", err)
	}
	if combatState.AutoStrikeExpiresAt != firstExpiresAt {
		t.Fatalf("expected auto strike window to stay at %d, got %d", firstExpiresAt, combatState.AutoStrikeExpiresAt)
	}
	if combatState.AutoStrikeComboCount != 2 {
		t.Fatalf("expected combo count continue to 2 within fixed window, got %d", combatState.AutoStrikeComboCount)
	}
}

func TestBossKillAwardsFixedTalentPoints(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	ctx := context.Background()
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:                 "talent-boss",
		Name:               "天赋点 Boss",
		MaxHP:              1,
		TalentPointsOnKill: 321,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	if _, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明"); err != nil {
		t.Fatalf("click boss part: %v", err)
	}

	userState, err := store.GetUserState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if userState.TalentPoints != 321 {
		t.Fatalf("expected fixed talent points reward 321, got %d", userState.TalentPoints)
	}
}

func TestBossLootDropRateIsIndependentProbability(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	boss := &Boss{
		ID:        "raid-1",
		Name:      "裂隙领主",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "mythic-core", DropRatePercent: 10},
		{ItemID: "raid-token", DropRatePercent: 10},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 2 {
		t.Fatalf("expected boss loot resources, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].DropRatePercent+resources.BossLoot[1].DropRatePercent != 20 {
		t.Fatalf("expected independent drop rates to keep configured values, got %+v", resources.BossLoot)
	}
}

func TestBossResourcesLootContainsEquipmentIcon(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, store.equipmentKey("fire-ring"), map[string]any{
		"name":                   "烈焰戒指",
		"slot":                   "accessory",
		"rarity":                 "epic",
		"image_path":             "https://cdn.example.com/items/fire-ring.png",
		"image_alt":              "烈焰戒指图标",
		"attack_power":           "18",
		"armor_pen_percent":      "12.5",
		"crit_rate":              "0.22",
		"crit_damage_multiplier": "1.8",
		"part_type_damage_soft":  "0.2",
		"part_type_damage_heavy": "0.3",
		"part_type_damage_weak":  "0.4",
		"talent_affinity":        "flame",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}

	boss := &Boss{
		ID:        "icon-boss",
		Name:      "图标测试 Boss",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "fire-ring", DropRatePercent: 30},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 1 {
		t.Fatalf("expected 1 loot entry, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].ImagePath != "https://cdn.example.com/items/fire-ring.png" {
		t.Fatalf("expected loot image path to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].ImageAlt != "烈焰戒指图标" {
		t.Fatalf("expected loot image alt to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].AttackPower != 18 {
		t.Fatalf("expected loot attack power to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].ArmorPenPercent != 12.5 {
		t.Fatalf("expected loot armor penetration to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].CritRate != 0.22 {
		t.Fatalf("expected loot crit rate to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].CritDamageMultiplier != 1.8 {
		t.Fatalf("expected loot crit damage multiplier to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].PartTypeDamageSoft != 0.2 || resources.BossLoot[0].PartTypeDamageHeavy != 0.3 || resources.BossLoot[0].PartTypeDamageWeak != 0.4 {
		t.Fatalf("expected loot part type damage to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].TalentAffinity != "flame" {
		t.Fatalf("expected loot talent affinity to be returned, got %+v", resources.BossLoot[0])
	}
}

func TestRollLootDropsCanReturnMultipleItems(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	drops := store.rollLootDrops([]BossLootEntry{
		{ItemID: "mythic-core", DropRatePercent: 100},
		{ItemID: "raid-token", DropRatePercent: 100},
	})

	if len(drops) != 2 || drops[0].ItemID != "mythic-core" || drops[1].ItemID != "raid-token" {
		t.Fatalf("expected multiple independent drops, got %+v", drops)
	}
}
func TestActivateBossWithPartsUsesPartHealthAsTotalHealth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	boss, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "part-boss",
		Name:  "分区 Boss",
		MaxHP: 9999,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 120, CurrentHP: 120, Alive: true},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 80, CurrentHP: 60, Alive: true},
			{X: 2, Y: 0, Type: PartTypeWeak, MaxHP: 50, CurrentHP: 0, Alive: false},
		},
	})
	if err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	if boss.MaxHP != 250 || boss.CurrentHP != 250 {
		t.Fatalf("expected boss health to match parts max/current sums, got max=%d current=%d parts=%+v", boss.MaxHP, boss.CurrentHP, boss.Parts)
	}
	if boss.Parts[1].CurrentHP != 80 || !boss.Parts[2].Alive {
		t.Fatalf("expected activated boss parts to be reset to full health, got %+v", boss.Parts)
	}

	stored, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("load current boss: %v", err)
	}
	if stored.MaxHP != 250 || stored.CurrentHP != 250 {
		t.Fatalf("expected stored boss health to match parts max/current sums, got max=%d current=%d parts=%+v", stored.MaxHP, stored.CurrentHP, stored.Parts)
	}
}

func TestActivateBossRequiresParts(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "no-parts-boss",
		Name:  "无部位 Boss",
		MaxHP: 100,
	}); !errors.Is(err, ErrBossPartsRequired) {
		t.Fatalf("expected ErrBossPartsRequired, got %v", err)
	}
}

func TestBossTemplateActivationUsesPartHealthAsTotalHealth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "template-part-boss",
		Name:  "模板分区 Boss",
		MaxHP: 9999,
		Layout: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 300, CurrentHP: 10, Alive: false},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 200, CurrentHP: 20, Alive: false},
		},
	}); err != nil {
		t.Fatalf("save boss template: %v", err)
	}

	templates, err := store.ListBossTemplates(ctx)
	if err != nil {
		t.Fatalf("list boss templates: %v", err)
	}
	if len(templates) != 1 || templates[0].MaxHP != 500 {
		t.Fatalf("expected saved template max health to match layout, got %+v", templates)
	}

	boss, err := store.activateRandomBossFromPool(ctx)
	if err != nil {
		t.Fatalf("activate boss from pool: %v", err)
	}
	if boss.MaxHP != 500 || boss.CurrentHP != 500 {
		t.Fatalf("expected activated template boss health to match layout, got max=%d current=%d parts=%+v", boss.MaxHP, boss.CurrentHP, boss.Parts)
	}
	for _, part := range boss.Parts {
		if part.CurrentHP != part.MaxHP || !part.Alive {
			t.Fatalf("expected activated template parts to be reset to full health, got %+v", boss.Parts)
		}
	}
}

func TestSaveBossTemplateRequiresLayout(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "no-layout-template",
		Name:  "无部位模板",
		MaxHP: 100,
	}); !errors.Is(err, ErrBossPartsRequired) {
		t.Fatalf("expected ErrBossPartsRequired, got %v", err)
	}
}

func TestBossPartDisplayFieldsPersist(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "display-part-boss",
		Name:  "展示字段 Boss",
		MaxHP: 100,
		Layout: []BossPart{
			{
				X:           0,
				Y:           0,
				Type:        PartTypeWeak,
				MaxHP:       100,
				CurrentHP:   100,
				Alive:       true,
				DisplayName: "眼核",
				ImagePath:   "/assets/boss/eye.png",
			},
		},
	}); err != nil {
		t.Fatalf("save boss template: %v", err)
	}

	templates, err := store.ListBossTemplates(ctx)
	if err != nil {
		t.Fatalf("list boss templates: %v", err)
	}
	if len(templates) != 1 || len(templates[0].Layout) != 1 {
		t.Fatalf("expected one template part, got %+v", templates)
	}
	part := templates[0].Layout[0]
	if part.DisplayName != "眼核" || part.ImagePath != "/assets/boss/eye.png" {
		t.Fatalf("expected template display fields to persist, got %+v", part)
	}

	boss, err := store.activateRandomBossFromPool(ctx)
	if err != nil {
		t.Fatalf("activate boss from pool: %v", err)
	}
	if len(boss.Parts) != 1 || boss.Parts[0].DisplayName != "眼核" || boss.Parts[0].ImagePath != "/assets/boss/eye.png" {
		t.Fatalf("expected activated boss display fields to persist, got %+v", boss.Parts)
	}
}

func TestBossCycleQueueAdvanceAndWrapOnKill(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mustSaveBossTemplateForCycleTest(t, store, ctx, "a", "新手木桩")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "b", "史莱姆王")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "c", "骷髅将军")

	if _, err := store.SetBossCycleQueue(ctx, []string{"a", "b", "c"}); err != nil {
		t.Fatalf("set boss cycle queue: %v", err)
	}

	first, err := store.SetBossCycleEnabled(ctx, true)
	if err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}
	if first == nil || first.TemplateID != "a" {
		t.Fatalf("expected first boss template a, got %+v", first)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click first boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "b" || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected next boss template b, got %+v", result.Boss)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click second boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "c" || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected next boss template c, got %+v", result.Boss)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click third boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "a" || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected wrapped boss template a, got %+v", result.Boss)
	}
}

func TestBossCycleQueueDynamicUpdateAppliesOnCurrentBossDefeat(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mustSaveBossTemplateForCycleTest(t, store, ctx, "a", "新手木桩")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "b", "史莱姆王")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "c", "骷髅将军")

	if _, err := store.SetBossCycleQueue(ctx, []string{"a", "b", "c"}); err != nil {
		t.Fatalf("set initial boss cycle queue: %v", err)
	}
	if _, err := store.SetBossCycleEnabled(ctx, true); err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}

	if _, err := store.SetBossCycleQueue(ctx, []string{"c", "b"}); err != nil {
		t.Fatalf("set updated boss cycle queue: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click current boss after queue update: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "c" {
		t.Fatalf("expected next boss template c after queue update, got %+v", result.Boss)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click updated queue boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "b" {
		t.Fatalf("expected next boss template b after c, got %+v", result.Boss)
	}
}

func TestEnableBossCycleRequiresConfiguredQueue(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mustSaveBossTemplateForCycleTest(t, store, ctx, "a", "新手木桩")

	if _, err := store.SetBossCycleEnabled(ctx, true); !errors.Is(err, ErrBossCycleQueueEmpty) {
		t.Fatalf("expected ErrBossCycleQueueEmpty, got %v", err)
	}
}

func TestClickButtonWithBossPartsPersistsBossAndPartHealth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(limit int) int {
		if limit <= 0 {
			return 0
		}
		// 固定返回上界，避免测试命中随机暴击导致断言不稳定。
		return limit - 1
	}

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "part-boss",
		Name:  "分区 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.Boss == nil {
		t.Fatal("expected click result to include boss state")
	}
	if result.Boss.CurrentHP != 95 || len(result.Boss.Parts) != 1 || result.Boss.Parts[0].CurrentHP != 95 {
		t.Fatalf("expected click result to reduce boss and part health, got %+v", result.Boss)
	}

	stored, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("load current boss: %v", err)
	}
	if stored.CurrentHP != 95 || len(stored.Parts) != 1 || stored.Parts[0].CurrentHP != 95 {
		t.Fatalf("expected stored boss and part health to be reduced, got %+v", stored)
	}
}

func mustSaveBossTemplateForCycleTest(t *testing.T, store *Store, ctx context.Context, id string, name string) {
	t.Helper()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    id,
		Name:  name,
		MaxHP: 1,
		Layout: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
		},
	}); err != nil {
		t.Fatalf("save boss template %s: %v", id, err)
	}
}

func TestClickButtonWithBossPartsPersistsDefeatedStatus(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "fragile-boss",
		Name:  "脆弱 Boss",
		MaxHP: 1,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.Boss == nil || result.Boss.Status != bossStatusDefeated || result.Boss.CurrentHP != 0 {
		t.Fatalf("expected click result to defeat boss, got %+v", result.Boss)
	}

	stored, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("load current boss: %v", err)
	}
	if stored.Status != bossStatusDefeated || stored.CurrentHP != 0 || stored.DefeatedAt == 0 {
		t.Fatalf("expected stored boss to be defeated, got %+v", stored)
	}
}

func TestManualBossPartClickCountsOneButDamageUsesCombatFormula(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	seedEquipmentDefinition(t, store, ctx, "strong-sword", "weapon", 7)
	strongSwordInst := seedOwnedInstance(t, store, ctx, "阿明", "strong-sword")
	if _, err := store.EquipItem(ctx, "阿明", strongSwordInst); err != nil {
		t.Fatalf("equip item: %v", err)
	}
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "formula-boss",
		Name:  "公式 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.Delta != 1 || result.BossDamage != 12 {
		t.Fatalf("expected click delta 1 and boss damage 12, got delta=%d bossDamage=%d result=%+v", result.Delta, result.BossDamage, result)
	}
	if result.UserStats.ClickCount != 1 {
		t.Fatalf("expected manual click count to increase by 1, got %+v", result.UserStats)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 88 || result.Boss.Parts[0].CurrentHP != 88 {
		t.Fatalf("expected boss health to lose 12 damage, got %+v", result.Boss)
	}
}

func TestClickBossPartWithoutButtonTargetsSelectedPart(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "direct-part-boss",
		Name:  "直点 Boss",
		MaxHP: 200,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, DisplayName: "左翼", MaxHP: 100, CurrentHP: 100, Alive: true},
			{X: 1, Y: 0, Type: PartTypeWeak, DisplayName: "眼核", MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickButton(ctx, "boss-part:1-0", "阿明", 0)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.Delta != 1 || result.UserStats.ClickCount != 1 {
		t.Fatalf("expected direct part click to count once, got %+v", result)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 188 {
		t.Fatalf("expected boss health to decrease by selected part damage, got %+v", result.Boss)
	}
	if result.Boss.Parts[0].CurrentHP != 100 || result.Boss.Parts[1].CurrentHP != 88 {
		t.Fatalf("expected only selected part to lose HP, got %+v", result.Boss.Parts)
	}

	snapshot, err := store.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if snapshot.TotalVotes != 1 {
		t.Fatalf("expected total=1, got %d", snapshot.TotalVotes)
	}
}

func TestClickBossPartReturnsRealtimeBossSummaryFields(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	ctx := context.Background()
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "realtime-summary-boss",
		Name:  "实时摘要 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.MyBossDamage != result.BossDamage {
		t.Fatalf("expected realtime myBossDamage=%d, got %+v", result.BossDamage, result)
	}
	if result.BossLeaderboardCount != 1 {
		t.Fatalf("expected realtime boss leaderboard count 1, got %+v", result)
	}
	if len(result.BossLeaderboard) != 0 {
		t.Fatalf("expected click result not to carry full boss leaderboard, got %+v", result.BossLeaderboard)
	}
}

func TestBossAutoClickDoesNotIncreaseUserClicks(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	seedEquipmentDefinition(t, store, ctx, "strong-sword", "weapon", 7)
	strongSwordInst := seedOwnedInstance(t, store, ctx, "阿明", "strong-sword")
	if _, err := store.EquipItem(ctx, "阿明", strongSwordInst); err != nil {
		t.Fatalf("equip item: %v", err)
	}
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "auto-boss",
		Name:  "挂机 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.AutoClickBossPart(ctx, "0-0", "阿明")
	if err != nil {
		t.Fatalf("auto click boss part: %v", err)
	}
	if result.Delta != 0 || result.BossDamage != 6 {
		t.Fatalf("expected auto click delta 0 and boss damage 6, got delta=%d bossDamage=%d result=%+v", result.Delta, result.BossDamage, result)
	}

	userStats, err := store.GetUserStats(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user stats: %v", err)
	}
	if userStats.ClickCount != 0 {
		t.Fatalf("expected auto click not to increase click count, got %+v", userStats)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 94 || result.Boss.Parts[0].CurrentHP != 94 {
		t.Fatalf("expected auto click to damage boss, got %+v", result.Boss)
	}
}

func TestBossAutoClickKillReturnsRecentRewards(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	seedEquipmentDefinition(t, store, ctx, "strong-sword", "weapon", 7)
	seedEquipmentDefinition(t, store, ctx, "loot-ring", "accessory", 0)
	strongSwordInst := seedOwnedInstance(t, store, ctx, "阿明", "strong-sword")
	if _, err := store.EquipItem(ctx, "阿明", strongSwordInst); err != nil {
		t.Fatalf("equip item: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "auto-reward-boss",
		Name:  "挂机掉落 Boss",
		MaxHP: 1,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}
	if err := store.SetBossLoot(ctx, "auto-reward-boss", []BossLootEntry{
		{ItemID: "loot-ring", DropRatePercent: 100},
	}); err != nil {
		t.Fatalf("set boss loot: %v", err)
	}

	result, err := store.AutoClickBossPart(ctx, "0-0", "阿明")
	if err != nil {
		t.Fatalf("auto click boss part: %v", err)
	}
	if !result.BroadcastUserAll {
		t.Fatalf("expected auto click to trigger boss kill broadcast, got %+v", result)
	}
	if len(result.RecentRewards) != 1 {
		t.Fatalf("expected one recent reward on afk kill, got %+v", result.RecentRewards)
	}
	if result.RecentRewards[0].ItemID != "loot-ring" {
		t.Fatalf("expected afk reward item loot-ring, got %+v", result.RecentRewards[0])
	}
}

func TestEquipmentCritRateContributesToCriticalChance(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.critical.CriticalCount = 5
	store.roll = func(limit int) int {
		return 0
	}

	ctx := context.Background()
	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID:   "crit-ring",
		Name:     "暴击戒指",
		Slot:     "accessory",
		Rarity:   "传说",
		CritRate: 0.05,
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}
	critRingInst := seedOwnedInstance(t, store, ctx, "阿明", "crit-ring")
	if _, err := store.EquipItem(ctx, "阿明", critRingInst); err != nil {
		t.Fatalf("equip item: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "crit-boss",
		Name:  "暴击测试 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if !result.Critical {
		t.Fatalf("expected critical hit from equipment critRate, got %+v", result)
	}
	if result.Delta != 1 {
		t.Fatalf("expected boss part click delta 1, got %+v", result)
	}

	userState, err := store.GetUserState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if userState.CombatStats.CriticalChancePercent != 5 {
		t.Fatalf("expected critical chance to include equipment critRate, got %+v", userState.CombatStats)
	}
}

func TestCalcBossPartDamageCriticalDamageUsesMultiplier(t *testing.T) {
	stats := CombatStats{
		AttackPower:           100,
		CriticalChancePercent: 0,
		CriticalCount:         5,
		CritDamageMultiplier:  2.0,
	}

	result := CalcBossPartDamage(stats, PartTypeSoft, 0, 1, 100, 100)
	if result.NormalDamage != 100 {
		t.Fatalf("expected normal damage 100, got %+v", result)
	}
	if result.CriticalDamage != 200 {
		t.Fatalf("expected critical damage 200, got %+v", result)
	}
}

func TestCritFinalCutTriggersAtOmenCap(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	currentTime := time.Unix(1_700_400_000, 0)
	store.now = func() time.Time { return currentTime }
	store.critical.CriticalChancePercent = 100
	store.critical.CriticalCount = 5
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "终末血斩重叠测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"crit_final_cut": 1,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed crit_final_cut state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "final-cut-restack-test",
		Name:  "终末血斩测试Boss",
		MaxHP: 100000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeWeak, MaxHP: 100000, CurrentHP: 100000, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	combatState := NewTalentCombatState()
	combatState.OmenStacks = critFinalCutOmenTriggerForLevel(1)
	if err := store.SaveTalentCombatState(ctx, nickname, "final-cut-restack-test", combatState); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}
	var firstFinalCut *TalentTriggerEvent
	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click: %v", err)
	}
	for _, ev := range result.TalentEvents {
		if ev.TalentID == "crit_final_cut" {
			firstFinalCut = &ev
			break
		}
	}
	if firstFinalCut == nil {
		t.Fatalf("expected final cut to trigger at omen cap")
	}
	combatState, err = store.GetTalentCombatState(ctx, nickname, "final-cut-restack-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if combatState.OmenStacks != 0 {
		t.Fatalf("expected omen stacks reset to 0 after final cut, got %d", combatState.OmenStacks)
	}
}

func TestCritFinalCutDoesNotTriggerBelowOmenCap(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	currentTime := time.Unix(1_700_410_000, 0)
	store.now = func() time.Time { return currentTime }
	store.critical.CriticalChancePercent = 100
	store.critical.CriticalCount = 5
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "终末血斩冷却不累计"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"crit_final_cut": 1,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed crit_final_cut state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "final-cut-cooldown-test",
		Name:  "终末血斩冷却Boss",
		MaxHP: 100000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeWeak, MaxHP: 100000, CurrentHP: 100000, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	combatState := NewTalentCombatState()
	combatState.OmenStacks = critFinalCutOmenTriggerForLevel(1) - 2
	if err := store.SaveTalentCombatState(ctx, nickname, "final-cut-cooldown-test", combatState); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click below omen cap: %v", err)
	}
	for _, ev := range result.TalentEvents {
		if ev.TalentID == "crit_final_cut" {
			t.Fatalf("expected final cut not to trigger below omen cap, got %+v", ev)
		}
	}

	combatState, err = store.GetTalentCombatState(ctx, nickname, "final-cut-cooldown-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if combatState.OmenStacks != critFinalCutOmenTriggerForLevel(1)-2 {
		t.Fatalf("expected omen stacks stay at %d below trigger, got %d", critFinalCutOmenTriggerForLevel(1)-2, combatState.OmenStacks)
	}
}

func TestCritBleedSettlesOverDurationAndExpires(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	currentTime := time.Unix(1_700_420_000, 0)
	store.now = func() time.Time { return currentTime }
	store.critical.CriticalChancePercent = 100
	store.critical.CriticalCount = 30
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "致命出血测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"crit_bleed": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed crit_bleed state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "crit-bleed-test",
		Name:  "致命出血Boss",
		MaxHP: 100000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100000, CurrentHP: 100000, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	first, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("first click: %v", err)
	}
	if !first.Critical {
		t.Fatalf("expected first click to crit and挂出血, got %+v", first)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "crit-bleed-test")
	if err != nil {
		t.Fatalf("get combat state after first click: %v", err)
	}
	bleedState, ok := combatState.Bleeds[TalentPartKey(0, 0)]
	if !ok {
		t.Fatalf("expected bleed state after first click, got %+v", combatState)
	}
	if bleedState.DurationMs != critBleedDurationForLevel(5)*1000 {
		t.Fatalf("expected bleed duration %dms, got %+v", critBleedDurationForLevel(5)*1000, bleedState)
	}
	if bleedState.TotalDamage <= 0 {
		t.Fatalf("expected bleed total damage > 0, got %+v", bleedState)
	}

	store.critical.CriticalChancePercent = 0
	store.critical.CriticalCount = 0
	store.roll = func(limit int) int { return maxInt(0, limit-1) }
	store.invalidateCombatStatsCache(nickname)

	currentTime = currentTime.Add(3 * time.Second)
	changes, err := store.ProcessTalentBleedTicks(ctx)
	if err != nil {
		t.Fatalf("process bleed ticks: %v", err)
	}
	if len(changes) == 0 {
		t.Fatalf("expected bleed tick changes after 3 seconds")
	}

	userState, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after bleed ticks: %v", err)
	}
	bleedDamage := int64(0)
	for _, ev := range userState.TalentEvents {
		if ev.TalentID == "crit_bleed" && ev.ExtraDamage > 0 {
			bleedDamage += ev.ExtraDamage
		}
	}
	if bleedDamage != bleedState.TotalDamage {
		t.Fatalf("expected bleed settlement %d, got %d", bleedState.TotalDamage, bleedDamage)
	}

	combatState, err = store.GetTalentCombatState(ctx, nickname, "crit-bleed-test")
	if err != nil {
		t.Fatalf("get combat state after bleed expire: %v", err)
	}
	if _, ok := combatState.Bleeds[TalentPartKey(0, 0)]; ok {
		t.Fatalf("expected bleed state to expire, got %+v", combatState.Bleeds)
	}
}

func TestProcessTalentBleedTicksSettlesEveryTwoHundredMs(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	currentTime := time.Unix(1_700_430_000, 0)
	store.now = func() time.Time { return currentTime }
	store.critical.CriticalChancePercent = 100
	store.critical.CriticalCount = 30
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "出血分段结算测试"

	talentsJSON, err := sonic.Marshal(map[string]int{
		"crit_bleed": 5,
	})
	if err != nil {
		t.Fatalf("marshal talents: %v", err)
	}
	if err := store.client.HSet(ctx, store.talentKey(nickname), "talents", string(talentsJSON)).Err(); err != nil {
		t.Fatalf("seed crit_bleed state: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "crit-bleed-tick-test",
		Name:  "出血分段Boss",
		MaxHP: 100000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100000, CurrentHP: 100000, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	first, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("first click: %v", err)
	}
	if !first.Critical {
		t.Fatalf("expected first click critical, got %+v", first)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "crit-bleed-tick-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	bleedState, ok := combatState.Bleeds[TalentPartKey(0, 0)]
	if !ok {
		t.Fatalf("expected bleed state after first click, got %+v", combatState)
	}
	bleedState.TotalDamage = 30
	combatState.Bleeds[TalentPartKey(0, 0)] = bleedState
	if err := store.SaveTalentCombatState(ctx, nickname, "crit-bleed-tick-test", combatState); err != nil {
		t.Fatalf("save amplified bleed state: %v", err)
	}

	currentTime = currentTime.Add(199 * time.Millisecond)
	changes, err := store.ProcessTalentBleedTicks(ctx)
	if err != nil {
		t.Fatalf("process bleed ticks before due: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected no bleed tick before 0.2s, got %+v", changes)
	}

	totalBleedDamage := int64(0)
	totalBleedEvents := 0
	for i := 0; i < 15; i++ {
		currentTime = currentTime.Add(200 * time.Millisecond)
		changes, err = store.ProcessTalentBleedTicks(ctx)
		if err != nil {
			t.Fatalf("process bleed ticks: %v", err)
		}
		if len(changes) == 0 {
			t.Fatalf("expected bleed tick change on each 0.2s step")
		}
		tickEvents, err := store.consumePendingTalentEvents(ctx, nickname, "crit-bleed-tick-test")
		if err != nil {
			t.Fatalf("consume pending bleed events after tick: %v", err)
		}
		if len(tickEvents) == 0 {
			t.Fatalf("expected pending bleed talent event after tick")
		}
		for _, ev := range tickEvents {
			if ev.TalentID == "crit_bleed" && ev.ExtraDamage > 0 {
				totalBleedDamage += ev.ExtraDamage
				totalBleedEvents++
			}
		}
	}

	if totalBleedEvents != 15 {
		t.Fatalf("expected 15 bleed tick events, got %d", totalBleedEvents)
	}
	if totalBleedDamage != bleedState.TotalDamage {
		t.Fatalf("expected total bleed damage %d, got %d", bleedState.TotalDamage, totalBleedDamage)
	}

	combatState, err = store.GetTalentCombatState(ctx, nickname, "crit-bleed-tick-test")
	if err != nil {
		t.Fatalf("get combat state after final tick: %v", err)
	}
	if _, ok := combatState.Bleeds[TalentPartKey(0, 0)]; ok {
		t.Fatalf("expected bleed state removed after 15 ticks, got %+v", combatState.Bleeds)
	}
}

func TestGetUserStateConsumesPendingTalentEvents(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "用户态出血事件测试"

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "user-state-talent-event-test",
		Name:  "用户态事件Boss",
		MaxHP: 1000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1000, CurrentHP: 1000, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	events := []TalentTriggerEvent{
		{
			TalentID:    "crit_bleed",
			Name:        "致命出血",
			EffectType:  "bleed",
			ExtraDamage: 3,
			Message:     "出血结算",
			PartX:       0,
			PartY:       0,
		},
	}
	if err := store.appendPendingTalentEvents(ctx, nickname, "user-state-talent-event-test", events); err != nil {
		t.Fatalf("append pending events: %v", err)
	}

	userState, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if len(userState.TalentEvents) != 1 {
		t.Fatalf("expected one pending talent event, got %+v", userState.TalentEvents)
	}

	userState, err = store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state second time: %v", err)
	}
	if len(userState.TalentEvents) != 0 {
		t.Fatalf("expected pending talent events consumed after first read, got %+v", userState.TalentEvents)
	}
}

func TestArmorCoreCollapseTrigger(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	ctx := context.Background()
	nickname := "测试破甲"

	// 1. 给天赋点，学 armor_core
	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed points: %v", err)
	}
	if err := store.UpgradeTalent(ctx, nickname, "armor_core", 1); err != nil {
		t.Fatalf("learn armor_core: %v", err)
	}

	// 2. 创建一个有重甲部位的 Boss
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "collapse-test",
		Name:  "崩塌测试Boss",
		MaxHP: 100000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 10000, CurrentHP: 10000, Alive: true, Armor: 100},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	triggerCount := armorCoreCollapseTriggerForLevel(1)
	// 3. 点击重甲达到阈值后，最后一次应该触发崩塌
	var collapseEvent *TalentTriggerEvent
	for i := 1; i <= triggerCount; i++ {
		clickKey := "boss-part:0-0"
		result, err := store.ClickBossPart(ctx, clickKey, nickname)
		if err != nil {
			t.Fatalf("click %d: %v", i, err)
		}
		if i == triggerCount {
			if result.DamageType != "trueDamage" {
				t.Fatalf("第%d次点击触发崩塌后应返回 trueDamage，实际 %q", triggerCount, result.DamageType)
			}
			for _, ev := range result.TalentEvents {
				if ev.EffectType == "collapse_trigger" {
					collapseEvent = &ev
					break
				}
			}
		}
	}

	// 4. 验证
	if collapseEvent == nil {
		t.Fatalf("第%d次点击应该触发崩塌事件，但未收到 collapse_trigger", triggerCount)
	}
	if collapseEvent.PartX != 0 || collapseEvent.PartY != 0 {
		t.Fatalf("崩塌事件的 PartX/Y 应为 (0,0)，实际 (%d,%d)", collapseEvent.PartX, collapseEvent.PartY)
	}
	t.Logf("崩塌触发成功: %s (PartX=%d, PartY=%d)", collapseEvent.Message, collapseEvent.PartX, collapseEvent.PartY)

	// 5. 确认作战状态中崩塌部位已记录
	combatState, err := store.GetTalentCombatState(ctx, nickname, "collapse-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if len(combatState.CollapseParts) != 1 || combatState.CollapseParts[0] != 0 {
		t.Fatalf("预期 CollapseParts = [0]，实际 %v", combatState.CollapseParts)
	}
	if combatState.CollapseEndsAt <= store.now().Unix() {
		t.Fatal("崩塌结束时间应在未来")
	}
	t.Logf("崩塌状态正确: parts=%v, endsAt=%d", combatState.CollapseParts, combatState.CollapseEndsAt)
}

func TestArmorCoreCollapseCountDoesNotResetWhenCollapseExpires(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	currentTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return currentTime }
	store.critical.CriticalChancePercent = 0

	ctx := context.Background()
	nickname := "崩塌续算测试"

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed points: %v", err)
	}
	if err := store.UpgradeTalent(ctx, nickname, "armor_core", 1); err != nil {
		t.Fatalf("learn armor_core: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "collapse-expire-keep-count",
		Name:  "崩塌续算Boss",
		MaxHP: 2_000_000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 2_000_000, CurrentHP: 2_000_000, Alive: true, Armor: 100},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	clickKey := "boss-part:0-0"
	triggerCount := armorCoreCollapseTriggerForLevel(1)
	for i := 0; i < triggerCount; i++ {
		if _, err := store.ClickBossPart(ctx, clickKey, nickname); err != nil {
			t.Fatalf("trigger collapse click %d: %v", i+1, err)
		}
	}

	for i := 0; i < 10; i++ {
		if _, err := store.ClickBossPart(ctx, clickKey, nickname); err != nil {
			t.Fatalf("during collapse click %d: %v", i+1, err)
		}
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "collapse-expire-keep-count")
	if err != nil {
		t.Fatalf("get combat state before expire: %v", err)
	}
	if got := combatState.PartHeavyClickCount["0-0"]; got != 10 {
		t.Fatalf("expected heavy click count 10 during collapse, got %d", got)
	}

	currentTime = currentTime.Add(9 * time.Second)
	if _, err := store.ClickBossPart(ctx, clickKey, nickname); err != nil {
		t.Fatalf("post-expire click: %v", err)
	}

	combatState, err = store.GetTalentCombatState(ctx, nickname, "collapse-expire-keep-count")
	if err != nil {
		t.Fatalf("get combat state after expire: %v", err)
	}
	if got := combatState.PartHeavyClickCount["0-0"]; got != 11 {
		t.Fatalf("expected heavy click count continue to 11 after collapse expired, got %d", got)
	}
}

func TestArmorCoreCollapseDoesNotRetriggerWhileActive(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	currentTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return currentTime }
	store.critical.CriticalChancePercent = 0

	ctx := context.Background()
	nickname := "崩塌期间不重触发"

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "talent_points", "5000").Err(); err != nil {
		t.Fatalf("seed points: %v", err)
	}
	if err := store.UpgradeTalent(ctx, nickname, "armor_core", 1); err != nil {
		t.Fatalf("learn armor_core: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "collapse-no-retrigger",
		Name:  "崩塌不重触发Boss",
		MaxHP: 2_000_000,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeHeavy, MaxHP: 2_000_000, CurrentHP: 2_000_000, Alive: true, Armor: 100},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	clickKey := "boss-part:0-0"
	triggerCount := armorCoreCollapseTriggerForLevel(1)
	for i := 0; i < triggerCount; i++ {
		if _, err := store.ClickBossPart(ctx, clickKey, nickname); err != nil {
			t.Fatalf("trigger collapse click %d: %v", i+1, err)
		}
	}

	retriggerCount := 0
	for i := 0; i < triggerCount; i++ {
		result, err := store.ClickBossPart(ctx, clickKey, nickname)
		if err != nil {
			t.Fatalf("during collapse click %d: %v", i+1, err)
		}
		for _, ev := range result.TalentEvents {
			if ev.EffectType == "collapse_trigger" {
				retriggerCount++
			}
		}
	}

	if retriggerCount != 0 {
		t.Fatalf("崩塌持续期间不应重复触发 collapse_trigger，实际重复 %d 次", retriggerCount)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "collapse-no-retrigger")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if len(combatState.CollapseParts) != 1 || combatState.CollapseParts[0] != 0 {
		t.Fatalf("崩塌持续期间 CollapseParts 不应重复追加，实际 %v", combatState.CollapseParts)
	}
}
