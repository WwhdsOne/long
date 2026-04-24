function normalizeId(value) {
  return String(value ?? '').trim()
}

export function createEmptyCosmeticLoadout() {
  return {
    trailId: '',
    impactId: '',
  }
}

export function normalizeCosmeticLoadout(loadout = {}) {
  return {
    trailId: normalizeId(loadout?.trailId),
    impactId: normalizeId(loadout?.impactId),
  }
}

export function buildCosmeticCollections(items = []) {
  const trails = []
  const impacts = []

  for (const item of Array.isArray(items) ? items : []) {
    if (item?.type === 'trail') {
      trails.push(item)
      continue
    }
    if (item?.type === 'impact') {
      impacts.push(item)
    }
  }

  return {
    trails,
    impacts,
  }
}

export function canEquipCosmeticSelection(items = [], loadout = {}) {
  const normalized = normalizeCosmeticLoadout(loadout)
  if (!normalized.trailId && !normalized.impactId) {
    return true
  }

  const owned = new Set(
    (Array.isArray(items) ? items : [])
      .filter((item) => item?.owned)
      .map((item) => normalizeId(item?.cosmeticId)),
  )

  return (!normalized.trailId || owned.has(normalized.trailId))
    && (!normalized.impactId || owned.has(normalized.impactId))
}

export function cosmeticStatusText(item, loadout = {}) {
  const normalized = normalizeCosmeticLoadout(loadout)
  if (!item?.owned) {
    return '未拥有'
  }
  if (item?.type === 'trail' && normalizeId(item?.cosmeticId) === normalized.trailId) {
    return '已装备'
  }
  if (item?.type === 'impact' && normalizeId(item?.cosmeticId) === normalized.impactId) {
    return '已装备'
  }
  return '可装备'
}

export function resolveCosmeticEffectConfig() {
  return {
    mode: 'normal',
    suppressed: true,
    trailTheme: '',
    impactTheme: '',
    trailClass: '',
    impactClass: '',
    durationMs: 900,
    particleCount: 0,
  }
}

export function salvageableCount(item, active = false) {
  const quantity = Math.max(0, Number(item?.quantity || 0))
  if (active) {
    return Math.max(0, quantity - 1)
  }
  return quantity
}

export function summarizeEquippedCosmetics(items = [], loadout = {}) {
  const normalized = normalizeCosmeticLoadout(loadout)
  const index = new Map((Array.isArray(items) ? items : []).map((item) => [normalizeId(item?.cosmeticId), item]))

  const parts = []
  if (normalized.trailId) {
    parts.push(index.get(normalized.trailId)?.name || normalized.trailId)
  }
  if (normalized.impactId) {
    parts.push(index.get(normalized.impactId)?.name || normalized.impactId)
  }

  return parts.length > 0 ? parts.join(' / ') : '未装备'
}
