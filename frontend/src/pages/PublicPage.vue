<script setup>
import {computed, onBeforeUnmount, onMounted, ref} from 'vue'

import {mergeBossState} from '../utils/bossState'
import {collectButtonTags, filterAndSortButtons, formatDropRate} from '../utils/buttonBoard'
import {buildClickRequestBody, mergeClickFallbackState} from '../utils/clickResponse'
import {
  buildCosmeticCollections,
  canEquipCosmeticSelection,
  cosmeticStatusText,
  createEmptyCosmeticLoadout,
  normalizeCosmeticLoadout,
  resolveCosmeticEffectConfig,
  salvageableCount,
  summarizeEquippedCosmetics,
} from '../utils/cosmetics'
import { formatRarityLabel, getRarityClassName, splitEquipmentName } from '../utils/rarity'
import {createRealtimeTransport} from '../utils/realtimeTransport'
import {resolveStarlightRefreshPlan} from '../utils/starlightRefresh'
import { buildFingerprintProof, collectFingerprintHash, createClickBehaviorTracker } from '../utils/manualClickSignals'

const ANNOUNCEMENT_READ_KEY = 'vote-wall-announcement-read'
const ANNOUNCEMENT_CACHE_KEY = 'vote-wall-announcement-cache'
const AUTO_CLICK_RATE_LABEL = '每秒固定 3 次'
const EQUIPMENT_ENHANCE_COST = 10
const HERO_AWAKEN_COST = 15
const GROWTH_FORMULA_TEXT = '点击 / 暴击单次成长 = ceil((当前点击 + 当前暴击 + 当前暴击率) / 4)，至少 +1'
const HERO_GROWTH_FORMULA_TEXT = '点击 / 暴击单次成长 = ceil((当前点击 + 当前暴击 + 当前暴击率 + 最终伤害提升百分比) / 4)，至少 +1'


const buttons = ref([])
const firstPageButtons = ref([])
const buttonPage = ref(1)
const buttonPageSize = ref(9)
const buttonTotalPages = ref(1)
const buttonTotalCount = ref(0)
const buttonTotalVotes = ref(0)
const leaderboard = ref([])
const boss = ref(null)
const bossLeaderboard = ref([])
const bossLoot = ref([])
const bossHeroLoot = ref([])
const starlight = ref({activeKeys: [], startedAt: 0, endsAt: 0})
const announcementVersion = ref('')
const latestAnnouncement = ref(null)
const announcements = ref([])
const myBossStats = ref(null)
const inventory = ref([])
const heroes = ref([])
const activeHero = ref(null)
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
const activeHudTab = ref('inventory')
const lastUpdatedAt = ref('')
const liveConnected = ref(false)
const criticalBursts = ref({})
const bossHistory = ref([])
const bossHistoryQuery = ref('')
const loadingBossHistory = ref(false)
const bossHistoryLoaded = ref(false)
const bossHistoryError = ref('')
const selectedButtonTag = ref('全部')
const buttonSearch = ref('')
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
const ownedCosmetics = ref([])
const equippedCosmetics = ref(createEmptyCosmeticLoadout())
const cosmeticDraft = ref(createEmptyCosmeticLoadout())
const shopCatalog = ref([])
const lastForgeResult = ref(null)
const cosmeticBursts = ref({})
const fingerprintHash = ref('')

let realtimeTransport
let starlightTimer = 0
let lastExpiredStarlightEndsAt = 0
let lastBossResourceVersion = ''
const burstTimers = new Map()
const cosmeticTimers = new Map()
const pendingClickSources = new Map()
const clickBehaviorTracker = createClickBehaviorTracker()
let fingerprintPromise

const buttonCount = computed(() => buttonTotalCount.value || buttons.value.length)
const totalVotes = computed(() =>
    buttonTotalVotes.value || buttons.value.reduce((total, button) => total + button.count, 0),
)
const buttonTags = computed(() => ['全部', ...collectButtonTags(buttons.value)])
const activeStarlightKeys = computed(() => starlight.value?.activeKeys ?? [])
const displayedButtons = computed(() =>
    filterAndSortButtons(buttons.value, {
      selectedTag: selectedButtonTag.value,
      query: buttonSearch.value,
      activeStarlightKeys: activeStarlightKeys.value,
    }),
)
const syncLabel = computed(() => {
  if (syncing.value) {
    return '同步中'
  }

  return liveConnected.value ? '全员在线' : '正在重连'
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
    return '先登录账号，再选择按钮并开启官方挂机。'
  }
  if (!autoClickTargetKey.value) {
    return '先手动点一次按钮锁定目标，再开启官方挂机。'
  }
  if (autoClickEnabled.value) {
    return `官方挂机正在服务端托管 ${autoClickTargetLabel.value}，你手动点别的按钮后会立刻切换目标。`
  }

  return `已锁定 ${autoClickTargetLabel.value}，开启后会按 ${AUTO_CLICK_RATE_LABEL} 在服务端持续结算。`
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
const equippedItems = computed(() => [loadout.value.weapon, loadout.value.armor, loadout.value.accessory].filter(Boolean))
const heroCount = computed(() => heroes.value.length)
const cosmeticCollections = computed(() => buildCosmeticCollections(shopCatalog.value))
const selectedCosmeticLoadout = computed(() => normalizeCosmeticLoadout(cosmeticDraft.value))
const selectedCosmeticSummary = computed(() =>
  summarizeEquippedCosmetics(shopCatalog.value, selectedCosmeticLoadout.value),
)
const equippedCosmeticSummary = computed(() =>
  summarizeEquippedCosmetics(shopCatalog.value, equippedCosmetics.value),
)
const canApplyCosmeticSelection = computed(() =>
  isLoggedIn.value && canEquipCosmeticSelection(shopCatalog.value, selectedCosmeticLoadout.value),
)
const previewEffectConfig = computed(() =>
  resolveCosmeticEffectConfig(shopCatalog.value, selectedCosmeticLoadout.value, {
    mode: 'normal',
    starlight: false,
  }),
)
const previewDots = computed(() => dotIndexes(previewEffectConfig.value.particleCount || 6))
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
        // 第一个：装备，第二个：小小英雄
        if (index === 0) return `${reward.itemName}（装备）`
        if (index === 1) return `${reward.itemName}（小小英雄）`
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
  return {
    weapon: null,
    armor: null,
    accessory: null,
  }
}

function defaultCombatStats() {
  return {
    baseIncrement: 1,
    bonusClicks: 0,
    effectiveIncrement: 1,
    normalDamage: 1,
    criticalDamage: 1,
    criticalChancePercent: 0,
    criticalCount: 1,
  }
}

function formatItemStats(item) {
  return [
    formatStatWithDelta('点击', item?.bonusClicks, item?.bonusClicksDelta),
    formatPercentWithDelta('暴击率', item?.bonusCriticalChancePercent, item?.bonusCriticalChancePercentDelta),
    formatStatWithDelta('暴击', item?.bonusCriticalCount, item?.bonusCriticalCountDelta),
  ].join(' ')
}

function formatItemStatLines(item) {
  return [
    formatStatWithDelta('点击', item?.bonusClicks, item?.bonusClicksDelta),
    formatPercentWithDelta('暴击率', item?.bonusCriticalChancePercent, item?.bonusCriticalChancePercentDelta),
    formatStatWithDelta('暴击', item?.bonusCriticalCount, item?.bonusCriticalCountDelta),
  ]
}

function equipmentNameParts(item) {
  return splitEquipmentName(item?.itemName || item?.name || item?.itemId || '')
}

function equipmentNameClass(item) {
  return getRarityClassName(item?.rarity)
}

function formatEnhanceCap(cap) {
  return Number(cap) > 0 ? `可强化 ${Number(cap)} 次` : '可强化∞次'
}

function formatAwakenCap(cap) {
  return Number(cap) > 0 ? `可觉醒 ${Number(cap)} 次` : '可觉醒∞次'
}

function formatHeroTrait(hero) {
  const effects = Array.isArray(hero?.effects) ? hero.effects : []
  if (effects.length === 0) {
    return '被动：暂无'
  }

  return `被动：${effects.map((effect) => formatHeroEffect(effect)).join(' / ')}`
}

function heroImageAlt(hero) {
  return hero?.imageAlt || hero?.heroName || hero?.name || hero?.heroId || '英雄头像'
}

function normalizeNickname(value) {
  return value.trim()
}

function isStarlightButton(key) {
  return activeStarlightKeys.value.includes(key)
}

function clearStarlightTimer() {
  if (starlightTimer) {
    window.clearTimeout(starlightTimer)
    starlightTimer = 0
  }
}

function scheduleStarlightRefresh() {
  clearStarlightTimer()
  const plan = resolveStarlightRefreshPlan({
    endsAtSeconds: starlight.value?.endsAt,
    lastExpiredEndsAtSeconds: lastExpiredStarlightEndsAt,
  })
  lastExpiredStarlightEndsAt = plan.expiredRetryEndsAtSeconds
  if (plan.delayMs === null) {
    return
  }

  starlightTimer = window.setTimeout(() => {
    if (!realtimeTransport?.requestSync?.()) {
      void loadState()
    }
  }, plan.delayMs)
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

function formatHeroEffect(effect) {
  switch (effect?.type) {
    case 'bonus_clicks':
      return `额外点击 +${formatNumber(effect?.value)}`
    case 'critical_chance_percent':
      return `暴击率 +${formatNumber(effect?.value)}%`
    case 'critical_count_bonus':
      return `暴击额外 +${formatNumber(effect?.value)}`
    case 'final_damage_percent':
      return `最终伤害 +${formatNumber(effect?.value)}%`
    default:
      return effect?.displayName || effect?.type || '未知效果'
  }
}

function salvageableEquipmentCount(item) {
  return salvageableCount(item)
}

function salvageableHeroCount(hero) {
  return salvageableCount(hero, hero?.active)
}

function equipmentEnhanceHint(item) {
  const currentLevel = Number(item?.enhanceLevel || 0)
  const cap = Number(item?.enhanceCap || 0)

  if (!isLoggedIn.value) {
    return '登录后才能消耗原石强化。'
  }
  if (cap > 0 && currentLevel >= cap) {
    return '已达模板上限，本件装备不能继续强化。'
  }
  if (gems.value < EQUIPMENT_ENHANCE_COST) {
    return `原石不足，还差 ${EQUIPMENT_ENHANCE_COST - gems.value}。`
  }

  return ``
}

function heroAwakenHint(hero) {
  const currentLevel = Number(hero?.awakenLevel || 0)
  const cap = Number(hero?.awakenCap || 0)

  if (!isLoggedIn.value) {
    return '登录后才能消耗原石觉醒。'
  }
  if (cap > 0 && currentLevel >= cap) {
    return '已达模板上限，本位英雄不能继续觉醒。'
  }
  if (gems.value < HERO_AWAKEN_COST) {
    return `原石不足，还差 ${HERO_AWAKEN_COST - gems.value}。`
  }

  return ``
}

function dotIndexes(count) {
  const normalized = Math.max(0, Math.min(Number(count) || 0, 6))
  return Array.from({length: normalized}, (_, index) => index)
}

function cosmeticModeClasses(effect) {
  return {
    'cosmetic-mode--auto': effect?.mode === 'auto',
    'cosmetic-mode--suppressed': Boolean(effect?.suppressed),
  }
}

function syncCosmeticDraft(loadout) {
  cosmeticDraft.value = normalizeCosmeticLoadout(loadout)
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

function normalizePageNumber(value, fallback = 1) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return Math.floor(parsed)
}

function updateCurrentPageButtons(nextButtons) {
  buttons.value = Array.isArray(nextButtons) ? nextButtons : []
  if (!buttonTags.value.includes(selectedButtonTag.value)) {
    selectedButtonTag.value = '全部'
  }
  syncAutoClickTarget()
}

function applyButtonPagePayload(payload, options = {}) {
  if (!payload || typeof payload !== 'object') {
    return
  }

  const nextButtons = Array.isArray(payload.buttons)
      ? payload.buttons
      : Array.isArray(payload.items)
          ? payload.items
          : []
  const nextPage = normalizePageNumber(payload.buttonPage ?? payload.page, options.defaultPage ?? 1)
  const nextPageSize = normalizePageNumber(payload.buttonPageSize ?? payload.pageSize, buttonPageSize.value || 9)
  const nextTotalPages = normalizePageNumber(payload.buttonTotalPages ?? payload.totalPages, buttonTotalPages.value || 1)
  const nextTotal = Number(payload.buttonTotal ?? payload.total ?? nextButtons.length)
  const nextTotalVotes = Number(payload.totalVotes ?? buttonTotalVotes.value)

  buttonPageSize.value = nextPageSize
  buttonTotalPages.value = Math.max(1, nextTotalPages)
  buttonTotalCount.value = Number.isFinite(nextTotal) ? nextTotal : nextButtons.length
  buttonTotalVotes.value = Number.isFinite(nextTotalVotes) ? nextTotalVotes : buttonTotalVotes.value

  if (nextPage === 1) {
    firstPageButtons.value = nextButtons
  }

  if (options.preserveCurrentPage && buttonPage.value !== nextPage) {
    return
  }

  buttonPage.value = nextPage
  updateCurrentPageButtons(nextButtons)
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
    bossHeroLoot.value = []
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
    bossHeroLoot.value = Array.isArray(payload?.bossHeroLoot) ? payload.bossHeroLoot : []
    lastBossResourceVersion = currentVersion
  } catch {
    if (force) {
      bossLoot.value = []
      bossHeroLoot.value = []
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
  applyUserState(payload)
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
    applyButtonPagePayload(payload, { defaultPage: 1, preserveCurrentPage: true })
    if (buttonPage.value > buttonTotalPages.value) {
      void loadButtonPage(buttonTotalPages.value)
    }
  } else {
    if ('buttonPage' in payload) {
      buttonPage.value = normalizePageNumber(payload.buttonPage, buttonPage.value)
    }
    if ('buttonPageSize' in payload) {
      buttonPageSize.value = normalizePageNumber(payload.buttonPageSize, buttonPageSize.value || 9)
    }
    if ('buttonTotalPages' in payload) {
      buttonTotalPages.value = normalizePageNumber(payload.buttonTotalPages, buttonTotalPages.value || 1)
    }
    if ('buttonTotal' in payload) {
      buttonTotalCount.value = Number(payload.buttonTotal ?? buttonTotalCount.value)
    }
    if ('totalVotes' in payload) {
      buttonTotalVotes.value = Number(payload.totalVotes ?? buttonTotalVotes.value)
    }
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
  if ('starlight' in payload) {
    starlight.value = payload.starlight ?? {activeKeys: [], startedAt: 0, endsAt: 0}
    scheduleStarlightRefresh()
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

  if ('userStats' in payload) {
    userStats.value = payload.userStats ?? null
  }
  if ('myBossStats' in payload) {
    myBossStats.value = payload.myBossStats ?? null
  }
  if ('inventory' in payload) {
    inventory.value = Array.isArray(payload.inventory) ? payload.inventory : []
  }
  if ('heroes' in payload) {
    heroes.value = Array.isArray(payload.heroes) ? payload.heroes : []
  }
  if ('activeHero' in payload) {
    activeHero.value = payload.activeHero ?? null
  }
  if ('loadout' in payload) {
    loadout.value = payload.loadout ?? emptyLoadout()
  }
  if ('combatStats' in payload) {
    combatStats.value = payload.combatStats ?? defaultCombatStats()
  }
  if ('gems' in payload) {
    gems.value = Number(payload.gems ?? 0)
  }
  if ('ownedCosmetics' in payload) {
    ownedCosmetics.value = Array.isArray(payload.ownedCosmetics) ? payload.ownedCosmetics : []
  }
  if ('equippedCosmetics' in payload) {
    equippedCosmetics.value = normalizeCosmeticLoadout(payload.equippedCosmetics)
    syncCosmeticDraft(payload.equippedCosmetics)
  }
  if ('lastForgeResult' in payload) {
    lastForgeResult.value = payload.lastForgeResult ?? null
  }
  if ('shopCatalog' in payload) {
    shopCatalog.value = Array.isArray(payload.shopCatalog) ? payload.shopCatalog : []
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
    firstPageButtons.value = firstPageButtons.value.map((button) =>
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
  heroes.value = []
  activeHero.value = null
  loadout.value = emptyLoadout()
  combatStats.value = defaultCombatStats()
  gems.value = 0
  ownedCosmetics.value = []
  equippedCosmetics.value = createEmptyCosmeticLoadout()
  syncCosmeticDraft(createEmptyCosmeticLoadout())
  shopCatalog.value = []
  lastForgeResult.value = null
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
    applyUserState(userState)
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
      applyUserState(payload)
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
        triggerCosmeticBurst(key, {mode: source})
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

function clearCosmeticBurst(key) {
  const timer = cosmeticTimers.get(key)
  if (timer) {
    window.clearTimeout(timer)
    cosmeticTimers.delete(key)
  }

  if (!cosmeticBursts.value[key]) {
    return
  }

  const nextBursts = {...cosmeticBursts.value}
  delete nextBursts[key]
  cosmeticBursts.value = nextBursts
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

function triggerCosmeticBurst(key, options = {}) {
  const effect = resolveCosmeticEffectConfig(shopCatalog.value, equippedCosmetics.value, {
    mode: options.mode === 'auto' ? 'auto' : 'normal',
    starlight: isStarlightButton(key),
  })
  if (!effect.trailTheme && !effect.impactTheme) {
    return
  }

  clearCosmeticBurst(key)
  cosmeticBursts.value = {
    ...cosmeticBursts.value,
    [key]: {
      ...effect,
      nonce: `${key}-${Date.now()}`,
      dots: dotIndexes(effect.particleCount),
    },
  }

  cosmeticTimers.set(
    key,
    window.setTimeout(() => {
      clearCosmeticBurst(key)
    }, Number(effect.durationMs || 900) + 240),
  )
}

function currentNicknameQuery() {
  return ''
}

async function loadButtonPage(page) {
  const nextPage = normalizePageNumber(page, 1)
  if (nextPage === 1 && firstPageButtons.value.length > 0) {
    buttonPage.value = 1
    updateCurrentPageButtons(firstPageButtons.value)
    return
  }

  syncing.value = true
  try {
    const response = await fetch(`/api/buttons/pages?page=${nextPage}&pageSize=${buttonPageSize.value || 9}`)
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '按钮分页加载失败'))
    }
    const payload = await response.json()
    applyButtonPagePayload(payload, { defaultPage: nextPage })
  } catch (error) {
    errorMessage.value = error.message || '按钮分页加载失败'
  } finally {
    syncing.value = false
  }
}

function syncAutoClickTarget() {
  // 挂机目标允许不在当前页显示，保留服务端状态即可。
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
    triggerCosmeticBurst(key, {mode: clearPendingClicks(key)})
    applyClickResult(data)
    errorMessage.value = ''
  } catch (error) {
    clearPendingClicks(key)
    errorMessage.value = error.message || '点击失败，请稍后重试。'
  }
}

async function postEquipmentAction(itemId, action, extraBody = {}) {
  if (!nickname.value || !itemId) {
    return
  }

  actioningItemId.value = itemId
  errorMessage.value = ''

  try {
    const response = await fetch(`/api/equipment/${encodeURIComponent(itemId)}/${action}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname: nickname.value,
        ...extraBody,
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '装备操作失败，请稍后重试。'))
    }

    const data = await response.json()
    applyState(data)
  } catch (error) {
    errorMessage.value = error.message || '装备操作失败，请稍后重试。'
  } finally {
    actioningItemId.value = ''
  }
}

async function postHeroAction(heroId, action, extraBody = {}) {
  if (!nickname.value || !heroId) {
    return
  }

  actioningItemId.value = heroId
  errorMessage.value = ''

  try {
    const response = await fetch(`/api/heroes/${encodeURIComponent(heroId)}/${action}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname: nickname.value,
        ...extraBody,
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '英雄操作失败，请稍后重试。'))
    }

    const data = await response.json()
    applyState(data)
  } catch (error) {
    errorMessage.value = error.message || '英雄操作失败，请稍后重试。'
  } finally {
    actioningItemId.value = ''
  }
}

async function salvageEquipment(item) {
  const quantity = salvageableEquipmentCount(item)
  if (!quantity) {
    return
  }

  await postEquipmentAction(item.itemId, 'salvage', {quantity})
}

async function enhanceEquipment(item) {
  await postEquipmentAction(item.itemId, 'enhance')
}

async function salvageHero(hero) {
  const quantity = salvageableHeroCount(hero)
  if (!quantity) {
    return
  }

  await postHeroAction(hero.heroId, 'salvage', {quantity})
}

async function awakenHero(hero) {
  await postHeroAction(hero.heroId, 'awaken')
}

async function purchaseCosmetic(item) {
  if (!nickname.value || !item?.cosmeticId) {
    return
  }

  actioningItemId.value = item.cosmeticId
  errorMessage.value = ''

  try {
    const response = await fetch(`/api/shop/cosmetics/${encodeURIComponent(item.cosmeticId)}/purchase`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname: nickname.value,
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '外观购买失败，请稍后重试。'))
    }

    const data = await response.json()
    applyState(data)
  } catch (error) {
    errorMessage.value = error.message || '外观购买失败，请稍后重试。'
  } finally {
    actioningItemId.value = ''
  }
}

function selectCosmeticItem(item) {
  if (!item?.owned) {
    return
  }

  if (item.type === 'trail') {
    cosmeticDraft.value = {
      ...selectedCosmeticLoadout.value,
      trailId: item.cosmeticId,
    }
    return
  }

  if (item.type === 'impact') {
    cosmeticDraft.value = {
      ...selectedCosmeticLoadout.value,
      impactId: item.cosmeticId,
    }
  }
}

async function equipSelectedCosmetics() {
  if (!nickname.value || !canApplyCosmeticSelection.value) {
    return
  }

  actioningItemId.value = 'cosmetic-loadout'
  errorMessage.value = ''

  try {
    const response = await fetch('/api/shop/cosmetics/equip', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname: nickname.value,
        trailId: selectedCosmeticLoadout.value.trailId,
        impactId: selectedCosmeticLoadout.value.impactId,
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '外观装备失败，请稍后重试。'))
    }

    const data = await response.json()
    applyState(data)
  } catch (error) {
    errorMessage.value = error.message || '外观装备失败，请稍后重试。'
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

onMounted(async () => {
  restoreCachedLatestAnnouncement()
  await loadPlayerSession()
  await loadState()
  connectRealtime(nickname.value)
})

onBeforeUnmount(() => {
  clickBehaviorTracker.clear()
  realtimeTransport?.close()
  clearStarlightTimer()
  burstTimers.forEach((timer) => window.clearTimeout(timer))
  burstTimers.clear()
  cosmeticTimers.forEach((timer) => window.clearTimeout(timer))
  cosmeticTimers.clear()
})
</script>

<template>
  <main class="page-shell">
    <div class="page-shell__glow page-shell__glow--pink"></div>
    <div class="page-shell__glow page-shell__glow--blue"></div>
    <div class="page-shell__glow page-shell__glow--yellow"></div>

    <section class="hero">
      <div class="hero__copy">
        <p class="hero__eyebrow">Long Vote Wall</p>
        <h1>登录账号，再狠狠干一票。</h1>
        <p class="hero__lede">
          平时点按钮照样冲榜；有活动 Boss 时，同一点击会把装备增量一起结算成伤害。
        </p>
      </div>

      <div class="hero__status">
        <span class="live-pill" :class="{ 'live-pill--syncing': syncing }">
          <span class="live-pill__dot"></span>
          {{ syncLabel }}
        </span>
        <span class="hero__time">最近刷新 {{ lastUpdatedAt || '--:--:--' }}</span>
        <a class="hero__admin-link" href="/admin">管理后台</a>
      </div>
    </section>

    <section v-if="announcementModalOpen && latestAnnouncement" class="announcement-modal" aria-label="更新公告">
      <div class="announcement-modal__backdrop" @click="closeAnnouncementModal"></div>
      <article class="announcement-modal__card">
        <p class="vote-stage__eyebrow">更新内容公告</p>
        <strong>{{ latestAnnouncement.title }}</strong>
        <p class="announcement-modal__time">{{ formatTime(latestAnnouncement.publishedAt) }}</p>
        <p class="social-card__copy social-card__copy--multiline">{{ latestAnnouncement.content }}</p>
        <div class="announcement-modal__actions">
          <button class="nickname-form__submit" type="button" @click="closeAnnouncementModal">我知道了</button>
        </div>
      </article>
    </section>

    <section class="stats-band stats-band--wide" aria-label="实时统计">
      <article class="stats-band__card">
        <span class="stats-band__label">当前按钮</span>
        <strong>{{ buttonCount }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">累计点击</span>
        <strong>{{ totalVotes }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">我的点击</span>
        <strong>{{ isLoggedIn ? myClicks : '先登录' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">我的排名</span>
        <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">单击增量</span>
        <strong>+{{ effectiveIncrement }}</strong>
      </article>
    </section>

    <section class="boss-stage social-card">
      <div class="boss-stage__head">
        <div>
          <p class="vote-stage__eyebrow">世界 Boss</p>
          <strong>{{ boss?.name || '当前休战中' }}</strong>
          <p class="social-card__copy">
            {{
              !boss
                  ? '现在没有活动 Boss，按钮依然能正常计票，装备加成也照常生效。'
                  : boss.status === 'active'
                      ? '全服正在集火当前 Boss，每次点击都会把装备加成一起折算成伤害。'
                      : '这只 Boss 已经倒下，等待后台手动开启下一只。'
            }}
          </p>
          <p>
            ⚠️小贴士⚠️：只有对Boss造成超过1%生命值的伤害，才有资格掉落装备哦～萌新也要努力输出呀！(๑•̀ㅂ•́)و✧ 
          </p>
        </div>
        <div class="boss-stage__meta">
          <span class="boss-stage__pill">{{ bossStatusLabel }}</span>
          <strong v-if="boss">HP {{ boss.currentHp }} / {{ boss.maxHp }}</strong>
          <strong v-else>我的伤害 {{ myBossDamage }}</strong>
        </div>
      </div>

      <div v-if="boss" class="boss-stage__progress">
        <div class="boss-stage__bar">
          <span class="boss-stage__bar-fill" :style="{ width: `${bossProgress}%` }"></span>
        </div>
        <div class="boss-stage__stats">
          <span>我的伤害 {{ myBossDamage }}</span>
          <span>当前 Boss 榜 {{ bossLeaderboard.length }} 人</span>
          <span>掉落池 {{ bossLoot.length }} 件</span>
          <span v-if="displayedRecentRewards.length > 0">最近掉落 {{ recentRewardTitle }}</span>
        </div>
      </div>

      <div v-if="boss" class="boss-stage__drops">
        <div class="boss-stage__drops-head">
          <div>
            <p class="vote-stage__eyebrow">Boss 掉落池</p>
            <strong>{{ bossLoot.length }} 件</strong>
          </div>
        </div>

        <div v-if="bossLoot.length === 0" class="leaderboard-list leaderboard-list--empty">
          <p>当前 Boss 还没配置掉落池。</p>
        </div>
        <ul v-else class="inventory-list inventory-list--loot">
          <li
              v-for="item in bossLoot"
              :key="item.itemId"
              class="inventory-item inventory-item--stacked inventory-item--loot"
          >
            <div>
              <strong>
                <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
              </strong>
              <p>{{ formatRarityLabel(item.rarity) }} · {{ item.slot || '未分类' }} · 掉落概率 {{ formatDropRate(item.dropRatePercent) }}</p>
              <p>{{ formatEnhanceCap(item.enhanceCap) }}</p>
              <p>{{ formatItemStats(item) }}</p>
            </div>
          </li>
        </ul>
      </div>

      <div v-if="boss && bossHeroLoot.length > 0" class="boss-stage__drops">
        <div class="boss-stage__drops-head">
          <div>
            <p class="vote-stage__eyebrow">Boss 英雄池</p>
            <strong>{{ bossHeroLoot.length }} 位</strong>
          </div>
        </div>

        <ul class="inventory-list inventory-list--loot">
          <li
              v-for="hero in bossHeroLoot"
              :key="hero.heroId"
              class="inventory-item inventory-item--stacked inventory-item--loot"
          >
            <div class="inventory-item__hero">
              <img
                  v-if="hero.imagePath"
                  class="inventory-item__avatar"
                  :src="hero.imagePath"
                  :alt="heroImageAlt(hero)"
              />
              <div class="inventory-item__hero-meta">
                <strong>{{ hero.heroName || hero.heroId }}</strong>
                <p>掉落概率 {{ formatDropRate(hero.dropRatePercent) }}</p>
                <p>{{ formatAwakenCap(hero.awakenCap) }}</p>
                <p>{{ formatItemStats(hero) }}</p>
                <p>{{ formatHeroTrait(hero) }}</p>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </section>

    <section class="stage-layout">
      <aside class="player-hud">
        <section class="player-hud__shell">
          <div class="player-hud__head">
            <div>
              <p class="vote-stage__eyebrow">Player HUD</p>
              <strong>{{ isLoggedIn ? nickname : '未登录角色' }}</strong>
            </div>
            <span class="player-hud__pill">{{ isLoggedIn ? '已上墙' : '访客' }}</span>
          </div>

          <p class="player-hud__copy">
            {{
              isLoggedIn ? `你现在登录的是 ${nickname}。背包、属性和装备都会跟着这个账号走。` : '先输入昵称和密码登录；第一次使用该昵称时会直接为它设置密码。'
            }}
          </p>

          <form class="nickname-form player-hud__form" @submit.prevent="submitNickname">
            <input
                v-model="nicknameDraft"
                class="nickname-form__input"
                type="text"
                maxlength="20"
                placeholder="比如：阿明"
            />
            <input
                v-model="passwordDraft"
                class="nickname-form__input"
                type="password"
                placeholder="输入密码"
            />
            <button class="nickname-form__submit" type="submit">
              {{ isLoggedIn ? '切换账号' : '登录 / 首次认领' }}
            </button>
          </form>

          <button
              v-if="isLoggedIn"
              class="nickname-form__ghost player-hud__reset"
              type="button"
              @click="resetNickname"
          >
            退出登录
          </button>

          <section class="player-hud__auto">
            <div class="player-hud__section-head">
              <div>
                <p class="vote-stage__eyebrow">挂机</p>
                <strong>{{ autoClickEnabled ? '进行中' : '未开启' }}</strong>
              </div>
              <span class="player-hud__pill" :class="{ 'player-hud__pill--active': autoClickEnabled }">
                {{ AUTO_CLICK_RATE_LABEL }}
              </span>
            </div>

            <p class="player-hud__note">{{ autoClickStatus }}</p>

            <div class="player-hud__auto-meta">
              <span class="player-hud__auto-chip">目标：{{ autoClickTargetLabel }}</span>
              <span class="player-hud__auto-chip">关闭页面自动停止</span>
            </div>

            <button
                class="nickname-form__submit player-hud__auto-button"
                type="button"
                :disabled="!autoClickEnabled && !canStartAutoClick"
                @click="toggleAutoClick"
            >
              {{ autoClickEnabled ? '关闭挂机' : '开启挂机' }}
            </button>
          </section>

          <div class="player-hud__tabs">
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'inventory' }"
                type="button"
                @click="selectHudTab('inventory')"
            >
              背包
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'stats' }"
                type="button"
                @click="selectHudTab('stats')"
            >
              属性
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'loadout' }"
                type="button"
                @click="selectHudTab('loadout')"
            >
              装备栏
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'heroes' }"
                type="button"
                @click="selectHudTab('heroes')"
            >
              小小英雄
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'forge' }"
                type="button"
                @click="selectHudTab('forge')"
            >
              强化
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'shop' }"
                type="button"
                @click="selectHudTab('shop')"
            >
              商店
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'info' }"
                type="button"
                @click="selectHudTab('info')"
            >
              信息
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'messages' }"
                type="button"
                @click="selectHudTab('messages')"
            >
              留言
            </button>
          </div>

          <div class="player-hud__content">
            <section v-if="activeHudTab === 'inventory'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">背包</p>
                <strong>{{ inventory.length }} 件</strong>
              </div>

              <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
                <p>先去打 Boss 或等后台发装备，背包就会慢慢满起来。</p>
              </div>

              <ul v-else class="inventory-list">
                <li v-for="item in inventory" :key="item.itemId" class="inventory-item inventory-item--panel">
                  <div class="inventory-item__top">
                    <div class="inventory-item__main">
                      <strong>
                        <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                        <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                      </strong>
                      <div class="inventory-item__meta">
                        <span class="inventory-item__chip">{{ formatRarityLabel(item.rarity) }}</span>
                        <span class="inventory-item__chip">类型:{{ item.slot || '未分类' }}</span>
                        <span class="inventory-item__chip">库存:{{ item.quantity }}</span>
                        <span class="inventory-item__chip">强化:{{
                            item.enhanceLevel ? `+${item.enhanceLevel}` : '未强化'
                          }}</span>
                        <span class="inventory-item__chip">可分解:{{ salvageableEquipmentCount(item) }}</span>
                      </div>
                    </div>
                  </div>

                  <ul class="inventory-item__stats inventory-item__stats--stacked">
                    <li v-for="line in formatItemStatLines(item)" :key="line">
                      {{ line }}
                    </li>
                  </ul>

                  <div class="inventory-item__footer">
                    <span
                        class="inventory-item__state"
                        :class="{ 'inventory-item__state--active': item.equipped }"
                    >
                      {{ item.equipped ? '已穿戴' : '待命中' }}
                    </span>

                    <div class="inventory-item__actions">
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || actioningItemId === item.itemId"
                          @click="item.equipped ? postEquipmentAction(item.itemId, 'unequip') : postEquipmentAction(item.itemId, 'equip')"
                      >
                        {{ item.equipped ? '卸下' : '穿戴' }}
                      </button>
                    </div>
                  </div>
                </li>
              </ul>
            </section>

            <section v-else-if="activeHudTab === 'stats'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">战斗属性</p>
                <strong>{{ isLoggedIn ? nickname : '未登录' }}</strong>
              </div>

              <div class="me-card__stats">
                <article>
                  <span>普通伤害</span>
                  <strong>{{ normalDamage }}</strong>
                </article>
                <article>
                  <span>暴击伤害</span>
                  <strong>{{ criticalDamage }}</strong>
                </article>
                <article>
                  <span>暴击率</span>
                  <strong>{{ formatNumber(combatStats.criticalChancePercent, 2) }}%</strong>
                </article>
                <article>
                  <span>我的 Boss 伤害</span>
                  <strong>{{ myBossDamage }}</strong>
                </article>
                <article>
                  <span>我的点击</span>
                  <strong>{{ isLoggedIn ? myClicks : '--' }}</strong>
                </article>
                <article>
                  <span>我的排名</span>
                  <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
                </article>
              </div>
            </section>

            <section v-else-if="activeHudTab === 'loadout'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">装备栏</p>
                <strong>{{ equippedItems.length }} / 3</strong>
              </div>

              <div class="loadout-grid">
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>武器</span>
                    <strong v-if="loadout.weapon">
                      <span v-if="equipmentNameParts(loadout.weapon).prefix">{{ equipmentNameParts(loadout.weapon).prefix }}</span>
                      <span :class="equipmentNameClass(loadout.weapon)">{{ equipmentNameParts(loadout.weapon).text }}</span>
                    </strong>
                    <strong v-else>未穿戴</strong>
                  </div>
                  <ul v-if="loadout.weapon" class="loadout-slot__attrs">
                    <li>{{ formatRarityLabel(loadout.weapon.rarity) }}</li>
                    <li v-for="line in formatItemStatLines(loadout.weapon)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>护甲</span>
                    <strong v-if="loadout.armor">
                      <span v-if="equipmentNameParts(loadout.armor).prefix">{{ equipmentNameParts(loadout.armor).prefix }}</span>
                      <span :class="equipmentNameClass(loadout.armor)">{{ equipmentNameParts(loadout.armor).text }}</span>
                    </strong>
                    <strong v-else>未穿戴</strong>
                  </div>
                  <ul v-if="loadout.armor" class="loadout-slot__attrs">
                    <li>{{ formatRarityLabel(loadout.armor.rarity) }}</li>
                    <li v-for="line in formatItemStatLines(loadout.armor)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>饰品</span>
                    <strong v-if="loadout.accessory">
                      <span v-if="equipmentNameParts(loadout.accessory).prefix">{{ equipmentNameParts(loadout.accessory).prefix }}</span>
                      <span :class="equipmentNameClass(loadout.accessory)">{{ equipmentNameParts(loadout.accessory).text }}</span>
                    </strong>
                    <strong v-else>未穿戴</strong>
                  </div>
                  <ul v-if="loadout.accessory" class="loadout-slot__attrs">
                    <li>{{ formatRarityLabel(loadout.accessory.rarity) }}</li>
                    <li v-for="line in formatItemStatLines(loadout.accessory)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
              </div>
            </section>

            <section v-else-if="activeHudTab === 'heroes'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">小小英雄</p>
                <strong>{{ heroCount }} 位</strong>
              </div>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>当前出战</span>
                  <strong>{{ activeHero?.name || '未派出' }}</strong>
                </div>
                <div v-if="activeHero?.imagePath" class="player-hud__hero-active">
                  <img
                      class="player-hud__hero-portrait"
                      :src="activeHero.imagePath"
                      :alt="heroImageAlt(activeHero)"
                  />
                </div>
                <p class="player-hud__note">
                  {{
                    activeHero
                        ? `${formatItemStats(activeHero)}，${formatHeroTrait(activeHero)}`
                        : '先去打 Boss 拿到一位英雄，再派出去陪你冲榜。'
                  }}
                </p>
              </section>

              <div v-if="heroes.length === 0" class="leaderboard-list leaderboard-list--empty">
                <p>你还没有招募到任何小小英雄。</p>
              </div>

              <ul v-else class="inventory-list">
                <li v-for="hero in heroes" :key="hero.heroId" class="inventory-item inventory-item--panel">
                  <div class="inventory-item__top">
                    <img
                        v-if="hero.imagePath"
                        class="inventory-item__avatar inventory-item__avatar--hero"
                        :src="hero.imagePath"
                        :alt="heroImageAlt(hero)"
                    />
                    <div class="inventory-item__main">
                      <strong>{{ hero.name }}</strong>
                      <div class="inventory-item__meta">
                        <span class="inventory-item__chip">库存:{{ hero.quantity }}</span>
                        <span class="inventory-item__chip">{{ hero.active ? '出战中' : '待命中' }}</span>
                        <span class="inventory-item__chip">觉醒:{{ hero.awakenLevel || 0 }} / {{ hero.awakenCap || '∞' }}</span>
                        <span class="inventory-item__chip">可分解:{{ salvageableHeroCount(hero) }}</span>
                      </div>
                    </div>
                  </div>

                  <ul class="inventory-item__stats inventory-item__stats--stacked">
                    <li>{{ formatItemStats(hero) }}</li>
                    <li>{{ formatHeroTrait(hero) }}</li>
                  </ul>

                  <div class="inventory-item__footer">
                    <span
                        class="inventory-item__state"
                        :class="{ 'inventory-item__state--active': hero.active }"
                    >
                      {{ hero.active ? '已出战' : '未出战' }}
                    </span>

                    <div class="inventory-item__actions">
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || actioningItemId === hero.heroId"
                          @click="hero.active ? postHeroAction(hero.heroId, 'unequip') : postHeroAction(hero.heroId, 'equip')"
                      >
                        {{ hero.active ? '收回' : '出战' }}
                      </button>
                    </div>
                  </div>
                </li>
              </ul>
            </section>

            <section v-else-if="activeHudTab === 'forge'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">原石强化</p>
                <strong>{{ gems }} 原石</strong>
              </div>

              <div class="forge-grid">
                <article class="forge-summary">
                  <span>当前余额</span>
                  <strong>{{ gems }} 原石</strong>
                  <p>重复装备和重复英雄都可以分解成原石，再投入强化和觉醒。</p>
                  <p>每件装备或者小小英雄可以分解为1原石</p>
                </article>
                <article class="forge-summary">
                  <span>本期价格</span>
                  <strong>强化 {{ EQUIPMENT_ENHANCE_COST }} · 觉醒 {{ HERO_AWAKEN_COST }}</strong>
                  <p>装备与英雄都按模板上限成长，每次只提升一个基础属性，并直接返回本次成长结果。</p>
                </article>
                <article class="forge-summary">
                  <span>强化规则</span>
                  <strong>三项基础属性等概率命中</strong>
                  <p class="forge-summary__tips">
                    <span>仅提升点击 / 暴击 / 暴击率中的一项。</span>
                    <span>{{ GROWTH_FORMULA_TEXT }}。</span>
                    <span>暴击率每次固定 +0.20%。</span>
                  </p>
                </article>
              </div>

              <article
                  v-if="lastForgeResult"
                  class="forge-result"
                  :class="{ 'forge-result--jackpot': lastForgeResult.jackpot }"
              >
                <span>{{ lastForgeResult.kind }}</span>
                <strong>{{ lastForgeResult.targetName || lastForgeResult.targetId }}</strong>
                <p class="player-hud__note">
                  {{ lastForgeResult.rewardSummary }} · 原石 {{ lastForgeResult.gemsDelta > 0 ? '+' : '' }}{{ lastForgeResult.gemsDelta }} · 余额 {{ lastForgeResult.remainingGems }}
                </p>
              </article>

              <section class="player-hud__info-block">
                <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>背包里还没有装备，先去打 Boss 再回来强化。</p>
                </div>
                <ul v-else class="forge-action-list">
                  <li v-for="item in inventory" :key="`forge-${item.itemId}`">
                    <div class="forge-action-list__copy">
                      <strong>
                        <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                        <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                      </strong>
                      <div class="forge-action-list__meta">
                        <span>{{ formatRarityLabel(item.rarity) }}</span>
                        <span>可分解 {{ salvageableEquipmentCount(item) }} 件</span>
                        <span>强化 {{ item.enhanceLevel || 0 }} / {{ item.enhanceCap || '∞' }}</span>
                        <span>每次 {{ EQUIPMENT_ENHANCE_COST }} 原石</span>
                      </div>
                      <p>{{ equipmentEnhanceHint(item) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn || !salvageableEquipmentCount(item) || actioningItemId === item.itemId"
                          @click="salvageEquipment(item)"
                      >
                        分解 x{{ salvageableEquipmentCount(item) }}
                      </button>
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < EQUIPMENT_ENHANCE_COST || actioningItemId === item.itemId || (item.enhanceCap > 0 && item.enhanceLevel >= item.enhanceCap)"
                          @click="enhanceEquipment(item)"
                      >
                        强化
                      </button>
                    </div>
                  </li>
                </ul>
              </section>

              <section class="player-hud__info-block">
                <div v-if="heroes.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>你还没有招募到英雄，先去 Boss 池碰碰运气。</p>
                </div>
                <ul v-else class="forge-action-list">
                  <li v-for="hero in heroes" :key="`awaken-${hero.heroId}`">
                    <div class="forge-action-list__copy">
                      <strong>{{ hero.name }}</strong>
                      <div class="forge-action-list__meta">
                        <span>可分解 {{ salvageableHeroCount(hero) }} 个</span>
                        <span>觉醒 {{ hero.awakenLevel || 0 }} / {{ hero.awakenCap || '∞' }}</span>
                        <span>每次 {{ HERO_AWAKEN_COST }} 原石</span>
                      </div>
                      <p>{{ heroAwakenHint(hero) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn || !salvageableHeroCount(hero) || actioningItemId === hero.heroId"
                          @click="salvageHero(hero)"
                      >
                        分解 x{{ salvageableHeroCount(hero) }}
                      </button>
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < HERO_AWAKEN_COST || actioningItemId === hero.heroId || (hero.awakenCap > 0 && hero.awakenLevel >= hero.awakenCap)"
                          @click="awakenHero(hero)"
                      >
                        觉醒
                      </button>
                    </div>
                  </li>
                </ul>
              </section>
            </section>

            <section v-else-if="activeHudTab === 'shop'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">外观商店</p>
                <strong>{{ gems }} 原石</strong>
              </div>

              <div class="forge-grid">
                <article class="forge-summary">
                  <span>已拥有外观</span>
                  <strong>{{ ownedCosmetics.length }} 件</strong>
                  <p>一期只卖轨迹和点击特效，全部拆件售卖，不碰任何数值。</p>
                </article>
                <article class="forge-summary">
                  <span>当前装备</span>
                  <strong>{{ equippedCosmeticSummary.trailName }} / {{ equippedCosmeticSummary.impactName }}</strong>
                  <p>轨迹和点击特效可以自由混搭，星光按钮上会自动降透明度。</p>
                </article>
              </div>

              <section class="cosmetic-preview">
                <div class="player-hud__mini-head">
                  <span>试衣预览</span>
                  <strong>{{ selectedCosmeticSummary.trailName }} / {{ selectedCosmeticSummary.impactName }}</strong>
                </div>
                <div class="cosmetic-preview__stage">
                  <div class="cosmetic-preview__copy">
                    <span>仅自己可见</span>
                    <strong>普通点击、挂机点击和星光按钮都会自动切换到对应表现。</strong>
                    <p>星光态会压制外观亮度，避免和系统提示抢焦点。</p>
                  </div>
                  <span
                      v-if="previewEffectConfig.trailTheme"
                      class="cosmetic-preview__trail"
                      :class="[previewEffectConfig.trailClass, cosmeticModeClasses(previewEffectConfig)]"
                  ></span>
                  <span
                      v-if="previewEffectConfig.impactTheme"
                      class="cosmetic-preview__impact"
                      :class="[previewEffectConfig.impactClass, cosmeticModeClasses(previewEffectConfig)]"
                  >
                    <span
                        v-for="dot in previewDots"
                        :key="`preview-${dot}`"
                        class="cosmetic-preview__dot"
                    ></span>
                  </span>
                </div>
                <div class="cosmetic-preview__actions">
                  <button
                      class="inventory-item__action"
                      type="button"
                      :disabled="!canApplyCosmeticSelection || actioningItemId === 'cosmetic-loadout'"
                      @click="equipSelectedCosmetics"
                  >
                    应用当前搭配
                  </button>
                  <button
                      class="nickname-form__ghost"
                      type="button"
                      :disabled="actioningItemId === 'cosmetic-loadout'"
                      @click="syncCosmeticDraft(equippedCosmetics)"
                  >
                    恢复已装备
                  </button>
                  <button
                      class="nickname-form__ghost"
                      type="button"
                      :disabled="actioningItemId === 'cosmetic-loadout'"
                      @click="syncCosmeticDraft(createEmptyCosmeticLoadout())"
                  >
                    清空搭配
                  </button>
                </div>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>轨迹</span>
                  <strong>{{ cosmeticCollections.trails.length }} 件</strong>
                </div>
                <ul class="shop-grid">
                  <li
                      v-for="item in cosmeticCollections.trails"
                      :key="item.cosmeticId"
                      class="shop-card"
                      :class="{
                        'shop-card--owned': item.owned,
                        'shop-card--equipped': item.equipped,
                        'shop-card--selected': selectedCosmeticLoadout.trailId === item.cosmeticId,
                      }"
                  >
                    <div class="shop-card__preview" :class="`cosmetic-theme--${item.preview?.theme || 'ribbon'}`">
                      <span class="shop-card__preview-mark"></span>
                    </div>
                    <div>
                      <strong>{{ item.name }}</strong>
                      <p>{{ formatRarityLabel(item.rarity) }} · 轨迹 · {{ cosmeticStatusText(item) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          v-if="!item.owned"
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < item.price || actioningItemId === item.cosmeticId"
                          @click="purchaseCosmetic(item)"
                      >
                        购买
                      </button>
                      <button
                          v-else
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn"
                          @click="selectCosmeticItem(item)"
                      >
                        {{ selectedCosmeticLoadout.trailId === item.cosmeticId ? '已选中' : '选这条轨迹' }}
                      </button>
                    </div>
                  </li>
                </ul>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>点击特效</span>
                  <strong>{{ cosmeticCollections.impacts.length }} 件</strong>
                </div>
                <ul class="shop-grid">
                  <li
                      v-for="item in cosmeticCollections.impacts"
                      :key="item.cosmeticId"
                      class="shop-card"
                      :class="{
                        'shop-card--owned': item.owned,
                        'shop-card--equipped': item.equipped,
                        'shop-card--selected': selectedCosmeticLoadout.impactId === item.cosmeticId,
                      }"
                  >
                    <div class="shop-card__preview" :class="`cosmetic-theme--${item.preview?.theme || 'ribbon'}`">
                      <span class="shop-card__preview-mark"></span>
                    </div>
                    <div>
                      <strong>{{ item.name }}</strong>
                      <p>{{ formatRarityLabel(item.rarity) }} · 点击特效 · {{ cosmeticStatusText(item) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          v-if="!item.owned"
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < item.price || actioningItemId === item.cosmeticId"
                          @click="purchaseCosmetic(item)"
                      >
                        购买
                      </button>
                      <button
                          v-else
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn"
                          @click="selectCosmeticItem(item)"
                      >
                        {{ selectedCosmeticLoadout.impactId === item.cosmeticId ? '已选中' : '选这个特效' }}
                      </button>
                    </div>
                  </li>
                </ul>
              </section>
            </section>

            <section v-else-if="activeHudTab === 'info'" class="player-hud__panel player-hud__panel--info">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">信息</p>
                <strong>{{ bossHistory.length }} 条战报</strong>
              </div>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>最新公告</span>
                  <strong>{{ latestAnnouncement?.title || '暂无' }}</strong>
                </div>
                <p class="player-hud__note player-hud__note--multiline">
                  {{ latestAnnouncement?.content || '当前还没有新的站内公告。' }}
                </p>
                <button
                    v-if="latestAnnouncement"
                    class="nickname-form__ghost player-hud__retry"
                    type="button"
                    @click="announcementModalOpen = true"
                >
                  再看一遍
                </button>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>最近掉落</span>
                  <strong>{{ recentRewardTitle }}</strong>
                </div>
                <p class="player-hud__note">{{ recentRewardNote }}</p>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>装备获取</span>
                  <strong>{{ bossLoot.length }} 件</strong>
                </div>
                <div v-if="bossLoot.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>当前 Boss 还没配置掉落池。</p>
                </div>
                <ul v-else class="inventory-list inventory-list--hud-loot">
                  <li
                      v-for="item in bossLoot"
                      :key="item.itemId"
                      class="inventory-item inventory-item--stacked inventory-item--loot"
                  >
                    <div>
                      <strong>
                        <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                        <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                      </strong>
                      <p>{{ formatRarityLabel(item.rarity) }} · {{ item.slot || '未分类' }} · 掉落概率 {{ formatDropRate(item.dropRatePercent) }}</p>
                      <p>{{ formatEnhanceCap(item.enhanceCap) }}</p>
                      <p>{{ formatItemStats(item) }}</p>
                    </div>
                  </li>
                </ul>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>英雄招募</span>
                  <strong>{{ bossHeroLoot.length }} 位</strong>
                </div>
                <div v-if="bossHeroLoot.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>当前 Boss 还没配置英雄掉落。</p>
                </div>
                <ul v-else class="inventory-list inventory-list--hud-loot">
                  <li
                      v-for="hero in bossHeroLoot"
                      :key="hero.heroId"
                      class="inventory-item inventory-item--stacked inventory-item--loot"
                  >
                    <div class="inventory-item__hero">
                      <img
                          v-if="hero.imagePath"
                          class="inventory-item__avatar"
                          :src="hero.imagePath"
                          :alt="heroImageAlt(hero)"
                      />
                      <div>
                        <strong>{{ hero.heroName || hero.heroId }}</strong>
                        <p>掉落概率 {{ formatDropRate(hero.dropRatePercent) }}</p>
                        <p>{{ formatAwakenCap(hero.awakenCap) }}</p>
                        <p>{{ formatItemStats(hero) }}</p>
                        <p>{{ formatHeroTrait(hero) }}</p>
                      </div>
                    </div>
                  </li>
                </ul>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>公告历史</span>
                  <strong>{{ announcements.length }} 条</strong>
                </div>
                <div v-if="loadingAnnouncements" class="leaderboard-list leaderboard-list--empty">
                  <p>公告加载中...</p>
                </div>
                <div v-else-if="announcementError" class="leaderboard-list leaderboard-list--empty">
                  <p>{{ announcementError }}</p>
                </div>
                <ul v-else-if="announcements.length > 0" class="history-list">
                  <li v-for="item in announcements" :key="item.id" class="history-item">
                    <div class="history-item__head">
                      <strong>{{ item.title }}</strong>
                      <span>{{ formatTime(item.publishedAt) }}</span>
                    </div>
                    <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
                  </li>
                </ul>
                <div v-else class="leaderboard-list leaderboard-list--empty">
                  <p>暂无公告历史。</p>
                </div>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>往届 Boss 查询</span>
                  <strong>{{ filteredBossHistory.length }} 条</strong>
                </div>
                <input
                    v-model="bossHistoryQuery"
                    class="nickname-form__input"
                    type="text"
                    placeholder="按 Boss 名称或 ID 搜索"
                />
                <div v-if="loadingBossHistory" class="leaderboard-list leaderboard-list--empty">
                  <p>历史 Boss 加载中...</p>
                </div>
                <div v-else-if="bossHistoryError" class="leaderboard-list leaderboard-list--empty">
                  <p>{{ bossHistoryError }}</p>
                  <button
                      class="nickname-form__ghost player-hud__retry"
                      type="button"
                      @click="loadBossHistory(true)"
                  >
                    重新加载
                  </button>
                </div>
                <div v-else-if="filteredBossHistory.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>没有匹配的 Boss 记录。</p>
                </div>
                <ul v-else class="history-list">
                  <li v-for="entry in filteredBossHistory" :key="entry.id" class="history-item">
                    <div class="history-item__head">
                      <strong>{{ entry.name || entry.id }}</strong>
                      <span>{{ formatBossTime(entry.startedAt) }}</span>
                    </div>
                    <p>
                      {{ entry.status === 'defeated' ? '已击败' : '已结束' }} · 掉落 {{ entry.loot?.length ?? 0 }} 件
                    </p>
                    <p v-if="topBossDamage(entry)">
                      伤害第一 {{ topBossDamage(entry).nickname }} · {{ topBossDamage(entry).damage }}
                    </p>
                    <p v-else>暂无伤害记录。</p>
                  </li>
                </ul>
              </section>
            </section>

            <section v-else class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">公共留言墙</p>
                <strong>{{ messages.length }} 条</strong>
              </div>

              <form class="admin-form player-hud__message-form" @submit.prevent="submitMessage">
                <textarea
                    v-model="messageDraft"
                    class="nickname-form__input admin-textarea"
                    rows="4"
                    maxlength="200"
                    placeholder="说点什么，所有人都能看到。"
                ></textarea>
                <button class="nickname-form__submit" type="submit" :disabled="postingMessage || !isLoggedIn">
                  {{ postingMessage ? '发送中...' : '发送留言' }}
                </button>
              </form>

              <p v-if="messageError" class="feedback feedback--error">{{ messageError }}</p>

              <div v-if="loadingMessages" class="leaderboard-list leaderboard-list--empty">
                <p>留言加载中...</p>
              </div>
              <div v-else-if="messages.length === 0" class="leaderboard-list leaderboard-list--empty">
                <p>还没有留言，先写第一条。</p>
              </div>
              <ul v-else class="history-list">
                <li v-for="item in messages" :key="item.id" class="history-item">
                  <div class="history-item__head">
                    <strong>{{ item.nickname }}</strong>
                    <span>{{ formatTime(item.createdAt) }}</span>
                  </div>
                  <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
                </li>
              </ul>

              <button
                  v-if="messageNextCursor"
                  class="nickname-form__ghost player-hud__retry"
                  type="button"
                  :disabled="loadingMessages"
                  @click="loadMessages(messageNextCursor, true)"
              >
                加载更多
              </button>
            </section>
          </div>
        </section>
      </aside>

      <section class="vote-stage">
        <div class="vote-stage__head">
          <div>
            <p class="vote-stage__eyebrow">现场投票墙</p>
            <h2>看见哪个想按，就直接拍下去。</h2>
          </div>
          <p v-if="!errorMessage" class="vote-stage__hint">
            {{ isLoggedIn ? `现在上墙的是 ${nickname}` : '先登录账号，再开始冲榜。' }}
          </p>
        </div>

        <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

        <section v-if="boss" class="vote-stage__boss-hud">
          <div class="vote-stage__boss-hud-head">
            <div>
              <p class="vote-stage__eyebrow">当前 Boss</p>
              <strong>{{ boss.name }}</strong>
            </div>
            <strong>HP {{ boss.currentHp }} / {{ boss.maxHp }}</strong>
          </div>
          <div class="boss-stage__bar boss-stage__bar--compact">
            <span class="boss-stage__bar-fill" :style="{ width: `${bossProgress}%` }"></span>
          </div>
          <div class="vote-stage__boss-hud-stats">
            <span>我的伤害 {{ myBossDamage }}</span>
            <span>Boss 榜 {{ bossLeaderboard.length }} 人</span>
            <span>掉落池 {{ bossLoot.length }} 件</span>
            <span v-if="displayedRecentRewards.length > 0">最近掉落 {{ recentRewardTitle }}</span>
          </div>
        </section>

        <div v-if="loading" class="feedback-panel">
          <p>正在把现场按钮搬上来...</p>
        </div>

        <div v-else-if="buttons.length === 0" class="feedback-panel">
          <p>还没有按钮上墙，先加一个再回来看看。</p>
        </div>

        <div v-else>
          <div class="vote-stage__filters">
            <input
                v-model="buttonSearch"
                class="nickname-form__input vote-stage__search"
                type="text"
                placeholder="搜按钮名或标签"
            />
            <div class="vote-stage__tags">
              <button
                  v-for="tag in buttonTags"
                  :key="tag"
                  class="vote-stage__tag"
                  :class="{ 'vote-stage__tag--active': selectedButtonTag === tag }"
                  type="button"
                  @click="selectedButtonTag = tag"
              >
                {{ tag }}
              </button>
            </div>
            <p class="vote-stage__hint">
              当前第 {{ buttonPage }} / {{ buttonTotalPages }} 页，搜索和标签只筛当前页。
            </p>
          </div>

          <div v-if="displayedButtons.length === 0" class="feedback-panel">
            <p>当前筛选下没有匹配按钮，换个标签或关键词试试。</p>
          </div>

          <div v-else class="button-grid">
            <button
                v-for="button in displayedButtons"
                :key="button.key"
                class="vote-card"
                :class="{
              'vote-card--image': button.imagePath,
              'vote-card--pending': pendingKeys.has(button.key),
              'vote-card--critical': Boolean(criticalBursts[button.key]),
              'vote-card--starlight': isStarlightButton(button.key),
              'vote-card--locked': !isLoggedIn,
            }"
                type="button"
                :disabled="pendingKeys.has(button.key) || !isLoggedIn"
                :aria-label="`${button.label}，当前 ${button.count} 票`"
                @pointerdown="handlePressStart(button.key, $event)"
                @pointerup="handlePressEnd(button.key, $event)"
                @pointercancel="handlePressCancel(button.key)"
                @click="clickButton(button.key)"
            >
              <span class="vote-card__shine"></span>
              <span
                  v-if="cosmeticBursts[button.key]?.trailTheme"
                  :key="`${cosmeticBursts[button.key].nonce}-trail`"
                  class="vote-card__cosmetic vote-card__cosmetic--trail"
                  :class="[cosmeticBursts[button.key].trailClass, cosmeticModeClasses(cosmeticBursts[button.key])]"
                  :style="{ animationDuration: `${cosmeticBursts[button.key].durationMs}ms` }"
              ></span>
              <span
                  v-if="cosmeticBursts[button.key]?.impactTheme"
                  :key="`${cosmeticBursts[button.key].nonce}-impact`"
                  class="vote-card__cosmetic vote-card__cosmetic--impact"
                  :class="[cosmeticBursts[button.key].impactClass, cosmeticModeClasses(cosmeticBursts[button.key])]"
                  :style="{ animationDuration: `${cosmeticBursts[button.key].durationMs}ms` }"
              >
                <span
                    v-for="dot in cosmeticBursts[button.key].dots"
                    :key="`${cosmeticBursts[button.key].nonce}-dot-${dot}`"
                    class="vote-card__cosmetic-dot"
                ></span>
              </span>
              <span
                  v-if="criticalBursts[button.key]"
                  :key="criticalBursts[button.key].nonce"
                  class="vote-card__burst"
                  aria-hidden="true"
              ></span>
              <span
                  v-if="criticalBursts[button.key]"
                  :key="`${criticalBursts[button.key].nonce}-label`"
                  class="vote-card__critical-text"
                  aria-hidden="true"
              >
              {{ criticalBursts[button.key].label }}
            </span>
              <span class="vote-card__badge">
              {{
                  !isLoggedIn
                      ? '先登录'
                      : pendingKeys.has(button.key)
                          ? '正在记票'
                          : isStarlightButton(button.key)
                              ? boss?.status === 'active'
                                  ? `星光双倍 · 拍一下 +${effectiveIncrement * 2}`
                                  : `星光双倍 · 拍一下 +${effectiveIncrement * 2}`
                              : boss?.status === 'active'
                                  ? `拍一下 +${effectiveIncrement} 并打 Boss`
                                  : `拍一下 +${effectiveIncrement}`
                }}
            </span>

              <img
                  v-if="button.imagePath"
                  class="vote-card__image"
                  :src="button.imagePath"
                  :alt="button.imageAlt || button.label"
              />
              <strong v-else class="vote-card__label">{{ button.label }}</strong>

              <span class="vote-card__count">{{ button.count }}</span>
            </button>
          </div>

          <div v-if="buttonTotalPages > 1" class="inventory-item__actions">
            <button
                class="nickname-form__ghost"
                type="button"
                :disabled="buttonPage <= 1"
                @click="loadButtonPage(buttonPage - 1)"
            >
              上一页
            </button>
            <span>第 {{ buttonPage }} / {{ buttonTotalPages }} 页</span>
            <button
                class="nickname-form__ghost"
                type="button"
                :disabled="buttonPage >= buttonTotalPages"
                @click="loadButtonPage(buttonPage + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </section>

      <aside class="social-panel social-panel--ranking">
        <section class="social-card leaderboard-card leaderboard-card--stacked">
          <section class="leaderboard-card__section">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">Boss 伤害榜</p>
              <strong>{{ bossLeaderboard.length || 0 }} 人</strong>
            </div>

            <ol v-if="bossLeaderboard.length > 0" class="leaderboard-list">
              <li
                  v-for="entry in bossLeaderboard"
                  :key="entry.nickname"
                  class="leaderboard-list__item"
                  :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }"
              >
                <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
                <span class="leaderboard-list__name">{{ entry.nickname }}</span>
                <strong class="leaderboard-list__count">{{ entry.damage }}</strong>
              </li>
            </ol>
            <div v-else class="leaderboard-list leaderboard-list--empty">
              <p>当前 Boss 还没人动手，或者正在休战。</p>
            </div>
          </section>

          <section class="leaderboard-card__section">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">实时排行榜</p>
              <strong>前 {{ leaderboard.length || 0 }} 名</strong>
            </div>

            <ol v-if="leaderboard.length > 0" class="leaderboard-list">
              <li
                  v-for="entry in leaderboard"
                  :key="entry.nickname"
                  class="leaderboard-list__item"
                  :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }"
              >
                <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
                <span class="leaderboard-list__name">{{ entry.nickname }}</span>
                <strong class="leaderboard-list__count">{{ entry.clickCount }}</strong>
              </li>
            </ol>
            <div v-else class="leaderboard-list leaderboard-list--empty">
              <p>还没人上榜，等你来开张。</p>
            </div>
          </section>
        </section>
      </aside>
    </section>
  </main>
</template>
