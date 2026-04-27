package vote

import (
	"context"
	"encoding/json"
	"fmt"
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
	// ===== 天赋点经济参数（可直接调）=====
	// tier0: 基石层成本
	TalentCostTier0 int64 = 100
	// tier1: 第一层成本
	TalentCostTier1 int64 = 2000
	// tier2: 第二层成本
	TalentCostTier2 int64 = 8000
	// tier3: 第三层成本
	TalentCostTier3 int64 = 30000
	// tier4: 终极层成本
	TalentCostTier4 int64 = 120000

	// ===== 小节点成本（层锁机制）=====
	TalentCostFillerTier1 int64 = 120
	TalentCostFillerTier2 int64 = 150
	TalentCostFillerTier3 int64 = 200

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

// TalentDef 天赋节点定义
type TalentDef struct {
	ID           string     `json:"id"`
	Tree         TalentTree `json:"tree"`
	Tier         int        `json:"tier"` // 0=基石, 1-3=中间, 4=终极
	Cost         int64      `json:"cost"` // 学习消耗的天赋点
	Name         string     `json:"name"`
	EffectType   string     `json:"effectType"`
	EffectValue  any        `json:"effectValue"`
	Prerequisite string     `json:"prerequisite,omitempty"` // 前置天赋 ID
}

// TalentState 玩家天赋状态
type TalentState struct {
	Talents []string `json:"talents"` // 已学习天赋 ID 列表
}

// talentPlayerData Redis 中存储的原始结构
type talentPlayerData struct {
	Talents string `json:"talents"` // JSON array of strings
}

// 三系天赋定义表
var talentDefs = map[string]TalentDef{
	// ===== 普攻：均衡攻势 =====
	"normal_core":      {ID: "normal_core", Tree: TalentTreeNormal, Tier: 0, Name: "暴风连击", EffectType: "storm_combo", EffectValue: map[string]any{"triggerCount": TalentNormalStormTriggerCount, "extraHits": TalentNormalStormExtraHits, "chaseRatio": TalentNormalStormChaseRatio, "maxChaseRatio": TalentNormalStormMaxChaseRatio}},
	"normal_atk_up":    {ID: "normal_atk_up", Tree: TalentTreeNormal, Tier: 1, Name: "攻击强化", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.25}, Prerequisite: "normal_core"},
	"normal_dmg_amp":   {ID: "normal_dmg_amp", Tree: TalentTreeNormal, Tier: 1, Name: "伤害增幅", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.20}, Prerequisite: "normal_core"},
	"normal_soft_atk":  {ID: "normal_soft_atk", Tree: TalentTreeNormal, Tier: 1, Name: "软组织特攻", EffectType: "part_type_damage", EffectValue: map[string]any{"partType": "soft", "percent": 0.40}, Prerequisite: "normal_core"},
	"normal_charge":    {ID: "normal_charge", Tree: TalentTreeNormal, Tier: 2, Name: "蓄力返还", EffectType: "charge_retain", EffectValue: map[string]any{"retainPercent": 0.30}, Prerequisite: "normal_atk_up"},
	"normal_chase_up":  {ID: "normal_chase_up", Tree: TalentTreeNormal, Tier: 2, Name: "追击强化", EffectType: "chase_upgrade", EffectValue: map[string]any{"chaseRatio": TalentNormalChaseUpgradeRatio}, Prerequisite: "normal_dmg_amp"},
	"normal_combo_ext": {ID: "normal_combo_ext", Tree: TalentTreeNormal, Tier: 2, Name: "连击扩展", EffectType: "combo_extend", EffectValue: map[string]any{"extraHits": TalentNormalComboExtendHits}, Prerequisite: "normal_soft_atk"},
	"normal_encircle":  {ID: "normal_encircle", Tree: TalentTreeNormal, Tier: 3, Name: "围剿", EffectType: "per_part_damage", EffectValue: map[string]any{"percentPerPart": 0.12}, Prerequisite: "normal_charge"},
	"normal_low_hp":    {ID: "normal_low_hp", Tree: TalentTreeNormal, Tier: 3, Name: "残血收割", EffectType: "low_hp_bonus", EffectValue: map[string]any{"hpThreshold": 0.25, "multiplier": 2.0}, Prerequisite: "normal_chase_up"},
	"normal_ultimate":  {ID: "normal_ultimate", Tree: TalentTreeNormal, Tier: 4, Name: "白银风暴", EffectType: "silver_storm", EffectValue: map[string]any{"triggerHits": 15, "treatAllAs": "soft"}, Prerequisite: "normal_encircle"},

	// ===== 破甲：碎盾攻坚 =====
	"armor_core":         {ID: "armor_core", Tree: TalentTreeArmor, Tier: 0, Name: "灭绝穿甲", EffectType: "permanent_armor_pen", EffectValue: map[string]any{"penPercent": 0.40, "collapseTrigger": 100, "collapseDuration": 8}},
	"armor_pen_up":       {ID: "armor_pen_up", Tree: TalentTreeArmor, Tier: 1, Name: "穿甲强化", EffectType: "armor_pen_extra", EffectValue: map[string]any{"extraPen": 0.25}, Prerequisite: "armor_core"},
	"armor_boss_hunter":  {ID: "armor_boss_hunter", Tree: TalentTreeArmor, Tier: 1, Name: "首领猎杀", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.30}, Prerequisite: "armor_core"},
	"armor_heavy_scale":  {ID: "armor_heavy_scale", Tree: TalentTreeArmor, Tier: 1, Name: "以强制强", EffectType: "armor_scaling", EffectValue: map[string]any{"damagePer100Armor": 0.02}, Prerequisite: "armor_core"},
	"armor_heavy_atk":    {ID: "armor_heavy_atk", Tree: TalentTreeArmor, Tier: 2, Name: "重甲特攻", EffectType: "part_type_damage", EffectValue: map[string]any{"partType": "heavy", "percent": 0.50}, Prerequisite: "armor_pen_up"},
	"armor_collapse_ext": {ID: "armor_collapse_ext", Tree: TalentTreeArmor, Tier: 2, Name: "崩塌延长", EffectType: "collapse_extend", EffectValue: map[string]any{"extraDuration": 7}, Prerequisite: "armor_boss_hunter"},
	"armor_auto_strike":  {ID: "armor_auto_strike", Tree: TalentTreeArmor, Tier: 2, Name: "自动打击", EffectType: "auto_strike", EffectValue: map[string]any{"interval": 20, "damageRatio": 3.0}, Prerequisite: "armor_heavy_scale"},
	"armor_ruin":         {ID: "armor_ruin", Tree: TalentTreeArmor, Tier: 3, Name: "废墟打击", EffectType: "collapse_damage_amp", EffectValue: map[string]any{"extraPercent": 1.0}, Prerequisite: "armor_heavy_atk"},
	"armor_pen_convert":  {ID: "armor_pen_convert", Tree: TalentTreeArmor, Tier: 3, Name: "破甲转化", EffectType: "pen_to_amplify", EffectValue: map[string]any{"convertRatio": 0.50}, Prerequisite: "armor_collapse_ext"},
	"armor_ultimate":     {ID: "armor_ultimate", Tree: TalentTreeArmor, Tier: 4, Name: "审判日", EffectType: "judgment_day", EffectValue: map[string]any{"triggerCount": 100, "hpCutPercent": 0.50}, Prerequisite: "armor_ruin"},

	// ===== 暴击：致命洞察 =====
	"crit_core":          {ID: "crit_core", Tree: TalentTreeCrit, Tier: 0, Name: "溢杀", EffectType: "overkill", EffectValue: map[string]any{"baseCritBonus": 0.20, "overflowToCritDmg": 0.02, "omenPerWeakCrit": 1}},
	"crit_omen_resonate": {ID: "crit_omen_resonate", Tree: TalentTreeCrit, Tier: 1, Name: "死兆共鸣", EffectType: "omen_crit_damage", EffectValue: map[string]any{"critDmgPerOmen": 0.003}, Prerequisite: "crit_core"},
	"crit_cruel":         {ID: "crit_cruel", Tree: TalentTreeCrit, Tier: 1, Name: "残忍", EffectType: "crit_damage_bonus", EffectValue: map[string]any{"percent": 0.60}, Prerequisite: "crit_core"},
	"crit_skinner":       {ID: "crit_skinner", Tree: TalentTreeCrit, Tier: 1, Name: "剥皮", EffectType: "force_weak", EffectValue: map[string]any{"chance": 0.30, "duration": 5}, Prerequisite: "crit_core"},
	"crit_bleed":         {ID: "crit_bleed", Tree: TalentTreeCrit, Tier: 2, Name: "致命出血", EffectType: "bleed", EffectValue: map[string]any{"duration": 4, "damageRatio": 0.60}, Prerequisite: "crit_omen_resonate"},
	"crit_omen_kill":     {ID: "crit_omen_kill", Tree: TalentTreeCrit, Tier: 2, Name: "斩杀预兆", EffectType: "omen_low_hp", EffectValue: map[string]any{"hpThreshold": 0.35, "dmgPerOmen": 0.01}, Prerequisite: "crit_cruel"},
	"crit_omen_reap":     {ID: "crit_omen_reap", Tree: TalentTreeCrit, Tier: 2, Name: "死兆收割", EffectType: "omen_harvest", EffectValue: map[string]any{"omenCost": 30}, Prerequisite: "crit_skinner"},
	"crit_death_ecstasy": {ID: "crit_death_ecstasy", Tree: TalentTreeCrit, Tier: 3, Name: "死亡狂喜", EffectType: "death_ecstasy", EffectValue: map[string]any{"omenCost": 50, "duration": 6, "critDmgBonus": 2.0}, Prerequisite: "crit_bleed"},
	"crit_final_cut":     {ID: "crit_final_cut", Tree: TalentTreeCrit, Tier: 3, Name: "终末血斩", EffectType: "final_cut", EffectValue: map[string]any{"critCount": 120, "hpCutPercent": 0.12, "cooldown": 30}, Prerequisite: "crit_omen_kill"},
	"crit_ultimate":      {ID: "crit_ultimate", Tree: TalentTreeCrit, Tier: 4, Name: "末日审判", EffectType: "doom_judgment", EffectValue: map[string]any{"markCount": 2, "omenReward": 100, "critDmgMult": 3, "dualKillMult": 6}, Prerequisite: "crit_death_ecstasy"},

	// ===== 均衡攻势 小节点（filler）=====
	"normal_filler_t1a": {ID: "normal_filler_t1a", Tree: TalentTreeNormal, Tier: 1, Name: "锐锋", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.03}},
	"normal_filler_t1b": {ID: "normal_filler_t1b", Tree: TalentTreeNormal, Tier: 1, Name: "乱舞", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.03}},
	"normal_filler_t2a": {ID: "normal_filler_t2a", Tree: TalentTreeNormal, Tier: 2, Name: "追猎", EffectType: "chase_ratio_bonus", EffectValue: map[string]any{"percent": 0.05}},
	"normal_filler_t2b": {ID: "normal_filler_t2b", Tree: TalentTreeNormal, Tier: 2, Name: "穿刺", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.03}},
	"normal_filler_t3a": {ID: "normal_filler_t3a", Tree: TalentTreeNormal, Tier: 3, Name: "狩猎", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.05}},
	"normal_filler_t3b": {ID: "normal_filler_t3b", Tree: TalentTreeNormal, Tier: 3, Name: "铁腕", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.03}},

	// ===== 碎盾攻坚 小节点（filler）=====
	"armor_filler_t1a": {ID: "armor_filler_t1a", Tree: TalentTreeArmor, Tier: 1, Name: "破岩", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.03}},
	"armor_filler_t1b": {ID: "armor_filler_t1b", Tree: TalentTreeArmor, Tier: 1, Name: "凿裂", EffectType: "armor_pen_extra", EffectValue: map[string]any{"extraPen": 0.03}},
	"armor_filler_t2a": {ID: "armor_filler_t2a", Tree: TalentTreeArmor, Tier: 2, Name: "瓦解", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.03}},
	"armor_filler_t2b": {ID: "armor_filler_t2b", Tree: TalentTreeArmor, Tier: 2, Name: "碾碎", EffectType: "armor_scaling", EffectValue: map[string]any{"damagePer100Armor": 0.005}},
	"armor_filler_t3a": {ID: "armor_filler_t3a", Tree: TalentTreeArmor, Tier: 3, Name: "碎颅", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.05}},
	"armor_filler_t3b": {ID: "armor_filler_t3b", Tree: TalentTreeArmor, Tier: 3, Name: "摧坚", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.03}},

	// ===== 致命洞察 小节点（filler）=====
	"crit_filler_t1a": {ID: "crit_filler_t1a", Tree: TalentTreeCrit, Tier: 1, Name: "锐眼", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.03}},
	"crit_filler_t1b": {ID: "crit_filler_t1b", Tree: TalentTreeCrit, Tier: 1, Name: "残酷", EffectType: "crit_damage_bonus", EffectValue: map[string]any{"percent": 0.05}},
	"crit_filler_t2a": {ID: "crit_filler_t2a", Tree: TalentTreeCrit, Tier: 2, Name: "深创", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.03}},
	"crit_filler_t2b": {ID: "crit_filler_t2b", Tree: TalentTreeCrit, Tier: 2, Name: "喋血", EffectType: "omen_crit_damage", EffectValue: map[string]any{"critDmgPerOmen": 0.001}},
	"crit_filler_t3a": {ID: "crit_filler_t3a", Tree: TalentTreeCrit, Tier: 3, Name: "追魂", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.05}},
	"crit_filler_t3b": {ID: "crit_filler_t3b", Tree: TalentTreeCrit, Tier: 3, Name: "暴虐", EffectType: "crit_damage_bonus", EffectValue: map[string]any{"percent": 0.05}},
}

var talentTierCosts = map[int]int64{
	0: TalentCostTier0,
	1: TalentCostTier1,
	2: TalentCostTier2,
	3: TalentCostTier3,
	4: TalentCostTier4,
}

var talentFillerTierCosts = map[int]int64{
	1: TalentCostFillerTier1,
	2: TalentCostFillerTier2,
	3: TalentCostFillerTier3,
}

// tierNodeCount 每层节点总数（主 + 小），用于层锁判定。
var tierNodeCount = map[int]int{
	0: 1, // 1 核心
	1: 5, // 3 主 + 2 小
	2: 5, // 3 主 + 2 小
	3: 4, // 2 主 + 2 小
	4: 1, // 1 终极
}

// tierCompletionBonuses 层满奖励表，key 格式 "tree:tier"。
var tierCompletionBonuses = map[string]TalentModifiers{
	"normal:0": {AllDamageAmplify: 0.05},
	"normal:1": {AttackPowerPercent: 0.08},
	"normal:2": {StormTriggerReduce: 20},
	"normal:3": {AllDamageAmplify: 0.10},
	"normal:4": {StormExtraHits: 3},
	"armor:0":  {AllDamageAmplify: 0.05},
	"armor:1":  {CollapseTriggerReduce: 30},
	"armor:2":  {ArmorPenExtra: 0.10},
	"armor:3":  {CollapseVulnerability: 0.10},
	"armor:4":  {JudgmentDayBoost: 0.05},
	"crit:0":   {AllDamageAmplify: 0.05},
	"crit:1":   {CritRateBonus: 0.05},
	"crit:2":   {OmenKillThresholdRaise: 0.03},
	"crit:3":   {OmenCritDmgExtra: 0.002},
	"crit:4":   {DoomMultBoost: 1.0},
}

// tierCompletionBonusLabels 层满奖励文案，供前端直接展示。
var tierCompletionBonusLabels = map[TalentTree]map[int]string{
	TalentTreeNormal: {
		0: "全伤害 +5%",
		1: "攻击力 +8%",
		2: "暴风连击触发 -20 次",
		3: "全伤害 +10%",
		4: "白银风暴 +3 段",
	},
	TalentTreeArmor: {
		0: "全伤害 +5%",
		1: "崩塌触发 -30 次",
		2: "护甲穿透 +10%",
		3: "崩塌易伤 +10%",
		4: "审判日削除 +5%",
	},
	TalentTreeCrit: {
		0: "全伤害 +5%",
		1: "暴击率 +5%",
		2: "斩杀血线 +3%",
		3: "每层死兆暴伤 +0.2%",
		4: "末日审判倍率 +1x",
	},
}

func isFillerTalentID(id string) bool {
	return strings.HasSuffix(id, "_t1a") || strings.HasSuffix(id, "_t1b") ||
		strings.HasSuffix(id, "_t2a") || strings.HasSuffix(id, "_t2b") ||
		strings.HasSuffix(id, "_t3a") || strings.HasSuffix(id, "_t3b")
}

func init() {
	for id, def := range talentDefs {
		var cost int64
		if isFillerTalentID(id) {
			cost = talentFillerTierCosts[def.Tier]
		} else {
			var ok bool
			cost, ok = talentTierCosts[def.Tier]
			if !ok {
				cost = 0
			}
		}
		def.Cost = cost
		talentDefs[id] = def
	}
}

func talentCostByTier(tier int) (int64, bool) {
	cost, ok := talentTierCosts[tier]
	if !ok || cost <= 0 {
		return 0, false
	}
	return cost, true
}

// isLearnedTierFull 检查指定天赋树某一层的所有节点（主 + 小）是否已全部学习。
func isLearnedTierFull(tree TalentTree, tier int, learned []string) bool {
	needed := tierNodeCount[tier]
	if needed == 0 {
		return true
	}
	count := 0
	for _, id := range learned {
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

// TalentPrerequisiteName 返回前置天赋名称，未配置则返回“无”。
func TalentPrerequisiteName(def TalentDef) string {
	prerequisite := strings.TrimSpace(def.Prerequisite)
	if prerequisite == "" {
		return "无"
	}
	preDef, ok := talentDefs[prerequisite]
	if !ok || strings.TrimSpace(preDef.Name) == "" {
		return prerequisite
	}
	return preDef.Name
}

// TalentEffectDescription 返回天赋效果中文描述，供前端直接展示。
func TalentEffectDescription(def TalentDef) string {
	effectType := strings.TrimSpace(def.EffectType)
	value, _ := def.EffectValue.(map[string]any)
	switch effectType {
	case "storm_combo":
		trigger := talentInt(value["triggerCount"])
		hits := talentInt(value["extraHits"])
		ratio := talentPercent(value["chaseRatio"])
		return fmt.Sprintf("每 %d 次点击触发追击爆发，造成 基础伤害 x %s x %d 段总伤。可无限触发。",
			trigger, ratio, hits,
		)
	case "attack_power_percent":
		return fmt.Sprintf("攻击力提升 %s", talentPercent(value["percent"]))
	case "all_damage_amplify":
		return fmt.Sprintf("所有伤害提升 %s", talentPercent(value["percent"]))
	case "part_type_damage":
		return fmt.Sprintf("%s伤害提升 %s", talentPartTypeLabel(value["partType"]), talentPercent(value["percent"]))
	case "charge_retain":
		return fmt.Sprintf("追击爆发触发后，该部位连击进度保留 %s（从30%%开始重新计数）。被动生效。", talentPercent(value["retainPercent"]))
	case "chase_upgrade":
		return fmt.Sprintf("追击爆发单段倍率从50%%提升到 %s。被动生效。", talentPercent(value["chaseRatio"]))
	case "combo_extend":
		return fmt.Sprintf("追击爆发段数从15增加到 %d。被动生效。", talentInt(value["extraHits"]))
	case "per_part_damage":
		return fmt.Sprintf("每个存活部位额外增加 %s 全伤害。被动生效。", talentPercent(value["percentPerPart"]))
	case "low_hp_bonus":
		return fmt.Sprintf("部位剩余血量低于 %s 时，伤害x%.0f。被动生效。", talentPercent(value["hpThreshold"]), talentFloat(value["multiplier"]))
	case "silver_storm":
		return fmt.Sprintf("任意部位被击碎时立即触发，持续15秒内所有部位视为%s（x1.0系数）。每部位击碎均可触发。", talentPartTypeLabel(value["treatAllAs"]))
	case "permanent_armor_pen":
		return fmt.Sprintf("常驻 %s 护甲穿透。对重甲部位累计 %d 次命中后该部位护甲归零 %d 秒（崩塌）。同一部位可多次触发。", talentPercent(value["penPercent"]), talentInt(value["collapseTrigger"]), talentInt(value["collapseDuration"]))
	case "armor_pen_extra":
		return fmt.Sprintf("额外护甲穿透 %s", talentPercent(value["extraPen"]))
	case "armor_scaling":
		return fmt.Sprintf("每 100 护甲额外获得 %s 伤害增幅", talentPercent(value["damagePer100Armor"]))
	case "collapse_extend":
		return fmt.Sprintf("崩塌持续时间从8秒延长到 %d 秒。被动生效。", talentInt(value["extraDuration"]))
	case "auto_strike":
		return fmt.Sprintf("每 %d 秒自动对血量最高的重甲造成 %.1fx攻击力必中真伤。无需点击。", talentInt(value["interval"]), talentFloat(value["damageRatio"]))
	case "collapse_damage_amp":
		return fmt.Sprintf("攻击处于崩塌状态的部位时，额外增伤 %s。被动生效。", talentPercent(value["extraPercent"]))
	case "pen_to_amplify":
		return fmt.Sprintf("将 %s 的护甲穿透值转化为全伤害加成。被动生效。", talentPercent(value["convertRatio"]))
	case "judgment_day":
		return fmt.Sprintf("对同一重甲部位累计 %d 次命中后，立即削除该部位 %s 最大生命值。每部位每场战斗仅一次。", talentInt(value["triggerCount"]), talentPercent(value["hpCutPercent"]))
	case "overkill":
		return fmt.Sprintf("基础暴击率 +%s。暴击率超过100%%的部分按 %s 比例转为暴伤。弱点暴击获得1层死兆。", talentPercent(value["baseCritBonus"]), talentPercent(value["overflowToCritDmg"]))
	case "omen_crit_damage":
		return fmt.Sprintf("每层死兆叠加 %s 暴击伤害（例：100层=+%.0f%%暴伤）。无上限。", talentPercent(value["critDmgPerOmen"]), talentFloat(value["critDmgPerOmen"])*100*100)
	case "crit_damage_bonus":
		return fmt.Sprintf("暴击伤害额外提升 %s", talentPercent(value["percent"]))
	case "force_weak":
		return fmt.Sprintf("暴击时有 %s 概率将当前部位视为弱点（x2.5系数），持续 %d 秒。", talentPercent(value["chance"]), talentInt(value["duration"]))
	case "bleed":
		return fmt.Sprintf("暴击时附加真伤 = 本次伤害 x %s。一次性结算。", talentPercent(value["damageRatio"]))
	case "omen_low_hp":
		return fmt.Sprintf("部位血量低于 %s 时，每层死兆额外 +%s 伤害（例：47层=+47%%）。被动生效。", talentPercent(value["hpThreshold"]), talentPercent(value["dmgPerOmen"]))
	case "omen_harvest":
		return fmt.Sprintf("死兆达到 %d 层时触发必暴伤害 = 本次伤害 x2.0。触发后死兆归零重算。", talentInt(value["omenCost"]))
	case "death_ecstasy":
		return fmt.Sprintf("死兆达到 %d 层时消耗 %d 层，%d 秒内暴伤 +%s 且所有攻击视为弱点。可多次触发。", talentInt(value["omenCost"]), talentInt(value["omenCost"]), talentInt(value["duration"]), talentPercent(value["critDmgBonus"]))
	case "final_cut":
		return fmt.Sprintf("累计 %d 次暴击后削除Boss最大生命值的 %s（%d 秒冷却）。", talentInt(value["critCount"]), talentPercent(value["hpCutPercent"]), talentInt(value["cooldown"]))
	case "doom_judgment":
		return fmt.Sprintf("开局随机标记 %d 个部位。击碎首标记获得%d死兆+暴伤x%d；持有x%d buff时击碎另一标记+暴伤x%d。", talentInt(value["markCount"]), talentInt(value["omenReward"]), talentInt(value["critDmgMult"]), talentInt(value["critDmgMult"]), talentInt(value["dualKillMult"]))
	case "chase_ratio_bonus":
		return fmt.Sprintf("追击爆发单段倍率额外 +%s。被动生效。", talentPercent(value["percent"]))
	default:
		return "该天赋效果说明暂未配置"
	}
}

// TalentTierCompletionBonusLabels 返回指定天赋树的层满奖励文案（key 为层级）。
func TalentTierCompletionBonusLabels(tree TalentTree) map[int]string {
	labels, ok := tierCompletionBonusLabels[tree]
	if !ok || len(labels) == 0 {
		return map[int]string{}
	}
	out := make(map[int]string, len(labels))
	for tier, label := range labels {
		out[tier] = label
	}
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
		return &TalentState{Talents: []string{}}, nil
	}

	state := &TalentState{}
	talentsRaw := values["talents"]
	if talentsRaw != "" {
		var talents []string
		if err := sonic.Unmarshal([]byte(talentsRaw), &talents); err != nil {
			return nil, err
		}
		state.Talents = talents
	} else {
		state.Talents = []string{}
	}

	return state, nil
}

// LearnTalent 学习一个天赋节点。
func (s *Store) LearnTalent(ctx context.Context, nickname string, talentID string) error {
	def, ok := talentDefs[talentID]
	if !ok {
		return ErrTalentNotFound
	}
	cost := def.Cost
	if cost <= 0 {
		return ErrTalentInvalidCost
	}

	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return err
	}
	// 检查是否已学习
	for _, t := range state.Talents {
		if t == talentID {
			return ErrTalentAlreadyLearned
		}
	}

	// 检查前置
	if def.Prerequisite != "" {
		hasPrereq := false
		for _, t := range state.Talents {
			if t == def.Prerequisite {
				hasPrereq = true
				break
			}
		}
		if !hasPrereq {
			return ErrTalentPrerequisite
		}
	}

	// 层锁校验：主系必须点满前一层才能学当前层
	if def.Tier > 0 {
		if !isLearnedTierFull(def.Tree, def.Tier-1, state.Talents) {
			return ErrTalentPrerequisite
		}
	}

	// 保存
	state.Talents = append(state.Talents, talentID)
	talentsJSON, err := sonic.Marshal(state.Talents)
	if err != nil {
		return err
	}
	resources, err := s.resourcesForNickname(ctx, nickname)
	if err != nil {
		return err
	}
	if resources.TalentPoints < cost {
		return ErrTalentPointsInsufficient
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.talentKey(nickname), "talents", string(talentsJSON))
	pipe.HIncrBy(ctx, s.resourceKey(nickname), "talent_points", -cost)
	_, err = pipe.Exec(ctx)
	return err
}

// ResetTalents 重置所有已学习天赋（保留主系副系选择）。
func (s *Store) ResetTalents(ctx context.Context, nickname string) error {
	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return err
	}
	if state == nil || len(state.Talents) == 0 {
		return s.client.HSet(ctx, s.talentKey(nickname), "talents", "[]").Err()
	}

	var refund int64
	for _, id := range state.Talents {
		def, ok := talentDefs[id]
		if !ok {
			continue
		}
		refund += def.Cost
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.talentKey(nickname), "talents", "[]")
	if refund > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(nickname), "talent_points", refund)
	}
	_, err = pipe.Exec(ctx)
	return err
}

// TalentModifiers 聚集所有已学习天赋的效果修改器。
type TalentModifiers struct {
	AttackPowerPercent     float64 `json:"attackPowerPercent"`
	AllDamageAmplify       float64 `json:"allDamageAmplify"`
	ArmorPenExtra          float64 `json:"armorPenExtra"`
	CritDamagePercentBonus float64 `json:"critDamagePercentBonus"`
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
	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return nil, err
	}
	if state == nil || len(state.Talents) == 0 {
		return &TalentModifiers{}, nil
	}

	mods := &TalentModifiers{
		PartTypeBonus: make(map[PartType]float64),
		Learned:       state.Talents,
	}

	for _, id := range state.Talents {
		def, ok := talentDefs[id]
		if !ok {
			continue
		}

		val, _ := def.EffectValue.(map[string]any)

		switch def.EffectType {
		case "attack_power_percent":
			if p, ok := val["percent"].(float64); ok {
				mods.AttackPowerPercent += p
			}
		case "all_damage_amplify":
			if p, ok := val["percent"].(float64); ok {
				mods.AllDamageAmplify += p
			}
		case "part_type_damage":
			partTypeStr, _ := val["partType"].(string)
			percent, _ := val["percent"].(float64)
			if partTypeStr != "" {
				mods.PartTypeBonus[PartType(partTypeStr)] += percent
			}
		case "armor_pen_extra":
			if p, ok := val["extraPen"].(float64); ok {
				mods.ArmorPenExtra += p
			}
		case "crit_damage_bonus":
			if p, ok := val["percent"].(float64); ok {
				mods.CritDamagePercentBonus += p
			}
		case "per_part_damage":
			if p, ok := val["percentPerPart"].(float64); ok {
				mods.PerPartDamagePercent += p
			}
		case "low_hp_bonus":
			if m, ok := val["multiplier"].(float64); ok {
				mods.LowHpMultiplier = m
			}
			if threshold, ok := val["hpThreshold"].(float64); ok {
				mods.LowHpThreshold = threshold
			}
		case "collapse_extend":
			if d, ok := val["extraDuration"].(float64); ok {
				mods.CollapseDuration = int(d)
			}
		case "pen_to_amplify":
			if r, ok := val["convertRatio"].(float64); ok {
				_ = r
			}
		// 小节点新增效果
		case "chase_ratio_bonus":
			if p, ok := val["percent"].(float64); ok {
				mods.ChaseRatioBonus += p
			}
		case "omen_crit_damage":
			if p, ok := val["critDmgPerOmen"].(float64); ok {
				mods.OmenCritDmgExtra += p
			}
		}
	}

	// 层满奖励检测
	// 计算所有已学天赋涉及的树的层满奖励
	learnedTrees := map[TalentTree]bool{}
	for _, id := range state.Talents {
		if def, ok := talentDefs[id]; ok {
			learnedTrees[def.Tree] = true
		}
	}
	for tree := range learnedTrees {
		treeStr := string(tree)
		for tier := 0; tier <= 4; tier++ {
			count := 0
			for _, id := range state.Talents {
				def, ok := talentDefs[id]
				if !ok {
					continue
				}
				if def.Tier == tier && string(def.Tree) == treeStr {
					count++
				}
			}
			needed, ok := tierNodeCount[tier]
			if !ok {
				continue
			}
			if count >= needed {
				switch {
				case treeStr == "normal" && tier == 0:
					mods.AllDamageAmplify += 0.05
				case treeStr == "normal" && tier == 1:
					mods.AttackPowerPercent += 0.08
				case treeStr == "normal" && tier == 2:
					mods.StormTriggerReduce += 20
				case treeStr == "normal" && tier == 3:
					mods.AllDamageAmplify += 0.10
				case treeStr == "normal" && tier == 4:
					mods.StormExtraHits += 3
				case treeStr == "armor" && tier == 0:
					mods.AllDamageAmplify += 0.05
				case treeStr == "armor" && tier == 1:
					mods.CollapseTriggerReduce += 30
				case treeStr == "armor" && tier == 2:
					mods.ArmorPenExtra += 0.10
				case treeStr == "armor" && tier == 3:
					mods.CollapseVulnerability += 0.10
				case treeStr == "armor" && tier == 4:
					mods.JudgmentDayBoost += 0.05
				case treeStr == "crit" && tier == 0:
					mods.AllDamageAmplify += 0.05
				case treeStr == "crit" && tier == 1:
					mods.CritRateBonus += 0.05
				case treeStr == "crit" && tier == 2:
					mods.OmenKillThresholdRaise += 0.03
				case treeStr == "crit" && tier == 3:
					mods.OmenCritDmgExtra += 0.002
				case treeStr == "crit" && tier == 4:
					mods.DoomMultBoost += 1.0
				}
			}
		}
	}

	return mods, nil
}

// HasTalent 检查玩家是否已学习指定天赋。
func (s *Store) HasTalent(ctx context.Context, nickname string, talentID string) (bool, error) {
	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return false, err
	}
	for _, t := range state.Talents {
		if t == talentID {
			return true, nil
		}
	}
	return false, nil
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
		stats.AllDamageAmplify += stats.ArmorPenPercent * 0.5
	}
}

// TalentCombatState 玩家在单场 Boss 战中的天赋战斗状态。
type TalentCombatState struct {
	OmenStacks           int              `json:"omenStacks"`
	CollapseParts        []int            `json:"collapseParts"`
	CollapseEndsAt       int64            `json:"collapseEndsAt"`
	DoomMarks            []int            `json:"doomMarks"`
	DoomDestroyed        int              `json:"doomDestroyed"`
	DoomCritBuff         bool             `json:"doomCritBuff"`
	SilverStormRemaining int              `json:"silverStormRemaining"`
	SilverStormActive    bool             `json:"silverStormActive"`
	LastAutoStrikeAt     int64            `json:"lastAutoStrikeAt"`
	LastFinalCutAt       int64            `json:"lastFinalCutAt"`
	JudgmentDayUsed      map[string]bool  `json:"judgmentDayUsed"`
	PartHeavyClickCount  map[string]int64 `json:"partHeavyClickCount"`
	PartRetainedClicks   map[string]int64 `json:"partRetainedClicks"`
	PartStormComboCount   map[string]int64 `json:"partStormComboCount"`
	CritCount            int64            `json:"critCount"`
	DeathEcstasyEndsAt   int64            `json:"deathEcstasyEndsAt"`
	SkinnerParts         map[string]int64 `json:"skinnerParts"`
}

// NewTalentCombatState 创建空天赋战斗状态。
func NewTalentCombatState() *TalentCombatState {
	return &TalentCombatState{
		JudgmentDayUsed:     make(map[string]bool),
		PartHeavyClickCount:  make(map[string]int64),
		PartStormComboCount:  make(map[string]int64),
		PartRetainedClicks:  make(map[string]int64),
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
	state.OmenStacks += delta
	if state.OmenStacks < 0 {
		state.OmenStacks = 0
	}
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
