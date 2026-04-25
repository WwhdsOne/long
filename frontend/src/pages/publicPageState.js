import {computed, onBeforeUnmount, onMounted, ref} from 'vue'

import {mergeBossState} from '../utils/bossState'
import {formatDropRate} from '../utils/buttonBoard'
import {buildClickRequestBody, mergeClickFallbackState} from '../utils/clickResponse'
import { formatRarityLabel, getRarityClassName, splitEquipmentName } from '../utils/rarity'
import { EQUIPMENT_SLOTS, normalizeLoadout } from '../utils/equipmentSlots'
import {createRealtimeTransport} from '../utils/realtimeTransport'
import { buildFingerprintProof, collectFingerprintHash, createClickBehaviorTracker } from '../utils/manualClickSignals'

const ANNOUNCEMENT_READ_KEY = 'vote-wall-announcement-read'
const ANNOUNCEMENT_CACHE_KEY = 'vote-wall-announcement-cache'
const AUTO_CLICK_RATE_LABEL = '每秒固定 3 次'
const EQUIPMENT_ENHANCE_COST = 10
const GROWTH_FORMULA_TEXT = '点击 / 暴击单次成长 = ceil((当前点击 + 当前暴击 + 当前暴击率) / 4)，至少 +1'

const publicPages = [
  {id: 'battle', label: '战斗', path: '/'},
  {id: 'talents', label: '天赋', path: '/talents'},
  {id: 'profile', label: '资料', path: '/profile'},
  {id: 'messages', label: '消息', path: '/messages'},
]

const buttons = ref([])
const buttonTotalCount = ref(0)
const buttonTotalVotes = ref(0)
const leaderboard = ref([])
const boss = ref(null)
const bossLeaderboard = ref([])
const bossLoot = ref([])
const announcementVersion = ref('')
const latestAnnouncement = ref(null)
const announcements = ref([])
const myBossStats = ref(null)
const inventory = ref([])
const loadout = ref(emptyLoadout())
const combatStats = ref(defaultCombatStats())
const recentRewards = ref([])
const lastReward = ref(null)
const userStats = ref(null)
const nickname = ref('')
const nicknameDraft = ref('')
const passwordDraft = ref('')
const loading = ref(true)
const syncing = ref(false)
const errorMessage = ref('')
const pendingKeys = ref(new Set())
const actioningItemId = ref('')
const activeHudTab = ref('account')
const lastUpdatedAt = ref('')
const liveConnected = ref(false)
const criticalBursts = ref({})
const onlineCount = ref(null)
const bossHistory = ref([])
const bossHistoryQuery = ref('')
const loadingBossHistory = ref(false)
const bossHistoryLoaded = ref(false)
const bossHistoryError = ref('')
const loadingAnnouncements = ref(false)
const announcementsLoaded = ref(false)
const announcementError = ref('')
const loadingBossResources = ref(false)
const latestAnnouncementLoaded = ref(false)
const announcementModalOpen = ref(false)
const messages = ref([])
const messageNextCursor = ref('')
const loadingMessages = ref(false)
const postingMessage = ref(false)
const messageDraft = ref('')
const messageError = ref('')
const autoClickEnabled = ref(false)
const autoClickTargetKey = ref('')
const gems = ref(0)
const fingerprintHash = ref('')
const currentPublicPage = ref(resolvePublicPage(window.location.pathname))
const profileLoading = ref(false)
const profileLoaded = ref(false)
const profileNotice = ref('')

let realtimeTransport
let lastBossResourceVersion = ''
const burstTimers = new Map()
const pendingClickSources = new Map()
const clickBehaviorTracker = createClickBehaviorTracker()
let fingerprintPromise

const buttonCount = computed(() => buttonTotalCount.value || buttons.value.length)
const totalVotes = computed(() =>
    buttonTotalVotes.value || buttons.value.reduce((total, button) => total + button.count, 0),
)
const displayedButtons = computed(() => buttons.value)
const syncLabel = computed(() => {
  if (syncing.value) {
    return '同步中'
  }

  if (!liveConnected.value) {
    return '正在重连'
  }
  return onlineCount.value !== null ? `${onlineCount.value} 人在线` : '已连接'
})
const isLoggedIn = computed(() => nickname.value !== '')
const myClicks = computed(() => userStats.value?.clickCount ?? 0)
const myRank = computed(() => {
  if (!nickname.value) {
    return null
  }

  const matched = leaderboard.value.find((entry) => entry.nickname === nickname.value)
  return matched?.rank ?? null
})
const myBossDamage = computed(() => myBossStats.value?.damage ?? 0)
const myBossRank = computed(() => {
  if (!nickname.value || !boss.value) return null
  if (myBossStats.value?.rank) return myBossStats.value.rank
  const matched = bossLeaderboard.value.find((entry) => entry.nickname === nickname.value)
  return matched?.rank ?? null
})
const effectiveIncrement = computed(() => combatStats.value?.effectiveIncrement ?? 1)
const normalDamage = computed(() => combatStats.value?.normalDamage ?? effectiveIncrement.value)
const criticalDamage = computed(() => combatStats.value?.criticalDamage ?? normalDamage.value)
const autoClickTargetButton = computed(() =>
    buttons.value.find((button) => button.key === autoClickTargetKey.value) ?? null,
)
const autoClickTargetLabel = computed(() => (autoClickTargetButton.value?.label ?? autoClickTargetKey.value) || '未选择')
const canStartAutoClick = computed(() => isLoggedIn.value && Boolean(autoClickTargetKey.value))
const autoClickStatus = computed(() => {
  if (!isLoggedIn.value) {
    return '请先登录账号，再选择按钮并开启官方挂机。'
  }
  if (!autoClickTargetKey.value) {
    return '请先手动点击一次按钮锁定目标，再开启官方挂机。'
  }
  if (autoClickEnabled.value) {
    return `✅ 官方挂机已开启：正在服务端托管【${autoClickTargetLabel.value}】，关闭页面、退出浏览器仍会持续自动挂机；手动点击其他按钮会立即切换挂机目标。`
  }

  return `已锁定目标：【${autoClickTargetLabel.value}】，开启后将按 ${AUTO_CLICK_RATE_LABEL} 在服务端持续自动结算，关闭页面也不会停止挂机。`
})
const bossStatusLabel = computed(() => {
  if (!boss.value) {
    return '休战中'
  }
  if (boss.value.status === 'active') {
    return '活动中'
  }
  if (boss.value.status === 'defeated') {
    return '已击败'
  }
  return boss.value.status || '待开启'
})
const bossProgress = computed(() => {
  if (!boss.value || !boss.value.maxHp) {
    return 0
  }

  return Math.max(0, Math.min(100, (boss.value.currentHp / boss.value.maxHp) * 100))
})
const loadoutSlots = EQUIPMENT_SLOTS
const equippedItems = computed(() => loadoutSlots.map((slot) => loadout.value[slot.value]).filter(Boolean))
const displayedRecentRewards = computed(() => {
  if (Array.isArray(recentRewards.value) && recentRewards.value.length > 0) {
    return recentRewards.value
  }
  return lastReward.value ? [lastReward.value] : []
})
const recentRewardTitle = computed(() => {
  if (displayedRecentRewards.value.length === 0) {
    return '暂无'
  }
  return displayedRecentRewards.value
      .map((reward, index) => {
        // 第一个装备掉落
        if (index === 0) return `${reward.itemName}`
        return reward.itemName
      })
      .join('、')
})
const recentRewardNote = computed(() => {
  if (displayedRecentRewards.value.length === 0) {
    return '还没有新的掉落记录。'
  }

  const bossName = displayedRecentRewards.value[0]?.bossName || displayedRecentRewards.value[0]?.bossId || '当前 Boss'
  if (displayedRecentRewards.value.length === 1) {
    return `来自 ${bossName}，已经放进你的背包。`
  }

  return `来自 ${bossName}，本次共掉落 ${displayedRecentRewards.value.length} 件：${displayedRecentRewards.value
      .map((reward) => reward.itemName)
      .join('、')}。`
})
const filteredBossHistory = computed(() => {
  const query = normalizeNickname(bossHistoryQuery.value).toLowerCase()
  if (!query) {
    return bossHistory.value.slice(0, 12)
  }

  return bossHistory.value
      .filter((entry) => [entry.name, entry.id].some((value) => String(value || '').toLowerCase().includes(query)))
      .slice(0, 12)
})

function emptyLoadout() {
  return normalizeLoadout(null)
}

function defaultCombatStats() {
  return {
    effectiveIncrement: 1,
    normalDamage: 1,
    criticalDamage: 1,
    criticalChancePercent: 0,
    criticalCount: 1,
    attackPower: 0,
    armorPenPercent: 0,
    critDamageMultiplier: 0,
    bossDamagePercent: 0,
    allDamageAmplify: 0,
  }
}

function formatItemStats(item) {
  const parts = []
  if (item?.attackPower) parts.push(`攻击 ${item.attackPower}`)
  if (item?.armorPenPercent) parts.push(`穿透 ${item.armorPenPercent}%`)
  if (item?.critDamageMultiplier) parts.push(`暴伤 ${item.critDamageMultiplier}`)
  if (item?.bossDamagePercent) parts.push(`首领伤 ${item.bossDamagePercent}%`)
  return parts.join(' ') || '无属性'
}

function formatItemStatLines(item) {
  const lines = []
  if (item?.attackPower) lines.push(`攻击力 ${item.attackPower}`)
  if (item?.armorPenPercent) lines.push(`护甲穿透 ${item.armorPenPercent}%`)
  if (item?.critDamageMultiplier) lines.push(`暴击伤害倍率 ${item.critDamageMultiplier}`)
  if (item?.bossDamagePercent) lines.push(`首领伤害 ${item.bossDamagePercent}%`)
  return lines
}

function equipmentNameParts(item) {
  return splitEquipmentName(item?.itemName || item?.name || item?.itemId || '')
}

function equipmentNameClass(item) {
  return getRarityClassName(item?.rarity)
}

function normalizeNickname(value) {
  return value.trim()
}

function resolvePublicPage(pathname) {
  if (pathname.startsWith('/messages')) {
    return 'messages'
  }
  if (pathname.startsWith('/profile')) {
    return 'profile'
  }
  return 'battle'
}

async function navigatePublicPage(page) {
  const target = publicPages.find((item) => item.id === page) ?? publicPages[0]
  if (currentPublicPage.value !== target.id) {
    window.history.pushState({}, '', target.path)
    currentPublicPage.value = target.id
  }
  await activatePublicPage(target.id)
}

async function activatePublicPage(page) {
  if (page === 'profile') {
    if (activeHudTab.value === 'messages' || activeHudTab.value === 'info') {
      activeHudTab.value = 'account'
    }
    await loadPlayerProfile(true)
    return
  }
  if (page === 'messages') {
    activeHudTab.value = 'messages'
    await loadMessages()
    await loadAnnouncements()
  }
}

function handlePublicRouteChange() {
  const nextPage = resolvePublicPage(window.location.pathname)
  if (nextPage === 'battle') {
    currentPublicPage.value = 'battle'
  } else {
    currentPublicPage.value = nextPage
  }
  void activatePublicPage(currentPublicPage.value)
}

function formatBossTime(timestamp) {
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

function topBossDamage(entry) {
  return entry?.damage?.[0] ?? null
}

function formatTime(timestamp) {
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

function formatNumber(value, digits = 0) {
  const normalized = Number(value ?? 0)
  return digits > 0 ? normalized.toFixed(digits) : `${normalized}`
}

function formatStatWithDelta(label, total, delta) {
  return `${label} ${formatNumber(total)}（+${formatNumber(delta)}）`
}

function formatPercentWithDelta(label, total, delta) {
  return `${label} ${formatNumber(total, 2)}%（+${formatNumber(delta, 2)}%）`
}

function dotIndexes(count) {
  const normalized = Math.max(0, Math.min(Number(count) || 0, 6))
  return Array.from({length: normalized}, (_, index) => index)
}

async function readErrorMessage(response, fallback) {
  try {
    const payload = await response.json()
    if (payload?.message) {
      return payload.message
    }
  } catch {
    // Ignore malformed error payloads and keep fallback copy.
  }

  return fallback
}

function bossResourceVersion(value = boss.value) {
  if (!value?.id) {
    return ''
  }
  return `${value.id}:${value.status || ''}`
}

function readCachedLatestAnnouncement() {
  try {
    const raw = window.localStorage.getItem(ANNOUNCEMENT_CACHE_KEY)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || !parsed.id) {
      return null
    }
    return parsed
  } catch {
    return null
  }
}

function writeCachedLatestAnnouncement(item) {
  if (!item?.id) {
    window.localStorage.removeItem(ANNOUNCEMENT_CACHE_KEY)
    return
  }
  window.localStorage.setItem(ANNOUNCEMENT_CACHE_KEY, JSON.stringify(item))
}

function restoreCachedLatestAnnouncement() {
  const cached = readCachedLatestAnnouncement()
  if (cached?.id) {
    latestAnnouncement.value = cached
  }
}

function maybePromptAnnouncement() {
  if (!latestAnnouncement.value?.id) {
    return
  }

  const readId = window.localStorage.getItem(ANNOUNCEMENT_READ_KEY)
  if (readId !== latestAnnouncement.value.id) {
    announcementModalOpen.value = true
  }
}

function closeAnnouncementModal() {
  if (latestAnnouncement.value?.id) {
    window.localStorage.setItem(ANNOUNCEMENT_READ_KEY, latestAnnouncement.value.id)
  }
  announcementModalOpen.value = false
}

async function loadBossResources(force = false) {
  const currentVersion = bossResourceVersion()
  if (!currentVersion) {
    bossLoot.value = []
    lastBossResourceVersion = ''
    return
  }
  if (loadingBossResources.value) {
    return
  }
  if (!force && lastBossResourceVersion === currentVersion) {
    return
  }

  loadingBossResources.value = true
  try {
    const response = await fetch('/api/boss/resources')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, 'Boss 掉落池加载失败'))
    }
    const payload = await response.json()
    bossLoot.value = Array.isArray(payload?.bossLoot) ? payload.bossLoot : []
    lastBossResourceVersion = currentVersion
  } catch {
    if (force) {
      bossLoot.value = []
    }
  } finally {
    loadingBossResources.value = false
  }
}

async function loadLatestAnnouncement(force = false) {
  if (!announcementVersion.value) {
    latestAnnouncement.value = null
    latestAnnouncementLoaded.value = true
    writeCachedLatestAnnouncement(null)
    announcementModalOpen.value = false
    return
  }

  const cached = readCachedLatestAnnouncement()
  if (cached?.id === announcementVersion.value) {
    latestAnnouncement.value = cached
    if (!force && latestAnnouncementLoaded.value) {
      maybePromptAnnouncement()
      return
    }
  }

  try {
    const response = await fetch('/api/announcements/latest')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '最新公告加载失败'))
    }
    const payload = await response.json()
    latestAnnouncement.value = payload?.id ? payload : null
    latestAnnouncementLoaded.value = true
    writeCachedLatestAnnouncement(latestAnnouncement.value)
    maybePromptAnnouncement()
  } catch {
    if (cached?.id === announcementVersion.value) {
      latestAnnouncement.value = cached
      latestAnnouncementLoaded.value = true
      maybePromptAnnouncement()
    }
  }
}

async function loadAnnouncements(force = false) {
  if ((announcementsLoaded.value || loadingAnnouncements.value) && !force) {
    return
  }

  loadingAnnouncements.value = true
  announcementError.value = ''
  try {
    const response = await fetch('/api/announcements')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '公告加载失败'))
    }
    const payload = await response.json()
    announcements.value = Array.isArray(payload) ? payload : []
    announcementsLoaded.value = true
  } catch (error) {
    announcementError.value = error.message || '公告加载失败'
  } finally {
    loadingAnnouncements.value = false
  }
}

async function loadMessages(cursor = '', append = false) {
  if (loadingMessages.value) {
    return
  }

  loadingMessages.value = true
  messageError.value = ''
  try {
    const query = cursor ? `?cursor=${encodeURIComponent(cursor)}` : ''
    const response = await fetch(`/api/messages${query}`)
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '留言加载失败'))
    }

    const payload = await response.json()
    const items = Array.isArray(payload?.items) ? payload.items : []
    messages.value = append ? [...messages.value, ...items] : items
    messageNextCursor.value = payload?.nextCursor || ''
  } catch (error) {
    messageError.value = error.message || '留言加载失败'
  } finally {
    loadingMessages.value = false
  }
}

async function submitMessage() {
  if (!nickname.value) {
    messageError.value = '先登录账号再留言。'
    return
  }

  postingMessage.value = true
  messageError.value = ''
  try {
    const response = await fetch('/api/messages', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname: nickname.value,
        content: messageDraft.value,
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '留言发送失败'))
    }

    const payload = await response.json()
    messages.value = [payload, ...messages.value]
    messageDraft.value = ''
  } catch (error) {
    messageError.value = error.message || '留言发送失败'
  } finally {
    postingMessage.value = false
  }
}

async function validateNicknameWithServer(nextNickname) {
  const response = await fetch('/api/nickname/validate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      nickname: nextNickname,
    }),
  })

  if (!response.ok) {
    const message = await readErrorMessage(response, '昵称校验失败，请稍后重试。')
    throw new Error(message)
  }
}

async function loadBossHistory(force = false) {
  if ((bossHistoryLoaded.value || loadingBossHistory.value) && !force) {
    return
  }

  loadingBossHistory.value = true
  bossHistoryError.value = ''

  try {
    const response = await fetch('/api/boss/history')
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('历史 Boss 接口还没生效，请重启后端服务。')
      }
      throw new Error(`历史 Boss 加载失败（${response.status}）`)
    }

    const payload = await response.json()
    bossHistory.value = Array.isArray(payload) ? payload : []
    bossHistoryLoaded.value = true
  } catch (error) {
    if (error instanceof TypeError) {
      bossHistoryError.value = '历史 Boss 接口不可达，请确认后端服务已启动。'
      return
    }

    bossHistoryError.value = error.message || '历史 Boss 加载失败'
  } finally {
    loadingBossHistory.value = false
  }
}

function markUpdated() {
  lastUpdatedAt.value = new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date())
}

function selectHudTab(tab) {
  activeHudTab.value = tab
  if (tab === 'auto') {
    loadAutoClickStatus()
  }
  if (tab === 'info') {
    loadBossHistory()
    loadAnnouncements()
  }
  if (tab === 'messages') {
    loadMessages()
  }
}

function applyState(payload) {
  applyPublicState(payload)
  applyBattleUserState(payload)
  pendingKeys.value = new Set()
  syncing.value = false
  markUpdated()
  maybePromptAnnouncement()
}

function applyPublicState(payload) {
  if (!payload || typeof payload !== 'object') {
    return
  }

  if ('buttons' in payload) {
    buttons.value = Array.isArray(payload.buttons) ? payload.buttons : []
    buttonTotalCount.value = buttons.value.length
    syncAutoClickTarget()
  }
  if ('buttonTotal' in payload) {
    buttonTotalCount.value = Number(payload.buttonTotal ?? buttonTotalCount.value)
  }
  if ('totalVotes' in payload) {
    buttonTotalVotes.value = Number(payload.totalVotes ?? buttonTotalVotes.value)
  }
  if ('leaderboard' in payload) {
    leaderboard.value = Array.isArray(payload.leaderboard) ? payload.leaderboard : []
  }
  const previousBoss = boss.value
  if ('boss' in payload) {
    boss.value = mergeBossState(boss.value, payload.boss)
  }
  if ('bossLeaderboard' in payload) {
    bossLeaderboard.value = Array.isArray(payload.bossLeaderboard) ? payload.bossLeaderboard : []
  }
  if ('announcementVersion' in payload) {
    const nextVersion = String(payload.announcementVersion || '').trim()
    const versionChanged = announcementVersion.value !== nextVersion
    announcementVersion.value = nextVersion
    if (versionChanged) {
      latestAnnouncementLoaded.value = false
    }
    if (!nextVersion) {
      latestAnnouncement.value = null
      latestAnnouncementLoaded.value = true
      writeCachedLatestAnnouncement(null)
      announcementModalOpen.value = false
    } else if (versionChanged || !latestAnnouncementLoaded.value) {
      void loadLatestAnnouncement(versionChanged)
    }
  }
  if (bossResourceVersion(previousBoss) !== bossResourceVersion()) {
    void loadBossResources(true)
  } else if (boss.value?.id && !lastBossResourceVersion) {
    void loadBossResources(true)
  }
  syncing.value = false
  markUpdated()
}

function applyUserState(payload) {
  if (!payload || typeof payload !== 'object') {
    return
  }

  applyBattleUserState(payload)
  applyPlayerProfileState(payload)
  syncing.value = false
  markUpdated()
}

function applyBattleUserState(payload) {
  if (!payload || typeof payload !== 'object') {
    return
  }

  if ('userStats' in payload) {
    userStats.value = payload.userStats ?? null
  }
  if ('myBossStats' in payload) {
    myBossStats.value = payload.myBossStats ?? null
  }
  if ('combatStats' in payload && !profileLoaded.value) {
    combatStats.value = payload.combatStats ?? defaultCombatStats()
  }
  if ('recentRewards' in payload) {
    recentRewards.value = Array.isArray(payload.recentRewards) ? payload.recentRewards : []
  }
  if ('lastReward' in payload) {
    lastReward.value = payload.lastReward ?? null
  }
  syncing.value = false
  markUpdated()
}

function applyPlayerProfileState(payload) {
  if (!payload || typeof payload !== 'object') {
    return
  }

  if ('inventory' in payload) {
    inventory.value = Array.isArray(payload.inventory) ? payload.inventory : []
  }
  if ('loadout' in payload) {
    loadout.value = normalizeLoadout(payload.loadout)
  }
  if ('combatStats' in payload) {
    combatStats.value = payload.combatStats ?? defaultCombatStats()
  }
  if ('gems' in payload) {
    gems.value = Number(payload.gems ?? 0)
  }
}

function applyClickResult(payload) {
  if (!payload || typeof payload !== 'object') {
    return
  }

  if (payload.button?.key) {
    buttons.value = buttons.value.map((button) =>
        button.key === payload.button.key
            ? {...button, ...payload.button}
            : button,
    )
    syncAutoClickTarget()
  }
  buttonTotalVotes.value = Math.max(0, buttonTotalVotes.value + Number(payload.delta || 0))
  const nextClickState = mergeClickFallbackState(
    {
      userStats: userStats.value,
      boss: boss.value,
      bossLeaderboard: bossLeaderboard.value,
      myBossStats: myBossStats.value,
      recentRewards: recentRewards.value,
      lastReward: lastReward.value,
    },
    payload,
  )
  userStats.value = nextClickState.userStats
  boss.value = nextClickState.boss
  bossLeaderboard.value = nextClickState.bossLeaderboard
  myBossStats.value = nextClickState.myBossStats
  recentRewards.value = nextClickState.recentRewards
  lastReward.value = nextClickState.lastReward
  syncing.value = false
  markUpdated()
}

function clearUserRealtimeState() {
  userStats.value = null
  inventory.value = []
  loadout.value = emptyLoadout()
  combatStats.value = defaultCombatStats()
  gems.value = 0
  myBossStats.value = null
  recentRewards.value = []
  lastReward.value = null
}

function clearPendingClicks(key = '') {
  if (!key) {
    pendingClickSources.clear()
    pendingKeys.value = new Set()
    return 'normal'
  }

  const normalizedKey = String(key).trim()
  const nextPending = new Set(pendingKeys.value)
  nextPending.delete(normalizedKey)
  pendingKeys.value = nextPending
  const source = pendingClickSources.get(normalizedKey) || 'normal'
  pendingClickSources.delete(normalizedKey)
  return source
}

function applyRealtimeSnapshot(publicState, userState) {
  applyPublicState(publicState)
  if (userState) {
    applyBattleUserState(userState)
  } else {
    clearUserRealtimeState()
  }
  pendingKeys.value = new Set()
  syncing.value = false
  loading.value = false
  errorMessage.value = ''
  markUpdated()
  maybePromptAnnouncement()
}

function ensureRealtimeTransport() {
  if (realtimeTransport) {
    return realtimeTransport
  }

  realtimeTransport = createRealtimeTransport({
    onSnapshot(publicState, userState) {
      applyRealtimeSnapshot(publicState, userState)
    },
    onPublicDelta(payload) {
      applyPublicState(payload)
      loading.value = false
      errorMessage.value = ''
    },
    onUserDelta(payload) {
      applyBattleUserState(payload)
      loading.value = false
      errorMessage.value = ''
    },
    onClickAck(payload) {
      const key = payload?.button?.key || ''
      if (payload?.critical && key) {
        triggerCriticalBurst(key, payload.delta)
      }
      const source = clearPendingClicks(key)
      if (key) {
        autoClickTargetKey.value = key
        if (autoClickEnabled.value) {
          void syncAutoClickTargetOnServer(key).catch((error) => {
            errorMessage.value = error.message || '挂机目标更新失败，请稍后重试。'
          })
        }
      }
      applyClickResult(payload)
      errorMessage.value = ''
    },
    onTransportState(nextState) {
      liveConnected.value = nextState.connected
      if (nextState.connected) {
        syncing.value = false
      }
    },
    onTransportError(message) {
      clearPendingClicks()
      if (message) {
        errorMessage.value = message
      }
    },
  })

  return realtimeTransport
}

function connectRealtime(nextNickname = nickname.value) {
  if (buttons.value.length === 0) {
    loading.value = true
  }
  syncing.value = true

  if (!realtimeTransport) {
    ensureRealtimeTransport().connect({nickname: nextNickname})
    return
  }

  realtimeTransport.reconnect({nickname: nextNickname})
}

function clearCriticalBurst(key) {
  const timer = burstTimers.get(key)
  if (timer) {
    window.clearTimeout(timer)
    burstTimers.delete(key)
  }

  if (!criticalBursts.value[key]) {
    return
  }

  const nextBursts = {...criticalBursts.value}
  delete nextBursts[key]
  criticalBursts.value = nextBursts
}

function triggerCriticalBurst(key, delta) {
  clearCriticalBurst(key)

  criticalBursts.value = {
    ...criticalBursts.value,
    [key]: {
      label: `暴击伤害 ${delta}`,
      nonce: `${key}-${Date.now()}`,
    },
  }

  burstTimers.set(
    key,
    window.setTimeout(() => {
      clearCriticalBurst(key)
    }, 1600),
  )
}

function handlePressStart(key, event) {
  clickBehaviorTracker.handlePressStart(key, event)
}

function handlePressEnd(key, event) {
  clickBehaviorTracker.handlePressEnd(key, event)
}

function handlePressCancel(key) {
  clickBehaviorTracker.handlePressCancel(key)
}

async function ensureFingerprintHash() {
  if (fingerprintHash.value) {
    return fingerprintHash.value
  }
  if (!fingerprintPromise) {
    fingerprintPromise = collectFingerprintHash()
  }
  fingerprintHash.value = await fingerprintPromise
  return fingerprintHash.value
}

function consumeClickBehavior(key) {
  const behavior = clickBehaviorTracker.consume(key)
  if (!behavior) {
    throw new Error('操作采样失败，请重试。')
  }
  return behavior
}

function currentNicknameQuery() {
  return ''
}

function syncAutoClickTarget() {
  // 挂机目标来自完整按钮列表，保持服务端状态即可。
}

function applyAutoClickStatus(payload, options = {}) {
  autoClickEnabled.value = Boolean(payload?.active)
  if (payload?.buttonKey) {
    autoClickTargetKey.value = payload.buttonKey
    return
  }
  if (options.clearTargetWhenMissing) {
    autoClickTargetKey.value = ''
  }
}

function clearAutoClickLocalState() {
  autoClickEnabled.value = false
  autoClickTargetKey.value = ''
}

async function loadAutoClickStatus() {
  if (!nickname.value) {
    clearAutoClickLocalState()
    return
  }

  try {
    const response = await fetch('/api/auto-click')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '挂机状态加载失败'))
    }
    applyAutoClickStatus(await response.json(), {clearTargetWhenMissing: true})
  } catch {
    autoClickEnabled.value = false
  }
}

async function syncAutoClickTargetOnServer(key) {
  if (!nickname.value || !key) {
    return
  }

  const response = await fetch('/api/auto-click/start', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({slug: key}),
  })
  if (!response.ok) {
    throw new Error(await readErrorMessage(response, '挂机目标更新失败，请稍后重试。'))
  }
  applyAutoClickStatus(await response.json(), {clearTargetWhenMissing: false})
}

async function startAutoClick() {
  if (!canStartAutoClick.value) {
    return
  }

  errorMessage.value = ''

  try {
    await syncAutoClickTargetOnServer(autoClickTargetKey.value)
  } catch (error) {
    errorMessage.value = error.message || '挂机开启失败，请稍后重试。'
  }
}

async function stopAutoClick() {
  if (!nickname.value) {
    autoClickEnabled.value = false
    return
  }

  errorMessage.value = ''
  try {
    const response = await fetch('/api/auto-click/stop', {
      method: 'POST',
    })
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '挂机关闭失败，请稍后重试。'))
    }
    autoClickEnabled.value = false
  } catch (error) {
    errorMessage.value = error.message || '挂机关闭失败，请稍后重试。'
  }
}

async function toggleAutoClick() {
  if (autoClickEnabled.value) {
    await stopAutoClick()
    return
  }

  await startAutoClick()
}

async function requestClickTicket(key) {
  const nextFingerprintHash = await ensureFingerprintHash()
  try {
    const realtimeTicket = await ensureRealtimeTransport().requestClickTicket(key, nextFingerprintHash)
    if (realtimeTicket?.ticket && realtimeTicket?.challengeNonce) {
      return {
        ticket: realtimeTicket.ticket,
        challengeNonce: realtimeTicket.challengeNonce,
        fingerprintHash: nextFingerprintHash,
      }
    }
  } catch {
    // ws 票据申请失败时退回 HTTP 兜底，避免点击体验被主链路抖动放大。
  }

  const response = await fetch('/api/click-tickets', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      slug: key,
      fingerprintHash: nextFingerprintHash,
    }),
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, '操作已过期，请重试。'))
  }

  const payload = await response.json()
  const ticket = String(payload?.ticket || '').trim()
  const challengeNonce = String(payload?.challengeNonce || '').trim()
  if (!ticket || !challengeNonce) {
    throw new Error('操作已过期，请重试。')
  }
  return {
    ticket,
    challengeNonce,
    fingerprintHash: nextFingerprintHash,
  }
}

async function loadState() {
  loading.value = true
  syncing.value = true

  try {
    const response = await fetch(`/api/buttons${currentNicknameQuery()}`)
    if (!response.ok) {
      throw new Error('按钮列表加载失败')
    }

    const data = await response.json()
    applyState(data)
  } catch (error) {
    errorMessage.value = error.message || '加载失败，请稍后重试。'
  } finally {
    loading.value = false
    syncing.value = false
  }
}

async function loadPlayerProfile(force = false) {
  if (!nickname.value) {
    profileLoaded.value = false
    profileNotice.value = '登录后进入资料页会刷新角色资料。'
    return
  }
  if (profileLoading.value || (profileLoaded.value && !force)) {
    return
  }

  profileLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetch('/api/player/profile')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '资料加载失败，请稍后重试。'))
    }

    const data = await response.json()
    applyBattleUserState(data)
    applyPlayerProfileState(data)
    profileLoaded.value = true
    profileNotice.value = '进入本页已刷新资料。'
  } catch (error) {
    errorMessage.value = error.message || '资料加载失败，请稍后重试。'
  } finally {
    profileLoading.value = false
  }
}

async function refreshProfileAfterMutation(data) {
  applyBattleUserState(data)
  if (currentPublicPage.value === 'profile') {
    await loadPlayerProfile(true)
    return
  }
  profileLoaded.value = false
}

async function clickButton(key, options = {}) {
  if (!nickname.value || pendingKeys.value.has(key)) {
    return
  }

  const nextPending = new Set(pendingKeys.value)
  nextPending.add(key)
  pendingKeys.value = nextPending
  pendingClickSources.set(key, options.source || 'normal')
  errorMessage.value = ''

  try {
    const ticketInfo = await requestClickTicket(key)
    const behavior = consumeClickBehavior(key)
    behavior.fingerprintHash = ticketInfo.fingerprintHash
    behavior.fingerprintProof = await buildFingerprintProof({
      fingerprintHash: ticketInfo.fingerprintHash,
      ticket: ticketInfo.ticket,
      challengeNonce: ticketInfo.challengeNonce,
    })

    if (ensureRealtimeTransport().sendClick(key, ticketInfo.ticket, behavior)) {
      return
    }

    const response = await fetch(`/api/buttons/${encodeURIComponent(key)}/click`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(buildClickRequestBody(ticketInfo.ticket, liveConnected.value, behavior)),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '点击失败，请稍后重试。'))
    }

    const data = await response.json()
    if (data.critical) {
      triggerCriticalBurst(key, data.delta)
    }
    autoClickTargetKey.value = key
    if (autoClickEnabled.value) {
      await syncAutoClickTargetOnServer(key)
    }
    clearPendingClicks(key)
    applyClickResult(data)
    errorMessage.value = ''
  } catch (error) {
    clearPendingClicks(key)
    errorMessage.value = error.message || '点击失败，请稍后重试。'
  }
}

async function toggleItemEquip(itemId, equipped) {
  if (!nickname.value || !itemId) {
    return
  }

  const action = equipped ? 'unequip' : 'equip'
  actioningItemId.value = itemId
  errorMessage.value = ''

  try {
    const response = await fetch(`/api/equipment/${encodeURIComponent(itemId)}/${action}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({nickname: nickname.value}),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '装备操作失败，请稍后重试。'))
    }

    const data = await response.json()
    await refreshProfileAfterMutation(data)
  } catch (error) {
    errorMessage.value = error.message || '装备操作失败，请稍后重试。'
  } finally {
    actioningItemId.value = ''
  }
}

async function submitNickname() {
  const nextNickname = normalizeNickname(nicknameDraft.value)
  if (!nextNickname) {
    errorMessage.value = '先填一个昵称。'
    return
  }
  if (!passwordDraft.value.trim()) {
    errorMessage.value = '再设一个密码。'
    return
  }

  errorMessage.value = ''

  try {
    const response = await fetch('/api/player/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname: nextNickname,
        password: passwordDraft.value,
      }),
    })
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '登录失败，请稍后重试。'))
    }

    const payload = await response.json()
    const resolvedNickname = normalizeNickname(payload?.nickname || nextNickname)

    nickname.value = resolvedNickname
    nicknameDraft.value = resolvedNickname
    passwordDraft.value = ''
    await loadAutoClickStatus()
    if (currentPublicPage.value === 'profile') {
      await loadPlayerProfile(true)
    }
    connectRealtime(resolvedNickname)
  } catch (error) {
    errorMessage.value = error.message || '登录失败，请稍后重试。'
  }
}

async function resetNickname() {
  try {
    await fetch('/api/player/auth/logout', {
      method: 'POST',
    })
  } catch {
    // 忽略异常，继续清理本地状态。
  }

  clearPlayerSessionState()
  connectRealtime('')
}

function clearPlayerSessionState() {
  nickname.value = ''
  nicknameDraft.value = ''
  passwordDraft.value = ''
  clearUserRealtimeState()
  clearAutoClickLocalState()
  clearPendingClicks()
  profileLoaded.value = false
  profileNotice.value = ''
}

async function loadPlayerSession() {
  try {
    const response = await fetch('/api/player/auth/session')
    if (!response.ok) {
      clearPlayerSessionState()
      return
    }

    const payload = await response.json()
    const resolvedNickname = normalizeNickname(payload?.nickname || '')
    if (!resolvedNickname) {
      clearPlayerSessionState()
      return
    }

    nickname.value = resolvedNickname
    nicknameDraft.value = resolvedNickname
    await loadAutoClickStatus()
  } catch {
    clearPlayerSessionState()
  }
}

function registerPublicPageLifecycle() {
  let onlineCountTimer

  async function fetchOnlineCount() {
    try {
      const response = await fetch('/api/online-count')
      if (response.ok) {
        const data = await response.json()
        onlineCount.value = data.count
      }
    } catch {
      // 静默失败，下次轮询重试
    }
  }

  onMounted(async () => {
    restoreCachedLatestAnnouncement()
    window.addEventListener('popstate', handlePublicRouteChange)
    await loadPlayerSession()
    await loadState()
    await activatePublicPage(currentPublicPage.value)
    connectRealtime(nickname.value)
    fetchOnlineCount()
    onlineCountTimer = setInterval(fetchOnlineCount, 30000)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('popstate', handlePublicRouteChange)
    clickBehaviorTracker.clear()
    realtimeTransport?.close()
    burstTimers.forEach((timer) => window.clearTimeout(timer))
    burstTimers.clear()
    if (onlineCountTimer) clearInterval(onlineCountTimer)
  })
}

export function usePublicPageState() {
  return {
    ANNOUNCEMENT_READ_KEY,
    ANNOUNCEMENT_CACHE_KEY,
    AUTO_CLICK_RATE_LABEL,
    EQUIPMENT_ENHANCE_COST,
    GROWTH_FORMULA_TEXT,
    publicPages,
    buttons,
    buttonTotalCount,
    buttonTotalVotes,
    leaderboard,
    boss,
    bossLeaderboard,
    bossLoot,
    announcementVersion,
    latestAnnouncement,
    announcements,
    myBossStats,
    inventory,
  loadout,
  loadoutSlots,
    combatStats,
    recentRewards,
    lastReward,
    userStats,
    nickname,
    nicknameDraft,
    passwordDraft,
    loading,
    syncing,
    errorMessage,
    pendingKeys,
    actioningItemId,
    activeHudTab,
    lastUpdatedAt,
    liveConnected,
    criticalBursts,
    bossHistory,
    bossHistoryQuery,
    loadingBossHistory,
    bossHistoryLoaded,
    bossHistoryError,
    loadingAnnouncements,
    announcementsLoaded,
    announcementError,
    loadingBossResources,
    latestAnnouncementLoaded,
    announcementModalOpen,
    messages,
    messageNextCursor,
    loadingMessages,
    postingMessage,
    messageDraft,
    messageError,
    autoClickEnabled,
    autoClickTargetKey,
    gems,
    fingerprintHash,
    currentPublicPage,
    profileLoading,
    profileLoaded,
    profileNotice,
    lastBossResourceVersion,
    burstTimers,
    pendingClickSources,
    clickBehaviorTracker,
    buttonCount,
    totalVotes,
    displayedButtons,
    syncLabel,
    onlineCount,
    isLoggedIn,
    myClicks,
    myRank,
    myBossDamage,
    myBossRank,
    effectiveIncrement,
    normalDamage,
    criticalDamage,
    autoClickTargetButton,
    autoClickTargetLabel,
    canStartAutoClick,
    autoClickStatus,
    bossStatusLabel,
    bossProgress,
    equippedItems,
    displayedRecentRewards,
    recentRewardTitle,
    recentRewardNote,
    filteredBossHistory,
    emptyLoadout,
    defaultCombatStats,
    formatDropRate,
    formatRarityLabel,
    formatItemStats,
    formatItemStatLines,
    equipmentNameParts,
    equipmentNameClass,
    normalizeNickname,
    resolvePublicPage,
    navigatePublicPage,
    activatePublicPage,
    handlePublicRouteChange,
    formatBossTime,
    topBossDamage,
    formatTime,
    formatNumber,
    formatStatWithDelta,
    formatPercentWithDelta,
    dotIndexes,
    readErrorMessage,
    bossResourceVersion,
    readCachedLatestAnnouncement,
    writeCachedLatestAnnouncement,
    restoreCachedLatestAnnouncement,
    maybePromptAnnouncement,
    closeAnnouncementModal,
    loadBossResources,
    loadLatestAnnouncement,
    loadAnnouncements,
    loadMessages,
    submitMessage,
    validateNicknameWithServer,
    loadBossHistory,
    markUpdated,
    selectHudTab,
    applyState,
    applyPublicState,
    applyUserState,
    applyBattleUserState,
    applyPlayerProfileState,
    applyClickResult,
    clearUserRealtimeState,
    clearPendingClicks,
    applyRealtimeSnapshot,
    ensureRealtimeTransport,
    connectRealtime,
    clearCriticalBurst,
    triggerCriticalBurst,
    handlePressStart,
    handlePressEnd,
    handlePressCancel,
    ensureFingerprintHash,
    consumeClickBehavior,
    currentNicknameQuery,
    syncAutoClickTarget,
    applyAutoClickStatus,
    clearAutoClickLocalState,
    loadAutoClickStatus,
    syncAutoClickTargetOnServer,
    startAutoClick,
    stopAutoClick,
    toggleAutoClick,
    requestClickTicket,
    loadState,
    loadPlayerProfile,
    refreshProfileAfterMutation,
    clickButton,
    toggleItemEquip,
    submitNickname,
    resetNickname,
    clearPlayerSessionState,
    loadPlayerSession,
    registerPublicPageLifecycle,
  }
}
