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
	UltimateTrigger      int64
	UltimateDamageRatio  float64
	UltimateCooldown     int64
}

type compiledCritTalents struct {
	OmenCap               int
	OmenPerWeakCrit       int
	SkinnerChance         float64
	SkinnerDuration       int64
	SkinnerCooldown       int64
	OmenKillThreshold     float64
	OmenKillDmgPerOmen    float64
	OmenResonatePerOmen   float64
	OmenReapThresholds    []int
	OmenReapDamageMults   []float64
	BleedRatio            float64
	BleedDuration         int64
	WeakspotInsightMult   float64
	FinalCutOmenTrigger   int
	FinalCutDamageRatio   float64
	DoomMarkThreshold     float64
	DoomMarkCount         int
	DoomOmenPerMark       int
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

		switch def.EffectType {
		case "attack_power_percent":
			if _, ok := val["percent"].(float64); ok {
				switch id {
				case "normal_atk_up":
					mods.AttackPowerPercent += normalAtkUpPercentForLevel(level)
				case "normal_filler_t1a":
					mods.AttackPowerPercent += normalFillerT1aAtkPercentForLevel(level)
				case "normal_filler_t3b":
					mods.AttackPowerPercent += normalFillerT3bAtkPercentForLevel(level)
				case "armor_filler_t1a":
					mods.AttackPowerPercent += armorFillerT1aAtkPercentForLevel(level)
				case "crit_filler_t1a":
					mods.AttackPowerPercent += critFillerT1aAtkPercentForLevel(level)
				}
			}
		case "all_damage_amplify":
			if _, ok := val["percent"].(float64); ok {
				switch id {
				case "normal_dmg_amp":
					mods.AllDamageAmplify += normalDmgAmpPercentForLevel(level)
				case "normal_filler_t1b":
					mods.AllDamageAmplify += normalFillerT1bDmgAmpForLevel(level)
				case "normal_filler_t2b":
					mods.AllDamageAmplify += normalFillerT2bDmgAmpForLevel(level)
				case "normal_filler_t3a":
					mods.AllDamageAmplify += normalFillerT3aDmgAmpForLevel(level)
				case "armor_boss_hunter":
					mods.AllDamageAmplify += armorBossHunterPercentForLevel(level)
				case "armor_filler_t2a":
					mods.AllDamageAmplify += armorFillerT2aDmgAmpForLevel(level)
				case "armor_filler_t3a":
					mods.AllDamageAmplify += armorFillerT3aDmgAmpForLevel(level)
				case "armor_filler_t3b":
					mods.AllDamageAmplify += armorFillerT3bDmgAmpForLevel(level)
				case "crit_filler_t2a":
					mods.AllDamageAmplify += critFillerT2aDmgAmpForLevel(level)
				case "crit_filler_t3a":
					mods.AllDamageAmplify += critFillerT3aDmgAmpForLevel(level)
				}
			}
		case "part_type_damage":
			partTypeStr, _ := val["partType"].(string)
			if partTypeStr != "" {
				switch id {
				case "normal_soft_atk":
					mods.PartTypeBonus[PartType(partTypeStr)] += normalSoftAtkPercentForLevel(level)
				case "armor_heavy_atk":
					mods.PartTypeBonus[PartType(partTypeStr)] += armorHeavyAtkPercentForLevel(level)
				}
			}
		case "armor_pen_extra":
			if _, ok := val["extraPen"].(float64); ok {
				switch id {
				case "armor_pen_up":
					mods.ArmorPenExtra += armorPenUpExtraForLevel(level)
				case "armor_filler_t1b":
					mods.ArmorPenExtra += armorFillerT1bPenForLevel(level)
				}
			}
		case "crit_damage_bonus":
			if _, ok := val["percent"].(float64); ok {
				switch id {
				case "crit_cruel":
					mods.CritDamagePercentBonus += critCruelBonusForLevel(level)
				case "crit_filler_t1b":
					mods.CritDamagePercentBonus += critFillerT1bCritDmgForLevel(level)
				case "crit_filler_t3b":
					mods.CritDamagePercentBonus += critFillerT3bCritDmgForLevel(level)
				}
			}
		case "per_part_damage":
			if _, ok := val["percentPerPart"].(float64); ok {
				if id == "normal_encircle" {
					mods.PerPartDamagePercent += normalEncirclePercentForLevel(level)
				}
			}
		case "low_hp_bonus":
			if _, ok := val["multiplier"].(float64); ok {
				if id == "normal_low_hp" {
					mods.LowHpMultiplier = normalLowHPMultiplierForLevel(level)
				}
			}
			if _, ok := val["hpThreshold"].(float64); ok {
				if id == "normal_low_hp" {
					t := normalLowHPThresholdForLevel(level)
					if t > mods.LowHpThreshold {
						mods.LowHpThreshold = t
					}
				}
			}
		case "pen_to_amplify":
			if _, ok := val["convertRatio"].(float64); ok {
				if id == "armor_pen_convert" {
					mods.PenToAmplifyRatio = armorPenConvertRatioForLevel(level)
				}
			}
		case "chase_ratio_bonus":
			if _, ok := val["percent"].(float64); ok {
				if id == "normal_filler_t2a" {
					mods.ChaseRatioBonus += normalFillerT2aChaseForLevel(level)
				}
			}
		case "omen_crit_damage":
			if _, ok := val["critDmgPerOmen"].(float64); ok {
				if id == "crit_filler_t2b" {
					mods.OmenCritDmgExtra += critFillerT2bOmenCritDmgForLevel(level)
				}
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
		armor.UltimateDamageRatio = armorUltimateDamageRatioForLevel(level)
		armor.UltimateCooldown = armorUltimateCooldownForLevel(level)
		if compiled.IsTierFull(TalentTreeArmor, 4) {
			armor.UltimateDamageRatio += 1.0
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
		crit.SkinnerCooldown = critSkinnerCooldownForLevel(level)
	}
	if compiled.Has("crit_omen_kill") {
		level := compiled.Level("crit_omen_kill")
		crit.OmenKillThreshold = critOmenKillThresholdForLevel(level)
		crit.OmenKillDmgPerOmen = critOmenKillDmgPerOmenForLevel(level)
	}
	if compiled.Has("crit_core") {
		crit.OmenCap = TalentOmenStackCap
		crit.OmenPerWeakCrit = talentInt(talentDefs["crit_core"].EffectValue.(map[string]any)["omenPerWeakCrit"])
		crit.OmenResonatePerOmen = critOmenResonateForLevel(compiled.Level("crit_core"))
	}
	if compiled.Has("crit_omen_reap") {
		crit.OmenReapThresholds = []int{15, 30, 60, 90, 120}
		crit.OmenReapDamageMults = []float64{1.10, 1.20, 1.30, 1.40, 1.50}
	}
	if compiled.Has("crit_bleed") {
		level := compiled.Level("crit_bleed")
		crit.BleedRatio = critBleedRatioForLevel(level)
		crit.BleedDuration = critBleedDurationForLevel(level)
	}
	if compiled.Has("crit_final_cut") {
		level := compiled.Level("crit_final_cut")
		crit.FinalCutOmenTrigger = critFinalCutOmenTriggerForLevel(level)
		crit.FinalCutDamageRatio = critFinalCutDamageRatioForLevel(level)
	}
	if compiled.Has("crit_weakspot_insight") {
		crit.WeakspotInsightMult = critWeakspotInsightMultiplierForLevel(compiled.Level("crit_weakspot_insight"))
	}
	if compiled.Has("crit_doom_judgment") {
		level := compiled.Level("crit_doom_judgment")
		crit.DoomMarkThreshold = 0.35
		crit.DoomMarkCount = critDoomMarkCountForLevel(level)
		crit.DoomOmenPerMark = critDoomOmenPerMarkForLevel(level)
	}

	return crit
}

func chaseRatioBonusForLevel(level int, talentID string) float64 {
	switch talentID {
	case "normal_filler_t2a":
		return normalFillerT2aChaseForLevel(level)
	}
	return 0
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
