package core

import (
	"context"
	"testing"
	"time"
)

func TestMagicCoreHitsMainTargetAndAdjacentPart(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "魔法溅射测试"

	if err := store.saveTalentState(ctx, nickname, &TalentState{
		Talents: map[string]int{
			"magic_core":      5,
			"magic_amp":       5,
			"magic_resonance": 5,
			"magic_splash":    5,
		},
	}); err != nil {
		t.Fatalf("seed magic talents: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "magic-core-test",
		Name:  "魔法核心Boss",
		MaxHP: 400,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeArcane, DamageAffinity: PartDamageAffinityMagicOnly, MaxHP: 200, CurrentHP: 200, Alive: true, Armor: 0},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 200, CurrentHP: 200, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}

	foundMain := false
	foundSplash := false
	for _, ev := range result.TalentEvents {
		if ev.EffectType != "magic_burst" {
			continue
		}
		if ev.PartX == 0 && ev.PartY == 0 {
			foundMain = true
		}
		if ev.PartX == 1 && ev.PartY == 0 {
			foundSplash = true
		}
	}
	if !foundMain || !foundSplash {
		t.Fatalf("expected magic burst on main and adjacent parts, got %+v", result.TalentEvents)
	}

	if len(result.PartStateDeltas) < 2 {
		t.Fatalf("expected magic deltas on main and adjacent parts, got %+v", result.PartStateDeltas)
	}
}

func TestMagicOnlyPartDoesNotTakeNormalClickDamage(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "奥核普通点击免疫测试"

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "magic-only-test",
		Name:  "奥核Boss",
		MaxHP: 200,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeArcane, DamageAffinity: PartDamageAffinityMagicOnly, MaxHP: 200, CurrentHP: 200, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.BossDamage != 0 {
		t.Fatalf("expected zero normal damage, got %+v", result)
	}
	if result.DamageType != "physicalInvalid" {
		t.Fatalf("expected physicalInvalid damage type, got %+v", result)
	}
	if len(result.PartStateDeltas) != 1 || result.PartStateDeltas[0].Damage != 0 {
		t.Fatalf("expected zero-damage delta only, got %+v", result.PartStateDeltas)
	}
}

func TestMagicEchoWindowGuaranteesRepeatedProcsAndKeepsWindow(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	baseNow := time.Unix(1_700_200_000, 0)
	store.now = func() time.Time { return baseNow }
	store.critical.CriticalChancePercent = 0
	store.roll = func(limit int) int { return maxInt(0, limit-1) }

	ctx := context.Background()
	nickname := "魔法回响测试"

	if err := store.saveTalentState(ctx, nickname, &TalentState{
		Talents: map[string]int{
			"magic_core":      5,
			"magic_echo_mark": 5,
		},
	}); err != nil {
		t.Fatalf("seed magic talents: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "magic-echo-test",
		Name:  "魔法回响Boss",
		MaxHP: 300,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 300, CurrentHP: 300, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	state := NewTalentCombatState()
	state.MagicEchoTargetPart = TalentPartKey(0, 0)
	state.MagicEchoExpiresAt = baseNow.Unix() + TalentMagicEchoWindowSec
	state.MagicEchoCooldownEndsAt = state.MagicEchoExpiresAt + magicEchoCooldownForLevel(5)
	if err := store.SaveTalentCombatState(ctx, nickname, "magic-echo-test", state); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}

	foundMagic := false
	for _, ev := range result.TalentEvents {
		if ev.EffectType == "magic_burst" {
			foundMagic = true
			break
		}
	}
	if !foundMagic {
		t.Fatalf("expected guaranteed magic burst, got %+v", result.TalentEvents)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "magic-echo-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if combatState.MagicEchoExpiresAt != baseNow.Unix()+TalentMagicEchoWindowSec {
		t.Fatalf("expected echo active window retained, got %+v", combatState)
	}
	if combatState.MagicEchoCooldownEndsAt != baseNow.Unix()+TalentMagicEchoWindowSec+magicEchoCooldownForLevel(5) {
		t.Fatalf("expected echo cooldown end at %d, got %d", baseNow.Unix()+TalentMagicEchoWindowSec+magicEchoCooldownForLevel(5), combatState.MagicEchoCooldownEndsAt)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part again: %v", err)
	}
	foundMagic = false
	for _, ev := range result.TalentEvents {
		if ev.EffectType == "magic_burst" {
			foundMagic = true
			break
		}
	}
	if !foundMagic {
		t.Fatalf("expected second guaranteed magic burst inside echo window, got %+v", result.TalentEvents)
	}
}

func TestMagicEchoWindowAppliesToAllAttacks(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	baseNow := time.Unix(1_700_210_000, 0)
	store.now = func() time.Time { return baseNow }
	store.critical.CriticalChancePercent = 0
	store.roll = func(limit int) int { return maxInt(0, limit-1) }

	ctx := context.Background()
	nickname := "魔法回响锁目标测试"

	if err := store.saveTalentState(ctx, nickname, &TalentState{
		Talents: map[string]int{
			"magic_core":      5,
			"magic_echo_mark": 5,
		},
	}); err != nil {
		t.Fatalf("seed magic talents: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "magic-echo-retarget-test",
		Name:  "魔法回响锁目标Boss",
		MaxHP: 300,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 300, CurrentHP: 300, Alive: true, Armor: 0},
			{X: 1, Y: 0, Type: PartTypeSoft, MaxHP: 300, CurrentHP: 300, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	state := NewTalentCombatState()
	state.MagicEchoTargetPart = TalentPartKey(0, 0)
	state.MagicEchoExpiresAt = baseNow.Unix() + TalentMagicEchoWindowSec
	state.MagicEchoCooldownEndsAt = state.MagicEchoExpiresAt + magicEchoCooldownForLevel(5)
	if err := store.SaveTalentCombatState(ctx, nickname, "magic-echo-retarget-test", state); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:1-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	foundMagic := false
	for _, ev := range result.TalentEvents {
		if ev.EffectType == "magic_burst" {
			foundMagic = true
			break
		}
	}
	if !foundMagic {
		t.Fatalf("expected off-target click to also trigger guaranteed burst, got %+v", result.TalentEvents)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "magic-echo-retarget-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if combatState.MagicEchoTargetPart != TalentPartKey(0, 0) {
		t.Fatalf("expected echo target to remain locked, got %+v", combatState)
	}
	if combatState.MagicEchoExpiresAt <= baseNow.Unix() {
		t.Fatalf("expected echo window to remain active, got %+v", combatState)
	}
}

func TestMagicUltimateAccumulatesWithoutCooldown(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	baseNow := time.Unix(1_700_220_000, 0)
	store.now = func() time.Time { return baseNow }
	store.critical.CriticalChancePercent = 0
	store.roll = func(limit int) int { return 0 }

	ctx := context.Background()
	nickname := "魔法终极无冷却测试"

	if err := store.saveTalentState(ctx, nickname, &TalentState{
		Talents: map[string]int{
			"magic_core":     5,
			"magic_ultimate": 5,
		},
	}); err != nil {
		t.Fatalf("seed magic talents: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "magic-ultimate-no-cooldown-test",
		Name:  "魔法终极无冷却Boss",
		MaxHP: 300,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 300, CurrentHP: 300, Alive: true, Armor: 0},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	state := NewTalentCombatState()
	state.PartMagicTriggerCount[TalentPartKey(0, 0)] = 1
	if err := store.SaveTalentCombatState(ctx, nickname, "magic-ultimate-no-cooldown-test", state); err != nil {
		t.Fatalf("seed combat state: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", nickname)
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	foundMagic := false
	for _, ev := range result.TalentEvents {
		if ev.EffectType == "magic_burst" {
			foundMagic = true
			break
		}
	}
	if !foundMagic {
		t.Fatalf("expected magic burst during no-cooldown test, got %+v", result.TalentEvents)
	}

	combatState, err := store.GetTalentCombatState(ctx, nickname, "magic-ultimate-no-cooldown-test")
	if err != nil {
		t.Fatalf("get combat state: %v", err)
	}
	if got := combatState.PartMagicTriggerCount[TalentPartKey(0, 0)]; got != 2 {
		t.Fatalf("expected ultimate count to continue accumulating without cooldown, got %d", got)
	}
	if combatState.MagicUltimateCooldownAt != 0 {
		t.Fatalf("expected no ultimate cooldown timestamp, got %+v", combatState)
	}
}
