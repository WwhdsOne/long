import {computed, onBeforeUnmount, onMounted, ref} from 'vue'

import {mergeBossState} from '../utils/bossState'
import {formatDropRate} from '../utils/buttonBoard'
import {mergeClickFallbackState} from '../utils/clickResponse'
import {formatRarityLabel, getRarityClassName, splitEquipmentName} from '../utils/rarity'
import {EQUIPMENT_SLOTS, normalizeLoadout} from '../utils/equipmentSlots'
import {createRealtimeTransport} from '../utils/realtimeTransport'

const ANNOUNCEMENT_READ_KEY = 'vote-wall-announcement-read'
const ANNOUNCEMENT_CACHE_KEY = 'vote-wall-announcement-cache'
const AUTO_CLICK_RATE_LABEL = '每秒固定 3 次'
const EQUIPMENT_ENHANCE_COST = 10
const GROWTH_FORMULA_TEXT = '点击 / 暴击单次成长 = ceil((当前点击 + 当前暴击 + 当前暴击率) / 4)，至少 +1'
const AFK_HEARTBEAT_INTERVAL_MS = 15000
const DAMAGE_PRIORITY = ['doomsday', 'judgement', 'weakCritical', 'critical', 'trueDamage', 'pursuit', 'normal']
const DAMAGE_VARIANTS = {
    normal: {
        scale: 1,
        ttl: 1250,
        shake: 0,
        stageFx: [],
    },
    pursuit: {
        scale: 1,
        ttl: 1300,
        shake: 0,
        stageFx: ['flash'],
    },
    trueDamage: {
        scale: 1,
        ttl: 1420,
        shake: 0,
        stageFx: ['flash'],
    },
    critical: {
        scale: 1.5,
        ttl: 1500,
        shake: 100,
        stageFx: ['shake'],
    },
    weakCritical: {
        scale: 2,
        ttl: 1600,
        shake: 150,
        stageFx: ['shake', 'flash'],
    },
    doomsday: {
        scale: 2.8,
        ttl: 1900,
        shake: 240,
        stageFx: ['shake', 'doom', 'blade'],
    },
    judgement: {
        scale: 4,
        ttl: 2200,
        shake: 180,
        stageFx: ['shake', 'slowMo', 'vignette'],
    },
}

const profilePageMap = {
    resources: 'resources',
    inventory: 'inventory',
    stats: 'stats',
    loadout: 'loadout',
}

// todo后续补上天赋
const publicPages = [
    {id: 'battle', label: '战斗', path: '/'},
    // {id: 'talents', label: '天赋', path: '/talents'},
    {id: 'resources', label: '资源', path: '/profile/resources'},
    {id: 'inventory', label: '背包', path: '/profile/inventory'},
    {id: 'stats', label: '属性', path: '/profile/stats'},
    {id: 'loadout', label: '装备栏', path: '/profile/loadout'},
    {id: 'messages', label: '消息', path: '/messages'},
]

const buttonTotalVotes = ref(0)
const leaderboard = ref([])
const boss = ref(null)
const bossLeaderboard = ref([])
const bossLoot = ref([])
const bossGoldRange = ref({min: 0, max: 0})
const bossStoneRange = ref({min: 0, max: 0})
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
const lastUpdatedAt = ref('')
const liveConnected = ref(false)
const damageBursts = ref({})
const damageStageFx = ref({
    shake: false,
    flash: false,
    doom: false,
    blade: false,
    slowMo: false,
    vignette: false,
})
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
const gold = ref(0)
const stones = ref(0)
const afkSettlement = ref(null)
const rewardModal = ref(null)
const currentPublicPage = ref(resolvePublicPage(window.location.pathname))

let realtimeTransport
let lastBossResourceVersion = ''
const burstTimers = new Map()
const stageFxTimers = new Map()
const burstFrameOffsets = new Map()
const pendingClickSources = new Map()
let rewardSignatureReady = false
let lastRewardSignature = ''
let lastKnownGold = 0
let lastKnownStones = 0
let presenceHeartbeatTimer = 0

const totalVotes = computed(() => buttonTotalVotes.value)
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
const canStartAutoClick = computed(() => isLoggedIn.value && Boolean(autoClickTargetKey.value))
const autoClickStatus = computed(() => {
    void autoClickEnabled.value
    void autoClickTargetKey.value
    void autoClickTargetLabel.value
    void canStartAutoClick.value
    void AUTO_CLICK_RATE_LABEL
    return '旧版挂机开关已下线，现改为离页自动挂机。'
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
    if (item?.critRate) parts.push(`暴击率 ${(item.critRate * 100).toFixed(1)}%`)
    if (item?.critDamageMultiplier) parts.push(`暴伤 ${item.critDamageMultiplier}`)
    if (item?.bossDamagePercent) parts.push(`首领伤 ${item.bossDamagePercent}%`)
    return parts.join(' ') || '无属性'
}

function formatItemStatLines(item) {
    const lines = []
    if (item?.attackPower) lines.push(`攻击力 ${item.attackPower}`)
    if (item?.armorPenPercent) lines.push(`护甲穿透 ${item.armorPenPercent}%`)
    if (item?.critRate) lines.push(`暴击率 ${(item.critRate * 100).toFixed(1)}%`)
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

function isProfilePublicPage(page) {
    return Boolean(profilePageMap[page])
}

function resolvePublicPage(pathname) {
    if (pathname.startsWith('/messages')) {
        return 'messages'
    }
    if (pathname.startsWith('/talents')) {
        return 'talents'
    }
    if (pathname.startsWith('/profile/resources')) {
        return 'resources'
    }
    if (pathname.startsWith('/profile/inventory')) {
        return 'inventory'
    }
    if (pathname.startsWith('/profile/stats')) {
        return 'stats'
    }
    if (pathname.startsWith('/profile/loadout')) {
        return 'loadout'
    }
    if (pathname.startsWith('/profile')) {
        return 'resources'
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
    if (isProfilePublicPage(page)) {
        // 资料子页整合为统一页面，不再通过独立资料接口刷新。
        return
    }
    if (page === 'messages') {
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

function rewardSignature(reward) {
    if (!reward || typeof reward !== 'object') {
        return ''
    }
    const itemID = String(reward.itemId || '').trim()
    const grantedAt = Number(reward.grantedAt || 0)
    const bossID = String(reward.bossId || '').trim()
    if (!itemID) {
        return ''
    }
    return `${bossID}:${itemID}:${grantedAt}`
}

function normalizeRewardList(list) {
    if (!Array.isArray(list)) {
        return []
    }
    return list
        .filter((item) => item && typeof item === 'object' && String(item.itemId || '').trim() !== '')
        .map((item) => ({
            bossId: String(item.bossId || '').trim(),
            bossName: String(item.bossName || '').trim(),
            itemId: String(item.itemId || '').trim(),
            itemName: String(item.itemName || item.itemId || '').trim(),
            grantedAt: Number(item.grantedAt || 0),
        }))
}

function rewardIconForItem(itemID) {
    const normalized = String(itemID || '').trim()
    if (!normalized) {
        return ''
    }
    const equipped = Object.values(loadout.value || {}).find((item) => item?.itemId === normalized)
    if (equipped?.imagePath) {
        return equipped.imagePath
    }
    const inventoryItem = inventory.value.find((item) => item?.itemId === normalized)
    return inventoryItem?.imagePath || ''
}

function buildRewardEntries(rewards) {
    return normalizeRewardList(rewards).map((reward) => ({
        ...reward,
        imagePath: rewardIconForItem(reward.itemId),
        imageAlt: reward.itemName || reward.itemId || '装备图标',
    }))
}

function openOnlineRewardModal(lastRewardItem, rewards, goldGain, stoneGain) {
    const rewardEntries = buildRewardEntries(rewards.length > 0 ? rewards : [lastRewardItem])
    rewardModal.value = {
        mode: 'online',
        title: '本次击杀战利品',
        bossName: lastRewardItem?.bossName || lastRewardItem?.bossId || boss.value?.name || '世界 Boss',
        kills: 1,
        goldTotal: Math.max(0, Number(goldGain || 0)),
        stoneTotal: Math.max(0, Number(stoneGain || 0)),
        rewards: rewardEntries,
    }
}

function openAfkRewardModal(settlement) {
    rewardModal.value = {
        mode: 'afk',
        title: '挂机战利品结算',
        bossName: '挂机期间',
        kills: Number(settlement?.kills || 0),
        goldTotal: Math.max(0, Number(settlement?.goldTotal || 0)),
        stoneTotal: Math.max(0, Number(settlement?.stoneTotal || 0)),
        rewards: buildRewardEntries(settlement?.rewards || []),
    }
}

function closeRewardModal() {
    rewardModal.value = null
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
        bossGoldRange.value = {min: 0, max: 0}
        bossStoneRange.value = {min: 0, max: 0}
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
        bossGoldRange.value = payload?.goldRange ?? {min: 0, max: 0}
        bossStoneRange.value = payload?.stoneRange ?? {min: 0, max: 0}
        lastBossResourceVersion = currentVersion
    } catch {
        if (force) {
            bossLoot.value = []
            bossGoldRange.value = {min: 0, max: 0}
            bossStoneRange.value = {min: 0, max: 0}
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

    applyPlayerProfileState(payload)
    applyBattleUserState(payload)
    syncing.value = false
    markUpdated()
}

function applyBattleUserState(payload) {
    if (!payload || typeof payload !== 'object') {
        return
    }
    const hasGold = 'gold' in payload
    const hasStones = 'stones' in payload
    const nextGold = hasGold ? Number(payload.gold ?? 0) : gold.value
    const nextStones = hasStones ? Number(payload.stones ?? 0) : stones.value

    if ('userStats' in payload) {
        userStats.value = payload.userStats ?? null
    }
    if ('myBossStats' in payload) {
        myBossStats.value = payload.myBossStats ?? null
    }
    if ('combatStats' in payload) {
        combatStats.value = payload.combatStats ?? defaultCombatStats()
    }
    if ('recentRewards' in payload) {
        recentRewards.value = Array.isArray(payload.recentRewards) ? payload.recentRewards : []
    }
    if ('lastReward' in payload) {
        lastReward.value = payload.lastReward ?? null
    }
    const signature = rewardSignature(lastReward.value)
    if (signature && rewardSignatureReady && signature !== lastRewardSignature) {
        const goldGain = hasGold ? Math.max(0, nextGold - lastKnownGold) : 0
        const stoneGain = hasStones ? Math.max(0, nextStones - lastKnownStones) : 0
        const rewards = Array.isArray(payload.recentRewards) && payload.recentRewards.length > 0
            ? payload.recentRewards
            : [lastReward.value]
        openOnlineRewardModal(lastReward.value, rewards, goldGain, stoneGain)
    }
    if (signature) {
        lastRewardSignature = signature
    }
    if (!rewardSignatureReady) {
        rewardSignatureReady = true
    }
    lastKnownGold = nextGold
    lastKnownStones = nextStones
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
    if ('gold' in payload) {
        gold.value = Number(payload.gold ?? 0)
    }
    if ('stones' in payload) {
        stones.value = Number(payload.stones ?? 0)
    }
}

function applyClickResult(payload) {
    if (!payload || typeof payload !== 'object') {
        return
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
    gold.value = 0
    stones.value = 0
    myBossStats.value = null
    recentRewards.value = []
    lastReward.value = null
    rewardModal.value = null
    rewardSignatureReady = false
    lastRewardSignature = ''
    lastKnownGold = 0
    lastKnownStones = 0
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

function resolveClickAckKey(payload) {
    const directKey = String(payload?.button?.key || payload?.key || '').trim()
    if (directKey) {
        return directKey
    }
    if (pendingKeys.value.size === 1) {
        return pendingKeys.value.values().next().value || ''
    }
    if (autoClickTargetKey.value && pendingKeys.value.has(autoClickTargetKey.value)) {
        return autoClickTargetKey.value
    }
    const firstPending = pendingKeys.value.values().next()
    return firstPending.done ? '' : String(firstPending.value || '').trim()
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
        onOnlineCount(payload) {
            const count = Number(payload?.count)
            if (Number.isFinite(count) && count >= 0) {
                onlineCount.value = count
            }
        },
        onClickAck(payload) {
            const key = resolveClickAckKey(payload)
            if (key) {
                triggerDamageBurst(key, payload)
            }
            clearPendingClicks(key)
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
    syncing.value = true
    if (!realtimeTransport) {
        ensureRealtimeTransport().connect({nickname: nextNickname})
        return
    }
    realtimeTransport.reconnect({nickname: nextNickname})
}

function clearDamageBurstTimer(id) {
    const timer = burstTimers.get(id)
    if (timer) {
        window.clearTimeout(timer)
        burstTimers.delete(id)
    }
}

function clearDamageBurst(key, burstID = '') {
    const normalizedKey = String(key || '').trim()
    if (!normalizedKey || !damageBursts.value[normalizedKey]) {
        return
    }
    if (!burstID) {
        damageBursts.value[normalizedKey].forEach((entry) => clearDamageBurstTimer(entry.id))
        const nextBursts = {...damageBursts.value}
        delete nextBursts[normalizedKey]
        damageBursts.value = nextBursts
        return
    }

    const remained = damageBursts.value[normalizedKey].filter((entry) => entry.id !== burstID)
    clearDamageBurstTimer(burstID)
    const nextBursts = {...damageBursts.value}
    if (remained.length === 0) {
        delete nextBursts[normalizedKey]
    } else {
        nextBursts[normalizedKey] = remained
    }
    damageBursts.value = nextBursts
}

function clearStageEffect(name) {
    const timer = stageFxTimers.get(name)
    if (timer) {
        window.clearTimeout(timer)
        stageFxTimers.delete(name)
    }
}

function triggerStageEffect(name, duration) {
    if (!Object.prototype.hasOwnProperty.call(damageStageFx.value, name)) {
        return
    }
    clearStageEffect(name)
    damageStageFx.value = {
        ...damageStageFx.value,
        [name]: true,
    }
    stageFxTimers.set(
        name,
        window.setTimeout(() => {
            damageStageFx.value = {
                ...damageStageFx.value,
                [name]: false,
            }
            stageFxTimers.delete(name)
        }, Math.max(80, Number(duration) || 0)),
    )
}

function parseButtonKey(key) {
    const matched = String(key || '').trim().match(/^boss-part:(\d+)-(\d+)$/)
    if (!matched) {
        return null
    }
    return {
        x: Number(matched[1]),
        y: Number(matched[2]),
    }
}

function findBossPartByKey(key) {
    const point = parseButtonKey(key)
    if (!point || !boss.value?.parts || !Array.isArray(boss.value.parts)) {
        return null
    }
    return boss.value.parts.find((part) => Number(part.x) === point.x && Number(part.y) === point.y) || null
}

function normalizeDamageVariant(rawType) {
    const normalized = String(rawType || '').trim().toLowerCase()
    if (!normalized) {
        return ''
    }
    const alias = {
        doomsday: 'doomsday',
        apocalypse: 'doomsday',
        maxhpcut: 'doomsday',
        max_hp_cut: 'doomsday',
        judgement: 'judgement',
        judgment: 'judgement',
        ultimate_critical: 'judgement',
        weak_critical: 'weakCritical',
        weakcritical: 'weakCritical',
        critical: 'critical',
        crit: 'critical',
        true: 'trueDamage',
        true_damage: 'trueDamage',
        truedamage: 'trueDamage',
        pursuit: 'pursuit',
        followup: 'pursuit',
        normal: 'normal',
    }
    return alias[normalized] || ''
}

function rankDamageVariant(variant) {
    const index = DAMAGE_PRIORITY.indexOf(variant)
    return index < 0 ? DAMAGE_PRIORITY.length : index
}

function resolveDamageVariant(payload, part, damageValue, source) {
    const explicit = normalizeDamageVariant(payload?.damageType || payload?.hitType || payload?.effectType)
    if (explicit) {
        return explicit
    }

    const critical = Boolean(payload?.critical)
    const weakCritical = critical && part?.type === 'weak'
    if (weakCritical) {
        return 'weakCritical'
    }
    if (critical) {
        return 'critical'
    }
    if (source === 'auto' || source === 'afk') {
        return 'pursuit'
    }
    void damageValue
    return 'normal'
}

function estimateBossDamage(payload, part) {
    const fromPayload = Number(payload?.bossDamage)
    if (Number.isFinite(fromPayload) && fromPayload > 0) {
        return Math.round(fromPayload)
    }
    const base = payload?.critical ? criticalDamage.value : normalDamage.value
    const partRatio = part?.type === 'weak' ? 2 : part?.type === 'heavy' ? 0.55 : 1
    return Math.max(1, Math.round((Number(base) || 1) * partRatio))
}

function nextBurstOffset(key, variant) {
    const now = Date.now()
    const frame = Math.floor(now / 90)
    const frameKey = `${key}:${frame}`
    const currentIndex = burstFrameOffsets.get(frameKey) || 0
    burstFrameOffsets.set(frameKey, currentIndex + 1)
    window.setTimeout(() => {
        if (burstFrameOffsets.get(frameKey) === currentIndex + 1) {
            burstFrameOffsets.delete(frameKey)
        }
    }, 260)

    const pattern = [
        [-66, -52],
        [58, -72],
        [-84, -98],
        [76, -118],
        [0, -136],
        [-104, -152],
        [95, -170],
    ]
    const base = pattern[currentIndex % pattern.length]
    const lane = Math.floor(currentIndex / pattern.length)
    const shift = lane * 22
    const variantBias = variant === 'doomsday' ? 16 : variant === 'judgement' ? 28 : 0
    return {
        x: base[0] + (currentIndex % 2 === 0 ? -shift : shift),
        y: base[1] - shift - variantBias,
    }
}

function buildDamageBurst(key, payload, part, source) {
    const damageValue = estimateBossDamage(payload, part)
    const variant = resolveDamageVariant(payload, part, damageValue, source)
    const config = DAMAGE_VARIANTS[variant] || DAMAGE_VARIANTS.normal
    const offset = nextBurstOffset(key, variant)
    return {
        id: `${key}-${variant}-${Date.now()}-${Math.floor(Math.random() * 100000)}`,
        type: variant,
        priority: rankDamageVariant(variant),
        value: formatNumber(damageValue),
        scale: config.scale,
        ttl: config.ttl,
        offsetX: offset.x,
        offsetY: offset.y,
    }
}

function triggerDamageBurst(key, payload = {}) {
    const normalizedKey = String(key || '').trim()
    if (!normalizedKey) {
        return
    }

    const source = pendingClickSources.get(normalizedKey) || 'normal'
    const part = findBossPartByKey(normalizedKey)
    const burst = buildDamageBurst(normalizedKey, payload, part, source)
    const config = DAMAGE_VARIANTS[burst.type] || DAMAGE_VARIANTS.normal
    const currentBursts = damageBursts.value[normalizedKey] || []
    const nextBursts = [...currentBursts, burst]
        .sort((left, right) => left.priority - right.priority)
        .slice(0, 6)
    damageBursts.value = {
        ...damageBursts.value,
        [normalizedKey]: nextBursts,
    }

    config.stageFx.forEach((fxName) => {
        const duration = fxName === 'slowMo' || fxName === 'vignette' ? 360 : 180
        triggerStageEffect(fxName, duration)
    })
    if (config.shake > 0) {
        triggerStageEffect('shake', config.shake)
    }

    clearDamageBurstTimer(burst.id)
    burstTimers.set(
        burst.id,
        window.setTimeout(() => {
            clearDamageBurst(normalizedKey, burst.id)
        }, config.ttl),
    )
}


function currentNicknameQuery() {
    return ''
}

function syncAutoClickTarget() {
    // 挂机模式已移除，保留函数防止旧调用报错。
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
    clearAutoClickLocalState()
}

async function syncAutoClickTargetOnServer(key) {
    void key
}

async function startAutoClick() {
    autoClickEnabled.value = false
}

async function stopAutoClick() {
    autoClickEnabled.value = false
}

async function toggleAutoClick() {
    if (autoClickEnabled.value) {
        await stopAutoClick()
        return
    }

    await startAutoClick()
}

async function loadState() {
    loading.value = true
    syncing.value = true

    try {
        const response = await fetch('/api/battle/state')
        if (!response.ok) {
            throw new Error('战斗状态加载失败')
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
    const matched = String(key || '').match(/^boss-part:(\d+)-(\d+)$/)
    if (!matched) {
        errorMessage.value = '仅支持攻击 Boss 部位。'
        return
    }
    const bossPartKey = `boss-part:${matched[1]}-${matched[2]}`

    const nextPending = new Set(pendingKeys.value)
    nextPending.add(key)
    pendingKeys.value = nextPending
    pendingClickSources.set(key, options.source || 'normal')
    errorMessage.value = ''

    try {
        const sent = ensureRealtimeTransport().sendClick(bossPartKey)
        if (!sent) {
            throw new Error('实时连接尚未建立，正在重连，请稍后再试。')
        }
    } catch (error) {
        clearPendingClicks(key)
        errorMessage.value = error.message || '点击失败，请稍后重试。'
    }
}

async function toggleItemEquip(instanceId, equipped) {
    if (!nickname.value || !instanceId) {
        return
    }

    const action = equipped ? 'unequip' : 'equip'
    actioningItemId.value = instanceId
    errorMessage.value = ''

    try {
        const response = await fetch(`/api/equipment/${encodeURIComponent(instanceId)}/${action}`, {
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
        applyUserState(data)
    } catch (error) {
        errorMessage.value = error.message || '装备操作失败，请稍后重试。'
    } finally {
        actioningItemId.value = ''
    }
}

async function salvageItem(instanceId) {
    if (!nickname.value || !instanceId) {
        return
    }

    actioningItemId.value = instanceId
    errorMessage.value = ''

    try {
        const response = await fetch(`/api/equipment/${encodeURIComponent(instanceId)}/salvage`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({nickname: nickname.value}),
        })

        if (!response.ok) {
            throw new Error(await readErrorMessage(response, '分解失败，请稍后重试。'))
        }

        // 分解接口返回结算结果，背包与属性由实时增量刷新。
        await response.json()
    } catch (error) {
        errorMessage.value = error.message || '分解失败，请稍后重试。'
    } finally {
        actioningItemId.value = ''
    }
}

async function enhanceItem(instanceId) {
    if (!nickname.value || !instanceId) {
        return
    }

    actioningItemId.value = instanceId
    errorMessage.value = ''

    try {
        const response = await fetch(`/api/equipment/${encodeURIComponent(instanceId)}/enhance`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({nickname: nickname.value}),
        })

        if (!response.ok) {
            const message = await readErrorMessage(response, '强化失败，请稍后重试。')
            errorMessage.value = message
            return {ok: false, message}
        }

        await response.json()
        return {ok: true}
    } catch (error) {
        const message = error.message || '强化失败，请稍后重试。'
        errorMessage.value = message
        return {ok: false, message}
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
        await reportPresence(true)
        await loadAfkSettlement()
        startPresenceHeartbeat()
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

    stopPresenceHeartbeat()
    await reportPresence(false)
    clearPlayerSessionState()
    connectRealtime('')
}

function clearPlayerSessionState() {
    stopPresenceHeartbeat()
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
        await reportPresence(true)
        await loadAfkSettlement()
        startPresenceHeartbeat()
    } catch {
        clearPlayerSessionState()
    }
}

async function reportPresence(visible) {
    if (!nickname.value) {
        return
    }
    try {
        await fetch('/api/player/presence', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({visible}),
        })
    } catch {
        // Presence 上报失败不阻断主流程。
    }
}

function stopPresenceHeartbeat() {
    if (!presenceHeartbeatTimer) {
        return
    }
    window.clearInterval(presenceHeartbeatTimer)
    presenceHeartbeatTimer = 0
}

function startPresenceHeartbeat() {
    if (!nickname.value || document.visibilityState !== 'visible' || presenceHeartbeatTimer) {
        return
    }
    presenceHeartbeatTimer = window.setInterval(() => {
        void reportPresence(true)
    }, AFK_HEARTBEAT_INTERVAL_MS)
}

async function loadAfkSettlement() {
    if (!nickname.value) {
        return
    }
    try {
        const response = await fetch('/api/player/afk/settlement')
        if (!response.ok) {
            return
        }
        const payload = await response.json()
        const kills = Number(payload?.kills ?? 0)
        const goldTotal = Number(payload?.goldTotal ?? 0)
        const stoneTotal = Number(payload?.stoneTotal ?? 0)
        const rewards = normalizeRewardList(payload?.rewards)
        if (kills <= 0 && goldTotal <= 0 && stoneTotal <= 0 && rewards.length === 0) {
            return
        }
        afkSettlement.value = {
            kills,
            goldTotal,
            stoneTotal,
            startedAt: Number(payload?.startedAt ?? 0),
            endedAt: Number(payload?.endedAt ?? 0),
            rewards,
        }
        openAfkRewardModal(afkSettlement.value)
    } catch {
        // 忽略结算拉取失败。
    }
}

function closeAfkSettlementModal() {
    afkSettlement.value = null
    closeRewardModal()
}

function registerPublicPageLifecycle() {
    const handleVisibilityChange = () => {
        const visible = document.visibilityState === 'visible'
        if (visible) {
            startPresenceHeartbeat()
        } else {
            stopPresenceHeartbeat()
        }
        void reportPresence(visible)
        if (visible) {
            void loadAfkSettlement()
        }
    }

    onMounted(async () => {
        restoreCachedLatestAnnouncement()
        window.addEventListener('popstate', handlePublicRouteChange)
        await loadPlayerSession()
        await loadState()
        await activatePublicPage(currentPublicPage.value)
        connectRealtime(nickname.value)
        document.addEventListener('visibilitychange', handleVisibilityChange)
        startPresenceHeartbeat()
        void reportPresence(true)
        void loadAfkSettlement()

        // 实时通道建立前先给出保守值，连接后会被实时在线人数事件覆盖。
        onlineCount.value = 1
    })

    onBeforeUnmount(() => {
        window.removeEventListener('popstate', handlePublicRouteChange)
        document.removeEventListener('visibilitychange', handleVisibilityChange)
        stopPresenceHeartbeat()
        realtimeTransport?.close()
        burstTimers.forEach((timer) => window.clearTimeout(timer))
        burstTimers.clear()
        stageFxTimers.forEach((timer) => window.clearTimeout(timer))
        stageFxTimers.clear()
        burstFrameOffsets.clear()
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
        buttonTotalVotes,
        leaderboard,
        boss,
        bossLeaderboard,
        bossLoot,
        bossGoldRange,
        bossStoneRange,
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
        lastUpdatedAt,
        liveConnected,
        damageBursts,
        damageStageFx,
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
        gold,
        stones,
        afkSettlement,
        rewardModal,
        currentPublicPage,
        lastBossResourceVersion,
        burstTimers,
        pendingClickSources,
        totalVotes,
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
        canStartAutoClick,
        autoClickStatus,
        bossStatusLabel,
        bossProgress,
        equippedItems,
        displayedRecentRewards,
        recentRewardTitle,
        recentRewardNote,
        filteredBossHistory,
        formatDropRate,
        formatRarityLabel,
        formatItemStats,
        formatItemStatLines,
        equipmentNameParts,
        equipmentNameClass,
        navigatePublicPage,
        formatTime,
        formatNumber,
        formatStatWithDelta,
        formatPercentWithDelta,
        readErrorMessage,
        closeAnnouncementModal,
        closeAfkSettlementModal,
        closeRewardModal,
        loadMessages,
        submitMessage,
        toggleAutoClick,
        clickButton,
        toggleItemEquip,
        salvageItem,
        enhanceItem,
        submitNickname,
        resetNickname,
        registerPublicPageLifecycle,
    }
}
