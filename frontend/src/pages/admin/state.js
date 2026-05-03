import { DEFAULT_RARITY, normalizeRarity } from '../../utils/rarity'
import { EQUIPMENT_SLOTS, normalizeEquipmentSlot, normalizeLoadout as normalizeEquipmentLoadout } from '../../utils/equipmentSlots'

export { EQUIPMENT_SLOTS }

export function emptyAdminState() {
  return {
    roomId: '1',
    queueId: '1',
    boss: null,
    bossLeaderboard: [],
    loot: [],
    bossCycleEnabled: false,
    bossPool: [],
    bossCycleQueue: [],
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
    critRate: '',
    critDamageMultiplier: '',
    bossDamagePercent: '',
    partTypeDamageSoft: '',
    partTypeDamageHeavy: '',
    partTypeDamageWeak: '',
    talentAffinity: '',
  }
}

export function emptyShopItemForm() {
  return {
    itemId: '',
    title: '',
    itemType: 'battle_click_skin',
    priceGold: 0,
    imagePath: '',
    imageAlt: '',
    previewImagePath: '',
    battleClickCursorImagePath: '',
    description: '',
    active: true,
    sortOrder: 0,
    autoEquipOnPurchase: true,
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

export function emptyTaskForm() {
  return {
    taskId: '',
    title: '',
    description: '',
    taskType: 'daily',
    eventKind: 'click',
    windowKind: 'daily',
    status: 'draft',
    conditionKind: 'daily_clicks',
    targetValue: 1,
    rewards: {
      gold: 0,
      stones: 0,
      talentPoints: 0,
      equipmentItems: [],
    },
    displayOrder: 0,
    startAt: 0,
    endAt: 0,
  }
}

function taskEventKindFromLegacy(conditionKind) {
  switch (conditionKind) {
    case 'boss_kills':
      return 'boss_kill'
    case 'enhance_count':
      return 'enhance'
    default:
      return 'click'
  }
}

function taskWindowKindFromLegacy(taskType, conditionKind) {
  if (conditionKind === 'weekly_clicks' && taskType !== 'limited') {
    return 'weekly'
  }
  if (conditionKind === 'daily_clicks' && taskType !== 'limited') {
    return 'daily'
  }
  switch (taskType) {
    case 'weekly':
      return 'weekly'
    case 'limited':
      return 'fixed_range'
    default:
      return 'daily'
  }
}

function legacyTaskTypeFromWindowKind(windowKind) {
  switch (windowKind) {
    case 'weekly':
      return 'weekly'
    case 'fixed_range':
      return 'limited'
    default:
      return 'daily'
  }
}

function legacyConditionKindFromModel(eventKind, windowKind) {
  switch (eventKind) {
    case 'boss_kill':
      return 'boss_kills'
    case 'enhance':
      return 'enhance_count'
    default:
      return windowKind === 'weekly' ? 'weekly_clicks' : 'daily_clicks'
  }
}

function normalizeTaskModelFields(payload) {
  const eventKind = payload?.eventKind || taskEventKindFromLegacy(payload?.conditionKind || '')
  const windowKind = payload?.windowKind || taskWindowKindFromLegacy(payload?.taskType || '', payload?.conditionKind || '')
  return {
    eventKind,
    windowKind,
    taskType: legacyTaskTypeFromWindowKind(windowKind),
    conditionKind: legacyConditionKindFromModel(eventKind, windowKind),
  }
}

export function emptyMessagePage() {
  return {
    items: [],
    nextCursor: '',
  }
}

export function emptyTaskCycleResults() {
  return {
    archive: null,
    items: [],
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
  return [{ itemId: '', dropRatePercent: '' }]
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
    dropRatePercent: Number(entry?.dropRatePercent ?? entry?.weight ?? 0),
    attackPower: Number(entry?.attackPower ?? 0),
    armorPenPercent: Number(entry?.armorPenPercent ?? 0),
    critRate: Number(entry?.critRate ?? 0),
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
    goldOnKill: Number(entry?.goldOnKill ?? 0),
    stoneOnKill: Number(entry?.stoneOnKill ?? 0),
    talentPointsOnKill: Number(entry?.talentPointsOnKill ?? 0),
    layout: Array.isArray(entry?.layout) ? entry.layout : [],
    loot: Array.isArray(entry?.loot) ? entry.loot.map(normalizeLootEntry) : [],
  }
}

export function normalizeAdminState(payload) {
  return {
    roomId: String(payload?.roomId || '1'),
    queueId: String(payload?.queueId || payload?.roomId || '1'),
    boss: payload?.boss ?? null,
    bossLeaderboard: Array.isArray(payload?.bossLeaderboard) ? payload.bossLeaderboard : [],
    loot: Array.isArray(payload?.loot) ? payload.loot.map(normalizeLootEntry) : [],
    bossCycleEnabled: Boolean(payload?.bossCycleEnabled),
    bossPool: Array.isArray(payload?.bossPool) ? payload.bossPool.map(normalizeBossTemplate) : [],
    bossCycleQueue: Array.isArray(payload?.bossCycleQueue) ? payload.bossCycleQueue.filter(Boolean) : [],
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

export function normalizeShopItem(payload) {
  return {
    ...emptyShopItemForm(),
    itemId: payload?.itemId || '',
    title: payload?.title || '',
    itemType: payload?.itemType || 'battle_click_skin',
    priceGold: Number(payload?.priceGold ?? 0),
    imagePath: payload?.imagePath || '',
    imageAlt: payload?.imageAlt || '',
    previewImagePath: payload?.previewImagePath || '',
    battleClickCursorImagePath: payload?.battleClickCursorImagePath || '',
    description: payload?.description || '',
    active: Boolean(payload?.active),
    sortOrder: Number(payload?.sortOrder ?? 0),
    autoEquipOnPurchase: payload?.autoEquipOnPurchase !== false,
    createdAt: Number(payload?.createdAt ?? 0),
    updatedAt: Number(payload?.updatedAt ?? 0),
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

export function normalizeTaskDefinition(payload) {
  const model = normalizeTaskModelFields(payload)
  return {
    ...emptyTaskForm(),
    taskId: payload?.taskId || '',
    title: payload?.title || '',
    description: payload?.description || '',
    taskType: model.taskType,
    eventKind: model.eventKind,
    windowKind: model.windowKind,
    status: payload?.status || 'draft',
    conditionKind: model.conditionKind,
    targetValue: Number(payload?.targetValue ?? 1),
    rewards: {
      gold: Number(payload?.rewards?.gold ?? 0),
      stones: Number(payload?.rewards?.stones ?? 0),
      talentPoints: Number(payload?.rewards?.talentPoints ?? 0),
      equipmentItems: Array.isArray(payload?.rewards?.equipmentItems)
        ? payload.rewards.equipmentItems.map((entry) => ({
            itemId: entry?.itemId || '',
            quantity: Number(entry?.quantity ?? 1),
          }))
        : [],
    },
    displayOrder: Number(payload?.displayOrder ?? 0),
    startAt: Number(payload?.startAt ?? 0),
    endAt: Number(payload?.endAt ?? 0),
    createdAt: Number(payload?.createdAt ?? 0),
    updatedAt: Number(payload?.updatedAt ?? 0),
  }
}

export function normalizeTaskArchive(payload) {
  const model = normalizeTaskModelFields(payload)
  return {
    taskId: payload?.taskId || '',
    cycleKey: payload?.cycleKey || '',
    taskType: model.taskType,
    eventKind: model.eventKind,
    windowKind: model.windowKind,
    conditionKind: model.conditionKind,
    targetValue: Number(payload?.targetValue ?? 0),
    startAt: Number(payload?.startAt ?? 0),
    endAt: Number(payload?.endAt ?? 0),
    participantsTotal: Number(payload?.participantsTotal ?? 0),
    completedTotal: Number(payload?.completedTotal ?? 0),
    claimedTotal: Number(payload?.claimedTotal ?? 0),
    expiredUnclaimedTotal: Number(payload?.expiredUnclaimedTotal ?? 0),
    unfinishedTotal: Number(payload?.unfinishedTotal ?? 0),
    notParticipatedTotal: Number(payload?.notParticipatedTotal ?? 0),
    archivedAt: Number(payload?.archivedAt ?? 0),
  }
}

export function normalizeTaskCycleResults(payload) {
  return {
    archive: payload?.archive ? normalizeTaskArchive(payload.archive) : null,
    items: Array.isArray(payload?.items)
      ? payload.items.map((entry) => ({
          taskId: entry?.taskId || '',
          cycleKey: entry?.cycleKey || '',
          nickname: entry?.nickname || '',
          progress: Number(entry?.progress ?? 0),
          targetValue: Number(entry?.targetValue ?? 0),
          status: entry?.status || 'unfinished',
          completedAt: Number(entry?.completedAt ?? 0),
          claimedAt: Number(entry?.claimedAt ?? 0),
          archivedAt: Number(entry?.archivedAt ?? 0),
        }))
      : [],
  }
}

export function formatItemStats(item) {
  const parts = []
  if (item.attackPower != null) parts.push(`攻击力 +${item.attackPower}`)
  if (item.armorPenPercent != null) parts.push(`破甲 +${(item.armorPenPercent * 100).toFixed(0)}%`)
  if (item.critRate != null) parts.push(`暴击率 +${(item.critRate * 100).toFixed(0)}%`)
  if (item.critDamageMultiplier != null) parts.push(`暴伤 +${item.critDamageMultiplier.toFixed(1)}`)
  if (item.partTypeDamageSoft != null) parts.push(`软组织 +${(item.partTypeDamageSoft * 100).toFixed(0)}%`)
  if (item.partTypeDamageHeavy != null) parts.push(`重甲 +${(item.partTypeDamageHeavy * 100).toFixed(0)}%`)
  if (item.partTypeDamageWeak != null) parts.push(`弱点 +${(item.partTypeDamageWeak * 100).toFixed(0)}%`)
  return parts.join('，') || '无主要属性'
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
