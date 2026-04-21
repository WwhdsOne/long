export function createEmptyCosmeticLoadout() {
  return {
    trailId: '',
    impactId: '',
  }
}

export function normalizeCosmeticLoadout(loadout) {
  return {
    trailId: loadout?.trailId || '',
    impactId: loadout?.impactId || '',
  }
}

export function buildCosmeticCollections(shopCatalog = []) {
  return {
    trails: shopCatalog.filter((item) => item?.type === 'trail'),
    impacts: shopCatalog.filter((item) => item?.type === 'impact'),
  }
}

export function findCosmeticItem(shopCatalog = [], cosmeticId = '') {
  if (!cosmeticId) {
    return null
  }

  return shopCatalog.find((item) => item?.cosmeticId === cosmeticId) ?? null
}

export function canEquipCosmeticSelection(shopCatalog = [], selection = {}) {
  const normalized = normalizeCosmeticLoadout(selection)
  const trail = findCosmeticItem(shopCatalog, normalized.trailId)
  const impact = findCosmeticItem(shopCatalog, normalized.impactId)

  if (normalized.trailId && (!trail || trail.type !== 'trail' || !trail.owned)) {
    return false
  }
  if (normalized.impactId && (!impact || impact.type !== 'impact' || !impact.owned)) {
    return false
  }

  return true
}

export function summarizeEquippedCosmetics(shopCatalog = [], loadout = {}) {
  const normalized = normalizeCosmeticLoadout(loadout)
  const trail = findCosmeticItem(shopCatalog, normalized.trailId)
  const impact = findCosmeticItem(shopCatalog, normalized.impactId)

  return {
    trailName: trail?.name || '未装备轨迹',
    impactName: impact?.name || '未装备点击特效',
  }
}

export function resolveCosmeticEffectConfig(shopCatalog = [], loadout = {}, options = {}) {
  const normalized = normalizeCosmeticLoadout(loadout)
  const trail = findCosmeticItem(shopCatalog, normalized.trailId)
  const impact = findCosmeticItem(shopCatalog, normalized.impactId)
  const mode = options.mode === 'auto' ? 'auto' : 'normal'
  const suppressed = Boolean(options.starlight)

  let particleCount = mode === 'auto' ? 3 : 6
  let durationMs = mode === 'auto' ? 900 : 1200
  if (suppressed) {
    particleCount = Math.max(2, particleCount - 2)
    durationMs = Math.max(520, durationMs - 380)
  }

  return {
    trailTheme: trail?.preview?.theme || '',
    impactTheme: impact?.preview?.theme || '',
    trailClass: trail ? `cosmetic-theme--${trail.preview?.theme || 'default'}` : '',
    impactClass: impact ? `cosmetic-theme--${impact.preview?.theme || 'default'}` : '',
    particleCount,
    durationMs,
    mode,
    suppressed,
  }
}

export function cosmeticStatusText(item) {
  if (item?.equipped) {
    return '已装备'
  }
  if (item?.owned) {
    return '已拥有'
  }
  return `${item?.price ?? 0} 原石`
}

export function salvageableCount(item, active = false) {
  const quantity = Number(item?.quantity ?? 0)
  const protectedCount = item?.equipped || active ? 1 : 0
  return Math.max(0, quantity - protectedCount)
}
