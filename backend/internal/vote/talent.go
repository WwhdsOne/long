package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/bytedance/sonic"
)

// TalentTree 天赋树类型
type TalentTree string

const (
	TalentTreeNormal TalentTree = "normal" // 普攻 - 均衡攻势
	TalentTreeArmor  TalentTree = "armor"  // 破甲 - 碎盾攻坚
	TalentTreeCrit   TalentTree = "crit"   // 暴击 - 致命洞察
)

const (
	TalentCostTier0Main       int64 = 20
	TalentCostTier1Main       int64 = 30
	TalentCostTier2Main       int64 = 80
	TalentCostTier3Main       int64 = 150
	TalentCostTier4Main       int64 = 200
	TalentCostTier1Filler     int64 = 15
	TalentCostTier2Filler     int64 = 35
	TalentCostTier3Filler     int64 = 60
	TalentDefaultMaxLevel           = 5
	TalentAutoStrikeWindowSec       = 5
	TalentOmenStackCap              = 100

	talentCostLevelExponent       = 0.85
	talentCostMultiplier          = 1.8
	talentTier0GrowthFactor       = 3.0
	TalentOmenOverflowDamageRatio = 0.02

	// ===== 普攻系关键参数（可直接调）=====
	// 暴风连击：触发所需点击次数
	TalentNormalStormTriggerCount = 100.0
	// 暴风连击：基础追加段数
	TalentNormalStormExtraHits = 15.0
	// 暴风连击：单段追击倍率
	TalentNormalStormChaseRatio = 0.50
	// 暴风连击：追击倍率上限（预留）
	TalentNormalStormMaxChaseRatio = 0.80
	// 追击强化：追击倍率提升后的目标值
	TalentNormalChaseUpgradeRatio = 0.80
	// 连击扩展：额外追加段数
	TalentNormalComboExtendHits = 10.0
)

func TalentLevelCost(base int64, targetLevel int) int64 {
	if base <= 0 || targetLevel <= 0 {
		return 0
	}
	if base == TalentCostTier0Main {
		return int64(math.Round(float64(base) * talentCostMultiplier * math.Pow(talentTier0GrowthFactor, float64(targetLevel-1))))
	}
	return int64(math.Round(float64(base) * math.Pow(float64(targetLevel), talentCostLevelExponent) * talentCostMultiplier))
}

func TalentLevelCostDiff(base int64, currentLevel, targetLevel int) int64 {
	if base <= 0 || targetLevel <= currentLevel {
		return 0
	}
	var total int64
	for level := currentLevel + 1; level <= targetLevel; level++ {
		total += TalentLevelCost(base, level)
	}
	return total
}

// TalentCumulativeCost 返回从 0 级升到 targetLevel 的累计实际消耗。
// 新公式按“单次升级成本”计费，因此需要逐级累加。
func TalentCumulativeCost(base int64, targetLevel int) int64 {
	return TalentLevelCostDiff(base, 0, targetLevel)
}

// TalentDef 天赋节点定义
type TalentDef struct {
	ID          string     `json:"id"`
	Tree        TalentTree `json:"tree"`
	Tier        int        `json:"tier"`     // 0=基石, 1-3=中间, 4=终极
	Cost        int64      `json:"cost"`     // Lv1 基准成本
	MaxLevel    int        `json:"maxLevel"` // 最高可学等级，默认 5
	Name        string     `json:"name"`
	EffectType  string     `json:"effectType"`
	EffectValue any        `json:"effectValue"`
}

// TalentState 玩家天赋状态
type TalentState struct {
	Talents map[string]int `json:"talents"` // talentID → 当前等级
}

type TalentEffectLine struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// talentPlayerData Redis 中存储的原始结构
type talentPlayerData struct {
	Talents string `json:"talents"` // JSON: map[string]int
}

func GetTalentLevel(state *TalentState, talentID string) int {
	if state == nil || state.Talents == nil {
		return 0
	}
	return state.Talents[talentID]
}

func HasTalentLearned(state *TalentState, talentID string) bool {
	return GetTalentLevel(state, talentID) > 0
}

// 三系天赋定义表
var talentDefs = map[string]TalentDef{
	// ===== 普攻：均衡攻势 =====
	"normal_core":      {ID: "normal_core", Tree: TalentTreeNormal, Tier: 0, MaxLevel: 5, Name: "暴风连击", EffectType: "storm_combo", EffectValue: map[string]any{"triggerCount": 50.0, "extraHits": 20.0, "chaseRatio": 0.50, "maxChaseRatio": 0.80}},
	"normal_atk_up":    {ID: "normal_atk_up", Tree: TalentTreeNormal, Tier: 1, MaxLevel: 5, Name: "攻击强化", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.60}},
	"normal_dmg_amp":   {ID: "normal_dmg_amp", Tree: TalentTreeNormal, Tier: 1, MaxLevel: 5, Name: "伤害增幅", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.50}},
	"normal_soft_atk":  {ID: "normal_soft_atk", Tree: TalentTreeNormal, Tier: 1, MaxLevel: 5, Name: "软组织特攻", EffectType: "part_type_damage", EffectValue: map[string]any{"partType": "soft", "percent": 0.80}},
	"normal_charge":    {ID: "normal_charge", Tree: TalentTreeNormal, Tier: 2, MaxLevel: 5, Name: "蓄力返还", EffectType: "charge_retain", EffectValue: map[string]any{"retainPercent": 0.40}},
	"normal_chase_up":  {ID: "normal_chase_up", Tree: TalentTreeNormal, Tier: 2, MaxLevel: 5, Name: "追击强化", EffectType: "chase_upgrade", EffectValue: map[string]any{"chaseRatio": 1.00}},
	"normal_combo_ext": {ID: "normal_combo_ext", Tree: TalentTreeNormal, Tier: 2, MaxLevel: 5, Name: "连击扩展", EffectType: "combo_extend", EffectValue: map[string]any{"extraHits": 30.0}},
	"normal_encircle":  {ID: "normal_encircle", Tree: TalentTreeNormal, Tier: 3, MaxLevel: 5, Name: "围剿", EffectType: "per_part_damage", EffectValue: map[string]any{"percentPerPart": 0.20}},
	"normal_low_hp":    {ID: "normal_low_hp", Tree: TalentTreeNormal, Tier: 3, MaxLevel: 5, Name: "残血收割", EffectType: "low_hp_bonus", EffectValue: map[string]any{"hpThreshold": 0.40, "multiplier": 3.0}},
	"normal_ultimate":  {ID: "normal_ultimate", Tree: TalentTreeNormal, Tier: 4, MaxLevel: 5, Name: "白银风暴", EffectType: "silver_storm", EffectValue: map[string]any{"triggerHits": 15, "treatAllAs": "soft"}},

	// ===== 破甲：碎盾攻坚 =====
	"armor_core":         {ID: "armor_core", Tree: TalentTreeArmor, Tier: 0, MaxLevel: 5, Name: "灭绝穿甲", EffectType: "permanent_armor_pen", EffectValue: map[string]any{"penPercent": 0.60, "collapseTrigger": 50, "collapseDuration": 8}},
	"armor_pen_up":       {ID: "armor_pen_up", Tree: TalentTreeArmor, Tier: 1, MaxLevel: 5, Name: "穿甲强化", EffectType: "armor_pen_extra", EffectValue: map[string]any{"extraPen": 0.50}},
	"armor_boss_hunter":  {ID: "armor_boss_hunter", Tree: TalentTreeArmor, Tier: 1, MaxLevel: 5, Name: "首领猎杀", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.60}},
	"armor_heavy_scale":  {ID: "armor_heavy_scale", Tree: TalentTreeArmor, Tier: 1, MaxLevel: 5, Name: "以强制强", EffectType: "armor_scaling", EffectValue: map[string]any{"damagePer100Armor": 0.04}},
	"armor_heavy_atk":    {ID: "armor_heavy_atk", Tree: TalentTreeArmor, Tier: 2, MaxLevel: 5, Name: "重甲特攻", EffectType: "part_type_damage", EffectValue: map[string]any{"partType": "heavy", "percent": 1.00}},
	"armor_collapse_ext": {ID: "armor_collapse_ext", Tree: TalentTreeArmor, Tier: 2, MaxLevel: 5, Name: "崩塌延长", EffectType: "collapse_extend", EffectValue: map[string]any{"extraDuration": 20.0}},
	"armor_auto_strike":  {ID: "armor_auto_strike", Tree: TalentTreeArmor, Tier: 2, MaxLevel: 5, Name: "自动打击", EffectType: "auto_strike", EffectValue: map[string]any{"triggerCount": 15.0, "damageRatio": 4.0}},
	"armor_ruin":         {ID: "armor_ruin", Tree: TalentTreeArmor, Tier: 3, MaxLevel: 5, Name: "废墟打击", EffectType: "collapse_damage_amp", EffectValue: map[string]any{"extraPercent": 2.0}},
	"armor_pen_convert":  {ID: "armor_pen_convert", Tree: TalentTreeArmor, Tier: 3, MaxLevel: 5, Name: "破甲转化", EffectType: "pen_to_amplify", EffectValue: map[string]any{"convertRatio": 0.60}},
	"armor_ultimate":     {ID: "armor_ultimate", Tree: TalentTreeArmor, Tier: 4, MaxLevel: 5, Name: "审判日", EffectType: "judgment_day", EffectValue: map[string]any{"triggerCount": 60.0, "hpCutPercent": 0.60}},

	// ===== 暴击：致命洞察 =====
	"crit_core":          {ID: "crit_core", Tree: TalentTreeCrit, Tier: 0, MaxLevel: 5, Name: "溢杀", EffectType: "overkill", EffectValue: map[string]any{"baseCritBonus": 0.35, "overflowToCritDmg": 0.02, "omenPerWeakCrit": 2, "critDmgPerOmen": 0.008}},
	"crit_cruel":         {ID: "crit_cruel", Tree: TalentTreeCrit, Tier: 1, MaxLevel: 5, Name: "残忍", EffectType: "crit_damage_bonus", EffectValue: map[string]any{"percent": 1.20}},
	"crit_skinner":       {ID: "crit_skinner", Tree: TalentTreeCrit, Tier: 1, MaxLevel: 5, Name: "剥皮", EffectType: "force_weak", EffectValue: map[string]any{"chance": 0.50, "duration": 8}},
	"crit_doom_judgment": {ID: "crit_doom_judgment", Tree: TalentTreeCrit, Tier: 1, MaxLevel: 5, Name: "末日审判", EffectType: "doom_mark", EffectValue: map[string]any{"markCount": 2.0, "omenPerMark": 25.0, "hpThreshold": 0.30}},
	"crit_bleed":         {ID: "crit_bleed", Tree: TalentTreeCrit, Tier: 2, MaxLevel: 5, Name: "致命出血", EffectType: "bleed", EffectValue: map[string]any{"duration": 4, "damageRatio": 1.00}},
	"crit_omen_kill":     {ID: "crit_omen_kill", Tree: TalentTreeCrit, Tier: 2, MaxLevel: 5, Name: "斩杀预兆", EffectType: "omen_low_hp", EffectValue: map[string]any{"hpThreshold": 0.50, "dmgPerOmen": 0.02}},
	"crit_omen_reap":     {ID: "crit_omen_reap", Tree: TalentTreeCrit, Tier: 2, MaxLevel: 5, Name: "死兆收割", EffectType: "omen_reap_passive", EffectValue: map[string]any{"thresholds": []float64{30, 60, 90, 120}, "damageMult": []float64{1.5, 2.0, 2.5, 3.0}}},
	"crit_final_cut":     {ID: "crit_final_cut", Tree: TalentTreeCrit, Tier: 3, MaxLevel: 5, Name: "终末血斩", EffectType: "final_cut", EffectValue: map[string]any{"critCount": 80.0, "hpCutPercent": 0.15, "cooldown": 30}},
	"crit_death_ecstasy": {ID: "crit_death_ecstasy", Tree: TalentTreeCrit, Tier: 4, MaxLevel: 5, Name: "死亡狂喜", EffectType: "death_ecstasy_ult", EffectValue: map[string]any{"omenCost": 100.0, "critDmgMult": 1.0}},

	// ===== 均衡攻势 小节点 =====
	"normal_filler_t1a": {ID: "normal_filler_t1a", Tree: TalentTreeNormal, Tier: 1, MaxLevel: 5, Name: "锐锋", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.15}},
	"normal_filler_t1b": {ID: "normal_filler_t1b", Tree: TalentTreeNormal, Tier: 1, MaxLevel: 5, Name: "乱舞", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.12}},
	"normal_filler_t2a": {ID: "normal_filler_t2a", Tree: TalentTreeNormal, Tier: 2, MaxLevel: 5, Name: "追猎", EffectType: "chase_ratio_bonus", EffectValue: map[string]any{"percent": 0.15}},
	"normal_filler_t2b": {ID: "normal_filler_t2b", Tree: TalentTreeNormal, Tier: 2, MaxLevel: 5, Name: "穿刺", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.12}},
	"normal_filler_t3a": {ID: "normal_filler_t3a", Tree: TalentTreeNormal, Tier: 3, MaxLevel: 5, Name: "狩猎", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.20}},
	"normal_filler_t3b": {ID: "normal_filler_t3b", Tree: TalentTreeNormal, Tier: 3, MaxLevel: 5, Name: "铁腕", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.15}},

	// ===== 碎盾攻坚 小节点 =====
	"armor_filler_t1a": {ID: "armor_filler_t1a", Tree: TalentTreeArmor, Tier: 1, MaxLevel: 5, Name: "破岩", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.15}},
	"armor_filler_t1b": {ID: "armor_filler_t1b", Tree: TalentTreeArmor, Tier: 1, MaxLevel: 5, Name: "凿裂", EffectType: "armor_pen_extra", EffectValue: map[string]any{"extraPen": 0.08}},
	"armor_filler_t2a": {ID: "armor_filler_t2a", Tree: TalentTreeArmor, Tier: 2, MaxLevel: 5, Name: "瓦解", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.12}},
	"armor_filler_t2b": {ID: "armor_filler_t2b", Tree: TalentTreeArmor, Tier: 2, MaxLevel: 5, Name: "碾碎", EffectType: "armor_scaling", EffectValue: map[string]any{"damagePer100Armor": 0.015}},
	"armor_filler_t3a": {ID: "armor_filler_t3a", Tree: TalentTreeArmor, Tier: 3, MaxLevel: 5, Name: "碎颅", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.20}},
	"armor_filler_t3b": {ID: "armor_filler_t3b", Tree: TalentTreeArmor, Tier: 3, MaxLevel: 5, Name: "摧坚", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.12}},

	// ===== 致命洞察 小节点 =====
	"crit_filler_t1a": {ID: "crit_filler_t1a", Tree: TalentTreeCrit, Tier: 1, MaxLevel: 5, Name: "锐眼", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.15}},
	"crit_filler_t1b": {ID: "crit_filler_t1b", Tree: TalentTreeCrit, Tier: 1, MaxLevel: 5, Name: "残酷", EffectType: "crit_damage_bonus", EffectValue: map[string]any{"percent": 0.15}},
	"crit_filler_t2a": {ID: "crit_filler_t2a", Tree: TalentTreeCrit, Tier: 2, MaxLevel: 5, Name: "深创", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.12}},
	"crit_filler_t2b": {ID: "crit_filler_t2b", Tree: TalentTreeCrit, Tier: 2, MaxLevel: 5, Name: "喋血", EffectType: "omen_crit_damage", EffectValue: map[string]any{"critDmgPerOmen": 0.003}},
	"crit_filler_t3a": {ID: "crit_filler_t3a", Tree: TalentTreeCrit, Tier: 3, MaxLevel: 5, Name: "追魂", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.20}},
	"crit_filler_t3b": {ID: "crit_filler_t3b", Tree: TalentTreeCrit, Tier: 3, MaxLevel: 5, Name: "暴虐", EffectType: "crit_damage_bonus", EffectValue: map[string]any{"percent": 0.15}},
}

var talentTierMainCosts = map[int]int64{
	0: TalentCostTier0Main,
	1: TalentCostTier1Main,
	2: TalentCostTier2Main,
	3: TalentCostTier3Main,
	4: TalentCostTier4Main,
}

var talentTierFillerCosts = map[int]int64{
	1: TalentCostTier1Filler,
	2: TalentCostTier2Filler,
	3: TalentCostTier3Filler,
}

// tierNodeCount 每层节点总数（主 + 小），用于层锁判定。
var tierNodeCount = map[int]int{
	0: 1, // 1 核心
	1: 5, // 3 主 + 2 小
	2: 5, // 3 主 + 2 小
	3: 4, // 2 主 + 2 小
	4: 1, // 1 终极
}

// tierCompletionBonusLabels 层满奖励文案，供前端直接展示。
var tierCompletionBonusLabels = map[TalentTree]map[int]string{
	TalentTreeNormal: {
		0: "全伤害 +10%",
		1: "攻击力 +15%",
		2: "触发 -20 次 + 全伤害 +10%",
		3: "全伤害 +15%",
		4: "+5 段 + 全伤害 +10%",
	},
	TalentTreeArmor: {
		0: "全伤害 +10%",
		1: "崩塌触发 -30 + 全伤害 +10%",
		2: "护甲穿透 +15%",
		3: "崩塌易伤 +15%",
		4: "审判日削除 +10%",
	},
	TalentTreeCrit: {
		0: "全伤害 +10%",
		1: "暴击率 +10%",
		2: "斩杀血线 +5%",
		3: "每层死兆暴伤 +0.5%",
		4: "狂喜倍率 +2x",
	},
}

func isFillerTalentID(id string) bool {
	return strings.HasSuffix(id, "_t1a") || strings.HasSuffix(id, "_t1b") ||
		strings.HasSuffix(id, "_t2a") || strings.HasSuffix(id, "_t2b") ||
		strings.HasSuffix(id, "_t3a") || strings.HasSuffix(id, "_t3b")
}

func init() {
	for id, def := range talentDefs {
		if def.MaxLevel <= 0 {
			def.MaxLevel = TalentDefaultMaxLevel
		}
		var baseCost int64
		if isFillerTalentID(id) {
			baseCost = talentTierFillerCosts[def.Tier]
		} else {
			var ok bool
			baseCost, ok = talentTierMainCosts[def.Tier]
			if !ok {
				baseCost = 0
			}
		}
		def.Cost = baseCost
		talentDefs[id] = def
	}
}

// isLearnedTierFull 检查指定天赋树某一层的所有节点（主 + 小）是否已全部学习。
func isLearnedTierFull(tree TalentTree, tier int, talents map[string]int) bool {
	needed := tierNodeCount[tier]
	if needed == 0 {
		return true
	}
	count := 0
	for id, level := range talents {
		if level <= 0 {
			continue
		}
		def, ok := talentDefs[id]
		if !ok {
			continue
		}
		if def.Tree == tree && def.Tier == tier {
			count++
		}
	}
	return count >= needed
}

// TalentEffectDescription 返回 Lv1 天赋效果中文描述，供静态定义展示。
func TalentEffectDescription(def TalentDef) string {
	return TalentEffectDescriptionForLevel(def, 1)
}

// TalentEffectDescriptionForLevel 返回指定等级下的天赋效果中文描述，供前端直接展示。
func TalentEffectDescriptionForLevel(def TalentDef, level int) string {
	effectType := strings.TrimSpace(def.EffectType)
	value, _ := def.EffectValue.(map[string]any)
	currentFactor := max(level, 1)
	switch effectType {
	case "storm_combo":
		trigger := talentInt(value["triggerCount"])
		hits := talentIntScaled(value["extraHits"], currentFactor)
		ratio := talentPercentScaled(value["chaseRatio"], currentFactor)
		if def.ID == "normal_core" {
			trigger = normalCoreTriggerCountForLevel(currentFactor)
			hits = normalCoreExtraHitsForLevel(currentFactor)
			ratio = talentPercent(value["chaseRatio"])
		}
		return fmt.Sprintf("每 %d 次点击触发追击爆发，造成 基础伤害 x %s x %d 段总伤。可无限触发。",
			trigger, ratio, hits,
		)
	case "attack_power_percent":
		return fmt.Sprintf("攻击力提升 %s", talentPercentScaled(value["percent"], currentFactor))
	case "all_damage_amplify":
		return fmt.Sprintf("所有伤害提升 %s", talentPercentScaled(value["percent"], currentFactor))
	case "part_type_damage":
		percent := talentPercentScaled(value["percent"], currentFactor)
		if def.ID == "normal_soft_atk" {
			percent = talentPercent(normalCoreScaledPartDamage(currentFactor, 0.80, 3.00))
		}
		if def.ID == "armor_heavy_atk" {
			percent = talentPercent(normalCoreScaledPartDamage(currentFactor, 1.00, 3.00))
		}
		return fmt.Sprintf("%s伤害提升 %s", talentPartTypeLabel(value["partType"]), percent)
	case "charge_retain":
		retain := talentPercentScaled(value["retainPercent"], currentFactor)
		if def.ID == "normal_charge" {
			retain = talentPercent(normalChargeRetainPercentForLevel(currentFactor))
		}
		return fmt.Sprintf("追击爆发触发后，该部位连击进度保留 %s（从30%%开始重新计数）。被动生效。", retain)
	case "chase_upgrade":
		ratio := talentPercentScaled(value["chaseRatio"], currentFactor)
		if def.ID == "normal_chase_up" {
			ratio = talentPercent(normalChaseUpgradeRatioForLevel(currentFactor))
		}
		return fmt.Sprintf("追击爆发单段倍率从50%%提升到 %s。被动生效。", ratio)
	case "combo_extend":
		hits := talentIntScaled(value["extraHits"], currentFactor)
		if def.ID == "normal_combo_ext" {
			hits = normalComboExtendHitsForLevel(currentFactor)
		}
		return fmt.Sprintf("追击爆发段数从15增加到 %d。被动生效。", hits)
	case "per_part_damage":
		return fmt.Sprintf("每个存活部位额外增加 %s 全伤害。被动生效。", talentPercentScaled(value["percentPerPart"], currentFactor))
	case "low_hp_bonus":
		threshold := talentPercentScaled(value["hpThreshold"], currentFactor)
		multiplier := talentFloat(value["multiplier"]) * float64(currentFactor)
		if def.ID == "normal_low_hp" {
			threshold = talentPercent(normalLowHPThresholdForLevel(currentFactor))
			multiplier = normalLowHPMultiplierForLevel(currentFactor)
		}
		return fmt.Sprintf("部位剩余血量低于 %s 时，伤害x%.0f。被动生效。", threshold, multiplier)
	case "silver_storm":
		return fmt.Sprintf("任意部位被击碎时立即触发，持续%d秒内所有部位视为%s（x1.0系数）。每部位击碎均可触发。", normalSilverStormDurationForLevel(currentFactor), talentPartTypeLabel(value["treatAllAs"]))
	case "permanent_armor_pen":
		penPercent := talentPercentScaled(value["penPercent"], currentFactor)
		collapseTrigger := talentInt(value["collapseTrigger"])
		if def.ID == "armor_core" {
			penPercent = talentPercent(armorCorePenPercentForLevel(currentFactor))
			collapseTrigger = armorCoreCollapseTriggerForLevel(currentFactor)
		}
		return fmt.Sprintf("常驻 %s 护甲穿透。对重甲部位累计 %d 次命中后该部位护甲归零 %d 秒（崩塌）。同一部位可多次触发。", penPercent, collapseTrigger, talentInt(value["collapseDuration"]))
	case "armor_pen_extra":
		extraPen := talentPercentScaled(value["extraPen"], currentFactor)
		if def.ID == "armor_pen_up" {
			extraPen = talentPercent(armorPenUpExtraForLevel(currentFactor))
		}
		return fmt.Sprintf("额外护甲穿透 %s", extraPen)
	case "armor_scaling":
		percent := talentPercentScaled(value["damagePer100Armor"], currentFactor)
		if def.ID == "armor_heavy_scale" {
			percent = talentPercent(armorHeavyScaleForLevel(currentFactor))
		}
		return fmt.Sprintf("每 100 护甲额外获得 %s 伤害增幅", percent)
	case "collapse_extend":
		duration := talentIntScaled(value["extraDuration"], currentFactor)
		if def.ID == "armor_collapse_ext" {
			duration = armorCollapseExtendForLevel(currentFactor)
		}
		return fmt.Sprintf("崩塌持续时间从8秒延长到 %d 秒。被动生效。", duration)
	case "auto_strike":
		triggerCount := talentInt(value["triggerCount"])
		ratio := talentFloat(value["damageRatio"]) * float64(currentFactor)
		if def.ID == "armor_auto_strike" {
			triggerCount = armorAutoStrikeTriggerCountForLevel(currentFactor)
			ratio = armorAutoStrikeRatioForLevel(currentFactor)
		}
		return fmt.Sprintf("5秒内连续命中同一重甲部位 %d 次后，追加一次 %.1fx 攻击力的碎甲重击。触发后重置累计。", triggerCount, ratio)
	case "collapse_damage_amp":
		percent := talentPercentScaled(value["extraPercent"], currentFactor)
		if def.ID == "armor_ruin" {
			percent = talentPercent(armorRuinAmpForLevel(currentFactor))
		}
		return fmt.Sprintf("攻击处于崩塌状态的部位时，额外增伤 %s。被动生效。", percent)
	case "pen_to_amplify":
		ratio := talentPercentScaled(value["convertRatio"], currentFactor)
		if def.ID == "armor_pen_convert" {
			ratio = talentPercent(armorPenConvertRatioForLevel(currentFactor))
		}
		return fmt.Sprintf("将 %s 的护甲穿透值转化为全伤害加成。被动生效。", ratio)
	case "judgment_day":
		triggerCount := talentInt(value["triggerCount"])
		hpCutPercent := talentPercentScaled(value["hpCutPercent"], currentFactor)
		if def.ID == "armor_ultimate" {
			triggerCount = armorUltimateTriggerCountForLevel(currentFactor)
			hpCutPercent = talentPercent(armorUltimateHpCutForLevel(currentFactor))
		}
		return fmt.Sprintf("对同一重甲部位累计 %d 次命中后，立即削除该部位 %s 最大生命值。每部位每场战斗仅一次。", triggerCount, hpCutPercent)
	case "overkill":
		baseCritBonus := talentPercentScaled(value["baseCritBonus"], currentFactor)
		if def.ID == "crit_core" {
			baseCritBonus = talentPercent(critCoreBaseCritBonusForLevel(currentFactor))
		}
		return fmt.Sprintf("基础暴击率 +%s。暴击率超过100%%的部分按 %s 比例转为暴伤；每层死兆额外提供 %s 暴击伤害。死兆最多累积至 %d 层，溢出转化为额外伤害。弱点暴击+2层死兆，普通暴击+1层，击碎部位+5层。", baseCritBonus, talentPercent(value["overflowToCritDmg"]), talentPercent(critOmenResonateForLevel(currentFactor)), TalentOmenStackCap)
	case "omen_crit_damage":
		critDmgPerOmen := talentFloat(value["critDmgPerOmen"]) * float64(currentFactor)
		return fmt.Sprintf("每层死兆叠加 %s 暴击伤害（例：100层=+%.0f%%暴伤）。", talentPercent(critDmgPerOmen), critDmgPerOmen*100)
	case "crit_damage_bonus":
		percent := talentPercentScaled(value["percent"], currentFactor)
		if def.ID == "crit_cruel" {
			percent = talentPercent(critCruelBonusForLevel(currentFactor))
		}
		return fmt.Sprintf("暴击伤害额外提升 %s", percent)
	case "force_weak":
		chance := talentPercentScaled(value["chance"], currentFactor)
		duration := talentIntScaled(value["duration"], currentFactor)
		if def.ID == "crit_skinner" {
			chance = talentPercent(critSkinnerChanceForLevel(currentFactor))
			duration = critSkinnerDurationForLevel(currentFactor)
		}
		return fmt.Sprintf("暴击时有 %s 概率将当前部位视为弱点（x2.5系数），持续 %d 秒。", chance, duration)
	case "bleed":
		ratio := talentPercentScaled(value["damageRatio"], currentFactor)
		if def.ID == "crit_bleed" {
			ratio = talentPercent(critBleedRatioForLevel(currentFactor))
		}
		return fmt.Sprintf("暴击时附加真伤 = 本次伤害 x %s。一次性结算。", ratio)
	case "omen_low_hp":
		hpThreshold := talentPercentScaled(value["hpThreshold"], currentFactor)
		dmgPerOmen := talentPercentScaled(value["dmgPerOmen"], currentFactor)
		if def.ID == "crit_omen_kill" {
			hpThreshold = talentPercent(critOmenKillThresholdForLevel(currentFactor))
			dmgPerOmen = talentPercent(critOmenKillDmgPerOmenForLevel(currentFactor))
		}
		return fmt.Sprintf("部位血量低于 %s 时，每层死兆额外 +%s 伤害（例：47层=+47%%）。被动生效。", hpThreshold, dmgPerOmen)
	case "omen_reap_passive":
		thresholds := "30/60/90/120"
		mults := "×1.5/×2.0/×2.5/×3.0"
		return fmt.Sprintf("死兆达%s层时，伤害自动提升至%s（不消耗层数）。被动生效。", thresholds, mults)
	case "death_ecstasy_ult":
		mult := talentMultiplierScaled(value["critDmgMult"], currentFactor)
		if def.ID == "crit_death_ecstasy" {
			mult = fmt.Sprintf("×%.1f", critDeathEcstasyMultForLevel(currentFactor))
		}
		return fmt.Sprintf("死兆达到%d层时消耗%d层，造成 baseDamage × 层数，再乘以 %s 的巨额伤害。", talentInt(value["omenCost"]), talentInt(value["omenCost"]), mult)
	case "final_cut":
		critCount := talentInt(value["critCount"])
		hpCutPercent := talentPercentScaled(value["hpCutPercent"], currentFactor)
		if def.ID == "crit_final_cut" {
			critCount = critFinalCutCountForLevel(currentFactor)
			hpCutPercent = talentPercent(critFinalCutHpCutForLevel(currentFactor))
		}
		return fmt.Sprintf("累计 %d 次暴击后削除Boss最大生命值的 %s（%d 秒冷却）。", critCount, hpCutPercent, talentInt(value["cooldown"]))
	case "doom_mark":
		markCount := talentIntScaled(value["markCount"], currentFactor)
		omenPerMark := talentIntScaled(value["omenPerMark"], currentFactor)
		if def.ID == "crit_doom_judgment" {
			markCount = critDoomMarkCountForLevel(currentFactor)
			omenPerMark = critDoomOmenPerMarkForLevel(currentFactor)
		}
		return fmt.Sprintf("开局随机标记%d个部位。被标记部位被击碎时触发+%d死兆。可升级增加标记数和层数。", markCount, omenPerMark)
	case "chase_ratio_bonus":
		return fmt.Sprintf("追击爆发单段倍率额外 +%s。被动生效。", talentPercentScaled(value["percent"], currentFactor))
	default:
		return "该天赋效果说明暂未配置"
	}
}

func lerpTalentValue(level int, lv1, lv5 float64) float64 {
	if level <= 1 {
		return lv1
	}
	if level >= 5 {
		return lv5
	}
	step := float64(level-1) / 4.0
	return lv1 + (lv5-lv1)*step
}

func lerpTalentInt(level int, lv1, lv5 int) int {
	return int(math.Round(lerpTalentValue(level, float64(lv1), float64(lv5))))
}

func normalCoreTriggerCountForLevel(level int) int {
	return lerpTalentInt(level, 50, 30)
}

func normalCoreExtraHitsForLevel(level int) int {
	return lerpTalentInt(level, 20, 35)
}

func normalChargeRetainPercentForLevel(level int) float64 {
	return lerpTalentValue(level, 0.40, 0.60)
}

func normalChaseUpgradeRatioForLevel(level int) float64 {
	return lerpTalentValue(level, 1.00, 1.50)
}

func normalComboExtendHitsForLevel(level int) int {
	return lerpTalentInt(level, 30, 50)
}

func normalLowHPThresholdForLevel(level int) float64 {
	return lerpTalentValue(level, 0.40, 0.50)
}

func normalLowHPMultiplierForLevel(level int) float64 {
	return lerpTalentValue(level, 3.0, 6.0)
}

func normalSilverStormDurationForLevel(level int) int {
	return lerpTalentInt(level, 15, 20)
}

func armorCorePenPercentForLevel(level int) float64 {
	return lerpTalentValue(level, 0.60, 0.90)
}

func armorCoreCollapseTriggerForLevel(level int) int {
	return lerpTalentInt(level, 50, 20)
}

func armorPenUpExtraForLevel(level int) float64 {
	return lerpTalentValue(level, 0.50, 1.50)
}

func armorHeavyScaleForLevel(level int) float64 {
	return lerpTalentValue(level, 0.04, 0.12)
}

func armorHeavyAtkForLevel(level int) float64 {
	return lerpTalentValue(level, 1.00, 3.00)
}

func armorCollapseExtendForLevel(level int) int {
	return lerpTalentInt(level, 20, 35)
}

func armorAutoStrikeTriggerCountForLevel(level int) int {
	return lerpTalentInt(level, 15, 8)
}

func armorAutoStrikeRatioForLevel(level int) float64 {
	return lerpTalentValue(level, 4.0, 8.0)
}

func armorRuinAmpForLevel(level int) float64 {
	return lerpTalentValue(level, 2.0, 5.0)
}

func armorPenConvertRatioForLevel(level int) float64 {
	return lerpTalentValue(level, 0.60, 1.00)
}

func armorUltimateTriggerCountForLevel(level int) int {
	return lerpTalentInt(level, 60, 30)
}

func armorUltimateHpCutForLevel(level int) float64 {
	return lerpTalentValue(level, 0.60, 0.80)
}

func critCoreBaseCritBonusForLevel(level int) float64 {
	return lerpTalentValue(level, 0.35, 0.75)
}

func critCruelBonusForLevel(level int) float64 {
	return lerpTalentValue(level, 1.20, 4.00)
}

func critDoomMarkCountForLevel(level int) int {
	return lerpTalentInt(level, 2, 15)
}

func critDoomOmenPerMarkForLevel(level int) int {
	return lerpTalentInt(level, 25, 40)
}

func critOmenResonateForLevel(level int) float64 {
	return lerpTalentValue(level, 0.008, 0.020)
}

func critSkinnerChanceForLevel(level int) float64 {
	return lerpTalentValue(level, 0.50, 0.80)
}

func critSkinnerDurationForLevel(level int) int {
	return lerpTalentInt(level, 8, 15)
}

func critBleedRatioForLevel(level int) float64 {
	return lerpTalentValue(level, 1.00, 3.00)
}

func critOmenKillThresholdForLevel(level int) float64 {
	return lerpTalentValue(level, 0.50, 0.65)
}

func critOmenKillDmgPerOmenForLevel(level int) float64 {
	return lerpTalentValue(level, 0.02, 0.05)
}

func critFinalCutCountForLevel(level int) int {
	return lerpTalentInt(level, 80, 40)
}

func critFinalCutHpCutForLevel(level int) float64 {
	return lerpTalentValue(level, 0.15, 0.25)
}

func critDeathEcstasyMultForLevel(level int) float64 {
	return lerpTalentValue(level, 1.0, 3.0)
}

func BuildTalentEffectLines(def TalentDef, currentLevel int) []TalentEffectLine {
	value, _ := def.EffectValue.(map[string]any)
	if len(value) == 0 {
		return nil
	}

	maxLevel := def.MaxLevel
	if maxLevel <= 0 {
		maxLevel = TalentDefaultMaxLevel
	}
	level := max(currentLevel, 0)
	currentFactor := max(level, 1)
	nextLevel := 1
	if level > 0 {
		nextLevel = level + 1
	}
	showNext := level > 0 && level < maxLevel

	lines := make([]TalentEffectLine, 0, 3)
	add := func(label, current, next string) {
		text := current
		if showNext && next != "" && next != current {
			text = current + " → " + next
		}
		lines = append(lines, TalentEffectLine{
			Label: label,
			Text:  text,
		})
	}

	switch strings.TrimSpace(def.EffectType) {
	case "attack_power_percent":
		add("攻击力", talentPercentScaled(value["percent"], currentFactor), talentPercentScaled(value["percent"], nextLevel))
	case "all_damage_amplify":
		add("全伤害", talentPercentScaled(value["percent"], currentFactor), talentPercentScaled(value["percent"], nextLevel))
	case "part_type_damage":
		if def.ID == "normal_soft_atk" {
			add(talentPartTypeLabel(value["partType"])+"伤害", talentPercent(normalCoreScaledPartDamage(currentFactor, 0.80, 3.00)), talentPercent(normalCoreScaledPartDamage(nextLevel, 0.80, 3.00)))
			break
		}
		if def.ID == "armor_heavy_atk" {
			add(talentPartTypeLabel(value["partType"])+"伤害", talentPercent(normalCoreScaledPartDamage(currentFactor, 1.00, 3.00)), talentPercent(normalCoreScaledPartDamage(nextLevel, 1.00, 3.00)))
			break
		}
		add(talentPartTypeLabel(value["partType"])+"伤害", talentPercentScaled(value["percent"], currentFactor), talentPercentScaled(value["percent"], nextLevel))
	case "charge_retain":
		if def.ID == "normal_charge" {
			add("追击保留", talentPercent(normalChargeRetainPercentForLevel(currentFactor)), talentPercent(normalChargeRetainPercentForLevel(nextLevel)))
			break
		}
		add("追击保留", talentPercentScaled(value["retainPercent"], currentFactor), talentPercentScaled(value["retainPercent"], nextLevel))
	case "chase_upgrade":
		if def.ID == "normal_chase_up" {
			add("追击倍率", talentPercent(normalChaseUpgradeRatioForLevel(currentFactor)), talentPercent(normalChaseUpgradeRatioForLevel(nextLevel)))
			break
		}
		add("追击倍率", talentPercentScaled(value["chaseRatio"], currentFactor), talentPercentScaled(value["chaseRatio"], nextLevel))
	case "combo_extend":
		if def.ID == "normal_combo_ext" {
			add("追击段数", fmt.Sprintf("%d", normalComboExtendHitsForLevel(currentFactor)), fmt.Sprintf("%d", normalComboExtendHitsForLevel(nextLevel)))
			break
		}
		add("追击段数", talentIntScaledString(value["extraHits"], currentFactor), talentIntScaledString(value["extraHits"], nextLevel))
	case "per_part_damage":
		add("每部位增伤", talentPercentScaled(value["percentPerPart"], currentFactor), talentPercentScaled(value["percentPerPart"], nextLevel))
	case "low_hp_bonus":
		if def.ID == "normal_low_hp" {
			add("低血阈值", talentPercent(normalLowHPThresholdForLevel(currentFactor)), talentPercent(normalLowHPThresholdForLevel(nextLevel)))
			add("伤害倍率", fmt.Sprintf("×%.1f", normalLowHPMultiplierForLevel(currentFactor)), fmt.Sprintf("×%.1f", normalLowHPMultiplierForLevel(nextLevel)))
			break
		}
		add("低血阈值", talentPercentScaled(value["hpThreshold"], currentFactor), talentPercentScaled(value["hpThreshold"], nextLevel))
		add("伤害倍率", talentMultiplierScaled(value["multiplier"], currentFactor), talentMultiplierScaled(value["multiplier"], nextLevel))
	case "silver_storm":
		add("持续轮次", fmt.Sprintf("%d", normalSilverStormDurationForLevel(currentFactor)), fmt.Sprintf("%d", normalSilverStormDurationForLevel(nextLevel)))
	case "permanent_armor_pen":
		if def.ID == "armor_core" {
			add("常驻破甲", talentPercent(armorCorePenPercentForLevel(currentFactor)), talentPercent(armorCorePenPercentForLevel(nextLevel)))
			add("崩塌需命中", fmt.Sprintf("%d", armorCoreCollapseTriggerForLevel(currentFactor)), fmt.Sprintf("%d", armorCoreCollapseTriggerForLevel(nextLevel)))
			break
		}
		add("常驻破甲", talentPercentScaled(value["penPercent"], currentFactor), talentPercentScaled(value["penPercent"], nextLevel))
		add("崩塌需命中", talentIntString(value["collapseTrigger"]), talentIntString(value["collapseTrigger"]))
	case "armor_pen_extra":
		if def.ID == "armor_pen_up" {
			add("额外破甲", talentPercent(armorPenUpExtraForLevel(currentFactor)), talentPercent(armorPenUpExtraForLevel(nextLevel)))
			break
		}
		add("额外破甲", talentPercentScaled(value["extraPen"], currentFactor), talentPercentScaled(value["extraPen"], nextLevel))
	case "armor_scaling":
		if def.ID == "armor_heavy_scale" {
			add("每100甲增伤", talentPercent(armorHeavyScaleForLevel(currentFactor)), talentPercent(armorHeavyScaleForLevel(nextLevel)))
			break
		}
		add("每100甲增伤", talentPercentScaled(value["damagePer100Armor"], currentFactor), talentPercentScaled(value["damagePer100Armor"], nextLevel))
	case "collapse_extend":
		if def.ID == "armor_collapse_ext" {
			add("崩塌持续", fmt.Sprintf("%ds", armorCollapseExtendForLevel(currentFactor)), fmt.Sprintf("%ds", armorCollapseExtendForLevel(nextLevel)))
			break
		}
		add("崩塌持续", talentDurationScaled(value["extraDuration"], currentFactor), talentDurationScaled(value["extraDuration"], nextLevel))
	case "auto_strike":
		if def.ID == "armor_auto_strike" {
			add("触发次数", fmt.Sprintf("%d", armorAutoStrikeTriggerCountForLevel(currentFactor)), fmt.Sprintf("%d", armorAutoStrikeTriggerCountForLevel(nextLevel)))
			add("伤害倍率", fmt.Sprintf("×%.1f", armorAutoStrikeRatioForLevel(currentFactor)), fmt.Sprintf("×%.1f", armorAutoStrikeRatioForLevel(nextLevel)))
			break
		}
		add("触发次数", talentIntString(value["triggerCount"]), talentIntString(value["triggerCount"]))
		add("伤害倍率", talentMultiplierScaled(value["damageRatio"], currentFactor), talentMultiplierScaled(value["damageRatio"], nextLevel))
	case "collapse_damage_amp":
		if def.ID == "armor_ruin" {
			add("崩塌增伤", talentPercent(armorRuinAmpForLevel(currentFactor)), talentPercent(armorRuinAmpForLevel(nextLevel)))
			break
		}
		add("崩塌增伤", talentPercentScaled(value["extraPercent"], currentFactor), talentPercentScaled(value["extraPercent"], nextLevel))
	case "pen_to_amplify":
		if def.ID == "armor_pen_convert" {
			add("破甲转增伤", talentPercent(armorPenConvertRatioForLevel(currentFactor)), talentPercent(armorPenConvertRatioForLevel(nextLevel)))
			break
		}
		add("破甲转增伤", talentPercentScaled(value["convertRatio"], currentFactor), talentPercentScaled(value["convertRatio"], nextLevel))
	case "judgment_day":
		if def.ID == "armor_ultimate" {
			add("触发命中", fmt.Sprintf("%d", armorUltimateTriggerCountForLevel(currentFactor)), fmt.Sprintf("%d", armorUltimateTriggerCountForLevel(nextLevel)))
			add("削除生命", talentPercent(armorUltimateHpCutForLevel(currentFactor)), talentPercent(armorUltimateHpCutForLevel(nextLevel)))
			break
		}
		add("触发命中", talentIntString(value["triggerCount"]), talentIntString(value["triggerCount"]))
		add("削除生命", talentPercentScaled(value["hpCutPercent"], currentFactor), talentPercentScaled(value["hpCutPercent"], nextLevel))
	case "overkill":
		if def.ID == "crit_core" {
			add("暴击率", talentPercent(critCoreBaseCritBonusForLevel(currentFactor)), talentPercent(critCoreBaseCritBonusForLevel(nextLevel)))
			add("溢出转暴伤", talentPercent(value["overflowToCritDmg"]), talentPercent(value["overflowToCritDmg"]))
			add("每层暴伤", talentPercent(critOmenResonateForLevel(currentFactor)), talentPercent(critOmenResonateForLevel(nextLevel)))
			add("弱点暴击获层", talentIntString(value["omenPerWeakCrit"]), talentIntString(value["omenPerWeakCrit"]))
			break
		}
		add("暴击率", talentPercentScaled(value["baseCritBonus"], currentFactor), talentPercentScaled(value["baseCritBonus"], nextLevel))
		add("溢出转暴伤", talentPercent(value["overflowToCritDmg"]), talentPercent(value["overflowToCritDmg"]))
		add("弱点暴击获层", talentIntString(value["omenPerWeakCrit"]), talentIntString(value["omenPerWeakCrit"]))
	case "omen_crit_damage":
		add("每层暴伤", talentPercentScaled(value["critDmgPerOmen"], currentFactor), talentPercentScaled(value["critDmgPerOmen"], nextLevel))
	case "crit_damage_bonus":
		if def.ID == "crit_cruel" {
			add("暴击伤害", talentPercent(critCruelBonusForLevel(currentFactor)), talentPercent(critCruelBonusForLevel(nextLevel)))
			break
		}
		add("暴击伤害", talentPercentScaled(value["percent"], currentFactor), talentPercentScaled(value["percent"], nextLevel))
	case "force_weak":
		if def.ID == "crit_skinner" {
			add("触发几率", talentPercent(critSkinnerChanceForLevel(currentFactor)), talentPercent(critSkinnerChanceForLevel(nextLevel)))
			add("弱点持续", fmt.Sprintf("%ds", critSkinnerDurationForLevel(currentFactor)), fmt.Sprintf("%ds", critSkinnerDurationForLevel(nextLevel)))
			break
		}
		add("触发几率", talentPercentScaled(value["chance"], currentFactor), talentPercentScaled(value["chance"], nextLevel))
		add("弱点持续", talentDurationScaled(value["duration"], currentFactor), talentDurationScaled(value["duration"], nextLevel))
	case "bleed":
		if def.ID == "crit_bleed" {
			add("真伤比例", talentPercent(critBleedRatioForLevel(currentFactor)), talentPercent(critBleedRatioForLevel(nextLevel)))
			break
		}
		add("真伤比例", talentPercentScaled(value["damageRatio"], currentFactor), talentPercentScaled(value["damageRatio"], nextLevel))
	case "omen_low_hp":
		if def.ID == "crit_omen_kill" {
			add("触发阈值", talentPercent(critOmenKillThresholdForLevel(currentFactor)), talentPercent(critOmenKillThresholdForLevel(nextLevel)))
			add("每层增伤", talentPercent(critOmenKillDmgPerOmenForLevel(currentFactor)), talentPercent(critOmenKillDmgPerOmenForLevel(nextLevel)))
			break
		}
		add("触发阈值", talentPercentScaled(value["hpThreshold"], currentFactor), talentPercentScaled(value["hpThreshold"], nextLevel))
		add("每层增伤", talentPercentScaled(value["dmgPerOmen"], currentFactor), talentPercentScaled(value["dmgPerOmen"], nextLevel))
	case "omen_reap_passive":
		add("档位增伤", "30层×1.5 / 60层×2.0 / 90层×2.5 / 120层×3.0", "")
	case "final_cut":
		if def.ID == "crit_final_cut" {
			add("需暴击次数", fmt.Sprintf("%d", critFinalCutCountForLevel(currentFactor)), fmt.Sprintf("%d", critFinalCutCountForLevel(nextLevel)))
			add("削除生命", talentPercent(critFinalCutHpCutForLevel(currentFactor)), talentPercent(critFinalCutHpCutForLevel(nextLevel)))
			break
		}
		add("需暴击次数", talentIntString(value["critCount"]), talentIntString(value["critCount"]))
		add("削除生命", talentPercentScaled(value["hpCutPercent"], currentFactor), talentPercentScaled(value["hpCutPercent"], nextLevel))
	case "death_ecstasy_ult":
		add("消耗层数", talentIntString(value["omenCost"]), talentIntString(value["omenCost"]))
		if def.ID == "crit_death_ecstasy" {
			add("暴伤倍率", fmt.Sprintf("×%.1f", critDeathEcstasyMultForLevel(currentFactor)), fmt.Sprintf("×%.1f", critDeathEcstasyMultForLevel(nextLevel)))
			break
		}
		add("暴伤倍率", talentMultiplierScaled(value["critDmgMult"], currentFactor), talentMultiplierScaled(value["critDmgMult"], nextLevel))
	case "doom_mark":
		if def.ID == "crit_doom_judgment" {
			add("标记数量", fmt.Sprintf("%d", critDoomMarkCountForLevel(currentFactor)), fmt.Sprintf("%d", critDoomMarkCountForLevel(nextLevel)))
			add("每标获层", fmt.Sprintf("%d", critDoomOmenPerMarkForLevel(currentFactor)), fmt.Sprintf("%d", critDoomOmenPerMarkForLevel(nextLevel)))
			break
		}
		add("标记数量", talentIntScaledString(value["markCount"], currentFactor), talentIntScaledString(value["markCount"], nextLevel))
		add("每标获层", talentIntScaledString(value["omenPerMark"], currentFactor), talentIntScaledString(value["omenPerMark"], nextLevel))
	case "chase_ratio_bonus":
		add("追击倍率", talentPercentScaled(value["percent"], currentFactor), talentPercentScaled(value["percent"], nextLevel))
	case "storm_combo":
		currentTrigger := talentIntString(value["triggerCount"])
		nextTrigger := currentTrigger
		if def.ID == "normal_core" {
			currentTrigger = fmt.Sprintf("%d", normalCoreTriggerCountForLevel(currentFactor))
			nextTrigger = fmt.Sprintf("%d", normalCoreTriggerCountForLevel(nextLevel))
		}
		add("触发次数", currentTrigger, nextTrigger)
		if def.ID == "normal_core" {
			add("追击段数", fmt.Sprintf("%d", normalCoreExtraHitsForLevel(currentFactor)), fmt.Sprintf("%d", normalCoreExtraHitsForLevel(nextLevel)))
		} else {
			add("追击段数", talentIntScaledString(value["extraHits"], currentFactor), talentIntScaledString(value["extraHits"], nextLevel))
		}
		add("追击倍率", talentPercentScaled(value["chaseRatio"], currentFactor), talentPercentScaled(value["chaseRatio"], nextLevel))
	}

	return lines
}

func BuildTalentEffectLineMap(state *TalentState) map[string][]TalentEffectLine {
	result := make(map[string][]TalentEffectLine, len(talentDefs))
	for id, def := range talentDefs {
		level := GetTalentLevel(state, id)
		result[id] = BuildTalentEffectLines(def, level)
	}
	return result
}

func BuildTalentEffectDescriptionMap(state *TalentState) map[string]string {
	result := make(map[string]string, len(talentDefs))
	for id, def := range talentDefs {
		level := GetTalentLevel(state, id)
		result[id] = TalentEffectDescriptionForLevel(def, level)
	}
	return result
}

// TalentTierCompletionBonusLabels 返回指定天赋树的层满奖励文案（key 为层级）。
func TalentTierCompletionBonusLabels(tree TalentTree) map[int]string {
	labels, ok := tierCompletionBonusLabels[tree]
	if !ok || len(labels) == 0 {
		return map[int]string{}
	}
	out := make(map[int]string, len(labels))
	maps.Copy(out, labels)
	return out
}

func talentFloat(v any) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case int32:
		return float64(value)
	case int16:
		return float64(value)
	case int8:
		return float64(value)
	case uint:
		return float64(value)
	case uint64:
		return float64(value)
	case uint32:
		return float64(value)
	case uint16:
		return float64(value)
	case uint8:
		return float64(value)
	case json.Number:
		f, err := value.Float64()
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func talentInt(v any) int {
	return int(talentFloat(v))
}

func talentIntScaled(v any, factor int) int {
	return int(math.Round(talentFloat(v) * float64(factor)))
}

func talentPercent(v any) string {
	pct := talentFloat(v) * 100
	abs := pct
	if abs < 0 {
		abs = -abs
	}
	if abs > 0 && abs < 1 {
		return fmt.Sprintf("%.2f%%", pct)
	}
	if abs >= 1 && abs < 10 {
		return fmt.Sprintf("%.1f%%", pct)
	}
	return fmt.Sprintf("%.0f%%", pct)
}

func talentPartTypeLabel(v any) string {
	part, _ := v.(string)
	switch strings.TrimSpace(part) {
	case "soft":
		return "软组织"
	case "heavy":
		return "重甲"
	case "weak":
		return "弱点"
	default:
		if part == "" {
			return "未知部位"
		}
		return part
	}
}

func talentPercentScaled(v any, factor int) string {
	return talentPercent(talentFloat(v) * float64(factor))
}

func talentIntScaledString(v any, factor int) string {
	return fmt.Sprintf("%d", int(math.Round(talentFloat(v)*float64(factor))))
}

func talentIntString(v any) string {
	return fmt.Sprintf("%d", talentInt(v))
}

func talentDurationScaled(v any, factor int) string {
	return fmt.Sprintf("%ds", int(math.Round(talentFloat(v)*float64(factor))))
}

func talentDurationString(v any) string {
	return fmt.Sprintf("%ds", talentInt(v))
}

func talentMultiplierScaled(v any, factor int) string {
	return fmt.Sprintf("×%.1f", talentFloat(v)*float64(factor))
}

// GetTalentDef 返回指定 ID 的天赋定义。
func GetTalentDef(id string) (TalentDef, bool) {
	def, ok := talentDefs[id]
	return def, ok
}

// GetTreeTalents 返回指定天赋树的所有定义。
func GetTreeTalents(tree TalentTree) []TalentDef {
	var result []TalentDef
	for _, def := range talentDefs {
		if def.Tree == tree {
			result = append(result, def)
		}
	}
	slices.SortFunc(result, func(a, b TalentDef) int {
		if a.Tier != b.Tier {
			return a.Tier - b.Tier
		}
		return strings.Compare(a.ID, b.ID)
	})
	return result
}

func (s *Store) talentKey(nickname string) string {
	return s.namespace + "player:talents:" + nickname
}

// GetTalentState 获取玩家天赋状态。
func (s *Store) GetTalentState(ctx context.Context, nickname string) (*TalentState, error) {
	values, err := s.client.HGetAll(ctx, s.talentKey(nickname)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return &TalentState{Talents: make(map[string]int)}, nil
	}

	state := &TalentState{}
	talentsRaw := values["talents"]
	if talentsRaw != "" {
		talents := make(map[string]int)
		if err := sonic.Unmarshal([]byte(talentsRaw), &talents); err != nil {
			// 兼容旧格式 []string → 迁移为 map[string]int (均为 Lv1)
			var oldTalents []string
			if err2 := sonic.Unmarshal([]byte(talentsRaw), &oldTalents); err2 != nil {
				return nil, err
			}
			for _, id := range oldTalents {
				talents[id] = 1
			}
		}
		state.Talents = talents
	} else {
		state.Talents = make(map[string]int)
	}

	return state, nil
}

func (s *Store) compiledTalentSetForNickname(ctx context.Context, nickname string) (*CompiledTalentSet, error) {
	if compiled, ok := s.cachedCompiledTalentSet(nickname); ok {
		return compiled, nil
	}

	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return nil, err
	}

	compiled := compileTalentSet(state)
	s.storeCompiledTalentCache(nickname, compiled)
	return compiled, nil
}

// UpgradeTalent 升级天赋节点到指定等级。
func (s *Store) UpgradeTalent(ctx context.Context, nickname string, talentID string, targetLevel int) error {
	if targetLevel < 1 {
		return ErrTalentInvalidLevel
	}

	def, ok := talentDefs[talentID]
	if !ok {
		return ErrTalentNotFound
	}
	if targetLevel > def.MaxLevel {
		return ErrTalentMaxLevel
	}

	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return err
	}
	currentLevel := GetTalentLevel(state, talentID)
	if currentLevel >= targetLevel {
		return ErrTalentAlreadyLearned
	}
	if currentLevel == 0 && targetLevel >= 1 {
		if def.Tier > 0 {
			if !isLearnedTierFull(def.Tree, def.Tier-1, state.Talents) {
				return ErrTalentTierLocked
			}
		}
	}

	diff := TalentLevelCostDiff(def.Cost, currentLevel, targetLevel)
	if diff <= 0 {
		return ErrTalentInvalidCost
	}

	resources, err := s.resourcesForNickname(ctx, nickname)
	if err != nil {
		return err
	}
	if resources.TalentPoints < diff {
		return ErrTalentPointsInsufficient
	}

	state.Talents[talentID] = targetLevel
	talentsJSON, err := sonic.Marshal(state.Talents)
	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.talentKey(nickname), "talents", string(talentsJSON))
	pipe.HIncrBy(ctx, s.resourceKey(nickname), "talent_points", -diff)
	_, err = pipe.Exec(ctx)
	if err == nil {
		s.invalidatePlayerCombatCaches(nickname)
	}
	return err
}

// ResetTalents 重置所有已学习天赋。
func (s *Store) ResetTalents(ctx context.Context, nickname string) error {
	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return err
	}
	if state == nil || len(state.Talents) == 0 {
		return s.client.HSet(ctx, s.talentKey(nickname), "talents", "{}").Err()
	}

	var refund int64
	for id, level := range state.Talents {
		def, ok := talentDefs[id]
		if !ok {
			continue
		}
		refund += TalentCumulativeCost(def.Cost, level)
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.talentKey(nickname), "talents", "{}")
	if refund > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(nickname), "talent_points", refund)
	}
	_, err = pipe.Exec(ctx)
	if err == nil {
		s.invalidatePlayerCombatCaches(nickname)
	}
	return err
}

// TalentModifiers 聚集所有已学习天赋的效果修改器。
type TalentModifiers struct {
	AttackPowerPercent     float64 `json:"attackPowerPercent"`
	AllDamageAmplify       float64 `json:"allDamageAmplify"`
	ArmorPenExtra          float64 `json:"armorPenExtra"`
	CritDamagePercentBonus float64 `json:"critDamagePercentBonus"`
	PenToAmplifyRatio      float64 `json:"penToAmplifyRatio"`
	OverflowToCritDmgRatio float64 `json:"overflowToCritDmgRatio"`
	PerPartDamagePercent   float64 `json:"perPartDamagePercent"`
	LowHpMultiplier        float64 `json:"lowHpMultiplier"`
	LowHpThreshold         float64 `json:"lowHpThreshold"`
	CollapseDuration       int     `json:"collapseDuration"`
	// 小节点新效果
	ChaseRatioBonus  float64 `json:"chaseRatioBonus"`  // 追击倍率加成
	OmenCritDmgExtra float64 `json:"omenCritDmgExtra"` // 每层死兆额外暴伤
	// 层满奖励效果
	StormTriggerReduce     float64 `json:"stormTriggerReduce"`
	StormExtraHits         int     `json:"stormExtraHits"`
	CollapseTriggerReduce  int     `json:"collapseTriggerReduce"`
	CollapseVulnerability  float64 `json:"collapseVulnerability"`
	JudgmentDayBoost       float64 `json:"judgmentDayBoost"`
	CritRateBonus          float64 `json:"critRateBonus"`
	OmenKillThresholdRaise float64 `json:"omenKillThresholdRaise"`
	DoomMultBoost          float64 `json:"doomMultBoost"`
	// 已学习天赋 ID 列表，供具体逻辑判断
	Learned       []string             `json:"-"`
	PartTypeBonus map[PartType]float64 `json:"-"`
}

// ComputeTalentModifiers 计算玩家天赋提供的全量修正。
func (s *Store) ComputeTalentModifiers(ctx context.Context, nickname string) (*TalentModifiers, error) {
	compiled, err := s.compiledTalentSetForNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}
	if compiled == nil || compiled.Modifiers == nil {
		return &TalentModifiers{PartTypeBonus: make(map[PartType]float64)}, nil
	}
	return cloneTalentModifiers(compiled.Modifiers), nil
}

func applyTierCompletionBonus(mods *TalentModifiers, treeStr string, tier int) {
	switch {
	case treeStr == "normal" && tier == 0:
		mods.AllDamageAmplify += 0.10
	case treeStr == "normal" && tier == 1:
		mods.AttackPowerPercent += 0.15
	case treeStr == "normal" && tier == 2:
		mods.StormTriggerReduce += 20
		mods.AllDamageAmplify += 0.10
	case treeStr == "normal" && tier == 3:
		mods.AllDamageAmplify += 0.15
	case treeStr == "normal" && tier == 4:
		mods.StormExtraHits += 5
		mods.AllDamageAmplify += 0.10
	case treeStr == "armor" && tier == 0:
		mods.AllDamageAmplify += 0.10
	case treeStr == "armor" && tier == 1:
		mods.CollapseTriggerReduce += 30
		mods.AllDamageAmplify += 0.10
	case treeStr == "armor" && tier == 2:
		mods.ArmorPenExtra += 0.15
	case treeStr == "armor" && tier == 3:
		mods.CollapseVulnerability += 0.15
	case treeStr == "armor" && tier == 4:
		mods.JudgmentDayBoost += 0.10
	case treeStr == "crit" && tier == 0:
		mods.AllDamageAmplify += 0.10
	case treeStr == "crit" && tier == 1:
		mods.CritRateBonus += 0.10
	case treeStr == "crit" && tier == 2:
		mods.OmenKillThresholdRaise += 0.05
	case treeStr == "crit" && tier == 3:
		mods.OmenCritDmgExtra += 0.005
	case treeStr == "crit" && tier == 4:
		mods.DoomMultBoost += 2.0
	}
}

func normalCoreScaledPartDamage(level int, lv1, lv5 float64) float64 {
	return lerpTalentValue(level, lv1, lv5)
}

// HasTalent 检查玩家是否已学习指定天赋。
func (s *Store) HasTalent(ctx context.Context, nickname string, talentID string) (bool, error) {
	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return false, err
	}
	return HasTalentLearned(state, talentID), nil
}

// ApplyTalentEffectsToCombatStats 将天赋效果应用到 CombatStats 上。
func (mods *TalentModifiers) ApplyTalentEffectsToCombatStats(stats *CombatStats, alivePartCount int, hasTalent func(string) bool) {
	if mods == nil {
		return
	}

	// 攻击力百分比加成
	if mods.AttackPowerPercent > 0 {
		stats.AttackPower = max(1, stats.AttackPower+int64(float64(stats.AttackPower)*mods.AttackPowerPercent))
	}

	// 全伤害增幅
	stats.AllDamageAmplify += mods.AllDamageAmplify

	// Boss 增伤

	// 破甲率额外
	stats.ArmorPenPercent = min(0.80, stats.ArmorPenPercent+mods.ArmorPenExtra)

	// 暴击伤害百分比加成
	if mods.CritDamagePercentBonus > 0 {
		stats.CritDamageMultiplier += mods.CritDamagePercentBonus
	}

	// 围剿：每存活一个部位 +12% 伤害
	if mods.PerPartDamagePercent > 0 && alivePartCount > 1 {
		stats.AllDamageAmplify += mods.PerPartDamagePercent * float64(alivePartCount)
	}

	// 破甲转化：破甲率的50%转为全伤害增幅
	if hasTalent != nil && hasTalent("armor_pen_convert") {
		stats.AllDamageAmplify += stats.ArmorPenPercent * 0.60
	}
}

// TalentCombatState 玩家在单场 Boss 战中的天赋战斗状态。
type TalentCombatState struct {
	OmenStacks             int              `json:"omenStacks"`
	CollapseParts          []int            `json:"collapseParts"`
	CollapseEndsAt         int64            `json:"collapseEndsAt"`
	CollapseDuration       int64            `json:"collapseDuration"`
	DoomMarks              []int            `json:"doomMarks"`
	DoomMarkCumDamage      map[string]int64 `json:"doomMarkCumDamage"`
	SilverStormRemaining   int              `json:"silverStormRemaining"`
	SilverStormEndsAt      int64            `json:"silverStormEndsAt"`
	SilverStormActive      bool             `json:"silverStormActive"`
	AutoStrikeTargetPart   string           `json:"autoStrikeTargetPart"`
	AutoStrikeComboCount   int64            `json:"autoStrikeComboCount"`
	AutoStrikeExpiresAt    int64            `json:"autoStrikeExpiresAt"`
	LastFinalCutAt         int64            `json:"lastFinalCutAt"`
	JudgmentDayUsed        map[string]bool  `json:"judgmentDayUsed"`
	PartHeavyClickCount    map[string]int64 `json:"partHeavyClickCount"`
	PartRetainedClicks     map[string]int64 `json:"partRetainedClicks"`
	PartStormComboCount    map[string]int64 `json:"partStormComboCount"`
	CritCount              int64            `json:"critCount"`
	SkinnerParts           map[string]int64 `json:"skinnerParts"`
	NormalTriggerCount     int64            `json:"normalTriggerCount"`
	ArmorTriggerCount      int64            `json:"armorTriggerCount"`
	AutoStrikeTriggerCount int64            `json:"autoStrikeTriggerCount"`
	AutoStrikeWindowSec    int64            `json:"autoStrikeWindowSec"`
}

// NewTalentCombatState 创建空天赋战斗状态。
func NewTalentCombatState() *TalentCombatState {
	return &TalentCombatState{
		JudgmentDayUsed:     make(map[string]bool),
		PartHeavyClickCount: make(map[string]int64),
		PartStormComboCount: make(map[string]int64),
		PartRetainedClicks:  make(map[string]int64),
		DoomMarkCumDamage:   make(map[string]int64),
		SkinnerParts:        make(map[string]int64),
	}
}

func (s *Store) talentCombatStateKey(nickname, bossID string) string {
	return s.namespace + "player:talent_state:" + nickname + ":" + bossID
}

// GetTalentCombatState 获取天赋战斗状态。
func (s *Store) GetTalentCombatState(ctx context.Context, nickname, bossID string) (*TalentCombatState, error) {
	raw, err := s.client.HGet(ctx, s.talentCombatStateKey(nickname, bossID), "state").Result()
	if err != nil || raw == "" {
		return NewTalentCombatState(), nil
	}
	var state TalentCombatState
	if err := sonic.Unmarshal([]byte(raw), &state); err != nil {
		return NewTalentCombatState(), nil
	}
	if state.JudgmentDayUsed == nil {
		state.JudgmentDayUsed = make(map[string]bool)
	}
	if state.PartHeavyClickCount == nil {
		state.PartHeavyClickCount = make(map[string]int64)
	}
	if state.PartStormComboCount == nil {
		state.PartStormComboCount = make(map[string]int64)
	}
	if state.PartRetainedClicks == nil {
		state.PartRetainedClicks = make(map[string]int64)
	}
	if state.SkinnerParts == nil {
		state.SkinnerParts = make(map[string]int64)
	}
	if state.DoomMarkCumDamage == nil {
		state.DoomMarkCumDamage = make(map[string]int64)
	}
	return &state, nil
}

// SaveTalentCombatState 保存天赋战斗状态。
func (s *Store) SaveTalentCombatState(ctx context.Context, nickname, bossID string, state *TalentCombatState) error {
	if state == nil {
		return nil
	}
	raw, err := sonic.Marshal(state)
	if err != nil {
		return err
	}
	return s.client.HSet(ctx, s.talentCombatStateKey(nickname, bossID), "state", string(raw)).Err()
}

// DeleteTalentCombatState Boss 战后清理天赋战斗状态。
func (s *Store) DeleteTalentCombatState(ctx context.Context, nickname, bossID string) error {
	return s.client.Del(ctx, s.talentCombatStateKey(nickname, bossID)).Err()
}

// AddOmenStacks 增加死兆层数并返回新值。
func (s *Store) AddOmenStacks(ctx context.Context, nickname, bossID string, delta int) (int, error) {
	state, err := s.GetTalentCombatState(ctx, nickname, bossID)
	if err != nil {
		return 0, err
	}
	state.OmenStacks, _ = applyOmenStackDelta(state.OmenStacks, delta)
	return state.OmenStacks, s.SaveTalentCombatState(ctx, nickname, bossID, state)
}

// ConsumeOmenStacks 消耗死兆层数，返回实际消耗量。
func (s *Store) ConsumeOmenStacks(ctx context.Context, nickname, bossID string, cost int) (int, error) {
	state, err := s.GetTalentCombatState(ctx, nickname, bossID)
	if err != nil {
		return 0, err
	}
	if state.OmenStacks < cost {
		return 0, nil
	}
	state.OmenStacks -= cost
	return cost, s.SaveTalentCombatState(ctx, nickname, bossID, state)
}

// TalentPartKey 生成部位标识 key。
func TalentPartKey(x, y int) string {
	return fmt.Sprintf("%d-%d", x, y)
}

func applyOmenStackDelta(current, delta int) (next int, overflow int) {
	next = current + delta
	if next < 0 {
		return 0, 0
	}
	if next > TalentOmenStackCap {
		return TalentOmenStackCap, next - TalentOmenStackCap
	}
	return next, 0
}
