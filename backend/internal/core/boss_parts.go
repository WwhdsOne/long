package core

import "strings"

func normalizeBossPartLayout(parts []BossPart) []BossPart {
	if len(parts) == 0 {
		return nil
	}

	normalized := make([]BossPart, 0, len(parts))
	for _, part := range parts {
		part.DisplayName = strings.TrimSpace(part.DisplayName)
		part.ImagePath = strings.TrimSpace(part.ImagePath)
		part.DamageAffinity = normalizeBossPartDamageAffinity(part.DamageAffinity)
		part.MaxHP = maxInt64(1, part.MaxHP)
		part.CurrentHP = part.MaxHP
		part.Alive = true
		normalized = append(normalized, part)
	}
	return normalized
}

func normalizeBossPartDamageAffinity(affinity PartDamageAffinity) PartDamageAffinity {
	switch affinity {
	case PartDamageAffinityMagicOnly:
		return PartDamageAffinityMagicOnly
	default:
		return PartDamageAffinityNormal
	}
}

func sumBossPartMaxHP(parts []BossPart) int64 {
	var total int64
	for _, part := range parts {
		total += maxInt64(1, part.MaxHP)
	}
	return maxInt64(1, total)
}

func sumBossPartCurrentHP(parts []BossPart) int64 {
	var total int64
	for _, part := range parts {
		if part.CurrentHP > 0 {
			total += part.CurrentHP
		}
	}
	return total
}
