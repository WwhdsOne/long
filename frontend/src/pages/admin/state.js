import { DEFAULT_RARITY, normalizeRarity } from '../../utils/rarity'
import { EQUIPMENT_SLOTS, normalizeEquipmentSlot, normalizeLoadout as normalizeEquipmentLoadout } from '../../utils/equipmentSlots'

export { EQUIPMENT_SLOTS }

export function emptyAdminState() {
  return {
    boss: null,
    bossLeaderboard: [],
    loot: [],
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
    slot: EQUIPMENT_SLOTS[0].value,
    rarity: DEFAULT_RARITY,
    imagePath: '',
    imageAlt: '',
    attackPower: '',
    armorPenPercent: '',
    critDamageMultiplier: '',
    bossDamagePercent: '',
    partTypeDamageSoft: '',
    partTypeDamageHeavy: '',
    partTypeDamageWeak: '',
    talentAffinity: '',
  }
}

export function emptyButtonForm() {
  return {
    slug: '',
    label: '',
    sort: '',
    enabled: true,
    tagsText: '',
    imagePath: '',
    imageAlt: '',
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

export function normalizeLoadout(loadout) {
  return normalizeEquipmentLoadout(loadout)
}

export function normalizeLootEntry(entry) {
  return {
    itemId: entry?.itemId || '',
    itemName: entry?.itemName || '',
    slot: normalizeEquipmentSlot(entry?.slot),
    rarity: normalizeRarity(entry?.rarity),
    weight: Number(entry?.weight ?? 0),
    attackPower: Number(entry?.attackPower ?? 0),
    armorPenPercent: Number(entry?.armorPenPercent ?? 0),
    critDamageMultiplier: Number(entry?.critDamageMultiplier ?? 0),
    bossDamagePercent: Number(entry?.bossDamagePercent ?? 0),
    partTypeDamageSoft: Number(entry?.partTypeDamageSoft ?? 0),
    partTypeDamageHeavy: Number(entry?.partTypeDamageHeavy ?? 0),
    partTypeDamageWeak: Number(entry?.partTypeDamageWeak ?? 0),
    talentAffinity: entry?.talentAffinity || '',
  }
}

export function normalizeBossTemplate(entry) {
  return {
    id: entry?.id || '',
    name: entry?.name || '',
    maxHp: Number(entry?.maxHp ?? 0),
    layout: Array.isArray(entry?.layout) ? entry.layout : [],
    loot: Array.isArray(entry?.loot) ? entry.loot.map(normalizeLootEntry) : [],
  }
}

export function normalizeAdminState(payload) {
  return {
    boss: payload?.boss ?? null,
    bossLeaderboard: Array.isArray(payload?.bossLeaderboard) ? payload.bossLeaderboard : [],
    loot: Array.isArray(payload?.loot) ? payload.loot.map(normalizeLootEntry) : [],
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
    items: Array.isArray(payload?.items)
      ? payload.items.map((item) => ({
          ...item,
          slot: normalizeEquipmentSlot(item?.slot),
          rarity: normalizeRarity(item?.rarity),
        }))
      : [],
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
  const parts = []
  if (Number(item?.attackPower ?? 0) > 0) parts.push(`攻击力+${item.attackPower}`)
  if (Number(item?.armorPenPercent ?? 0) > 0) parts.push(`破甲+${(item.armorPenPercent * 100).toFixed(0)}%`)
  if (Number(item?.critDamageMultiplier ?? 0) > 1) parts.push(`暴伤×${item.critDamageMultiplier}`)
  if (Number(item?.bossDamagePercent ?? 0) > 0) parts.push(`Boss增伤+${(item.bossDamagePercent * 100).toFixed(0)}%`)
  if (Number(item?.partTypeDamageSoft ?? 0) > 0) parts.push(`软组织+${(item.partTypeDamageSoft * 100).toFixed(0)}%`)
  if (Number(item?.partTypeDamageHeavy ?? 0) > 0) parts.push(`重甲+${(item.partTypeDamageHeavy * 100).toFixed(0)}%`)
  if (Number(item?.partTypeDamageWeak ?? 0) > 0) parts.push(`弱点+${(item.partTypeDamageWeak * 100).toFixed(0)}%`)
  if (item?.talentAffinity) parts.push(`天赋:${item.talentAffinity}`)
  return parts.join(' ')
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
