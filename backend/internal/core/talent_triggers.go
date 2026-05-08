package core

import (
	"fmt"
	"math"
	"slices"
	"sort"
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
	combatStats        CombatStats
	effectivePartType  PartType
	compiledTalents    *CompiledTalentSet
	combatState        *TalentCombatState
	now                int64
	nowMs              int64
	totalExtra         int64
	events             []TalentTriggerEvent
	deltas             []BossPartStateDelta
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
	appendTrigger(c.Has("magic_core"), "magic_core", applyMagicCoreTrigger)

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
	tc.deltas = append(tc.deltas, BossPartStateDelta{
		X: tc.part.X, Y: tc.part.Y, Damage: burst, BeforeHP: tc.part.CurrentHP + burst, AfterHP: tc.part.CurrentHP, PartType: string(tc.part.Type),
	})
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
	applyCollapseState(tc, cd)
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "armor_core", Name: "灭绝穿甲", EffectType: "collapse_trigger",
		Message: fmt.Sprintf("结构崩塌！护甲归零 %d 秒", cd),
		PartX:   tc.part.X,
		PartY:   tc.part.Y,
	})
	tc.combatState.PartHeavyClickCount[partKey] = 0
}

func applyCollapseState(tc *talentTriggerContext, duration int64) {
	if tc == nil || tc.combatState == nil || tc.part == nil || duration <= 0 {
		return
	}
	if !slices.Contains(tc.combatState.CollapseParts, tc.partIndex) {
		tc.combatState.CollapseParts = append(tc.combatState.CollapseParts, tc.partIndex)
	}
	tc.combatState.CollapseEndsAt = tc.now + duration
	tc.combatState.CollapseDuration = duration
}

func judgmentDayBaseDamage(tc *talentTriggerContext) int64 {
	if tc == nil || tc.part == nil {
		return 0
	}

	activeCollapse := tc.combatState.CollapseEndsAt > tc.now && slices.Contains(tc.combatState.CollapseParts, tc.partIndex)
	if activeCollapse {
		return maxInt64(1, tc.baseDamage)
	}

	aliveCount := 0
	for _, part := range tc.boss.Parts {
		if part.Alive {
			aliveCount++
		}
	}

	currentStats := CalcBossPartDamage(tc.combatStats, tc.effectivePartType, tc.part.Armor, aliveCount, tc.boss.CurrentHP, tc.boss.MaxHP)
	collapsedStats := CalcBossPartDamage(tc.combatStats, tc.effectivePartType, 0, aliveCount, tc.boss.CurrentHP, tc.boss.MaxHP)

	currentResolved := currentStats.NormalDamage
	collapsedResolved := collapsedStats.NormalDamage
	if tc.isCritical {
		currentResolved = currentStats.CriticalDamage
		collapsedResolved = collapsedStats.CriticalDamage
	}
	if tc.compiledTalents.Armor.CollapseAmp > 1 {
		collapsedResolved = int64(float64(collapsedResolved) * tc.compiledTalents.Armor.CollapseAmp)
	}
	if currentResolved <= 0 {
		return maxInt64(1, tc.baseDamage)
	}

	return maxInt64(1, int64(math.Round(float64(maxInt64(1, tc.baseDamage))*float64(collapsedResolved)/float64(currentResolved))))
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
	tc.deltas = append(tc.deltas, BossPartStateDelta{
		X: tc.part.X, Y: tc.part.Y, Damage: sd, BeforeHP: tc.part.CurrentHP + sd, AfterHP: tc.part.CurrentHP, PartType: string(tc.part.Type),
	})
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

	baseDamage := judgmentDayBaseDamage(tc)
	applyCollapseState(tc, tc.compiledTalents.Armor.CollapseDuration)
	ratio := tc.compiledTalents.Armor.UltimateDamageRatio
	dmg := min(int64(float64(baseDamage)*ratio), tc.part.CurrentHP)
	_, dmg, _ = applyBossPartDamageDelta(tc.boss, tc.part, dmg)
	tc.totalExtra += dmg
	tc.deltas = append(tc.deltas, BossPartStateDelta{
		X: tc.part.X, Y: tc.part.Y, Damage: dmg, BeforeHP: tc.part.CurrentHP + dmg, AfterHP: tc.part.CurrentHP, PartType: string(tc.part.Type),
	})
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
	tc.deltas = append(tc.deltas, BossPartStateDelta{
		X: tc.part.X, Y: tc.part.Y, Damage: actualDamage, BeforeHP: tc.part.CurrentHP + actualDamage, AfterHP: tc.part.CurrentHP, PartType: string(tc.part.Type),
	})
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

func applyMagicCoreTrigger(tc *talentTriggerContext) {
	if tc.compiledTalents.Magic.MainRatio <= 0 || tc.combatStats.AttackPower <= 0 {
		return
	}

	if tc.combatState.MagicEchoExpiresAt > 0 && tc.now > tc.combatState.MagicEchoExpiresAt {
		tc.combatState.MagicEchoExpiresAt = 0
	}

	partKey := TalentPartKey(tc.part.X, tc.part.Y)
	echoActive := tc.combatState.MagicEchoExpiresAt > tc.now
	if !echoActive && tc.combatState.MagicEchoTargetPart != partKey {
		tc.combatState.MagicEchoTargetPart = partKey
		tc.combatState.MagicEchoStacks = 0
	}

	if tc.compiledTalents.Magic.EchoCooldownSec > 0 && tc.now >= tc.combatState.MagicEchoCooldownEndsAt && tc.combatState.MagicEchoExpiresAt == 0 {
		tc.combatState.MagicEchoStacks++
		if tc.compiledTalents.Magic.EchoRequiredHits > 0 && tc.combatState.MagicEchoStacks >= tc.compiledTalents.Magic.EchoRequiredHits {
			tc.combatState.MagicEchoStacks = 0
			tc.combatState.MagicEchoExpiresAt = tc.now + TalentMagicEchoWindowSec
			tc.combatState.MagicEchoCooldownEndsAt = tc.combatState.MagicEchoExpiresAt + tc.compiledTalents.Magic.EchoCooldownSec
			tc.events = append(tc.events, TalentTriggerEvent{
				TalentID: "magic_echo_mark", Name: "回响刻印", EffectType: "magic_rupture",
				Message: "奥术裂解已就绪",
				PartX:   tc.part.X, PartY: tc.part.Y,
			})
			echoActive = true
		}
	}

	procRate := tc.combatStats.MagicProcRate
	guaranteed := echoActive
	if guaranteed {
		procRate = 1
	}
	if procRate <= 0 {
		return
	}
	if !guaranteed {
		threshold := int(math.Round(procRate * 10000))
		if threshold <= 0 {
			return
		}
		if tc.roll == nil {
			if procRate < 1 {
				return
			}
		} else if tc.roll(10000) >= threshold {
			return
		}
	}

	mainDamage := calcMagicDamage(tc.combatStats, tc.compiledTalents.Magic.MainRatio, tc.compiledTalents.Magic.DamageMultiplier, tc.part.Armor, tc.compiledTalents.Magic.ArmorBluntPercent)
	mainDamage = min(mainDamage, tc.part.CurrentHP)
	if mainDamage > 0 {
		beforeHP, actualDamage, _ := applyBossPartDamageDelta(tc.boss, tc.part, mainDamage)
		if actualDamage > 0 {
			tc.totalExtra += actualDamage
			tc.deltas = append(tc.deltas, BossPartStateDelta{
				X: tc.part.X, Y: tc.part.Y, Damage: actualDamage, BeforeHP: beforeHP, AfterHP: tc.part.CurrentHP, PartType: string(tc.part.Type),
			})
			tc.events = append(tc.events, TalentTriggerEvent{
				TalentID: "magic_core", Name: "奥术爆裂", EffectType: "magic_burst",
				ExtraDamage: actualDamage, Message: "主目标奥术爆裂",
				PartX: tc.part.X, PartY: tc.part.Y,
			})
			if tc.now >= tc.combatState.MagicUltimateCooldownAt {
				tc.combatState.PartMagicTriggerCount[partKey]++
			}
		}
	}

	for _, idx := range magicAdjacentAliveTargets(tc.boss.Parts, tc.partIndex, 1) {
		target := &tc.boss.Parts[idx]
		damage := calcMagicDamage(tc.combatStats, tc.compiledTalents.Magic.SplashRatio, tc.compiledTalents.Magic.SplashMultiplier, target.Armor, tc.compiledTalents.Magic.ArmorBluntPercent)
		damage = min(damage, target.CurrentHP)
		if damage <= 0 {
			continue
		}
		beforeHP, actualDamage, _ := applyBossPartDamageDelta(tc.boss, target, damage)
		if actualDamage <= 0 {
			continue
		}
		tc.totalExtra += actualDamage
		tc.deltas = append(tc.deltas, BossPartStateDelta{
			X: target.X, Y: target.Y, Damage: actualDamage, BeforeHP: beforeHP, AfterHP: target.CurrentHP, PartType: string(target.Type),
		})
		tc.events = append(tc.events, TalentTriggerEvent{
			TalentID: "magic_core", Name: "奥术余波", EffectType: "magic_burst",
			ExtraDamage: actualDamage, Message: "邻近部位受击",
			PartX: target.X, PartY: target.Y,
		})
	}

	if tc.compiledTalents.Magic.UltimateTriggerCount > 0 && tc.compiledTalents.Magic.UltimateMainRatio > 0 {
		tc.tryTriggerMagicUltimate(partKey)
	}
}

func calcMagicDamage(stats CombatStats, baseRatio float64, damageMultiplier float64, armor int64, bluntPercent float64) int64 {
	attackBase := maxInt64(1, stats.AttackPower)
	magicBase := float64(attackBase) * max(0, baseRatio)
	if magicBase <= 0 {
		return 0
	}
	effectiveArmor := max(int64(float64(maxInt64(0, armor))*(1-max(0, min(0.95, bluntPercent)))), 0)
	damage := (magicBase - float64(effectiveArmor)) * (1 + max(0, stats.AllDamageAmplify)) * max(0, damageMultiplier)
	return maxInt64(1, int64(math.Round(damage)))
}

func magicAdjacentAliveTargets(parts []BossPart, centerIndex int, limit int) []int {
	if centerIndex < 0 || centerIndex >= len(parts) || limit <= 0 {
		return nil
	}
	center := parts[centerIndex]
	indices := make([]int, 0, limit)
	for index, part := range parts {
		if index == centerIndex || !part.Alive || part.CurrentHP <= 0 {
			continue
		}
		dx := absInt(part.X - center.X)
		dy := absInt(part.Y - center.Y)
		if dx <= 1 && dy <= 1 {
			indices = append(indices, index)
		}
	}
	sort.Slice(indices, func(i, j int) bool {
		left := parts[indices[i]]
		right := parts[indices[j]]
		leftDist := absInt(left.X-center.X) + absInt(left.Y-center.Y)
		rightDist := absInt(right.X-center.X) + absInt(right.Y-center.Y)
		if leftDist != rightDist {
			return leftDist < rightDist
		}
		if left.Y != right.Y {
			return left.Y < right.Y
		}
		return left.X < right.X
	})
	if len(indices) > limit {
		return indices[:limit]
	}
	return indices
}

func (tc *talentTriggerContext) tryTriggerMagicUltimate(partKey string) {
	if tc == nil || tc.combatState == nil || tc.compiledTalents.Magic.UltimateTriggerCount <= 0 {
		return
	}
	if tc.now < tc.combatState.MagicUltimateCooldownAt {
		return
	}
	if tc.combatState.PartMagicTriggerCount[partKey] < tc.compiledTalents.Magic.UltimateTriggerCount {
		return
	}
	tc.combatState.PartMagicTriggerCount[partKey] = 0
	tc.combatState.MagicUltimateCooldownAt = tc.now + tc.compiledTalents.Magic.UltimateCooldownSec

	mainDamage := calcMagicDamage(tc.combatStats, tc.compiledTalents.Magic.UltimateMainRatio, tc.compiledTalents.Magic.DamageMultiplier, tc.part.Armor, tc.compiledTalents.Magic.ArmorBluntPercent)
	for index := range tc.boss.Parts {
		target := &tc.boss.Parts[index]
		if !target.Alive || target.CurrentHP <= 0 {
			continue
		}
		damage := mainDamage
		if index != tc.partIndex {
			damage = int64(math.Round(float64(mainDamage) * tc.compiledTalents.Magic.UltimateSplashShare))
		}
		damage = min(damage, target.CurrentHP)
		if damage <= 0 {
			continue
		}
		beforeHP, actualDamage, _ := applyBossPartDamageDelta(tc.boss, target, damage)
		if actualDamage <= 0 {
			continue
		}
		tc.totalExtra += actualDamage
		tc.deltas = append(tc.deltas, BossPartStateDelta{
			X: target.X, Y: target.Y, Damage: actualDamage, BeforeHP: beforeHP, AfterHP: target.CurrentHP, PartType: string(target.Type),
		})
	}
	tc.events = append(tc.events, TalentTriggerEvent{
		TalentID: "magic_ultimate", Name: "星陨潮爆", EffectType: "magic_starfall",
		Message: "星陨潮爆席卷全场",
		PartX:   tc.part.X, PartY: tc.part.Y,
	})
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
