import {computed, onBeforeUnmount, onMounted, ref} from 'vue'

import {mergeBossState} from '../utils/bossState'
import {ratioPercent} from '../utils/formatNumber'
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
const TASK_POLL_INTERVAL_MS = 10000
const DAMAGE_PRIORITY = ['doomsday', 'judgement', 'weakCritical', 'critical', 'trueDamage', 'pursuit', 'heavy', 'normal']
const DAMAGE_VARIANTS = {
    normal: {
        scale: 1,
        ttl: 1400,
        shake: 0,
        stageFx: [],
        particles: 6,
        colors: ['#f8fafc', '#e2e8f0', '#cbd5e1'],
        label: '',
    },
    pursuit: {
        scale: 1,
        ttl: 1400,
        shake: 0,
        stageFx: ['flash'],
        particles: 8,
        colors: ['#60a5fa', '#93bbfd', '#3b82f6', '#bfdbfe'],
        label: '',
    },
    heavy: {
        scale: 0.9,
        ttl: 1200,
        shake: 0,
        stageFx: [],
        particles: 5,
        colors: ['#b0b0c0', '#9ca3af', '#787888', '#64748b'],
        label: '',
    },
    trueDamage: {
        scale: 1.05,
        ttl: 1200,
        shake: 0,
        stageFx: [],
        particles: 5,
        colors: ['#c084fc', '#a78bfa', '#8b5cf6', '#7c3aed'],
        label: '⚡',
    },
    critical: {
        scale: 1.5,
        ttl: 1600,
        shake: 100,
        stageFx: ['shake'],
        particles: 14,
        colors: ['#facc15', '#fbbf24', '#f59e0b', '#fef08a', '#eab308'],
        label: 'CRIT!',
    },
    weakCritical: {
        scale: 2.0,
        ttl: 1700,
        shake: 150,
        stageFx: ['shake', 'flash'],
        particles: 20,
        colors: ['#facc15', '#ef4444', '#f87171', '#fbbf24', '#f59e0b', '#dc2626'],
        label: 'WEAK!',
    },
    doomsday: {
        scale: 2.15,
        ttl: 2000,
        shake: 240,
        stageFx: ['shake', 'doom', 'blade'],
        particles: 28,
        colors: ['#c084fc', '#a78bfa', '#8b5cf6', '#7c3aed', '#ddd6fe', '#ede9fe'],
        label: '💀',
    },
    judgement: {
        scale: 2.75,
        ttl: 2400,
        shake: 180,
        stageFx: ['shake', 'slowMo', 'vignette'],
        particles: 36,
        colors: ['#fde047', '#facc15', '#fbbf24', '#fef08a', '#eab308', '#f59e0b'],
        label: 'K.O.',
    },
    bleed: {
        scale: 0.78,
        ttl: 280,
        shake: 0,
        stageFx: [],
        particles: 1,
        colors: ['#ef4444', '#dc2626', '#b91c1c'],
        label: '出血',
    },
}

const profilePageMap = {
    resources: 'resources',
    tasks: 'tasks',
    inventory: 'inventory',
    stats: 'stats',
    loadout: 'loadout',
}

const publicPages = [
    {id: 'battle', label: '战斗', path: '/'},
    {id: 'shop', label: '外观商店', path: '/shop'},
    {id: 'resources', label: '资源', path: '/profile/resources'},
    {id: 'inventory', label: '背包', path: '/profile/inventory'},
    {id: 'stats', label: '属性', path: '/profile/stats'},
    {id: 'loadout', label: '装备栏', path: '/profile/loadout'},
    {id: 'talents', label: '天赋', path: '/talents'},
    {id: 'tasks', label: '任务', path: '/profile/tasks'},
    {id: 'messages', label: '消息', path: '/messages'},
]

const buttonTotalVotes = ref(0)
const leaderboard = ref([])
const boss = ref(null)
const currentRoomId = ref('1')
const rooms = ref([])
const roomSwitching = ref(false)
const roomError = ref('')
const roomSwitchCooldownEndsAt = ref(0)

const bossLeaderboard = ref([])
const bossLoot = ref([])
const bossGoldRange = ref({min: 0, max: 0})
const bossStoneRange = ref({min: 0, max: 0})
const bossTalentPointsOnKill = ref(0)
const announcementVersion = ref('')
const latestAnnouncement = ref(null)
const announcements = ref([])
const myBossStats = ref(null)
const myBossDamageValue = ref(0)
const myBossKills = ref(0)
const totalBossKills = ref(0)
const bossLeaderboardCountValue = ref(-1)
const inventory = ref([])
const tasks = ref([])
const hasClaimableTasks = computed(() => tasks.value.some((item) => Boolean(item?.canClaim)))
const currentRoom = computed(() => rooms.value.find((item) => item.id === currentRoomId.value) || null)
const loadout = ref(emptyLoadout())
const combatStats = ref(defaultCombatStats())
const recentRewards = ref([])
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
function defaultTalentVisualState() {
  return {
    omenStacks: 0,
    omenCap: 150,
    silverStormActive: false,
    silverStormRemaining: 0,
    silverStormEndsAt: 0,
    silverStormDuration: 0,
    skinnerDurationByPart: {},
    skinnerCooldownEndsAt: 0,
    skinnerCooldownDuration: 0,
    collapsePartKeys: [],
    collapseEndsAt: 0,
    collapseDuration: 8,
    doomMarks: [],
    doomMarkCumDamage: {},
  }
}

const damageBursts = ref({})
const bleedBursts = ref({})
const talentTriggerFeed = ref([])
const bleedTriggerFeed = ref([])
const damageStageFx = ref({
    shake: false,
    flash: false,
    doom: false,
    blade: false,
    slowMo: false,
    vignette: false,
})
const talentVisualState = ref(defaultTalentVisualState())
const talentCombatState = ref(null)

// 连击计数系统
const COMBO_TIMEOUT_MS = 5000
const DEFAULT_STORM_TRIGGER = 100
const DEFAULT_ARMOR_TRIGGER = 100
const stormTrigger = computed(() => talentCombatState.value?.normalTriggerCount || DEFAULT_STORM_TRIGGER)
const armorTrigger = computed(() => talentCombatState.value?.armorTriggerCount || DEFAULT_ARMOR_TRIGGER)
const autoStrikeTrigger = computed(() => Math.max(0, Number(talentCombatState.value?.autoStrikeTriggerCount) || 0))
const judgmentDayTrigger = computed(() => Math.max(0, Number(talentCombatState.value?.judgmentDayTriggerCount) || 0))
const autoStrikeWindowSec = computed(() => Math.max(0, Number(talentCombatState.value?.autoStrikeWindowSec) || 0))
const nowTick = ref(0)
const comboCount = ref(0)
const stormCombo = ref(0)
const armorCombo = ref(0)
const comboLastClickAt = ref(0)
const comboTriggerFlash = ref(false)
const comboTimeoutPercent = ref(100)
let comboTimer = 0
let comboTickTimer = 0
let taskPollingTimer = 0

function partTypeForKey(key) {
  if (!key || !boss.value?.parts) return ''
  const m = String(key).match(/boss-part:(\d+)-(\d+)/)
  if (!m) return ''
  const x = parseInt(m[1]), y = parseInt(m[2])
  const part = boss.value.parts.find(p => p.x === x && p.y === y)
  return part?.type || ''
}

function advanceCombo(key, partType) {
  const now = Date.now()
  if (now - comboLastClickAt.value > COMBO_TIMEOUT_MS) { comboCount.value = 0 }
  comboCount.value++
  stormCombo.value++
  if (partType === 'heavy') armorCombo.value++
  comboLastClickAt.value = now
  scheduleComboClear()
  startComboTick()
}

function onStormComboTrigger() {
  comboTriggerFlash.value = true
  setTimeout(() => { comboTriggerFlash.value = false }, 800)
  stormCombo.value = 0
}

function onArmorTrigger() {
  armorCombo.value = 0
}

function clearComboState() {
  clearTimeout(comboTimer)
  clearInterval(comboTickTimer)
  comboTimeoutPercent.value = 100
  comboCount.value = 0
  comboLastClickAt.value = 0
  comboTriggerFlash.value = false
}

function startComboTick() {
  clearInterval(comboTickTimer)
  comboTickTimer = setInterval(() => {
    if (comboCount.value <= 0) { comboTimeoutPercent.value = 100; clearInterval(comboTickTimer); return }
    const elapsed = Date.now() - comboLastClickAt.value
    comboTimeoutPercent.value = Math.max(0, 100 - (elapsed / COMBO_TIMEOUT_MS) * 100)
  }, 200)
}

function scheduleComboClear() {
  clearTimeout(comboTimer)
  comboTimer = setTimeout(() => { if (Date.now() - comboLastClickAt.value >= COMBO_TIMEOUT_MS) clearComboState() }, COMBO_TIMEOUT_MS + 300)
}

const stormProgress = computed(() => Math.min(100, Math.round((stormCombo.value / stormTrigger.value) * 100)))
const armorProgress = computed(() => Math.min(100, Math.round((armorCombo.value / armorTrigger.value) * 100)))
const autoStrikeCountdown = computed(() => {
  void nowTick.value
  const expiresAt = Number(talentCombatState.value?.autoStrikeExpiresAt) || 0
  if (!expiresAt) return 0
  return Math.max(0, expiresAt - Date.now() / 1000)
})
const safeAutoStrikeCountdown = computed(() => Number.isFinite(autoStrikeCountdown.value) ? autoStrikeCountdown.value : 0)
const autoStrikeTimeoutPercent = computed(() => {
  const windowSec = autoStrikeWindowSec.value
  if (windowSec <= 0) return 0
  const ratio = safeAutoStrikeCountdown.value / windowSec
  return Number.isFinite(ratio)
    ? Math.min(100, Math.max(0, ratio * 100))
    : 0
})
const safeAutoStrikeTimeoutPercent = computed(() => Number.isFinite(autoStrikeTimeoutPercent.value) ? autoStrikeTimeoutPercent.value : 0)

const partProgressList = computed(() => {
  const parts = boss.value?.parts
  const cs = talentCombatState.value
  if (!Array.isArray(parts) || parts.length === 0) return []
  const stormMap = cs?.partStormComboCount || {}
  const heavyMap = cs?.partHeavyClickCount || {}
  const judgmentDayMap = cs?.partJudgmentDayCount || {}
  const jdUsedMap = cs?.judgmentDayUsed || {}
  const jdCooldown = Math.max(0, Number(cs?.judgmentDayCooldownSec) || 0)
  const autoStrikeTargetPart = String(cs?.autoStrikeTargetPart || '')
  const autoStrikeComboCount = Number(cs?.autoStrikeComboCount) || 0
  const jdTrigger = judgmentDayTrigger.value
  const nowSec = Date.now() / 1000
  const result = []
  for (const part of parts) {
    const key = `${part.x}-${part.y}`
    const storm = Number(stormMap[key]) || 0
    const armor = Number(heavyMap[key]) || 0
    const autoStrike = key === autoStrikeTargetPart ? autoStrikeComboCount : 0
    const rawJudgmentDay = Number(judgmentDayMap[key]) || 0
    const lastJdTrigger = Number(jdUsedMap[key]) || 0
    const jdOnCooldown = part.type === 'heavy' && lastJdTrigger > 0 && jdCooldown > 0 && (nowSec - lastJdTrigger) < jdCooldown
    const jdCount = (part.type === 'heavy' && !jdOnCooldown) ? rawJudgmentDay : 0
    const jdProgress = jdTrigger > 0 ? Math.min(100, Math.round((jdCount / jdTrigger) * 100)) : 0
    if (storm <= 0 && armor <= 0 && autoStrike <= 0 && jdCount <= 0) continue
    if (!part.alive) continue
    result.push({
      key,
      name: part.displayName || partTypeLabel(part.type),
      type: part.type,
      x: part.x,
      y: part.y,
      storm,
      stormProgress: Math.min(100, Math.round((storm / stormTrigger.value) * 100)),
      armor,
      armorProgress: Math.min(100, Math.round((armor / armorTrigger.value) * 100)),
      autoStrike,
      autoStrikeProgress: autoStrikeTrigger.value > 0
        ? Math.min(100, Math.round((autoStrike / autoStrikeTrigger.value) * 100))
        : 0,
      autoStrikeCountdown: autoStrike > 0 ? safeAutoStrikeCountdown.value : 0,
      autoStrikeTimeoutPercent: autoStrike > 0 ? safeAutoStrikeTimeoutPercent.value : 0,
      judgmentDay: jdCount,
      judgmentDayTrigger: jdTrigger,
      judgmentDayOnCooldown: jdOnCooldown,
      judgmentDayProgress: jdOnCooldown ? 100 : jdProgress,
      alive: part.alive,
    })
  }
  return result
})

const partStatusList = computed(() => {
  void nowTick.value
  const parts = boss.value?.parts
  const cs = talentCombatState.value
  if (!Array.isArray(parts) || parts.length === 0) return []
  const skinnerMap = cs?.skinnerParts || {}
  const bleedMap = cs?.bleeds && typeof cs.bleeds === 'object' ? cs.bleeds : {}
  const nowSec = Date.now() / 1000
  const nowMs = Date.now()
  const result = []
  const collapseEndsAt = Number(talentVisualState.value?.collapseEndsAt) || 0
  const collapsePartKeys = Array.isArray(talentVisualState.value?.collapsePartKeys)
    ? talentVisualState.value.collapsePartKeys
    : []
  const collapseRemainingSec = collapseEndsAt > nowSec ? Math.max(0, Math.ceil(collapseEndsAt - nowSec)) : 0
  const collapseRemainingMs = collapseEndsAt > nowSec ? Math.max(0, collapseEndsAt * 1000 - nowMs) : 0
  const collapseDurationSec = Math.max(0, Number(talentVisualState.value?.collapseDuration) || 0)
  const doomMarkKeys = Array.isArray(talentVisualState.value?.doomMarks)
    ? talentVisualState.value.doomMarks
    : []
  for (const part of parts) {
    if (!part.alive) continue
    const key = `${part.x}-${part.y}`
    if (collapseRemainingSec > 0 && collapsePartKeys.includes(key)) {
      result.push({
        key: `${key}:collapse`,
        partKey: key,
        name: part.displayName || partTypeLabel(part.type),
        type: part.type,
        statusKey: 'collapse',
        statusLabel: '护甲崩塌',
        remainingSec: collapseRemainingSec,
        progress: collapseDurationSec > 0
          ? Math.min(100, Math.max(0, (collapseRemainingMs / (collapseDurationSec * 1000)) * 100))
          : 0,
      })
    }
    const bleedState = bleedMap[key]
    const bleedEndsAtMs = Number(bleedState?.endsAtMs) || 0
    if (bleedEndsAtMs > nowMs) {
      const bleedDurationMs = Math.max(0, Number(bleedState?.durationMs) || 0)
      const bleedRemainingMs = Math.max(0, bleedEndsAtMs - nowMs)
      result.push({
        key: `${key}:bleed`,
        partKey: key,
        name: part.displayName || partTypeLabel(part.type),
        type: part.type,
        statusKey: 'bleed',
        statusLabel: '致命出血',
        remainingSec: Math.max(0, Math.ceil(bleedRemainingMs / 1000)),
        progress: bleedDurationMs > 0
          ? Math.min(100, Math.max(0, (bleedRemainingMs / bleedDurationMs) * 100))
          : 0,
      })
    }
    const endsAt = Number(skinnerMap[key]) || 0
    if (endsAt <= nowSec) continue
    const remainingMs = Math.max(0, endsAt * 1000 - nowMs)
    const skinnerDurationSec = Math.max(0, Number(cs?.skinnerDurationByPart?.[key]) || Number(talentVisualState.value?.skinnerDurationByPart?.[key]) || 0)
    result.push({
      key: `${key}:skinner`,
      partKey: key,
      name: part.displayName || partTypeLabel(part.type),
      type: part.type,
      statusKey: 'skinner',
      statusLabel: '变为弱点',
      remainingSec: Math.max(0, Math.ceil(endsAt - nowSec)),
      showProgress: false,
      progress: skinnerDurationSec > 0
        ? Math.min(100, Math.max(0, (remainingMs / (skinnerDurationSec * 1000)) * 100))
        : 0,
    })
  }
  for (const part of parts) {
    if (!part.alive) continue
    const key = `${part.x}-${part.y}`
    if (!doomMarkKeys.includes(key)) continue
    result.push({
      key: `${key}:doom-mark`,
      partKey: key,
      name: part.displayName || partTypeLabel(part.type),
      type: part.type,
      statusKey: 'doom-mark',
      statusLabel: '末日审判',
      statusMeta: '击碎后结算死兆',
      showCountdown: false,
      showProgress: false,
      remainingSec: 0,
      progress: 0,
    })
  }
	  // 审判日冷却状态
	  const jdCooldownCheckMap = cs?.judgmentDayUsed || {}
	  const jdCooldownCheckSec = Math.max(0, Number(cs?.judgmentDayCooldownSec) || 0)
	  if (jdCooldownCheckSec > 0) {
	    for (const part of parts) {
	      if (!part.alive) continue
	      if (part.type !== 'heavy') continue
	      const key = `${part.x}-${part.y}`
	      const lastTrigger = Number(jdCooldownCheckMap[key]) || 0
	      if (lastTrigger <= 0) continue
	      const remaining = Math.max(0, lastTrigger + jdCooldownCheckSec - nowSec)
	      if (remaining <= 0) continue
	      result.push({
	        key: `${key}:jd-cooldown`,
	        partKey: key,
	        name: part.displayName || partTypeLabel(part.type),
	        type: part.type,
	        statusKey: 'judgment-day',
	        statusLabel: '审判日',
	        statusMeta: '冷却中',
	        remainingSec: Math.ceil(remaining),
	        progress: Math.min(100, Math.max(0, ((jdCooldownCheckSec - remaining) / jdCooldownCheckSec) * 100)),
	        showCountdown: true,
	        showProgress: true,
	      })
	    }
	  }
  return result
})

const globalStatusList = computed(() => {
  void nowTick.value
  const result = []

  const comboValue = Number(comboCount.value) || 0
  const comboBonus = Math.floor(comboValue / 25) * 10
  const comboT = Math.min(comboValue / 200, 1)
  const comboHue = 120 - Math.min(comboValue / 200, 1) * 120
  const comboColor = `hsl(${comboHue}, 90%, ${55 - Math.min(comboValue / 200, 1) * 15}%)`
  const comboIsGold = comboValue >= 200
  const comboGoldGradientStops = ['#fde047', '#fbbf24', '#f59e0b', '#fef08a', '#fde047']
  const comboGoldGradient = `linear-gradient(135deg, ${comboGoldGradientStops.join(', ')})`
  const comboPanelStyle = comboValue > 0
    ? {
        borderColor: comboIsGold ? 'rgba(251, 191, 36, 0.6)' : `${comboColor}40`,
        boxShadow: comboIsGold
          ? '0 0 16px rgba(251, 191, 36, 0.2), inset 0 0 10px rgba(251, 191, 36, 0.06)'
          : `0 0 16px ${comboColor}22`,
      }
    : null
  const comboPrimaryStyle = comboValue > 0
    ? {
        background: comboIsGold ? comboGoldGradient : undefined,
        backgroundSize: comboIsGold ? '300% 300%' : undefined,
        WebkitBackgroundClip: comboIsGold ? 'text' : undefined,
        backgroundClip: comboIsGold ? 'text' : undefined,
        color: comboIsGold ? 'transparent' : comboColor,
        textShadow: comboIsGold ? 'none' : `0 0 14px ${comboColor}80`,
        filter: comboIsGold ? 'drop-shadow(0 0 8px rgba(251, 191, 36, 0.6)) drop-shadow(0 0 18px rgba(245, 158, 11, 0.3))' : undefined,
        animation: comboIsGold ? 'combo-gold-shimmer 2s linear infinite' : undefined,
      }
    : null
  const comboHintStyle = comboValue > 0
    ? {
        background: comboIsGold ? comboGoldGradient : undefined,
        backgroundSize: comboIsGold ? '300% 300%' : undefined,
        WebkitBackgroundClip: comboIsGold ? 'text' : undefined,
        backgroundClip: comboIsGold ? 'text' : undefined,
        color: comboIsGold ? 'transparent' : comboColor,
        textShadow: comboIsGold ? 'none' : `0 0 10px ${comboColor}55`,
        filter: comboIsGold ? 'drop-shadow(0 0 6px rgba(251, 191, 36, 0.45))' : undefined,
        animation: comboIsGold ? 'combo-gold-shimmer 2s linear infinite' : undefined,
      }
    : null
  const comboSecondaryStyle = comboBonus > 0
    ? {
        background: comboIsGold ? comboGoldGradient : undefined,
        backgroundSize: comboIsGold ? '300% 300%' : undefined,
        WebkitBackgroundClip: comboIsGold ? 'text' : undefined,
        backgroundClip: comboIsGold ? 'text' : undefined,
        color: comboIsGold ? 'transparent' : comboColor,
        textShadow: comboIsGold ? 'none' : `0 0 10px ${comboColor}55`,
        filter: comboIsGold ? 'drop-shadow(0 0 6px rgba(251, 191, 36, 0.45))' : undefined,
        animation: comboIsGold ? 'combo-gold-shimmer 2s linear infinite' : undefined,
      }
    : null
  const comboBarStyle = comboValue > 0
    ? {
        background: comboIsGold ? `linear-gradient(90deg, ${comboGoldGradientStops.join(', ')})` : `linear-gradient(90deg, ${comboColor}, ${comboColor}cc)`,
        backgroundSize: '300% 300%',
        boxShadow: comboIsGold ? '0 0 10px rgba(251, 191, 36, 0.65)' : `0 0 8px ${comboColor}80`,
        animation: 'combo-gold-shimmer 2s linear infinite',
      }
    : null
  result.push({
    key: 'combo',
    kind: 'combo',
    title: '连击',
    primary: `x${comboValue}`,
    secondary: comboBonus > 0 ? `伤害 +${comboBonus}%` : '',
    hint: comboValue > 0 ? `${Math.ceil(comboTimeoutPercent.value / 20)}s` : '待命',
    progress: comboValue > 0 ? Math.max(0, Number(comboTimeoutPercent.value) || 0) : 0,
    isGold: comboIsGold,
    panelStyle: comboPanelStyle,
    primaryStyle: comboPrimaryStyle,
    hintStyle: comboHintStyle,
    secondaryStyle: comboSecondaryStyle,
    barStyle: comboBarStyle,
  })

  const omenStacks = Math.max(0, Number(talentVisualState.value?.omenStacks) || 0)
  const omenCap = Math.max(1, Number(talentVisualState.value?.omenCap) || 150)
  const skinnerCooldownEndsAt = Math.max(0, Number(talentVisualState.value?.skinnerCooldownEndsAt) || 0)
  const skinnerCooldownRemaining = skinnerCooldownEndsAt > 0
    ? Math.max(0, Math.ceil(skinnerCooldownEndsAt - Date.now() / 1000))
    : 0
  const skinnerCooldownRemainingMs = skinnerCooldownEndsAt > 0
    ? Math.max(0, skinnerCooldownEndsAt * 1000 - Date.now())
    : 0
  const skinnerCooldownDuration = Math.max(0, Number(talentVisualState.value?.skinnerCooldownDuration) || 0)
  const skinnerActiveCount = partStatusList.value.filter((status) => status.statusKey === 'skinner').length
  if (skinnerCooldownRemaining > 0 || skinnerActiveCount > 0) {
    result.push({
      key: 'skinner',
      kind: 'skinner',
      title: '剥皮',
      primary: skinnerCooldownRemaining > 0 ? `${skinnerCooldownRemaining}s` : '待命',
      secondary: skinnerActiveCount > 0 ? `临时弱点 ${skinnerActiveCount} 处` : '暴击触发',
      hint: skinnerCooldownRemaining > 0 ? '冷却中' : '命中后制造弱点',
      showProgress: false,
      progress: skinnerCooldownDuration > 0
        ? Math.min(100, Math.max(0, (skinnerCooldownRemainingMs / (skinnerCooldownDuration * 1000)) * 100))
        : 0,
    })
  }
  if (omenStacks > 0) {
    result.push({
      key: 'omen',
      kind: 'omen',
      title: '死兆',
      primary: `${omenStacks} / ${omenCap}`,
      secondary: '',
      hint: `${omenCap} 层自动触发终末血斩`,
      progress: Math.min(100, Math.max(0, (omenStacks / omenCap) * 100)),
    })
  }

  const finalCutLastTriggerAt = Math.max(0, Number(talentCombatState.value?.lastFinalCutAt) || 0)
  const finalCutRecentWindowSec = 3
  const finalCutRecentlyTriggered = finalCutLastTriggerAt > 0
    ? Math.max(0, Math.ceil(finalCutLastTriggerAt + finalCutRecentWindowSec - Date.now() / 1000))
    : 0

  const silverStormEndsAt = Number(talentVisualState.value?.silverStormEndsAt) || 0
  const silverStormRemaining = silverStormEndsAt
    ? Math.max(0, Math.ceil(silverStormEndsAt - Date.now() / 1000))
    : Math.max(0, Number(talentVisualState.value?.silverStormRemaining) || 0)
  const silverStormRemainingMs = silverStormEndsAt
    ? Math.max(0, silverStormEndsAt * 1000 - Date.now())
    : Math.max(0, Number(talentVisualState.value?.silverStormRemaining) || 0) * 1000
  const silverStormDuration = Math.max(0, Number(talentVisualState.value?.silverStormDuration) || 0)
  const silverStormActive = Boolean(talentVisualState.value?.silverStormActive) && silverStormRemaining > 0
  if (silverStormActive) {
    result.push({
      key: 'silver-storm',
      kind: 'silver_storm',
      title: '白银风暴',
      primary: `${silverStormRemaining}s`,
      secondary: '最终伤害额外追加白银风暴',
      hint: '',
      progress: silverStormDuration > 0
        ? Math.min(100, Math.max(0, (silverStormRemainingMs / (silverStormDuration * 1000)) * 100))
        : 0,
      panelStyle: {
        borderColor: 'rgba(191, 219, 254, 0.36)',
        boxShadow: '0 0 16px rgba(191, 219, 254, 0.16)',
      },
      primaryStyle: {
        color: '#f8fafc',
        textShadow: '0 0 12px rgba(191, 219, 254, 0.5)',
      },
      secondaryStyle: {
        color: '#bfdbfe',
        textShadow: '0 0 10px rgba(125, 211, 252, 0.35)',
      },
      barStyle: {
        background: 'linear-gradient(90deg, #e2e8f0, #bfdbfe, #7dd3fc, #e2e8f0)',
        backgroundSize: '300% 300%',
        boxShadow: '0 0 10px rgba(125, 211, 252, 0.45)',
        animation: 'combo-gold-shimmer 2s linear infinite',
      },
    })
  }

  return result
})

function partTypeLabel(type) {
  const labels = { soft: '软组织', heavy: '重甲', weak: '弱点' }
  return labels[type] || type || '未知'
}

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
const shopItems = ref([])
const loadingShopItems = ref(false)
const gold = ref(0)
const stones = ref(0)
const talentPoints = ref(0)
const equippedBattleClickSkinId = ref('')
const equippedBattleClickCursorImagePath = ref('')
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
let lastRecentRewardSignature = ''
let lastKnownGold = 0
let lastKnownStones = 0
let presenceHeartbeatTimer = 0
let talentTickTimer = 0

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
const myBossDamage = computed(() => {
    if (Number.isFinite(myBossDamageValue.value)) {
        return myBossDamageValue.value
    }
    return myBossStats.value?.damage ?? 0
})
const bossLeaderboardCount = computed(() => {
    if (Number.isFinite(bossLeaderboardCountValue.value) && bossLeaderboardCountValue.value >= 0) {
        return bossLeaderboardCountValue.value
    }
    return bossLeaderboard.value.length
})
const myBossRank = computed(() => {
    if (!nickname.value || !boss.value) return null
    if (myBossStats.value?.rank) return myBossStats.value.rank
    const matched = bossLeaderboard.value.find((entry) => entry.nickname === nickname.value)
    return matched?.rank ?? null
})
const effectiveIncrement = computed(() => combatStats.value?.effectiveIncrement ?? 1)
const normalDamage = computed(() => combatStats.value?.normalDamage ?? effectiveIncrement.value)
const criticalDamage = computed(() => combatStats.value?.criticalDamage ?? normalDamage.value)

function setRoomSwitchCooldown(remainingSeconds) {
    const seconds = Math.max(0, Number(remainingSeconds ?? 0))
    roomSwitchCooldownEndsAt.value = seconds > 0 ? Date.now() + seconds * 1000 : 0
}

function applyRoomState(payload) {
    currentRoomId.value = String(payload?.currentRoomId || currentRoomId.value || '1')
    rooms.value = Array.isArray(payload?.rooms) ? payload.rooms : []
    setRoomSwitchCooldown(payload?.cooldownRemainingSeconds ?? payload?.switchCooldownRemainingSeconds ?? 0)
    roomError.value = ''
}

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

    return ratioPercent(boss.value.currentHp, boss.value.maxHp)
})
const loadoutSlots = EQUIPMENT_SLOTS
const equippedItems = computed(() => loadoutSlots.map((slot) => loadout.value[slot.value]).filter(Boolean))
const displayedRecentRewards = computed(() => normalizeRewardList(recentRewards.value))
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
        attackPower: 0,
        armorPenPercent: 0,
        critDamageMultiplier: 0,
        bossDamagePercent: 0,
        allDamageAmplify: 0,
        perPartDamagePercent: 0,
        lowHpMultiplier: 1,
        lowHpThreshold: 0,
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

function normalizeDisplayPercent(value) {
    const normalized = Number(value ?? 0)
    if (!Number.isFinite(normalized)) return 0
    return Math.abs(normalized) <= 1 ? normalized * 100 : normalized
}

function formatDisplayPercent(value) {
    return formatNumber(normalizeDisplayPercent(value), 2)
}

function formatItemStatLines(item) {
    const lines = []
    if (item?.attackPower) lines.push(`攻击力 ${formatNumber(item.attackPower)}`)
    if (item?.armorPenPercent) {
        lines.push(`护甲穿透 ${formatDisplayPercent(item.armorPenPercent)}%`)
    }
    if (item?.critRate) lines.push(`暴击率 ${formatDisplayPercent(item.critRate)}%`)
    if (item?.critDamageMultiplier) lines.push(`暴击倍率 +${formatDisplayPercent(item.critDamageMultiplier)}%`)
    if (item?.bossDamagePercent) lines.push(`首领伤害 ${formatDisplayPercent(item.bossDamagePercent)}%`)
    if (item?.partTypeDamageSoft) lines.push(`软组织伤害 ${formatDisplayPercent(item.partTypeDamageSoft)}%`)
    if (item?.partTypeDamageHeavy) lines.push(`重甲伤害 ${formatDisplayPercent(item.partTypeDamageHeavy)}%`)
    if (item?.partTypeDamageWeak) lines.push(`弱点伤害 ${formatDisplayPercent(item.partTypeDamageWeak)}%`)

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
    if (pathname.startsWith('/shop')) {
        return 'shop'
    }
    if (pathname.startsWith('/messages')) {
        return 'messages'
    }
    if (pathname.startsWith('/talents')) {
        return 'talents'
    }
    if (pathname.startsWith('/profile/resources')) {
        return 'resources'
    }
    if (pathname.startsWith('/profile/tasks')) {
        return 'tasks'
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
    if (page === 'shop') {
        await loadShopItems()
        return
    }
    if (isProfilePublicPage(page)) {
        try {
            await loadPlayerProfile()
        } catch (error) {
            errorMessage.value = error.message || '资料加载失败，请稍后重试。'
        }
        return
    }
    if (page === 'messages') {
        await loadMessages()
        await loadAnnouncements()
    }
}

async function loadPlayerProfile() {
    if (!nickname.value) {
        return
    }

    const response = await fetch('/api/player/profile')
    if (!response.ok) {
        throw new Error(await readErrorMessage(response, '资料加载失败，请稍后重试。'))
    }

    const payload = await response.json()
    applyPlayerProfileState(payload)
}

async function loadTasks() {
    if (!nickname.value) {
        tasks.value = []
        return
    }

    const response = await fetch('/api/tasks')
    if (!response.ok) {
        throw new Error(await readErrorMessage(response, '任务列表加载失败，请稍后重试。'))
    }

    const payload = await response.json()
    tasks.value = Array.isArray(payload) ? payload : []
}

async function loadShopItems() {
    loadingShopItems.value = true
    try {
        const response = await fetch('/api/shop/items')
        if (!response.ok) {
            throw new Error(await readErrorMessage(response, '商店列表加载失败，请稍后重试。'))
        }
        const payload = await response.json()
        shopItems.value = Array.isArray(payload) ? payload : []
    } finally {
        loadingShopItems.value = false
    }
}

function stopTaskPolling() {
    if (!taskPollingTimer) {
        return
    }
    window.clearInterval(taskPollingTimer)
    taskPollingTimer = 0
}

function startTaskPolling() {
    stopTaskPolling()
    if (!nickname.value) {
        return
    }
    void loadTasks().catch(() => {
        // 首次刷新失败不打断主流程，继续依赖后续轮询重试。
    })
    taskPollingTimer = window.setInterval(() => {
        void loadTasks().catch(() => {
            // 红点轮询失败不打断主流程，下次继续重试。
        })
    }, TASK_POLL_INTERVAL_MS)
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

function latestRewardFromList(list) {
    const normalized = normalizeRewardList(list)
    if (normalized.length === 0) {
        return null
    }
    return normalized[normalized.length - 1]
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

function openOnlineRewardModal(rewards, goldGain, stoneGain) {
    const rewardEntries = buildRewardEntries(rewards)
    const latestReward = rewardEntries.length > 0 ? rewardEntries[rewardEntries.length - 1] : null
    rewardModal.value = {
        mode: 'online',
        title: '本次击杀战利品',
        bossName: latestReward?.bossName || latestReward?.bossId || boss.value?.name || '世界 Boss',
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
        bossTalentPointsOnKill.value = 0
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
        const resourcesUrl = nickname.value ? `/api/boss/resources?nickname=${encodeURIComponent(nickname.value)}` : '/api/boss/resources'
        const response = await fetch(resourcesUrl)
        if (!response.ok) {
            throw new Error(await readErrorMessage(response, 'Boss 掉落池加载失败'))
        }
        const payload = await response.json()
        bossLoot.value = Array.isArray(payload?.bossLoot) ? payload.bossLoot : []
        bossGoldRange.value = payload?.goldRange ?? {min: 0, max: 0}
        bossStoneRange.value = payload?.stoneRange ?? {min: 0, max: 0}
        bossTalentPointsOnKill.value = Math.max(0, Number(payload?.talentPointsOnKill ?? 0))
        lastBossResourceVersion = currentVersion
    } catch {
        if (force) {
            bossLoot.value = []
            bossGoldRange.value = {min: 0, max: 0}
            bossStoneRange.value = {min: 0, max: 0}
            bossTalentPointsOnKill.value = 0
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
        bossHistory.value = Array.isArray(payload) ? payload.map((entry) => mergeBossState(null, entry)) : []
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
    if ('roomId' in payload) {
        currentRoomId.value = String(payload.roomId || currentRoomId.value || '1')
    }
    const previousBoss = boss.value
    if ('boss' in payload) {
        boss.value = mergeBossState(boss.value, payload.boss)
    }
    applyPublicMeta(payload)
    if ('bossGoldRange' in payload && payload.bossGoldRange) {
        bossGoldRange.value = payload.bossGoldRange
    }
    if ('bossStoneRange' in payload && payload.bossStoneRange) {
        bossStoneRange.value = payload.bossStoneRange
    }
    if ('bossTalentPointsOnKill' in payload) {
        bossTalentPointsOnKill.value = Math.max(0, Number(payload.bossTalentPointsOnKill ?? 0))
    }
    if (bossResourceVersion(previousBoss) !== bossResourceVersion()) {
        void loadBossResources(true)
        if (previousBoss?.id && boss.value?.id && previousBoss.id !== boss.value.id) {
            clearTalentVisualState()
            clearComboState()
            talentCombatState.value = null
        }
    } else if (boss.value?.id && !lastBossResourceVersion) {
        void loadBossResources(true)
    }
    syncing.value = false
    markUpdated()
}

function applyPublicMeta(payload) {
    if (!payload || typeof payload !== 'object') {
        return
    }

    if ('leaderboard' in payload) {
        const nextLeaderboard = Array.isArray(payload.leaderboard) ? payload.leaderboard : null
        if (nextLeaderboard) {
            leaderboard.value = nextLeaderboard
        }
    }
    if ('bossLeaderboard' in payload) {
        bossLeaderboard.value = Array.isArray(payload.bossLeaderboard) ? payload.bossLeaderboard : []
        bossLeaderboardCountValue.value = bossLeaderboard.value.length
    }
    if ('bossLeaderboardCount' in payload) {
        bossLeaderboardCountValue.value = Math.max(0, Number(payload.bossLeaderboardCount ?? 0))
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
    if ('roomId' in payload) {
        currentRoomId.value = String(payload.roomId || currentRoomId.value || '1')
    }
    const hasStones = 'stones' in payload
    const hasTalentPoints = 'talentPoints' in payload
    const nextGold = hasGold ? Number(payload.gold ?? 0) : gold.value
    const nextStones = hasStones ? Number(payload.stones ?? 0) : stones.value
    const nextTalentPoints = hasTalentPoints ? Number(payload.talentPoints ?? 0) : talentPoints.value

    if ('userStats' in payload) {
        userStats.value = payload.userStats ?? null
    }
    if ('myBossKills' in payload) {
        myBossKills.value = Math.max(0, Number(payload.myBossKills ?? 0))
    }
    if ('totalBossKills' in payload) {
        totalBossKills.value = Math.max(0, Number(payload.totalBossKills ?? 0))
    }
    if ('myBossStats' in payload) {
        myBossStats.value = payload.myBossStats ?? null
        myBossDamageValue.value = myBossStats.value?.damage ?? 0
    }
    if ('myBossDamage' in payload) {
        myBossDamageValue.value = Math.max(0, Number(payload.myBossDamage ?? 0))
    }
    if ('combatStats' in payload) {
        combatStats.value = payload.combatStats ?? defaultCombatStats()
    }
    if ('recentRewards' in payload) {
        recentRewards.value = Array.isArray(payload.recentRewards) ? payload.recentRewards : []
    }
    if (Array.isArray(payload.talentEvents) && payload.talentEvents.length > 0) {
        triggerTalentEventDamageBursts(payload.talentEvents)
        appendTalentTriggerEvents(payload.talentEvents)
    }
    if ('talentCombatState' in payload && payload.talentCombatState) {
        applyTalentCombatState(payload.talentCombatState)
    }
    const latestReward = latestRewardFromList(recentRewards.value)
    const signature = rewardSignature(latestReward)
    if (signature && rewardSignatureReady && signature !== lastRecentRewardSignature) {
        const goldGain = hasGold ? Math.max(0, nextGold - lastKnownGold) : 0
        const stoneGain = hasStones ? Math.max(0, nextStones - lastKnownStones) : 0
        const rewards = normalizeRewardList(recentRewards.value)
        openOnlineRewardModal(rewards, goldGain, stoneGain)
    }
    if (signature) {
        lastRecentRewardSignature = signature
    }
    if (!rewardSignatureReady) {
        rewardSignatureReady = true
    }
    lastKnownGold = nextGold
    lastKnownStones = nextStones
    talentPoints.value = nextTalentPoints
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
    if ('talentPoints' in payload) {
        talentPoints.value = Number(payload.talentPoints ?? 0)
    }
    if ('tasks' in payload) {
        tasks.value = Array.isArray(payload.tasks) ? payload.tasks : []
    }
    if ('equippedBattleClickSkinId' in payload) {
        equippedBattleClickSkinId.value = payload.equippedBattleClickSkinId || ''
    }
    if ('equippedBattleClickCursorImagePath' in payload) {
        equippedBattleClickCursorImagePath.value = payload.equippedBattleClickCursorImagePath || ''
    }
}

async function claimTask(taskId) {
    if (!nickname.value) {
        errorMessage.value = '先登录账号再领取任务奖励。'
        return {ok: false, message: errorMessage.value}
    }

    const response = await fetch(`/api/tasks/${encodeURIComponent(taskId)}/claim`, {
        method: 'POST',
    })
    if (!response.ok) {
        const message = await readErrorMessage(response, '任务奖励领取失败，请稍后重试。')
        errorMessage.value = message
        return {ok: false, message}
    }

    const payload = await response.json()
    applyUserState(payload)
    if (!('tasks' in payload)) {
        await loadTasks()
    }
    errorMessage.value = ''
    return {ok: true}
}

async function purchaseShopItem(itemId) {
    if (!nickname.value) {
        errorMessage.value = '先登录账号再购买。'
        return {ok: false, message: errorMessage.value}
    }

    try {
        const response = await fetch(`/api/shop/items/${encodeURIComponent(itemId)}/purchase`, {
            method: 'POST',
        })
        if (!response.ok) {
            const message = await readErrorMessage(response, '购买失败，请稍后重试。')
            errorMessage.value = message
            return {ok: false, message}
        }
        const payload = await response.json()
        applyUserState(payload.userState)
        await loadShopItems()
        errorMessage.value = ''
        return {ok: true}
    } catch (error) {
        const message = error.message || '购买失败，请稍后重试。'
        errorMessage.value = message
        return {ok: false, message}
    }
}

async function equipShopItem(itemId) {
    if (!nickname.value) {
        errorMessage.value = '先登录账号再使用。'
        return {ok: false, message: errorMessage.value}
    }

    try {
        const response = await fetch(`/api/shop/items/${encodeURIComponent(itemId)}/equip`, {
            method: 'POST',
        })
        if (!response.ok) {
            const message = await readErrorMessage(response, '切换失败，请稍后重试。')
            errorMessage.value = message
            return {ok: false, message}
        }
        const payload = await response.json()
        applyUserState(payload.userState)
        await loadShopItems()
        errorMessage.value = ''
        return {ok: true}
    } catch (error) {
        const message = error.message || '切换失败，请稍后重试。'
        errorMessage.value = message
        return {ok: false, message}
    }
}

async function unequipShopItem() {
    if (!nickname.value) {
        errorMessage.value = '先登录账号再操作。'
        return {ok: false, message: errorMessage.value}
    }

    try {
        const response = await fetch('/api/shop/items/unequip', {
            method: 'POST',
        })
        if (!response.ok) {
            const message = await readErrorMessage(response, '恢复默认失败，请稍后重试。')
            errorMessage.value = message
            return {ok: false, message}
        }
        const payload = await response.json()
        applyUserState(payload.userState)
        await loadShopItems()
        errorMessage.value = ''
        return {ok: true}
    } catch (error) {
        const message = error.message || '恢复默认失败，请稍后重试。'
        errorMessage.value = message
        return {ok: false, message}
    }
}

function applyClickResult(payload) {
    if (!payload || typeof payload !== 'object') {
        return
    }

    buttonTotalVotes.value = Math.max(0, buttonTotalVotes.value + Number(payload.delta || 0))
    if (payload.userDelta && typeof payload.userDelta === 'object') {
        if (payload.userDelta.gold !== undefined) {
            gold.value = Number(payload.userDelta.gold)
        }
        if (payload.userDelta.stones !== undefined) {
            stones.value = Number(payload.userDelta.stones)
        }
        if (payload.userDelta.talentPoints !== undefined) {
            talentPoints.value = Number(payload.userDelta.talentPoints)
        }
    }
    const nextClickState = mergeClickFallbackState(
        {
            userStats: userStats.value,
            boss: boss.value,
            bossLeaderboard: bossLeaderboard.value,
            bossLeaderboardCount: bossLeaderboardCountValue.value,
            myBossStats: myBossStats.value,
            myBossDamage: myBossDamageValue.value,
            recentRewards: recentRewards.value,
        },
        payload,
    )
    userStats.value = nextClickState.userStats
    boss.value = nextClickState.boss
    bossLeaderboard.value = nextClickState.bossLeaderboard
    bossLeaderboardCountValue.value = nextClickState.bossLeaderboardCount
    myBossStats.value = nextClickState.myBossStats
    myBossDamageValue.value = nextClickState.myBossDamage
    recentRewards.value = nextClickState.recentRewards
    triggerTalentEventDamageBursts(payload.talentEvents)
    appendTalentTriggerEvents(payload.talentEvents)
    if (payload.talentCombatState) {
      applyTalentCombatState(payload.talentCombatState)
    }
    syncing.value = false
    markUpdated()
}

function indexToPartKey(index) {
  const parts = boss.value?.parts
  if (!Array.isArray(parts) || index < 0 || index >= parts.length) return null
  return `${parts[index].x}-${parts[index].y}`
}

function applyTalentCombatState(state) {
  if (!state || typeof state !== 'object') return
  talentCombatState.value = state
  const vs = talentVisualState.value
  vs.omenStacks = Number(state.omenStacks) || 0
  vs.omenCap = 150

  const prevSilverStormEndsAt = Number(vs.silverStormEndsAt) || 0
  const prevSilverStormActive = Boolean(vs.silverStormActive)
  vs.silverStormEndsAt = Number(state.silverStormEndsAt) || 0
  vs.silverStormActive = Boolean(state.silverStormActive) && (!vs.silverStormEndsAt || vs.silverStormEndsAt > Date.now() / 1000)
  vs.silverStormRemaining = vs.silverStormEndsAt
    ? Math.max(0, Math.ceil(vs.silverStormEndsAt - Date.now() / 1000))
    : (Number(state.silverStormRemaining) || 0)
  const shouldResetSilverStormDuration = vs.silverStormActive && (!prevSilverStormActive || vs.silverStormEndsAt > prevSilverStormEndsAt)
  vs.silverStormDuration = vs.silverStormActive
    ? (shouldResetSilverStormDuration
        ? Math.max(Number(state.silverStormRemaining) || 0, vs.silverStormRemaining)
        : Math.max(Number(vs.silverStormDuration) || 0, Number(state.silverStormRemaining) || 0, vs.silverStormRemaining))
    : 0

  vs.skinnerCooldownEndsAt = Number(state.skinnerCooldownEndsAt) || 0
  vs.skinnerCooldownDuration = Number(state.skinnerCooldownDuration) || 0
  const skinnerParts = state.skinnerParts && typeof state.skinnerParts === 'object' ? state.skinnerParts : {}
  const skinnerDurationByPart = state.skinnerDurationByPart && typeof state.skinnerDurationByPart === 'object'
    ? state.skinnerDurationByPart
    : {}
  const nextSkinnerDurationByPart = {}
  const nowSec = Date.now() / 1000
  for (const [partKey, endsAtRaw] of Object.entries(skinnerParts)) {
    const endsAt = Number(endsAtRaw) || 0
    if (endsAt <= nowSec) continue
    const incomingDurationSec = Math.max(0, Number(skinnerDurationByPart[partKey]) || 0)
    const cachedDurationSec = Math.max(0, Number(vs.skinnerDurationByPart?.[partKey]) || 0)
    const observedRemainingSec = Math.max(0, Math.ceil(endsAt - nowSec))
    nextSkinnerDurationByPart[partKey] = incomingDurationSec || cachedDurationSec || observedRemainingSec
  }
  vs.skinnerDurationByPart = nextSkinnerDurationByPart
  vs.doomMarkCumDamage = state.doomMarkCumDamage || {}

  // doomMarks: indices → "x-y" keys
  vs.doomMarks = Array.isArray(state.doomMarks)
    ? state.doomMarks.map(indexToPartKey).filter(Boolean)
    : []
  // collapsePartKeys: indices → "x-y" keys
  vs.collapsePartKeys = Array.isArray(state.collapseParts)
    ? state.collapseParts.map(indexToPartKey).filter(Boolean)
    : []
  vs.collapseEndsAt = Number(state.collapseEndsAt) || 0
  vs.collapseDuration = Number(state.collapseDuration) || 8
}

function applyTalentVisualState(events) {
    if (!Array.isArray(events)) return
    const vs = talentVisualState.value
    for (const event of events) {
        switch (event?.effectType) {
            case 'storm_combo':
                onStormComboTrigger()
                break
            case 'silver_storm':
                onStormComboTrigger()
                break
            case 'collapse_trigger':
                onArmorTrigger()
                if (event.partX !== undefined && event.partY !== undefined) {
                    const key = `${event.partX}-${event.partY}`
                    if (!vs.collapsePartKeys.includes(key)) {
                        vs.collapsePartKeys.push(key)
                    }
                }
                break
            case 'doom_mark':
                break
            case 'auto_strike':
            case 'bleed':
            case 'omen_harvest':
            case 'final_cut':
                break
        }
    }
}

function appendBleedTriggerEvents(events) {
    if (!Array.isArray(events) || events.length === 0) {
        return
    }
    const now = Date.now()
    const nextEntries = events
        .filter((event) => event?.effectType === 'bleed')
        .map((event, index) => ({
            id: `${now}-${index}-${event.talentId || 'talent'}`,
            name: event.name || event.talentId || '天赋',
            message: event.message || '天赋触发',
            extraDamage: Number(event.extraDamage || 0),
            triggeredAt: Number(event.triggeredAt || now),
            effectType: event.effectType || '',
            partX: event.partX,
            partY: event.partY,
        }))
    if (nextEntries.length === 0) {
        return
    }
    bleedTriggerFeed.value = [
        ...nextEntries,
        ...bleedTriggerFeed.value,
    ]
        .sort((left, right) => Number(right.triggeredAt || 0) - Number(left.triggeredAt || 0))
        .slice(0, 12)
}

function appendTalentTriggerEvents(events) {
    if (!Array.isArray(events) || events.length === 0) {
        return
    }
    const now = Date.now()
    const nextEntries = events.map((event, index) => ({
            id: `${now}-${index}-${event.talentId || 'talent'}`,
            name: event.name || event.talentId || '天赋',
            message: event.message || '天赋触发',
            extraDamage: Number(event.extraDamage || 0),
            triggeredAt: Number(event.triggeredAt || now),
            effectType: event.effectType || '',
            partX: event.partX,
            partY: event.partY,
        }))
    const mergedFeed = [
        ...nextEntries,
        ...talentTriggerFeed.value,
    ]
    const otherEvents = mergedFeed.filter((entry) => entry.effectType !== 'bleed').slice(0, 6)
    talentTriggerFeed.value = otherEvents
        .sort((left, right) => Number(right.triggeredAt || 0) - Number(left.triggeredAt || 0))
        .slice(0, 6)
    appendBleedTriggerEvents(events)
    applyTalentVisualState(events)
}

function triggerTalentEventDamageBursts(events) {
    if (!Array.isArray(events) || events.length === 0) {
        return
    }
    for (const event of events) {
        if (event.effectType !== 'bleed' || Number(event.extraDamage || 0) <= 0) {
            continue
        }
        if (event.partX === undefined || event.partY === undefined) {
            continue
        }
        triggerBleedBurst(`boss-part:${event.partX}-${event.partY}`, {
            bossDamage: Number(event.extraDamage || 0),
            damageType: 'bleed',
            effectType: 'bleed',
        })
    }
}

function clearTalentVisualState() {
    talentVisualState.value = defaultTalentVisualState()
}

function clearUserRealtimeState() {
    userStats.value = null
    inventory.value = []
    tasks.value = []
    loadout.value = emptyLoadout()
    combatStats.value = defaultCombatStats()
    gold.value = 0
    stones.value = 0
    talentPoints.value = 0
    equippedBattleClickSkinId.value = ''
    equippedBattleClickCursorImagePath.value = ''
    myBossStats.value = null
    myBossDamageValue.value = 0
    bossLeaderboardCountValue.value = -1
    recentRewards.value = []
    rewardModal.value = null
    rewardSignatureReady = false
    lastRecentRewardSignature = ''
    lastKnownGold = 0
    lastKnownStones = 0
    clearTalentVisualState()
    bleedBursts.value = {}
    bleedTriggerFeed.value = []
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
        onPublicMeta(payload) {
            applyPublicMeta(payload)
            loading.value = false
            errorMessage.value = ''
        },
        onUserDelta(payload) {
            applyUserState(payload)
            loading.value = false
            errorMessage.value = ''
        },
        onRoomState(payload) {
            applyRoomState(payload)
            loading.value = false
            roomError.value = ''
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
                advanceCombo(key, partTypeForKey(key))
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

function clearBleedBurst(key, burstID = '') {
    const normalizedKey = String(key || '').trim()
    if (!normalizedKey || !bleedBursts.value[normalizedKey]) {
        return
    }
    if (!burstID) {
        bleedBursts.value[normalizedKey].forEach((entry) => clearDamageBurstTimer(entry.id))
        const nextBursts = {...bleedBursts.value}
        delete nextBursts[normalizedKey]
        bleedBursts.value = nextBursts
        return
    }

    const remained = bleedBursts.value[normalizedKey].filter((entry) => entry.id !== burstID)
    clearDamageBurstTimer(burstID)
    const nextBursts = {...bleedBursts.value}
    if (remained.length === 0) {
        delete nextBursts[normalizedKey]
    } else {
        nextBursts[normalizedKey] = remained
    }
    bleedBursts.value = nextBursts
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
        bleed: 'bleed',
        weak_critical: 'weakCritical',
        weakcritical: 'weakCritical',
        critical: 'critical',
        crit: 'critical',
        heavy: 'heavy',
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
    if (part?.x != null && part?.y != null && talentVisualState.value.collapsePartKeys.includes(`${part.x}-${part.y}`)) {
        return 'trueDamage'
    }
    if (part?.type === 'heavy') {
        return 'heavy'
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

    if (variant === 'bleed') {
        const bleedLane = currentIndex % 6
        const bleedOffsets = [
            [-18, -28],
            [18, -42],
            [-22, -56],
            [22, -70],
            [-14, -84],
            [14, -98],
        ]
        const base = bleedOffsets[bleedLane]
        const stack = Math.floor(currentIndex / bleedOffsets.length)
        return {
            x: base[0],
            y: base[1] - stack * 18,
        }
    }

    const pattern = [
        [0, -36],
        [0, -52],
        [0, -44],
        [0, -60],
        [0, -48],
        [0, -56],
        [0, -40],
    ]
    const base = pattern[currentIndex % pattern.length]
    const lane = Math.floor(currentIndex / pattern.length)
    const shift = lane * 22
    const variantBias = variant === 'doomsday' ? 16 : variant === 'judgement' ? 28 : 0
    return {
        x: 0,
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
        label: config.label || '',
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
    burst.particleCount = config.particles || 0
    burst.particleColors = config.colors || []
    burst.particles = []
    for (let i = 0; i < burst.particleCount; i++) {
      const angle = Math.random() * Math.PI * 2
      const dist = 18 + Math.random() * 48
      const size = 2 + Math.random() * 4
      burst.particles.push({
        id: `${burst.id}-p${i}`,
        color: burst.particleColors[Math.floor(Math.random() * burst.particleColors.length)],
        x: Math.cos(angle) * dist,
        y: Math.sin(angle) * dist,
        size,
      })
    }
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

function triggerBleedBurst(partKey, payload = {}) {
    const normalizedKey = String(partKey || '').trim()
    if (!normalizedKey) {
        return
    }

    const part = findBossPartByKey(normalizedKey)
    const burst = buildDamageBurst(normalizedKey, {
        ...payload,
        damageType: 'bleed',
        effectType: 'bleed',
    }, part, 'bleed')
    const currentBursts = bleedBursts.value[normalizedKey] || []
    const nextBursts = [...currentBursts, burst].slice(-6)
    bleedBursts.value = {
        ...bleedBursts.value,
        [normalizedKey]: nextBursts,
    }

    clearDamageBurstTimer(burst.id)
    burstTimers.set(
        burst.id,
        window.setTimeout(() => {
            clearBleedBurst(normalizedKey, burst.id)
        }, burst.ttl),
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
        const stateUrl = nickname.value ? `/api/battle/state?nickname=${encodeURIComponent(nickname.value)}` : '/api/battle/state'
        const response = await fetch(stateUrl)
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

async function loadRooms() {
    try {
        const roomsUrl = nickname.value ? `/api/rooms?nickname=${encodeURIComponent(nickname.value)}` : '/api/rooms'
        const response = await fetch(roomsUrl)
        if (!response.ok) {
            throw new Error('房间列表加载失败')
        }
        applyRoomState(await response.json())
    } catch (error) {
        roomError.value = error.message || '房间列表加载失败。'
    }
}

async function joinRoom(roomId) {
    const targetRoomId = String(roomId || '').trim()
    if (!nickname.value) {
        roomError.value = '请先登录后再切换房间。'
        return {ok: false, message: roomError.value}
    }
    if (!targetRoomId || targetRoomId === currentRoomId.value) {
        return {ok: true}
    }
    const targetRoom = rooms.value.find((item) => item.id === targetRoomId)
    if (targetRoom && targetRoom.joinable === false) {
        roomError.value = '这个房间暂时不可加入。'
        return {ok: false, message: roomError.value}
    }

    roomSwitching.value = true
    roomError.value = ''
    try {
        const response = await fetch('/api/rooms/join', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                nickname: nickname.value,
                roomId: targetRoomId,
            }),
        })
        if (!response.ok) {
            const payload = await response.json().catch(() => ({}))
            const message = payload?.message || (
                payload?.error === 'ROOM_SWITCH_COOLDOWN'
                    ? '切房冷却中。'
                    : '切换房间失败。'
            )
            throw new Error(message)
        }
        const payload = await response.json()
        applyRoomState({
            currentRoomId: payload?.currentRoomId || targetRoomId,
            rooms: Array.isArray(payload?.rooms) ? payload.rooms : rooms.value,
            cooldownRemainingSeconds: payload?.cooldownRemainingSeconds ?? payload?.switchCooldownRemainingSeconds ?? 0,
        })
        connectRealtime(nickname.value)
        await loadState()
        return {ok: true}
    } catch (error) {
        roomError.value = error.message || '切换房间失败。'
        return {ok: false, message: roomError.value}
    } finally {
        roomSwitching.value = false
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
        const sent = ensureRealtimeTransport().sendClick(bossPartKey, comboCount.value)
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

        await response.json()
        await loadPlayerProfile()
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

        await response.json()
        await loadPlayerProfile()
    } catch (error) {
        errorMessage.value = error.message || '分解失败，请稍后重试。'
    } finally {
        actioningItemId.value = ''
    }
}

async function toggleItemLock(instanceId, locked) {
    if (!nickname.value || !instanceId) {
        return
    }

    const action = locked ? 'unlock' : 'lock'
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
            throw new Error(await readErrorMessage(response, '锁定状态更新失败，请稍后重试。'))
        }

        await response.json()
        await loadPlayerProfile()
    } catch (error) {
        errorMessage.value = error.message || '锁定状态更新失败，请稍后重试。'
    } finally {
        actioningItemId.value = ''
    }
}

async function salvageUnequippedItems() {
    if (!nickname.value) {
        return null
    }

    errorMessage.value = ''

    try {
        const response = await fetch('/api/equipment/salvage/unequipped', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({nickname: nickname.value}),
        })

        if (!response.ok) {
            throw new Error(await readErrorMessage(response, '一键分解失败，请稍后重试。'))
        }

        const payload = await response.json()
        await loadPlayerProfile()
        return payload
    } catch (error) {
        errorMessage.value = error.message || '一键分解失败，请稍后重试。'
        return null
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
        await loadPlayerProfile()
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
        try {
            await loadPlayerProfile()
        } catch (error) {
            errorMessage.value = error.message || '资料加载失败，请稍后重试。'
        }
        if (currentPublicPage.value === 'shop') {
            await loadShopItems()
        }
        startPresenceHeartbeat()
        startTaskPolling()
        await loadRooms()
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
        if (currentPublicPage.value === 'shop') {
            void loadShopItems()
        }
        void loadRooms()
        connectRealtime('')
}

function clearPlayerSessionState() {
    stopPresenceHeartbeat()
    stopTaskPolling()
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
        try {
            await loadPlayerProfile()
        } catch (error) {
            errorMessage.value = error.message || '资料加载失败，请稍后重试。'
        }
        if (currentPublicPage.value === 'shop') {
            await loadShopItems()
        }
        startPresenceHeartbeat()
        startTaskPolling()
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

function startTalentTick() {
    if (talentTickTimer) {
        return
    }
    talentTickTimer = window.setInterval(() => {
        nowTick.value++
    }, 200)
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
         startTalentTick()
         await loadPlayerSession()
        await loadState()
        await loadRooms()
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
         window.clearInterval(talentTickTimer)
         talentTickTimer = 0
         stopPresenceHeartbeat()
         stopTaskPolling()
        realtimeTransport?.close()
        burstTimers.forEach((timer) => window.clearTimeout(timer))
        burstTimers.clear()
        stageFxTimers.forEach((timer) => window.clearTimeout(timer))
        stageFxTimers.clear()
        burstFrameOffsets.clear()
        talentTriggerFeed.value = []
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
        currentRoomId,
        currentRoom,
        rooms,
        roomSwitching,
        roomError,
        roomSwitchCooldownEndsAt,
        bossLeaderboard,
        bossLoot,
        bossGoldRange,
        bossStoneRange,
        bossTalentPointsOnKill,
        announcementVersion,
        latestAnnouncement,
        announcements,
        myBossStats,
        tasks,
        hasClaimableTasks,
        inventory,
        loadout,
        loadoutSlots,
        combatStats,
        recentRewards,
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
        bleedBursts,
        talentTriggerFeed,
        bleedTriggerFeed,
        talentVisualState,
        comboCount,
        stormCombo,
        armorCombo,
        stormProgress,
        armorProgress,
        stormTrigger,
        armorTrigger,
        judgmentDayTrigger,
        autoStrikeTrigger,
        globalStatusList,
        partProgressList,
        partStatusList,
        talentCombatState,
        comboTriggerFlash,
        comboTimeoutPercent,
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
        shopItems,
        loadingShopItems,
        gold,
        stones,
        talentPoints,
        equippedBattleClickSkinId,
        equippedBattleClickCursorImagePath,
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
        myBossKills,
        totalBossKills,
        myRank,
        myBossDamage,
        bossLeaderboardCount,
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
        loadShopItems,
        submitMessage,
        claimTask,
        purchaseShopItem,
        equipShopItem,
        unequipShopItem,
        toggleAutoClick,
        loadRooms,
        joinRoom,
        clickButton,
        toggleItemEquip,
        salvageItem,
        toggleItemLock,
        salvageUnequippedItems,
        enhanceItem,
        submitNickname,
        resetNickname,
        registerPublicPageLifecycle,
    }
}
