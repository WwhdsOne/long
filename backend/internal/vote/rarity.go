package vote

import "strings"

const defaultEquipmentRarity = "普通"

var equipmentRarityOrder = map[string]struct{}{
	"普通": {},
	"优秀": {},
	"稀有": {},
	"史诗": {},
	"传说": {},
	"至臻": {},
}

func normalizeEquipmentRarity(rarity string) string {
	trimmed := strings.TrimSpace(rarity)
	if _, ok := equipmentRarityOrder[trimmed]; ok {
		return trimmed
	}
	return defaultEquipmentRarity
}
