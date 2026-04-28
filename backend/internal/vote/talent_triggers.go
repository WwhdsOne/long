package vote

import (
	"fmt"
	"math"
)

type compiledTalentTrigger func(*talentTriggerContext)

type talentTriggerContext struct {
	boss               *Boss
	part               *BossPart
	nickname           string
	clickCount         int64
	baseDamage         int64
	isCritical         bool
	partIndex          int
	compiledTalents    *CompiledTalentSet
	combatState        *TalentCombatState
	now                int64
	totalExtra         int64
	events             []TalentTriggerEvent
	damageTypeOverride string
}

func (c *CompiledTalentSet) buildTriggerHandlers() ([]compiledTalentTrigger, []string) {
	if c == nil {
		return nil, nil
	}

	var handlers []compiledTalentTrigger
	var names []string
	appendTrigger := func(enabled bool, name string, handler compiledTalentTrigger) {
		if !enabled {
			return
		}
		handlers = append(handlers, handler)
		names = append(names, name)
	}

	appendTrigger(c.Has("normal_core"), "normal_core", applyNormalCoreTrigger)
	appendTrigger(c.Has("armor_core"), "armor_core", applyArmorCoreTrigger)
	appendTrigger(c.Has("armor_auto_strike"), "armor_auto_strike", applyArmorAutoStrikeTrigger)
	appendTrigger(c.Has("armor_ultimate"), "armor_ultimate", applyArmorUltimateTrigger)
	appendTrigger(c.Has("crit_bleed"), "crit_bleed", applyCritBleedTrigger)
	appendTrigger(c.Has("crit_final_cut"), "crit_final_cut", applyCritFinalCutTrigger)
	appendTrigger(c.Has("crit_death_ecstasy"), "crit_death_ecstasy", applyCritDeathEcstasyTrigger)
	appendTrigger(c.Has("crit_doom_judgment"), "crit_doom_judgment", applyCritDoomJudgmentTrigger)

	return handlers, names
}

func applyNormalCoreTrigger(tc *talentTriggerContext) {
	def, ok := talentDefs["normal_core"]
	if !ok {
		return
	}

	triggerCount := tc.compiledTalents.Normal.TriggerCount
	extraHits := tc.compiledTalents.Normal.ExtraHits
	chaseRatio := tc.compiledTalents.Normal.ChaseRatio

	partKey := TalentPartKey(tc.part.X, tc.part.Y)
	tc.combatState.PartStormComboCount[partKey]++
	if tc.combatState.PartStormComboCount[partKey] < triggerCount {
		return
	}

	burst := int64(math.Floor(float64(maxInt64(1, tc.baseDamage)) * chaseRatio * float64(maxInt64(1, extraHits))))
	if burst <= 0 {
		return
	}

	_, burst, _ = applyBossPartDamageDelta(tc.boss, tc.part, burst)
	tc.totalExtra += burst
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "normal_core", Name: def.Name, EffectType: def.EffectType,
		ExtraDamage: burst, Message: fmt.Sprintf("追击爆发 %d 段伤害", extraHits),
		PartX: tc.part.X, PartY: tc.part.Y,
	})
	if tc.compiledTalents.Has("normal_charge") {
		tc.combatState.PartStormComboCount[partKey] = int64(float64(triggerCount) * tc.compiledTalents.Normal.RetainPercent)
		return
	}
	tc.combatState.PartStormComboCount[partKey] = 0
}

func applyArmorCoreTrigger(tc *talentTriggerContext) {
	if tc.part.Type != PartTypeHeavy {
		return
	}

	partKey := TalentPartKey(tc.part.X, tc.part.Y)
	tc.combatState.PartHeavyClickCount[partKey]++
	collapseTrigger := tc.compiledTalents.Armor.CollapseTrigger
	if tc.combatState.PartHeavyClickCount[partKey] < collapseTrigger {
		return
	}

	cd := tc.compiledTalents.Armor.CollapseDuration
	tc.combatState.CollapseParts = append(tc.combatState.CollapseParts, tc.partIndex)
	tc.combatState.CollapseEndsAt = tc.now + cd
	tc.combatState.CollapseDuration = cd
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "armor_core", Name: "灭绝穿甲", EffectType: "collapse_trigger",
		Message: fmt.Sprintf("结构崩塌！护甲归零 %d 秒", cd),
		PartX:   tc.part.X,
		PartY:   tc.part.Y,
	})
	tc.combatState.PartHeavyClickCount[partKey] = 0
}

func applyArmorAutoStrikeTrigger(tc *talentTriggerContext) {
	asInterval := tc.compiledTalents.Armor.AutoStrikeInterval
	asRatio := tc.compiledTalents.Armor.AutoStrikeRatio
	if tc.now-tc.combatState.LastAutoStrikeAt < asInterval {
		return
	}

	var best *BossPart
	for i := range tc.boss.Parts {
		p := &tc.boss.Parts[i]
		if !p.Alive || p.Type != PartTypeHeavy {
			continue
		}
		if best == nil || p.CurrentHP > best.CurrentHP {
			best = p
		}
	}
	if best == nil {
		return
	}

	sd := int64(float64(tc.baseDamage) * asRatio)
	_, sd, _ = applyBossPartDamageDelta(tc.boss, best, sd)
	tc.combatState.LastAutoStrikeAt = tc.now
	tc.totalExtra += sd
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "armor_auto_strike", Name: "自动打击触发", EffectType: "auto_strike",
		ExtraDamage: sd, Message: "自动打击触发",
	})
	tc.damageTypeOverride = "trueDamage"
}

func applyArmorUltimateTrigger(tc *talentTriggerContext) {
	if tc.part.Type != PartTypeHeavy {
		return
	}

	pk := TalentPartKey(tc.part.X, tc.part.Y)
	jdTrigger := tc.compiledTalents.Armor.UltimateTrigger
	hpCut := tc.compiledTalents.Armor.UltimateHpCut
	if tc.combatState.JudgmentDayUsed[pk] || tc.combatState.PartHeavyClickCount[pk] < jdTrigger {
		return
	}

	tc.combatState.JudgmentDayUsed[pk] = true
	cd := min(int64(float64(tc.boss.MaxHP)*hpCut), tc.part.CurrentHP)
	_, cd, _ = applyBossPartDamageDelta(tc.boss, tc.part, cd)
	tc.totalExtra += cd
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "armor_ultimate", Name: "审判日触发！削除 50% 最大生命", EffectType: "judgment_day",
		ExtraDamage: cd, Message: "审判日触发！削除 50% 最大生命",
	})
	tc.damageTypeOverride = "judgement"
}

func applyCritBleedTrigger(tc *talentTriggerContext) {
	if !tc.isCritical {
		return
	}
	if bd := int64(float64(tc.baseDamage) * tc.compiledTalents.Crit.BleedRatio); bd > 0 {
		tc.totalExtra += bd
		tc.events = append(tc.events, TalentTriggerEvent{
			TalentID: "crit_bleed", Name: "致命出血", EffectType: "bleed",
			ExtraDamage: bd, Message: "致命出血",
			PartX: tc.part.X, PartY: tc.part.Y,
		})
	}
}

func applyCritFinalCutTrigger(tc *talentTriggerContext) {
	if !tc.isCritical {
		return
	}

	tc.combatState.CritCount++
	triggerCount := tc.compiledTalents.Crit.FinalCutTrigger
	hpCut := tc.compiledTalents.Crit.FinalCutHpCut
	if tc.combatState.CritCount < triggerCount || tc.now-tc.combatState.LastFinalCutAt < 30 {
		return
	}

	tc.combatState.LastFinalCutAt = tc.now
	cd := min(int64(float64(tc.boss.MaxHP)*hpCut), tc.part.CurrentHP)
	_, cd, _ = applyBossPartDamageDelta(tc.boss, tc.part, cd)
	tc.totalExtra += cd
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "crit_final_cut", Name: "终末血斩！", EffectType: "final_cut",
		ExtraDamage: cd, Message: "终末血斩！",
		PartX: tc.part.X, PartY: tc.part.Y,
	})
	tc.damageTypeOverride = "doomsday"
}

func applyCritDeathEcstasyTrigger(tc *talentTriggerContext) {
	if tc.combatState.OmenStacks < 100 || tc.compiledTalents.Crit.DeathEcstasyMult <= 0 {
		return
	}

	consumed := min(tc.combatState.OmenStacks, 100)
	tc.combatState.OmenStacks -= consumed

	effStacks := min(consumed, 100)
	ed := min(int64(float64(tc.baseDamage)*float64(effStacks)*tc.compiledTalents.Crit.DeathEcstasyMult), tc.part.CurrentHP)
	_, ed, _ = applyBossPartDamageDelta(tc.boss, tc.part, ed)
	tc.totalExtra += ed
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "crit_death_ecstasy", Name: "死亡狂喜", EffectType: "death_ecstasy_ult",
		ExtraDamage: ed, Message: "死亡狂喜！",
		PartX: tc.part.X, PartY: tc.part.Y,
	})
	tc.damageTypeOverride = "doomsday"
}

func applyCritDoomJudgmentTrigger(tc *talentTriggerContext) {
	if tc.part.Alive || len(tc.combatState.DoomMarks) == 0 {
		return
	}

	for _, idx := range tc.combatState.DoomMarks {
		if idx != tc.partIndex {
			continue
		}
		omenReward := tc.compiledTalents.Crit.DoomOmenPerMark
		tc.combatState.OmenStacks += omenReward
		tc.events = append(tc.events, TalentTriggerEvent{
			TalentID: "crit_doom_judgment", Name: "末日审判", EffectType: "doom_mark",
			Message: fmt.Sprintf("标记触发！+%d 死兆", omenReward),
			PartX:   tc.part.X, PartY: tc.part.Y,
		})
		tc.damageTypeOverride = "doomsday"
		return
	}
}
