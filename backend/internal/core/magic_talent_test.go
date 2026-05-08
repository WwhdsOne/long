package core

import (
	"strings"
	"testing"
)

func TestMagicTreeDefinitionsAndDescriptions(t *testing.T) {
	defs := GetTreeTalents(TalentTreeMagic)
	if len(defs) != 10 {
		t.Fatalf("expected 10 magic talents, got %d", len(defs))
	}

	coreDef, ok := talentDefs["magic_core"]
	if !ok {
		t.Fatal("expected magic_core definition")
	}
	if coreDef.Cost != 12600 {
		t.Fatalf("expected magic_core cost 12600, got %d", coreDef.Cost)
	}

	ultimateDef, ok := talentDefs["magic_ultimate"]
	if !ok {
		t.Fatal("expected magic_ultimate definition")
	}
	if ultimateDef.Cost != 26000 {
		t.Fatalf("expected magic_ultimate cost 26000, got %d", ultimateDef.Cost)
	}

	description := TalentEffectDescription(coreDef)
	if !strings.Contains(description, "攻击力基础值") {
		t.Fatalf("expected magic core description mention attack base power, got %q", description)
	}
	if !strings.Contains(description, "全伤害加成") {
		t.Fatalf("expected magic core description mention all damage amplify, got %q", description)
	}

	labels := TalentTierCompletionBonusLabels(TalentTreeMagic)
	if len(labels) != 5 {
		t.Fatalf("expected 5 magic tier completion labels, got %d", len(labels))
	}
}

func TestCompileTalentSetBuildsMagicThresholds(t *testing.T) {
	compiled := compileTalentSet(&TalentState{
		Talents: map[string]int{
			"magic_core":        5,
			"magic_amp":         3,
			"magic_resonance":   2,
			"magic_splash":      4,
			"magic_focus":       3,
			"magic_echo_mark":   2,
			"magic_static_flux": 1,
			"magic_pierce":      5,
			"magic_chain_bound": 4,
			"magic_ultimate":    2,
		},
	})

	if compiled == nil {
		t.Fatal("expected compiled talent set")
	}
	if !compiled.IsTierFull(TalentTreeMagic, 1) {
		t.Fatalf("expected magic tier 1 full, got %+v", compiled.tierFull)
	}
	if compiled.Magic.ProcRate <= 0 {
		t.Fatalf("expected magic proc rate > 0, got %f", compiled.Magic.ProcRate)
	}
	if compiled.Magic.MainRatio <= 0 || compiled.Magic.SplashRatio <= 0 {
		t.Fatalf("expected magic damage ratios > 0, got %+v", compiled.Magic)
	}
	if compiled.Magic.EchoRequiredHits <= 0 || compiled.Magic.EchoCooldownSec <= 0 {
		t.Fatalf("expected echo thresholds initialized, got %+v", compiled.Magic)
	}
	if compiled.Magic.UltimateTriggerCount != 65 {
		t.Fatalf("expected magic ultimate trigger count 65 at lv2, got %+v", compiled.Magic)
	}
	if compiled.Magic.UltimateMainRatio != 52.4 {
		t.Fatalf("expected magic ultimate main ratio 52.4 at lv2, got %+v", compiled.Magic)
	}
	if compiled.Magic.UltimateCooldownSec != 0 {
		t.Fatalf("expected magic ultimate cooldown disabled, got %+v", compiled.Magic)
	}
}
