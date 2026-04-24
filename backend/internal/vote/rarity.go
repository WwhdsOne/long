package vote

import (
	"math/rand/v2"
	"strings"
)

var globalRand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

const defaultEquipmentRarity = "普通"

var equipmentRarityOrder = map[string]struct{}{
	"普通": {},
	"优秀": {},
	"稀有": {},
	"史诗": {},
	"传说": {},
	"至臻": {},
}

// RarityStats 稀有度对应的属性范围与成长上限
type RarityStats struct {
	AttackPowerMin      int64
	AttackPowerMax      int64
	ArmorPenMin         float64
	ArmorPenMax         float64
	CritDamageMultMin   float64
	CritDamageMultMax   float64
	BossDamageMin       float64
	BossDamageMax       float64
	EnhanceCapExtra     int     // 额外强化上限
	DropWeightMult      float64 // 掉落权重倍率
	SalvageGemValue     int64   // 分解获得原石数
}

var rarityStatTable = map[string]RarityStats{
	"普通": {
		AttackPowerMin: 1, AttackPowerMax: 2,
		ArmorPenMin: 0, ArmorPenMax: 0,
		CritDamageMultMin: 0, CritDamageMultMax: 0,
		BossDamageMin: 0, BossDamageMax: 0,
		EnhanceCapExtra: 0, DropWeightMult: 1.0,
		SalvageGemValue: 1,
	},
	"优秀": {
		AttackPowerMin: 2, AttackPowerMax: 4,
		ArmorPenMin: 0.01, ArmorPenMax: 0.03,
		CritDamageMultMin: 0.05, CritDamageMultMax: 0.10,
		BossDamageMin: 0, BossDamageMax: 0.02,
		EnhanceCapExtra: 2, DropWeightMult: 0.8,
		SalvageGemValue: 2,
	},
	"稀有": {
		AttackPowerMin: 4, AttackPowerMax: 8,
		ArmorPenMin: 0.03, ArmorPenMax: 0.06,
		CritDamageMultMin: 0.10, CritDamageMultMax: 0.20,
		BossDamageMin: 0.02, BossDamageMax: 0.05,
		EnhanceCapExtra: 5, DropWeightMult: 0.5,
		SalvageGemValue: 5,
	},
	"史诗": {
		AttackPowerMin: 8, AttackPowerMax: 15,
		ArmorPenMin: 0.05, ArmorPenMax: 0.10,
		CritDamageMultMin: 0.20, CritDamageMultMax: 0.35,
		BossDamageMin: 0.05, BossDamageMax: 0.10,
		EnhanceCapExtra: 10, DropWeightMult: 0.3,
		SalvageGemValue: 10,
	},
	"传说": {
		AttackPowerMin: 15, AttackPowerMax: 25,
		ArmorPenMin: 0.08, ArmorPenMax: 0.15,
		CritDamageMultMin: 0.35, CritDamageMultMax: 0.50,
		BossDamageMin: 0.08, BossDamageMax: 0.15,
		EnhanceCapExtra: 15, DropWeightMult: 0.15,
		SalvageGemValue: 20,
	},
	"至臻": {
		AttackPowerMin: 25, AttackPowerMax: 40,
		ArmorPenMin: 0.12, ArmorPenMax: 0.20,
		CritDamageMultMin: 0.50, CritDamageMultMax: 0.80,
		BossDamageMin: 0.12, BossDamageMax: 0.25,
		EnhanceCapExtra: 20, DropWeightMult: 0.05,
		SalvageGemValue: 50,
	},
}

// RarityStatsForRarity 返回指定稀有度的属性表；如果未知稀有度返回普通。
func RarityStatsForRarity(rarity string) RarityStats {
	normalized := normalizeEquipmentRarity(rarity)
	stats, ok := rarityStatTable[normalized]
	if !ok {
		return rarityStatTable[defaultEquipmentRarity]
	}
	return stats
}

// RarityDropWeightMultiplier 返回该稀有度的掉落权重倍率。
func RarityDropWeightMultiplier(rarity string) float64 {
	return RarityStatsForRarity(rarity).DropWeightMult
}

// SalvageGemValue 返回该稀有度的分解原石价值。
func SalvageGemValue(rarity string) int64 {
	return RarityStatsForRarity(rarity).SalvageGemValue
}

func normalizeEquipmentRarity(rarity string) string {
	trimmed := strings.TrimSpace(rarity)
	if _, ok := equipmentRarityOrder[trimmed]; ok {
		return trimmed
	}
	return defaultEquipmentRarity
}

// GenerateRarityStats 根据稀有度随机生成一条装备额外属性配置。
func GenerateRarityStats(rarity string) (attackPower int64, armorPen, critDmgMult, bossDmg float64) {
	stats := RarityStatsForRarity(rarity)
	attackPower = stats.AttackPowerMin + randInt64(stats.AttackPowerMax-stats.AttackPowerMin+1)
	armorPen = stats.ArmorPenMin
	if stats.ArmorPenMax > stats.ArmorPenMin {
		armorPen = stats.ArmorPenMin + float64(randInt64(1000))/1000.0*(stats.ArmorPenMax-stats.ArmorPenMin)
	}
	critDmgMult = stats.CritDamageMultMin
	if stats.CritDamageMultMax > stats.CritDamageMultMin {
		critDmgMult = stats.CritDamageMultMin + float64(randInt64(1000))/1000.0*(stats.CritDamageMultMax-stats.CritDamageMultMin)
	}
	bossDmg = stats.BossDamageMin
	if stats.BossDamageMax > stats.BossDamageMin {
		bossDmg = stats.BossDamageMin + float64(randInt64(1000))/1000.0*(stats.BossDamageMax-stats.BossDamageMin)
	}
	return
}

func randInt64(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return int64(globalRand.Int64N(n))
}
