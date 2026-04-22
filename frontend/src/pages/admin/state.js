export function emptyAdminState() {
  return {
    boss: null,
    bossLeaderboard: [],
    heroes: [],
    loot: [],
    heroLoot: [],
    bossCycleEnabled: false,
    bossPool: [],
    playerCount: 0,
    recentPlayerCount: 0,
  }
}

export function emptyButtonPage() {
  return {
    items: [],
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  }
}

export function emptyEquipmentForm() {
  return {
    itemId: '',
    name: '',
    slot: 'weapon',
    bonusClicks: '',
    bonusCriticalChancePercent: '',
    bonusCriticalCount: '',
  }
}

export function emptyButtonForm() {
  return {
    slug: '',
    label: '',
    sort: '',
    enabled: true,
    tagsText: '',
    starlightEligible: false,
    imagePath: '',
    imageAlt: '',
  }
}

export function emptyHeroForm() {
  return {
    heroId: '',
    name: '',
    imagePath: '',
    imageAlt: '',
    bonusClicks: '',
    bonusCriticalChancePercent: '',
    bonusCriticalCount: '',
    traitType: 'bonus_clicks',
    traitValue: '',
  }
}

export function emptyAnnouncementForm() {
  return {
    title: '',
    content: '',
    active: true,
  }
}

export function emptyEquipmentPage() {
  return {
    items: [],
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  }
}

export function emptyBossHistoryPage() {
  return {
    items: [],
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  }
}

export function emptyMessagePage() {
  return {
    items: [],
    nextCursor: '',
  }
}

export function emptyPlayerPage() {
  return {
    items: [],
    nextCursor: '',
    total: 0,
  }
}

export function emptyLootRows() {
  return [{ itemId: '', weight: '' }]
}

export function emptyHeroLootRows() {
  return [{ heroId: '', weight: '' }]
}

export function normalizeLoadout(loadout) {
  return {
    weapon: loadout?.weapon ?? null,
    armor: loadout?.armor ?? null,
    accessory: loadout?.accessory ?? null,
  }
}

export function normalizeLootEntry(entry) {
  return {
    itemId: entry?.itemId || '',
    itemName: entry?.itemName || '',
    slot: entry?.slot || '',
    weight: Number(entry?.weight ?? 0),
    bonusClicks: Number(entry?.bonusClicks ?? 0),
    bonusCriticalChancePercent: Number(entry?.bonusCriticalChancePercent ?? 0),
    bonusCriticalCount: Number(entry?.bonusCriticalCount ?? 0),
  }
}

export function normalizeHeroLootEntry(entry) {
  return {
    heroId: entry?.heroId || '',
    heroName: entry?.heroName || '',
    imagePath: entry?.imagePath || '',
    imageAlt: entry?.imageAlt || '',
    weight: Number(entry?.weight ?? 0),
    dropRatePercent: Number(entry?.dropRatePercent ?? 0),
    bonusClicks: Number(entry?.bonusClicks ?? 0),
    bonusCriticalChancePercent: Number(entry?.bonusCriticalChancePercent ?? 0),
    bonusCriticalCount: Number(entry?.bonusCriticalCount ?? 0),
    traitType: entry?.traitType || '',
    traitValue: Number(entry?.traitValue ?? 0),
  }
}

export function normalizeHeroDefinition(entry) {
  return {
    heroId: entry?.heroId || '',
    name: entry?.name || '',
    imagePath: entry?.imagePath || '',
    imageAlt: entry?.imageAlt || '',
    bonusClicks: Number(entry?.bonusClicks ?? 0),
    bonusCriticalChancePercent: Number(entry?.bonusCriticalChancePercent ?? 0),
    bonusCriticalCount: Number(entry?.bonusCriticalCount ?? 0),
    traitType: entry?.traitType || 'bonus_clicks',
    traitValue: Number(entry?.traitValue ?? 0),
  }
}

export function normalizeBossTemplate(entry) {
  return {
    id: entry?.id || '',
    name: entry?.name || '',
    maxHp: Number(entry?.maxHp ?? 0),
    loot: Array.isArray(entry?.loot) ? entry.loot.map(normalizeLootEntry) : [],
    heroLoot: Array.isArray(entry?.heroLoot) ? entry.heroLoot.map(normalizeHeroLootEntry) : [],
  }
}

export function normalizeAdminState(payload) {
  return {
    boss: payload?.boss ?? null,
    bossLeaderboard: Array.isArray(payload?.bossLeaderboard) ? payload.bossLeaderboard : [],
    heroes: Array.isArray(payload?.heroes) ? payload.heroes.map(normalizeHeroDefinition) : [],
    loot: Array.isArray(payload?.loot) ? payload.loot.map(normalizeLootEntry) : [],
    heroLoot: Array.isArray(payload?.heroLoot) ? payload.heroLoot.map(normalizeHeroLootEntry) : [],
    bossCycleEnabled: Boolean(payload?.bossCycleEnabled),
    bossPool: Array.isArray(payload?.bossPool) ? payload.bossPool.map(normalizeBossTemplate) : [],
    playerCount: Number(payload?.playerCount ?? 0),
    recentPlayerCount: Number(payload?.recentPlayerCount ?? 0),
  }
}

export function normalizeButtonPage(payload) {
  return {
    items: Array.isArray(payload?.items) ? payload.items : [],
    page: Number(payload?.page ?? 1),
    pageSize: Number(payload?.pageSize ?? 20),
    total: Number(payload?.total ?? 0),
    totalPages: Number(payload?.totalPages ?? 0),
  }
}

export function normalizeEquipmentPage(payload) {
  return {
    items: Array.isArray(payload?.items) ? payload.items : [],
    page: Number(payload?.page ?? 1),
    pageSize: Number(payload?.pageSize ?? 20),
    total: Number(payload?.total ?? 0),
    totalPages: Number(payload?.totalPages ?? 0),
  }
}

export function normalizeBossHistoryPage(payload) {
  return {
    items: Array.isArray(payload?.items)
      ? payload.items.map((entry) => ({
          ...entry,
          loot: Array.isArray(entry?.loot) ? entry.loot.map(normalizeLootEntry) : [],
          damage: Array.isArray(entry?.damage) ? entry.damage : [],
        }))
      : [],
    page: Number(payload?.page ?? 1),
    pageSize: Number(payload?.pageSize ?? 20),
    total: Number(payload?.total ?? 0),
    totalPages: Number(payload?.totalPages ?? 0),
  }
}

export function normalizeAnnouncements(payload) {
  return Array.isArray(payload)
    ? payload.map((item) => ({
        id: item?.id || '',
        title: item?.title || '',
        content: item?.content || '',
        publishedAt: Number(item?.publishedAt ?? 0),
        active: Boolean(item?.active),
      }))
    : []
}

export function normalizeMessagePage(payload) {
  return {
    items: Array.isArray(payload?.items)
      ? payload.items.map((item) => ({
          id: item?.id || '',
          nickname: item?.nickname || '',
          content: item?.content || '',
          createdAt: Number(item?.createdAt ?? 0),
        }))
      : [],
    nextCursor: payload?.nextCursor || '',
  }
}

export function normalizePlayerPage(payload) {
  return {
    items: Array.isArray(payload?.items)
      ? payload.items.map((player) => ({
          nickname: player?.nickname || '',
          clickCount: Number(player?.clickCount ?? 0),
          inventory: Array.isArray(player?.inventory) ? player.inventory : [],
          loadout: normalizeLoadout(player?.loadout),
        }))
      : [],
    nextCursor: payload?.nextCursor || '',
    total: Number(payload?.total ?? 0),
  }
}

export function formatItemStats(item) {
  return `点击+${item?.bonusClicks ?? 0} 暴击率+${item?.bonusCriticalChancePercent ?? 0}% 暴击+${item?.bonusCriticalCount ?? 0}`
}

export function formatHeroTrait(hero) {
  switch (hero?.traitType) {
    case 'bonus_clicks':
      return `被动：额外点击 +${hero?.traitValue ?? 0}`
    case 'critical_chance_percent':
      return `被动：暴击率 +${hero?.traitValue ?? 0}%`
    case 'critical_count_bonus':
      return `被动：暴击额外 +${hero?.traitValue ?? 0}`
    case 'final_damage_percent':
      return `被动：最终伤害 +${hero?.traitValue ?? 0}%`
    default:
      return '被动：暂无'
  }
}

export function heroImageAlt(hero) {
  return hero?.imageAlt || hero?.name || hero?.heroId || '英雄头像'
}

export function formatTime(timestamp) {
  if (!timestamp) {
    return '未记录'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(timestamp * 1000))
}
