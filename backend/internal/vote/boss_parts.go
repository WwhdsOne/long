package vote

func normalizeBossPartLayout(parts []BossPart) []BossPart {
	if len(parts) == 0 {
		return nil
	}

	normalized := make([]BossPart, 0, len(parts))
	for _, part := range parts {
		part.MaxHP = maxInt64(1, part.MaxHP)
		part.CurrentHP = part.MaxHP
		part.Alive = true
		normalized = append(normalized, part)
	}
	return normalized
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
