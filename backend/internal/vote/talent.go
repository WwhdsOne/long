package vote

import (
	"context"
	"errors"
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

// TalentDef 天赋节点定义
type TalentDef struct {
	ID           string     `json:"id"`
	Tree         TalentTree `json:"tree"`
	Tier         int        `json:"tier"`    // 0=基石, 1-3=中间, 4=终极
	Name         string     `json:"name"`
	EffectType   string     `json:"effectType"`
	EffectValue  any        `json:"effectValue"`
	Prerequisite string     `json:"prerequisite,omitempty"` // 前置天赋 ID
}

// TalentState 玩家天赋状态
type TalentState struct {
	Tree     TalentTree `json:"tree"`
	SubTree  TalentTree `json:"subTree,omitempty"`
	Talents  []string   `json:"talents"`  // 已学习天赋 ID 列表
}

// talentPlayerData Redis 中存储的原始结构
type talentPlayerData struct {
	Tree    string `json:"tree"`
	SubTree string `json:"subTree,omitempty"`
	Talents string `json:"talents"` // JSON array of strings
}

// 三系天赋定义表
var talentDefs = map[string]TalentDef{
	// ===== 普攻：均衡攻势 =====
	"normal_core":      {ID: "normal_core", Tree: TalentTreeNormal, Tier: 0, Name: "暴风连击", EffectType: "storm_combo", EffectValue: map[string]any{"triggerCount": 200, "extraHits": 15, "chaseRatio": 0.5, "maxChaseRatio": 0.8}},
	"normal_atk_up":    {ID: "normal_atk_up", Tree: TalentTreeNormal, Tier: 1, Name: "攻击强化", EffectType: "attack_power_percent", EffectValue: map[string]any{"percent": 0.25}, Prerequisite: "normal_core"},
	"normal_dmg_amp":   {ID: "normal_dmg_amp", Tree: TalentTreeNormal, Tier: 1, Name: "伤害增幅", EffectType: "all_damage_amplify", EffectValue: map[string]any{"percent": 0.20}, Prerequisite: "normal_core"},
	"normal_soft_atk":  {ID: "normal_soft_atk", Tree: TalentTreeNormal, Tier: 1, Name: "软组织特攻", EffectType: "part_type_damage", EffectValue: map[string]any{"partType": "soft", "percent": 0.40}, Prerequisite: "normal_core"},
	"normal_charge":    {ID: "normal_charge", Tree: TalentTreeNormal, Tier: 2, Name: "蓄力返还", EffectType: "charge_retain", EffectValue: map[string]any{"retainPercent": 0.30}, Prerequisite: "normal_atk_up"},
	"normal_chase_up":  {ID: "normal_chase_up", Tree: TalentTreeNormal, Tier: 2, Name: "追击强化", EffectType: "chase_upgrade", EffectValue: map[string]any{"chaseRatio": 0.80}, Prerequisite: "normal_dmg_amp"},
	"normal_combo_ext": {ID: "normal_combo_ext", Tree: TalentTreeNormal, Tier: 2, Name: "连击扩展", EffectType: "combo_extend", EffectValue: map[string]any{"extraHits": 10}, Prerequisite: "normal_soft_atk"},
	"normal_encircle":  {ID: "normal_encircle", Tree: TalentTreeNormal, Tier: 3, Name: "围剿", EffectType: "per_part_damage", EffectValue: map[string]any{"percentPerPart": 0.12}, Prerequisite: "normal_charge"},
	"normal_low_hp":    {ID: "normal_low_hp", Tree: TalentTreeNormal, Tier: 3, Name: "残血收割", EffectType: "low_hp_bonus", EffectValue: map[string]any{"hpThreshold": 0.25, "multiplier": 2.0}, Prerequisite: "normal_chase_up"},
	"normal_ultimate":  {ID: "normal_ultimate", Tree: TalentTreeNormal, Tier: 4, Name: "白银风暴", EffectType: "silver_storm", EffectValue: map[string]any{"triggerHits": 15, "treatAllAs": "soft"}, Prerequisite: "normal_encircle"},

	// ===== 破甲：碎盾攻坚 =====
	"armor_core":          {ID: "armor_core", Tree: TalentTreeArmor, Tier: 0, Name: "灭绝穿甲", EffectType: "permanent_armor_pen", EffectValue: map[string]any{"penPercent": 0.40, "collapseTrigger": 200, "collapseDuration": 8}},
	"armor_pen_up":        {ID: "armor_pen_up", Tree: TalentTreeArmor, Tier: 1, Name: "穿甲强化", EffectType: "armor_pen_extra", EffectValue: map[string]any{"extraPen": 0.25}, Prerequisite: "armor_core"},
	"armor_boss_hunter":   {ID: "armor_boss_hunter", Tree: TalentTreeArmor, Tier: 1, Name: "首领猎杀", EffectType: "boss_damage", EffectValue: map[string]any{"percent": 0.30}, Prerequisite: "armor_core"},
	"armor_heavy_scale":   {ID: "armor_heavy_scale", Tree: TalentTreeArmor, Tier: 1, Name: "以强制强", EffectType: "armor_scaling", EffectValue: map[string]any{"damagePer100Armor": 0.02}, Prerequisite: "armor_core"},
	"armor_heavy_atk":     {ID: "armor_heavy_atk", Tree: TalentTreeArmor, Tier: 2, Name: "重甲特攻", EffectType: "part_type_damage", EffectValue: map[string]any{"partType": "heavy", "percent": 0.50}, Prerequisite: "armor_pen_up"},
	"armor_collapse_ext":  {ID: "armor_collapse_ext", Tree: TalentTreeArmor, Tier: 2, Name: "崩塌延长", EffectType: "collapse_extend", EffectValue: map[string]any{"extraDuration": 7}, Prerequisite: "armor_boss_hunter"},
	"armor_auto_strike":   {ID: "armor_auto_strike", Tree: TalentTreeArmor, Tier: 2, Name: "自动打击", EffectType: "auto_strike", EffectValue: map[string]any{"interval": 20, "damageRatio": 3.0}, Prerequisite: "armor_heavy_scale"},
	"armor_ruin":          {ID: "armor_ruin", Tree: TalentTreeArmor, Tier: 3, Name: "废墟打击", EffectType: "collapse_damage_amp", EffectValue: map[string]any{"extraPercent": 1.0}, Prerequisite: "armor_heavy_atk"},
	"armor_pen_convert":   {ID: "armor_pen_convert", Tree: TalentTreeArmor, Tier: 3, Name: "破甲转化", EffectType: "pen_to_amplify", EffectValue: map[string]any{"convertRatio": 0.50}, Prerequisite: "armor_collapse_ext"},
	"armor_ultimate":      {ID: "armor_ultimate", Tree: TalentTreeArmor, Tier: 4, Name: "审判日", EffectType: "judgment_day", EffectValue: map[string]any{"triggerCount": 500, "hpCutPercent": 0.50}, Prerequisite: "armor_ruin"},

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

// IsSubTreeAllowed 检查天赋是否可作为副系学习（不能是基石或终极）。
func IsSubTreeAllowed(def TalentDef) bool {
	return def.Tier > 0 && def.Tier < 4
}

func (s *Store) talentKey(nickname string) string {
	return s.namespace + "player:talents:" + nickname
}

// SelectTalentTree 选择主系和副系。
func (s *Store) SelectTalentTree(ctx context.Context, nickname string, tree, subTree TalentTree) error {
	if tree != TalentTreeNormal && tree != TalentTreeArmor && tree != TalentTreeCrit {
		return ErrInvalidTalentTree
	}
	if subTree != "" && subTree != TalentTreeNormal && subTree != TalentTreeArmor && subTree != TalentTreeCrit {
		return ErrInvalidTalentTree
	}
	if subTree != "" && subTree == tree {
		subTree = "" // 不能和主系相同
	}

	data := talentPlayerData{
		Tree:    string(tree),
		SubTree: string(subTree),
		Talents: "[]",
	}

	values := map[string]any{
		"tree":    data.Tree,
		"subTree": data.SubTree,
		"talents": data.Talents,
	}

	return s.client.HSet(ctx, s.talentKey(nickname), values).Err()
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

	state := &TalentState{
		Tree:    TalentTree(values["tree"]),
		SubTree: TalentTree(values["subTree"]),
	}

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

	state, err := s.GetTalentState(ctx, nickname)
	if err != nil {
		return err
	}
	if state.Tree == "" {
		return ErrTalentTreeNotSet
	}

	// 检查是否已学习
	for _, t := range state.Talents {
		if t == talentID {
			return ErrTalentAlreadyLearned
		}
	}

	// 检查天赋树归属
	if def.Tree != state.Tree && def.Tree != state.SubTree {
		return ErrTalentNotFound
	}

	// 副系限制：不能学基石或终极
	if def.Tree == state.SubTree && !IsSubTreeAllowed(def) {
		return ErrTalentNotFound
	}

	// 副系最多 2 个节点
	if def.Tree == state.SubTree {
		subCount := 0
		for _, t := range state.Talents {
			if d, ok := talentDefs[t]; ok && d.Tree == state.SubTree {
				subCount++
			}
		}
		if subCount >= 2 {
			return ErrTalentMaxLevel
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

	// 保存
	state.Talents = append(state.Talents, talentID)
	talentsJSON, err := sonic.Marshal(state.Talents)
	if err != nil {
		return err
	}
	return s.client.HSet(ctx, s.talentKey(nickname), "talents", string(talentsJSON)).Err()
}

// ResetTalents 重置所有已学习天赋（保留主系副系选择）。
func (s *Store) ResetTalents(ctx context.Context, nickname string) error {
	return s.client.HSet(ctx, s.talentKey(nickname), "talents", "[]").Err()
}

// TalentModifiers 聚集所有已学习天赋的效果修改器。
type TalentModifiers struct {
	AttackPowerPercent     float64 `json:"attackPowerPercent"`
	AllDamageAmplify       float64 `json:"allDamageAmplify"`
	ArmorPenExtra          float64 `json:"armorPenExtra"`
	CritDamagePercentBonus float64 `json:"critDamagePercentBonus"`
	PerPartDamagePercent   float64 `json:"perPartDamagePercent"`
	LowHpMultiplier        float64 `json:"lowHpMultiplier"`
	CollapseDuration       int     `json:"collapseDuration"`
	// 已学习天赋 ID 列表，供具体逻辑判断
	Learned       []string          `json:"-"`
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
		case "collapse_extend":
			if d, ok := val["extraDuration"].(float64); ok {
				mods.CollapseDuration = int(d)
			}
		case "pen_to_amplify":
			if r, ok := val["convertRatio"].(float64); ok {
				// 会在使用时基于实际破甲率换算
				_ = r
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

// TalentBossPartTypePreference 根据天赋返回偏好攻击的部位类型。
func (state *TalentState) TalentBossPartTypePreference() PartType {
	switch state.Tree {
	case TalentTreeNormal:
		return PartTypeSoft
	case TalentTreeArmor:
		return PartTypeHeavy
	case TalentTreeCrit:
		return PartTypeWeak
	default:
		return ""
	}
}

// Validate 检查 talentPlayerData 是否合法。
func (s *Store) validateTalentData(data talentPlayerData) error {
	if data.Tree == "" {
		return errors.New("talent tree is required")
	}
	return nil
}
