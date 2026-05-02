package core

import (
	"fmt"
	"math"
	"slices"
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
	nowMs              int64
	totalExtra         int64
	events             []TalentTriggerEvent
	damageTypeOverride string
	roll               func(int) int
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
	if tc.combatState.CollapseEndsAt > tc.now && slices.Contains(tc.combatState.CollapseParts, tc.partIndex) {
		return
	}

	collapseTrigger := tc.compiledTalents.Armor.CollapseTrigger
	if tc.combatState.PartHeavyClickCount[partKey] < collapseTrigger {
		return
	}

	cd := tc.compiledTalents.Armor.CollapseDuration
	if !slices.Contains(tc.combatState.CollapseParts, tc.partIndex) {
		tc.combatState.CollapseParts = append(tc.combatState.CollapseParts, tc.partIndex)
	}
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
	partKey := TalentPartKey(tc.part.X, tc.part.Y)
	if tc.combatState.AutoStrikeExpiresAt > 0 && tc.now > tc.combatState.AutoStrikeExpiresAt {
		resetAutoStrikeCombo(tc.combatState)
	}
	if tc.part.Type != PartTypeHeavy {
		resetAutoStrikeCombo(tc.combatState)
		return
	}

	if tc.combatState.AutoStrikeTargetPart != partKey {
		tc.combatState.AutoStrikeTargetPart = partKey
		tc.combatState.AutoStrikeComboCount = 0
		tc.combatState.AutoStrikeExpiresAt = 0
	}

	if tc.combatState.AutoStrikeExpiresAt == 0 {
		tc.combatState.AutoStrikeExpiresAt = tc.now + TalentAutoStrikeWindowSec
	}
	tc.combatState.AutoStrikeComboCount++

	asTrigger := tc.compiledTalents.Armor.AutoStrikeTrigger
	asRatio := tc.compiledTalents.Armor.AutoStrikeRatio
	if tc.combatState.AutoStrikeComboCount < asTrigger {
		return
	}
	if !tc.part.Alive {
		resetAutoStrikeCombo(tc.combatState)
		return
	}

	sd := int64(float64(tc.baseDamage) * asRatio)
	_, sd, _ = applyBossPartDamageDelta(tc.boss, tc.part, sd)
	tc.totalExtra += sd
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "armor_auto_strike", Name: "自动打击触发", EffectType: "auto_strike",
		ExtraDamage: sd, Message: "碎甲重击触发",
		PartX: tc.part.X, PartY: tc.part.Y,
	})
	resetAutoStrikeCombo(tc.combatState)
	tc.damageTypeOverride = "pursuit"
}

func resetAutoStrikeCombo(state *TalentCombatState) {
	if state == nil {
		return
	}
	state.AutoStrikeTargetPart = ""
	state.AutoStrikeComboCount = 0
	state.AutoStrikeExpiresAt = 0
}

func applyArmorUltimateTrigger(tc *talentTriggerContext) {
	if tc.part.Type != PartTypeHeavy {
		return
	}

	pk := TalentPartKey(tc.part.X, tc.part.Y)
	jdTrigger := tc.compiledTalents.Armor.UltimateTrigger
	now := tc.now

	// 冷却中不累计，冷却结束则重置
	cooldown := tc.compiledTalents.Armor.UltimateCooldown
	if lastTrigger, ok := tc.combatState.JudgmentDayUsed[pk]; ok && lastTrigger > 0 {
		if now-lastTrigger < cooldown {
			return
		}
		delete(tc.combatState.JudgmentDayUsed, pk)
		tc.combatState.PartJudgmentDayCount[pk] = 0
	}

	tc.combatState.PartJudgmentDayCount[pk]++
	if tc.combatState.PartJudgmentDayCount[pk] < jdTrigger {
		return
	}

	ratio := tc.compiledTalents.Armor.UltimateDamageRatio
	dmg := min(int64(float64(maxInt64(1, tc.baseDamage))*ratio), tc.part.CurrentHP)
	_, dmg, _ = applyBossPartDamageDelta(tc.boss, tc.part, dmg)
	tc.totalExtra += dmg
	tc.combatState.JudgmentDayUsed[pk] = now
	tc.combatState.PartJudgmentDayCount[pk] = 0
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "armor_ultimate", Name: "审判日", EffectType: "judgment_day",
		ExtraDamage: dmg, Message: fmt.Sprintf("审判日！×%.1f 攻击力", ratio),
	})
	tc.damageTypeOverride = "judgement"
}

func applyCritBleedTrigger(tc *talentTriggerContext) {
	if !tc.isCritical {
		return
	}
	if !tc.part.Alive || tc.compiledTalents.Crit.BleedDuration <= 0 {
		return
	}
	if bd := int64(float64(tc.baseDamage) * tc.compiledTalents.Crit.BleedRatio); bd > 0 {
		durationMs := tc.compiledTalents.Crit.BleedDuration * 1000
		tickIntervalMs := int64(200)
		totalTicks := durationMs / tickIntervalMs
		if totalTicks <= 0 {
			totalTicks = 1
		}
		partKey := TalentPartKey(tc.part.X, tc.part.Y)
		tc.combatState.Bleeds[partKey] = TalentBleedState{
			StartedAtMs:    tc.nowMs,
			NextTickAtMs:   tc.nowMs + tickIntervalMs,
			EndsAtMs:       tc.nowMs + durationMs,
			DurationMs:     durationMs,
			TickIntervalMs: tickIntervalMs,
			TotalTicks:     totalTicks,
			TotalDamage:    bd,
		}
		tc.events = append(tc.events, TalentTriggerEvent{
			TalentID: "crit_bleed", Name: "致命出血", EffectType: "bleed",
			Message: "致命出血（持续3秒）",
			PartX:   tc.part.X, PartY: tc.part.Y,
		})
	}
}

func applyCritFinalCutTrigger(tc *talentTriggerContext) {
	if tc.compiledTalents.Crit.FinalCutOmenTrigger <= 0 || tc.compiledTalents.Crit.FinalCutDamageRatio <= 0 {
		return
	}
	if tc.combatState.OmenStacks < tc.compiledTalents.Crit.FinalCutOmenTrigger || !tc.part.Alive {
		return
	}
	tc.combatState.OmenStacks = 0
	tc.combatState.LastFinalCutAt = tc.now
	cd := min(int64(float64(maxInt64(1, tc.baseDamage))*tc.compiledTalents.Crit.FinalCutDamageRatio), tc.part.CurrentHP)
	_, actualDamage, _ := applyBossPartDamageDelta(tc.boss, tc.part, cd)
	tc.totalExtra += actualDamage
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "crit_final_cut", Name: "终末血斩！", EffectType: "final_cut",
		ExtraDamage: actualDamage, Message: "终末血斩！",
		PartX: tc.part.X, PartY: tc.part.Y,
	})
	tc.damageTypeOverride = "doomsday"
}

func applyCritDoomJudgmentTrigger(tc *talentTriggerContext) {
	if !tc.combatState.HasTriggeredDoom && tc.compiledTalents.Crit.DoomMarkCount > 0 {
		if tc.boss.CurrentHP <= int64(float64(maxInt64(1, tc.boss.MaxHP))*tc.compiledTalents.Crit.DoomMarkThreshold) {
			tc.combatState.DoomMarks = randomMarkIndices(len(tc.boss.Parts), min(tc.compiledTalents.Crit.DoomMarkCount, len(tc.boss.Parts)), tc.roll)
			tc.combatState.HasTriggeredDoom = true
		}
	}
	if tc.part.Alive || len(tc.combatState.DoomMarks) == 0 {
		return
	}

	for _, idx := range tc.combatState.DoomMarks {
		if idx != tc.partIndex {
			continue
		}
		omenReward := tc.compiledTalents.Crit.DoomOmenPerMark
		tc.combatState.OmenStacks, _ = applyOmenStackDelta(tc.combatState.OmenStacks, omenReward)
		tc.events = append(tc.events, TalentTriggerEvent{
			TalentID: "crit_doom_judgment", Name: "末日审判", EffectType: "doom_mark",
			Message: fmt.Sprintf("标记触发！+%d 死兆", omenReward),
			PartX:   tc.part.X, PartY: tc.part.Y,
		})
		tc.combatState.DoomMarks = slices.DeleteFunc(tc.combatState.DoomMarks, func(marked int) bool {
			return marked == tc.partIndex
		})
		tc.damageTypeOverride = "doomsday"
		return
	}
}
