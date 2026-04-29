package vote

import "maps"

type compiledNormalTalents struct {
	TriggerCount           int64
	ExtraHits              int64
	ChaseRatio             float64
	RetainPercent          float64
	SilverStormDuration    int64
	SilverStormDamageRatio float64
}

type compiledArmorTalents struct {
	CollapseTrigger   int64
	CollapseDuration  int64
	AutoStrikeTrigger int64
	AutoStrikeRatio   float64
	CollapseAmp       float64
	UltimateTrigger   int64
	UltimateHpCut     float64
}

type compiledCritTalents struct {
	SkinnerChance       float64
	SkinnerDuration     int64
	OmenKillThreshold   float64
	OmenKillDmgPerOmen  float64
	OmenResonatePerOmen float64
	BleedRatio          float64
	FinalCutTrigger     int64
	FinalCutHpCut       float64
	DeathEcstasyMult    float64
	DoomMarkCount       int
	DoomOmenPerMark     int
}

type CompiledTalentSet struct {
	levels   map[string]int
	learned  map[string]struct{}
	tierFull map[TalentTree]map[int]bool

	Modifiers    *TalentModifiers
	Normal       compiledNormalTalents
	Armor        compiledArmorTalents
	Crit         compiledCritTalents
	triggers     []compiledTalentTrigger
	triggerNames []string
}

func (c *CompiledTalentSet) Has(talentID string) bool {
	if c == nil {
		return false
	}
	_, ok := c.learned[talentID]
	return ok
}

func (c *CompiledTalentSet) Level(talentID string) int {
	if c == nil {
		return 0
	}
	return c.levels[talentID]
}

func (c *CompiledTalentSet) IsTierFull(tree TalentTree, tier int) bool {
	if c == nil {
		return false
	}
	return c.tierFull[tree][tier]
}

func compileTalentSet(state *TalentState) *CompiledTalentSet {
	compiled := &CompiledTalentSet{
		levels:   make(map[string]int),
		learned:  make(map[string]struct{}),
		tierFull: make(map[TalentTree]map[int]bool),
		Modifiers: &TalentModifiers{
			PartTypeBonus: make(map[PartType]float64),
			Learned:       []string{},
		},
		Normal: compiledNormalTalents{
			ChaseRatio: TalentNormalStormChaseRatio,
		},
		Armor: compiledArmorTalents{
			CollapseDuration: 8,
		},
	}
	if state == nil || len(state.Talents) == 0 {
		return compiled
	}

	for id, level := range state.Talents {
		if level <= 0 {
			continue
		}
		compiled.levels[id] = level
		compiled.learned[id] = struct{}{}
		compiled.Modifiers.Learned = append(compiled.Modifiers.Learned, id)
	}

	compiled.tierFull = compileTierFull(state.Talents)
	compiled.Modifiers = buildTalentModifiersFromCompiled(compiled)
	compiled.Normal = compileNormalTalents(compiled)
	compiled.Armor = compileArmorTalents(compiled)
	compiled.Crit = compileCritTalents(compiled)
	compiled.triggers, compiled.triggerNames = compiled.buildTriggerHandlers()

	return compiled
}

func compileTierFull(talents map[string]int) map[TalentTree]map[int]bool {
	result := make(map[TalentTree]map[int]bool)
	for tree := range tierCompletionBonusLabels {
		result[tree] = make(map[int]bool)
		for tier := range tierCompletionBonusLabels[tree] {
			result[tree][tier] = isLearnedTierFull(tree, tier, talents)
		}
	}
	return result
}

func buildTalentModifiersFromCompiled(compiled *CompiledTalentSet) *TalentModifiers {
	mods := &TalentModifiers{
		PartTypeBonus: make(map[PartType]float64),
		Learned:       make([]string, 0, len(compiled.levels)),
	}
	mods.Learned = append(mods.Learned, compiled.Modifiers.Learned...)

	for id, level := range compiled.levels {
		def, ok := talentDefs[id]
		if !ok {
			continue
		}

		val, _ := def.EffectValue.(map[string]any)
		levelFactor := float64(level)

		switch def.EffectType {
		case "attack_power_percent":
			if p, ok := val["percent"].(float64); ok {
				mods.AttackPowerPercent += p * levelFactor
			}
		case "all_damage_amplify":
			if p, ok := val["percent"].(float64); ok {
				mods.AllDamageAmplify += p * levelFactor
			}
		case "part_type_damage":
			partTypeStr, _ := val["partType"].(string)
			percent, _ := val["percent"].(float64)
			if id == "normal_soft_atk" {
				percent = normalCoreScaledPartDamage(level, 0.80, 3.00)
			}
			if id == "armor_heavy_atk" {
				percent = normalCoreScaledPartDamage(level, 1.00, 3.00)
			}
			if partTypeStr != "" {
				mods.PartTypeBonus[PartType(partTypeStr)] += percent
			}
		case "armor_pen_extra":
			if p, ok := val["extraPen"].(float64); ok {
				if id == "armor_pen_up" {
					mods.ArmorPenExtra += armorPenUpExtraForLevel(level)
				} else {
					mods.ArmorPenExtra += p * levelFactor
				}
			}
		case "crit_damage_bonus":
			if p, ok := val["percent"].(float64); ok {
				if id == "crit_cruel" {
					mods.CritDamagePercentBonus += critCruelBonusForLevel(level)
				} else {
					mods.CritDamagePercentBonus += p * levelFactor
				}
			}
		case "per_part_damage":
			if p, ok := val["percentPerPart"].(float64); ok {
				mods.PerPartDamagePercent += p * levelFactor
			}
		case "low_hp_bonus":
			if m, ok := val["multiplier"].(float64); ok {
				if id == "normal_low_hp" {
					mods.LowHpMultiplier = normalLowHPMultiplierForLevel(level)
				} else {
					mods.LowHpMultiplier += m * levelFactor
				}
			}
			if threshold, ok := val["hpThreshold"].(float64); ok {
				t := threshold * levelFactor
				if id == "normal_low_hp" {
					t = normalLowHPThresholdForLevel(level)
				}
				if t > mods.LowHpThreshold {
					mods.LowHpThreshold = t
				}
			}
		case "pen_to_amplify":
			if r, ok := val["convertRatio"].(float64); ok {
				if id == "armor_pen_convert" {
					mods.PenToAmplifyRatio = armorPenConvertRatioForLevel(level)
				} else {
					mods.PenToAmplifyRatio = r * levelFactor
				}
			}
		case "chase_ratio_bonus":
			if p, ok := val["percent"].(float64); ok {
				mods.ChaseRatioBonus += p * levelFactor
			}
		case "omen_crit_damage":
			if p, ok := val["critDmgPerOmen"].(float64); ok {
				mods.OmenCritDmgExtra += p * levelFactor
			}
		case "overkill":
			if _, ok := val["baseCritBonus"].(float64); ok {
				mods.CritRateBonus += critCoreBaseCritBonusForLevel(level) * 100
			}
			if ratio, ok := val["overflowToCritDmg"].(float64); ok {
				mods.OverflowToCritDmgRatio = ratio
			}
			if _, ok := val["critDmgPerOmen"].(float64); ok {
				mods.OmenCritDmgExtra += critOmenResonateForLevel(level)
			}
		}
	}

	for tree, tiers := range compiled.tierFull {
		for tier, full := range tiers {
			if full {
				applyTierCompletionBonus(mods, string(tree), tier)
			}
		}
	}

	return mods
}

func compileNormalTalents(compiled *CompiledTalentSet) compiledNormalTalents {
	normal := compiledNormalTalents{
		ChaseRatio: TalentNormalStormChaseRatio,
	}
	if compiled.Has("normal_core") {
		level := compiled.Level("normal_core")
		normal.TriggerCount = int64(normalCoreTriggerCountForLevel(level))
		normal.ExtraHits = int64(normalCoreExtraHitsForLevel(level))
		if compiled.IsTierFull(TalentTreeNormal, 2) {
			normal.TriggerCount -= 20
		}
		if compiled.Has("normal_chase_up") {
			normal.ChaseRatio = max(normal.ChaseRatio, normalChaseUpgradeRatioForLevel(compiled.Level("normal_chase_up")))
		}
		if compiled.Has("normal_combo_ext") {
			normal.ExtraHits += int64(normalComboExtendHitsForLevel(compiled.Level("normal_combo_ext")))
		}
		if compiled.Has("normal_filler_t2a") {
			normal.ChaseRatio += chaseRatioBonusForLevel(compiled.Level("normal_filler_t2a"), "normal_filler_t2a")
		}
		if compiled.IsTierFull(TalentTreeNormal, 4) {
			normal.ExtraHits += 5
		}
		if normal.TriggerCount < 1 {
			normal.TriggerCount = 1
		}
	}
	if compiled.Has("normal_charge") {
		normal.RetainPercent = normalChargeRetainPercentForLevel(compiled.Level("normal_charge"))
	}
	if compiled.Has("normal_ultimate") {
		normal.SilverStormDuration = int64(normalSilverStormDurationForLevel(compiled.Level("normal_ultimate")))
		normal.SilverStormDamageRatio = normalSilverStormDamageRatioForLevel(compiled.Level("normal_ultimate"))
	}

	return normal
}

func compileArmorTalents(compiled *CompiledTalentSet) compiledArmorTalents {
	armor := compiledArmorTalents{
		CollapseDuration: 8,
		CollapseAmp:      1.0,
	}
	if compiled.Has("armor_core") {
		armor.CollapseTrigger = int64(armorCoreCollapseTriggerForLevel(compiled.Level("armor_core")))
		if armor.CollapseTrigger < 1 {
			armor.CollapseTrigger = 1
		}
	}
	if compiled.Has("armor_collapse_ext") {
		armor.CollapseAmp *= armorCollapseResonanceAmpForLevel(compiled.Level("armor_collapse_ext"))
	}
	if compiled.Has("armor_auto_strike") {
		level := compiled.Level("armor_auto_strike")
		armor.AutoStrikeTrigger = int64(armorAutoStrikeTriggerCountForLevel(level))
		armor.AutoStrikeRatio = armorAutoStrikeRatioForLevel(level)
	}
	if compiled.Has("armor_ruin") {
		armor.CollapseAmp *= armorRuinAmpForLevel(compiled.Level("armor_ruin"))
	}
	if compiled.Has("armor_ultimate") {
		level := compiled.Level("armor_ultimate")
		armor.UltimateTrigger = int64(armorUltimateTriggerCountForLevel(level))
		armor.UltimateHpCut = armorUltimateHpCutForLevel(level)
		if compiled.IsTierFull(TalentTreeArmor, 4) {
			armor.UltimateHpCut += 0.10
		}
	}

	return armor
}

func compileCritTalents(compiled *CompiledTalentSet) compiledCritTalents {
	crit := compiledCritTalents{}
	if compiled.Has("crit_skinner") {
		level := compiled.Level("crit_skinner")
		crit.SkinnerChance = critSkinnerChanceForLevel(level)
		crit.SkinnerDuration = int64(critSkinnerDurationForLevel(level))
	}
	if compiled.Has("crit_omen_kill") {
		level := compiled.Level("crit_omen_kill")
		crit.OmenKillThreshold = critOmenKillThresholdForLevel(level)
		crit.OmenKillDmgPerOmen = critOmenKillDmgPerOmenForLevel(level)
	}
	if compiled.Has("crit_core") {
		crit.OmenResonatePerOmen = critOmenResonateForLevel(compiled.Level("crit_core"))
	}
	if compiled.Has("crit_bleed") {
		crit.BleedRatio = critBleedRatioForLevel(compiled.Level("crit_bleed"))
	}
	if compiled.Has("crit_final_cut") {
		level := compiled.Level("crit_final_cut")
		crit.FinalCutTrigger = int64(critFinalCutCountForLevel(level))
		crit.FinalCutHpCut = critFinalCutHpCutForLevel(level)
	}
	if compiled.Has("crit_death_ecstasy") {
		crit.DeathEcstasyMult = critDeathEcstasyMultForLevel(compiled.Level("crit_death_ecstasy"))
		if compiled.IsTierFull(TalentTreeCrit, 4) {
			crit.DeathEcstasyMult += 2
		}
	}
	if compiled.Has("crit_doom_judgment") {
		level := compiled.Level("crit_doom_judgment")
		crit.DoomMarkCount = critDoomMarkCountForLevel(level)
		crit.DoomOmenPerMark = critDoomOmenPerMarkForLevel(level)
	}

	return crit
}

func chaseRatioBonusForLevel(level int, talentID string) float64 {
	def, ok := talentDefs[talentID]
	if !ok {
		return 0
	}
	val, _ := def.EffectValue.(map[string]any)
	percent, _ := val["percent"].(float64)
	if percent <= 0 {
		return 0
	}
	return percent * float64(max(level, 1))
}

func cloneTalentModifiers(mods *TalentModifiers) *TalentModifiers {
	if mods == nil {
		return &TalentModifiers{PartTypeBonus: make(map[PartType]float64)}
	}
	cloned := *mods
	cloned.Learned = append([]string(nil), mods.Learned...)
	cloned.PartTypeBonus = maps.Clone(mods.PartTypeBonus)
	return &cloned
}
