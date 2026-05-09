<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {usePublicPageState} from './publicPageState'
import PixelEffectCanvas from '../components/PixelEffectCanvas.vue'
import RoomSelector from '../components/RoomSelector.vue'
import RoomSwitchCooldownTag from '../components/RoomSwitchCooldownTag.vue'
import {formatCompact, formatIntegerExact, ratioPercent} from '../utils/formatNumber.js'
import wechatGroupImage from '../assets/community/wechat-group.png'

const {
  boss,
  bossLeaderboard,
  bossLoot,
  bossGoldRange,
  bossStoneRange,
  bossTalentPointsOnKill,
  currentRoomId,
  currentRoom,
  rooms,
  hallLeaderboardSnapshot,
  hallLeaderboardLoading,
  hallLeaderboardError,
  hallLeaderboardPage,
  roomSwitching,
  roomError,
  roomSwitchCooldownEndsAt,
  leaderboard,
  nickname,
  loading,
  syncing,
  syncLabel,
  errorMessage,
  pendingKeys,
  damageBursts,
  bleedBursts,
  talentTriggerFeed,
  bleedTriggerFeed,
  talentVisualState,
  comboCount,
  stormTrigger,
  armorTrigger,
  autoStrikeTrigger,
  judgmentDayTrigger,
  damageStageFx,
  totalVotes,
  isLoggedIn,
  myClicks,
  myBossKills,
  totalBossKills,
  myBossDamage,
  bossLeaderboardCount,
  combatStats,
  stamina,
  bossStatusLabel,
  bossProgress,
  formatDropRate,
  formatRarityLabel,
  formatItemStatLines,
  equipmentNameParts,
  equipmentNameClass,
  navigatePublicPage,
  rewardModal,
  closeRewardModal,
  loadHallLeaderboardSnapshot,
  resetHallLeaderboardSnapshot,
  joinRoom,
  clickButton,
  globalStatusList,
  partProgressList,
  partStatusList,
  equippedBattleClickCursorImagePath,
} = usePublicPageState()

const bossDropModalOpen = ref(false)
const bossGridRef = ref(null)
const swordCursorRef = ref(null)
const swordCursorImageRef = ref(null)
const bossCursorVisible = ref(false)
const bossCellSizePx = ref(56)
const DEFAULT_BOSS_SWORD_CURSOR_URL = 'https://hai-world2.oss-cn-beijing.aliyuncs.com/effects/click-sword_basic.png'
const HALL_ROOM_ID = 'hall'
const HALL_LEADERBOARD_START_RANK = 11
const HALL_LEADERBOARD_PAGE_SIZE = 40
const HALL_LEADERBOARD_COLUMN_COUNT = 4
const STAMINA_RECOVER_INTERVAL_SECONDS = 300
const bossSwordCursorUrl = computed(() => equippedBattleClickCursorImagePath.value || DEFAULT_BOSS_SWORD_CURSOR_URL)
const isHallRoom = computed(() => String(currentRoomId.value || '') === HALL_ROOM_ID)
const currentRoomDisplay = computed(() => {
  if (isHallRoom.value) return '大厅'
  const displayName = String(currentRoom.value?.displayName || '').trim()
  if (displayName) return displayName
  return `房间 ${currentRoom.value?.id || currentRoomId.value || '1'}`
})
const currentRoomSeal = computed(() => isHallRoom.value ? 'HL' : String(currentRoom.value?.id || currentRoomId.value || '1').padStart(2, '0'))
const myBattlePower = computed(() => (
    Math.max(0, Number(combatStats.value?.attackPower || 0)) +
    Math.max(0, Number(combatStats.value?.normalDamage || 0)) +
    Math.max(0, Number(combatStats.value?.criticalDamage || 0))
))
const battlePowerLabel = computed(() => (
    isLoggedIn.value ? formatCompact(myBattlePower.value) : '登录后激活'
))
const staminaRecoveryPreview = computed(() => {
  void nowTick.value
  const state = stamina.value || {}
  const max = Math.max(1, Number(state.max || 50))
  const current = Math.max(0, Number(state.current || 0))
  const nextRecoverAt = Number(state.nextRecoverAt || 0)
  if (!nextRecoverAt || current >= max) {
    return {
      current,
      nextRecoverAt: 0,
      countdown: '已满',
    }
  }
  const nowUnix = Date.now() / 1000
  const elapsed = nowUnix - nextRecoverAt
  if (elapsed < 0) {
    const seconds = Math.max(0, Math.ceil(nextRecoverAt - nowUnix))
    return {
      current,
      nextRecoverAt,
      countdown: formatStaminaRecoverCountdown(seconds),
    }
  }
  const recoveredPoints = 1 + Math.floor(elapsed / STAMINA_RECOVER_INTERVAL_SECONDS)
  const previewCurrent = Math.min(max, current + recoveredPoints)
  if (previewCurrent >= max) {
    return {
      current: max,
      nextRecoverAt: 0,
      countdown: '已满',
    }
  }
  const previewNextRecoverAt = nextRecoverAt + (recoveredPoints * STAMINA_RECOVER_INTERVAL_SECONDS)
  const seconds = Math.max(0, Math.ceil(previewNextRecoverAt - nowUnix))
  return {
    current: previewCurrent,
    nextRecoverAt: previewNextRecoverAt,
    countdown: formatStaminaRecoverCountdown(seconds),
  }
})
const recoveredStamina = computed(() => staminaRecoveryPreview.value.current)
const staminaRecoverCountdown = computed(() => {
  return staminaRecoveryPreview.value.countdown
})
const staminaTooltipText = computed(() => {
  if (recoveredStamina.value >= Number(stamina.value?.max || 50)) {
    return '体力已满'
  }
  return `下一点恢复还需要 ${staminaRecoverCountdown.value}`
})
const isStaminaRiskBanned = computed(() => {
  void nowTick.value
  return Number(stamina.value?.riskBanUntil || 0) > Date.now() / 1000
})
const staminaFloatLabel = computed(() => `${recoveredStamina.value}/${stamina.value?.max || 50}`)
const hallLeaderboardPageEntries = computed(() => {
  const start = hallLeaderboardPage.value * HALL_LEADERBOARD_PAGE_SIZE
  return hallLeaderboardSnapshot.value.slice(start, start + HALL_LEADERBOARD_PAGE_SIZE)
})
const hallLeaderboardTotalPages = computed(() => (
    Math.ceil(hallLeaderboardSnapshot.value.length / HALL_LEADERBOARD_PAGE_SIZE)
))
const hallLeaderboardHasPagination = computed(() => hallLeaderboardSnapshot.value.length > 0)
const hallLeaderboardRangeStart = computed(() => (
    HALL_LEADERBOARD_START_RANK + (hallLeaderboardPage.value * HALL_LEADERBOARD_PAGE_SIZE)
))
const hallLeaderboardRangeEnd = computed(() => (
    hallLeaderboardRangeStart.value + Math.max(0, hallLeaderboardPageEntries.value.length - 1)
))
const hallLeaderboardColumns = computed(() => {
  const columnSize = Math.max(1, Math.ceil(hallLeaderboardPageEntries.value.length / HALL_LEADERBOARD_COLUMN_COUNT))
  return Array.from({length: HALL_LEADERBOARD_COLUMN_COUNT}, (_, index) => (
      hallLeaderboardPageEntries.value.slice(index * columnSize, (index + 1) * columnSize)
  ))
})
const hallLeaderboardRangeLabel = computed(() => (
    hallLeaderboardPageEntries.value.length > 0
        ? `${hallLeaderboardRangeStart.value}-${hallLeaderboardRangeEnd.value}`
        : '11-50'
))
const bossBattlePower = computed(() => {
  if (!boss.value) return 0
  const armorPower = Array.isArray(boss.value.parts)
      ? boss.value.parts.reduce((sum, part) => {
        try {
          return sum + BigInt(String(part?.armor ?? 0))
        } catch {
          return sum
        }
      }, 0n)
      : 0n
  try {
    return (BigInt(String(boss.value.maxHp || 0)) + armorPower).toString()
  } catch {
    return '0'
  }
})

const comboMilestoneText = ref('')
const comboMilestoneTick = ref(0)
const pendingScrollToTopAfterExit = ref(false)
let comboMilestoneTimer = 0

watch(() => comboCount.value, (next, prev) => {
  const nextMilestone = Math.floor(next / 25)
  const prevMilestone = Math.floor((prev || 0) / 25)
  if (nextMilestone <= 0 || nextMilestone === prevMilestone) return
  comboMilestoneText.value = `+${nextMilestone * 10}%`
  comboMilestoneTick.value++
  clearTimeout(comboMilestoneTimer)
  comboMilestoneTimer = setTimeout(() => {
    comboMilestoneText.value = ''
  }, 900)
})

watch(() => currentRoomId.value, (next, prev) => {
  if (!pendingScrollToTopAfterExit.value) return
  if (String(prev || '') === HALL_ROOM_ID) {
    pendingScrollToTopAfterExit.value = false
    return
  }
  if (String(next || '') !== HALL_ROOM_ID) return
  pendingScrollToTopAfterExit.value = false
  window.scrollTo({top: 0, behavior: 'smooth'})
})

watch(() => currentRoomId.value, async (next, prev) => {
  if (String(next || '') === HALL_ROOM_ID && String(prev || '') !== HALL_ROOM_ID) {
    await loadHallLeaderboardSnapshot()
    return
  }
  if (String(next || '') !== HALL_ROOM_ID && String(prev || '') === HALL_ROOM_ID) {
    resetHallLeaderboardSnapshot()
  }
}, {immediate: true})

const talentEffectOverlayRef = ref(null)
const talentOverlayCanvasSize = ref({width: 1, height: 1})
const swordSwingTick = ref(0)
const swordRecoverTick = ref(0)
const swordAnimationPhase = ref('idle')
const zoneHitFlashTicks = ref({})
const clickSparkFeed = ref([])
const bossZoneElementMap = new Map()
let bossGridResizeObserver = null
let bossGridRect = null
let bossCursorFrame = 0
let pendingBossCursorPoint = null
let appliedBossCursorPoint = null

// 每秒 tick 驱动倒计时刷新
const nowTick = ref(0)
let tickTimer = 0

let recoverTimer = 0
let lastAttackTime = 0
let lastPointerDown = 0
const CLICK_SPARK_WINDOW_MS = 420

const roomJoinCooldownRemainingSeconds = computed(() => {
  void nowTick.value
  const endsAt = Number(roomSwitchCooldownEndsAt.value || 0)
  if (!Number.isFinite(endsAt) || endsAt <= 0) return 0
  return Math.max(0, Math.ceil((endsAt - Date.now()) / 1000))
})

function formatStaminaRecoverCountdown(seconds) {
  const minutes = Math.floor(seconds / 60)
  const remain = seconds % 60
  return `${minutes} 分 ${remain} 秒`
}

function exitCurrentRoom() {
  if (roomSwitching.value) return
  pendingScrollToTopAfterExit.value = true
  joinRoom(HALL_ROOM_ID)
}

function showPreviousHallLeaderboardPage() {
  hallLeaderboardPage.value = Math.max(0, hallLeaderboardPage.value - 1)
}

function showNextHallLeaderboardPage() {
  const lastPage = Math.max(0, hallLeaderboardTotalPages.value - 1)
  hallLeaderboardPage.value = Math.min(lastPage, hallLeaderboardPage.value + 1)
}

function openStaminaShop() {
  void navigatePublicPage('shop')
}

function applyBossCursorTransform() {
  const swordCursor = swordCursorRef.value
  if (!swordCursor || !appliedBossCursorPoint) return
  const {x, y} = appliedBossCursorPoint
  swordCursor.style.transform = `translate(${x}px, ${y}px) translate(-50%, -50%)`
}

function flushBossCursorFrame() {
  bossCursorFrame = 0
  if (!pendingBossCursorPoint) return
  appliedBossCursorPoint = pendingBossCursorPoint
  pendingBossCursorPoint = null
  applyBossCursorTransform()
}

function queueBossCursorPosition(x, y) {
  pendingBossCursorPoint = {x, y}
  if (bossCursorFrame) return
  bossCursorFrame = requestAnimationFrame(() => {
    flushBossCursorFrame()
  })
}

function measureBossGridRect() {
  const grid = bossGridRef.value
  if (!grid) {
    bossGridRect = null
    return null
  }
  bossGridRect = grid.getBoundingClientRect()
  return bossGridRect
}

function invalidateBossGridRect() {
  bossGridRect = null
}

function updateCursorPos(e) {
  const rect = measureBossGridRect()
  if (!rect) return
  queueBossCursorPosition(e.clientX - rect.left, e.clientY - rect.top)
}

function measureTalentOverlaySize() {
  const overlay = talentEffectOverlayRef.value
  if (!(overlay instanceof HTMLElement)) {
    talentOverlayCanvasSize.value = {width: 1, height: 1}
    return talentOverlayCanvasSize.value
  }
  const rect = overlay.getBoundingClientRect()
  talentOverlayCanvasSize.value = {
    width: Math.max(1, Math.round(rect.width || 0)),
    height: Math.max(1, Math.round(rect.height || 0)),
  }
  return talentOverlayCanvasSize.value
}

function measureBossCellSize() {
  const grid = bossGridRef.value
  const cell = grid?.querySelector?.('.boss-part-cell')
  if (!(cell instanceof HTMLElement)) return
  const rect = cell.getBoundingClientRect()
  if (rect.width > 0) {
    bossCellSizePx.value = Math.round(rect.width)
  }
}

function effectCanvasSize(scale) {
  return Math.max(1, Math.round(bossCellSizePx.value * scale))
}

function ultimateEffectCanvasSize() {
  return 90
}

function bossGridEffectSize() {
  const rect = bossGridRef.value?.getBoundingClientRect?.()
  const width = Math.round(rect?.width || 0)
  const height = Math.round(rect?.height || 0)
  return Math.max(1, Math.max(width, height, Math.round(bossCellSizePx.value * 5)))
}

function effectOffset(scale) {
  return Math.round(bossCellSizePx.value * scale)
}

function effectFallback(scale, fallback = {}) {
  const size = effectCanvasSize(scale)
  return {
    marginLeft: `-${Math.round(size / 2)}px`,
    marginTop: `-${Math.round(size / 2)}px`,
    ...fallback,
  }
}

function parseFallbackCoord(value, totalSize) {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  const raw = String(value || '').trim()
  if (!raw) return totalSize / 2
  if (raw.endsWith('%')) {
    const percent = Number(raw.slice(0, -1))
    if (Number.isFinite(percent)) {
      return totalSize * percent / 100
    }
  }
  const px = Number(raw.replace(/px$/, ''))
  if (Number.isFinite(px)) return px
  return totalSize / 2
}

function resolveFallbackAnchor(fallback = {}) {
  const size = measureTalentOverlaySize()
  return {
    left: parseFallbackCoord(fallback.left, size.width),
    top: parseFallbackCoord(fallback.top, size.height),
  }
}

function effectEntryLayout(type, options = {}) {
  const {
    anchor = 'part',
    scale = 1,
    offsetXScale = 0,
    offsetYScale = 0,
    entryOverride = null,
    fallback = {},
  } = options
  if (anchor === 'grid') {
    const size = bossGridEffectSize()
    const center = gridOverlayAnchor()
    if (!center) return null
    return {
      left: Math.round(center.left - size / 2),
      top: Math.round(center.top - size / 2),
      width: size,
      height: size,
      size,
      effect: type,
      id: `${type}-${size}-${Math.round(center.left)}-${Math.round(center.top)}`,
    }
  }
  const size = effectCanvasSize(scale)
  const center = triggerAnchor(type, effectWindowMs(type), entryOverride) || resolveFallbackAnchor(fallback)
  if (!center) return null
  return {
    left: Math.round(center.left + effectOffset(offsetXScale) - size / 2),
    top: Math.round(center.top + effectOffset(offsetYScale) - size / 2),
    width: size,
    height: size,
    size,
    effect: type,
    id: `${type}-${size}-${Math.round(center.left)}-${Math.round(center.top)}`,
  }
}

function pushBattleEffect(entries, effect, config = {}) {
  const {
    id = '',
    anchor = 'part',
    size = 0,
    width = size,
    height = size,
    left = 0,
    top = 0,
    ...extra
  } = config
  entries.push({
    id: id || `${effect}-${entries.length}`,
    effect,
    size: Math.max(1, Math.round(size || width || height || 1)),
    width: Math.max(1, Math.round(width || size || 1)),
    height: Math.max(1, Math.round(height || size || 1)),
    left: Math.round(left),
    top: Math.round(top),
    anchor,
    ...extra,
  })
}

function effectOverlayStyle(type, options = {}) {
  const {
    anchor = 'part',
    scale = 1,
    offsetXScale = 0,
    offsetYScale = 0,
    fallback = {},
  } = options
  if (anchor === 'grid') {
    const size = bossGridEffectSize()
    return triggerOverlayStyle(type, {
      width: size,
      height: size,
      anchor: gridOverlayAnchor(),
      fallback: {
        top: '50%',
        left: '50%',
        marginLeft: `-${Math.round(size / 2)}px`,
        marginTop: `-${Math.round(size / 2)}px`,
      },
    })
  }
  const size = effectCanvasSize(scale)
  return triggerOverlayStyle(type, {
    width: size,
    height: size,
    offsetX: effectOffset(offsetXScale),
    offsetY: effectOffset(offsetYScale),
    fallback: effectFallback(scale, fallback),
  })
}

// 剑挥火花（按部位类型变色，参照 DAMAGE_VARIANTS）
const sparkPresets = {
  weak: {count: 12, size: 112},
  heavy: {count: 5, size: 88},
  soft: {count: 6, size: 92},
}

function detectCellType(el) {
  const cell = el?.closest?.('.boss-part-cell')
  if (!cell) return 'soft'
  if (cell.classList.contains('boss-part-cell--weak')) return 'weak'
  if (cell.classList.contains('boss-part-cell--heavy')) return 'heavy'
  return 'soft'
}

function setBossZoneElement(key, el) {
  const normalizedKey = String(key || '').trim()
  if (!normalizedKey) return
  if (el instanceof HTMLElement) {
    bossZoneElementMap.set(normalizedKey, el)
    return
  }
  bossZoneElementMap.delete(normalizedKey)
}

function queueClickSpark(e, cellType) {
  const overlay = talentEffectOverlayRef.value
  if (!(overlay instanceof HTMLElement)) return
  const rect = overlay.getBoundingClientRect()
  const preset = sparkPresets[cellType] || sparkPresets.soft
  const spreadSize = Math.max(effectCanvasSize(3.05), 156)
  const now = Date.now()
  clickSparkFeed.value = [
    {
      id: `click-spark-${now}-${Math.floor(Math.random() * 100000)}`,
      triggeredAt: now,
      left: e.clientX - rect.left,
      top: e.clientY - rect.top,
      size: preset.size,
      spreadSize,
      cellType,
      count: preset.count,
    },
    ...clickSparkFeed.value,
  ].slice(0, 18)
}

function queueHitFlash(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return
  zoneHitFlashTicks.value = {
    ...zoneHitFlashTicks.value,
    [key]: (Number(zoneHitFlashTicks.value[key] || 0) + 1),
  }
}

function hitFlashClass(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return ''
  const tick = Number(zoneHitFlashTicks.value[key] || 0)
  if (tick <= 0) return ''
  return tick % 2 === 0 ? 'boss-part-cell--hit-flash-b' : 'boss-part-cell--hit-flash-a'
}

function doCursorAttack(e) {
  updateCursorPos(e)
  const now = Date.now()
  if (now - lastAttackTime < 16) return
  lastAttackTime = now
  e.preventDefault()
  const swordCursorImage = swordCursorImageRef.value
  if (!swordCursorImage) return

  clearTimeout(recoverTimer)
  swordAnimationPhase.value = 'swinging'
  swordSwingTick.value++

  recoverTimer = setTimeout(() => {
    if (!swordCursorImage) return
    swordAnimationPhase.value = 'recovering'
    swordRecoverTick.value++
  }, 50)

  queueClickSpark(e, detectCellType(e.target))
}

function handleBossGridPointerMove(e) {
  updateCursorPos(e)
}

function handleBossGridPointerEnter(e) {
  measureBossGridRect()
  bossCursorVisible.value = true
  updateCursorPos(e)
}

function handleBossGridPointerLeave() {
  bossCursorVisible.value = false
  pendingBossCursorPoint = null
}

function handleBossGridPointerDown(e) {
  lastPointerDown = Date.now()
  doCursorAttack(e)
}

function handleBossGridClick(e) {
  if (Date.now() - lastPointerDown < 48) return
  doCursorAttack(e)
}

onMounted(() => {
  measureBossGridRect()
  measureBossCellSize()
  measureTalentOverlaySize()
  window.addEventListener('scroll', invalidateBossGridRect, {passive: true})
  window.addEventListener('resize', invalidateBossGridRect)
  if (typeof ResizeObserver !== 'undefined') {
    bossGridResizeObserver = new ResizeObserver(() => {
      measureBossGridRect()
      measureBossCellSize()
      measureTalentOverlaySize()
    })
    if (bossGridRef.value instanceof HTMLElement) {
      bossGridResizeObserver.observe(bossGridRef.value)
    }
    if (talentEffectOverlayRef.value instanceof HTMLElement) {
      bossGridResizeObserver.observe(talentEffectOverlayRef.value)
    }
  }
  tickTimer = setInterval(() => {
    nowTick.value++
  }, 250)
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', invalidateBossGridRect)
  window.removeEventListener('resize', invalidateBossGridRect)
  bossGridResizeObserver?.disconnect?.()
  bossGridResizeObserver = null
  clearInterval(tickTimer)
  cancelAnimationFrame(bossCursorFrame)
  clearTimeout(recoverTimer)
  clearTimeout(comboMilestoneTimer)
  bossZoneElementMap.clear()
})

const bossZones = computed(() => {
  if (!boss.value?.parts || !Array.isArray(boss.value.parts)) return []
  const grid = Array.from({length: 5}, () => Array(5).fill(null))
  boss.value.parts.forEach((part) => {
    if (part.x >= 0 && part.x < 5 && part.y >= 0 && part.y < 5) {
      grid[part.y][part.x] = {
        ...part,
        healthPercent: getPartHealthPercent(part),
        zoneKey: `${part.x}-${part.y}`,
      }
    }
  })
  return grid
})

const partTypeLabels = {
  soft: '软组织',
  heavy: '重甲',
  weak: '弱点',
}

const partTypeColors = {
  soft: '#4ade80',
  heavy: '#9ca3af',
  weak: '#ef4444',
}

const bossDropPool = computed(() =>
    bossLoot.value.map((item) => ({
      id: `equipment:${item.itemId}`,
      type: 'equipment',
      label: '装备',
      item,
    })),
)

function openBossDropPool() {
  bossDropModalOpen.value = true
}

function closeBossDropPool() {
  bossDropModalOpen.value = false
}

function getPartHealthPercent(part) {
  if (!part?.maxHp) return 0
  return ratioPercent(part.currentHp, part.maxHp)
}

function getBossZoneButtonKey(zone) {
  if (!zone) return ''
  return `boss-part:${zone.x}-${zone.y}`
}

// 纯点击
function clickBossZone(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return
  queueHitFlash(zone)
  clickButton(key)
}

function isBossZoneDisabled(zone) {
  const key = getBossZoneButtonKey(zone)
  return !key || !isLoggedIn.value || !zone?.alive || pendingKeys.value.has(key)
}

function bossZoneAriaLabel(zone) {
  if (!zone) return '空 Boss 分区'
  const label = zone.displayName || partTypeLabels[zone.type] || zone.type
  return `${label} 分区，血量 ${formatIntegerExact(zone.currentHp)}/${formatIntegerExact(zone.maxHp)}`
}

function zoneDamageBursts(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return []
  return damageBursts.value?.[key] || []
}

function zoneBleedBursts(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return []
  return bleedBursts.value?.[key] || []
}

// 天赋视觉状态
const talentEdgeGlowClass = computed(() => {
  if (hasRecentTrigger('final_cut', ULTIMATE_EFFECT_WINDOW_MS)) return 'talent-edge-glow--crit'
  if (silverStormActive.value) return ''
  const recent = talentTriggerFeed.value[0]
  if (recent && recent.name === '暴风连击') {
    return 'talent-edge-glow--normal'
  }
  return ''
})

const showSilverFlash = computed(() => silverStormActive.value && silverStormCountdown.value >= 14)
const TALENT_EFFECT_WINDOW_MS = 1350
const BLEED_EFFECT_WINDOW_MS = 1200
const ULTIMATE_EFFECT_WINDOW_MS = 3200
const JUDGMENT_DAY_EFFECT_WINDOW_MS = 5000
const STARFALL_EFFECT_WINDOW_MS = 9200
const MAGIC_BURST_EFFECT_WINDOW_MS = 1600
const MAGIC_RUPTURE_EFFECT_WINDOW_MS = 1800

function latestTrigger(type) {
  return talentTriggerFeed.value.find((e) => e.effectType === type) || null
}

function effectWindowMs(type) {
  if (type === 'final_cut' || type === 'silver_storm') {
    return ULTIMATE_EFFECT_WINDOW_MS
  }
  if (type === 'judgment_day') {
    return JUDGMENT_DAY_EFFECT_WINDOW_MS
  }
  return TALENT_EFFECT_WINDOW_MS
}

function recentBleedTriggers(windowMs = BLEED_EFFECT_WINDOW_MS) {
  return bleedTriggerFeed.value.filter((entry) => isTriggerFresh(entry, windowMs)).slice(0, 12)
}

function isTriggerFresh(entry, windowMs = TALENT_EFFECT_WINDOW_MS) {
  void nowTick.value
  if (!entry) return false
  const at = Number(entry.triggeredAt || 0)
  if (!Number.isFinite(at) || at <= 0) return false
  return Date.now() - at <= windowMs
}

function hasRecentTrigger(type, windowMs = null) {
  return isTriggerFresh(latestTrigger(type), windowMs ?? effectWindowMs(type))
}

function triggerKey(type, windowMs = null) {
  const entry = latestTrigger(type)
  if (!isTriggerFresh(entry, windowMs ?? effectWindowMs(type))) return ''
  return `${type}-${entry.id}`
}

const clickSparkEntries = computed(() => {
  void nowTick.value
  return clickSparkFeed.value
      .filter((entry) => isTriggerFresh(entry, CLICK_SPARK_WINDOW_MS))
      .map((entry) => ({
        id: entry.id,
        effect: 'click_spark',
        size: entry.size,
        width: entry.spreadSize,
        height: entry.spreadSize,
        left: Math.round(entry.left - entry.spreadSize / 2),
        top: Math.round(entry.top - entry.spreadSize / 2),
        anchor: 'spark',
        cellType: entry.cellType,
      }))
})

const battleEffectEntries = computed(() => {
  void nowTick.value
  measureTalentOverlaySize()
  const entries = []

  const stormComboLayout = hasRecentTrigger('storm_combo') ? effectEntryLayout('storm_combo', {scale: 1.65, fallback: { top: '50%', left: '50%' }}) : null
  if (stormComboLayout) {
    void effectOverlayStyle('storm_combo', { scale: 1.65, fallback: { top: '50%', left: '50%' } })
    pushBattleEffect(entries, 'storm_combo', {
      id: triggerKey('storm_combo'),
      size: effectCanvasSize(1.65),
      width: stormComboLayout.width,
      height: stormComboLayout.height,
      left: stormComboLayout.left,
      top: stormComboLayout.top,
    })
  }

  const autoStrikeLayout = hasRecentTrigger('auto_strike') ? effectEntryLayout('auto_strike', {scale: 2.05, fallback: { top: '50%', left: '50%' }}) : null
  if (autoStrikeLayout) {
    pushBattleEffect(entries, 'auto_strike', {
      id: triggerKey('auto_strike'),
      size: effectCanvasSize(2.05),
      width: autoStrikeLayout.width,
      height: autoStrikeLayout.height,
      left: autoStrikeLayout.left,
      top: autoStrikeLayout.top,
    })
  }

  for (const entry of recentBleedTriggers(BLEED_EFFECT_WINDOW_MS)) {
    const bleedLayout = effectEntryLayout('bleed', {scale: 2.6, entryOverride: entry, fallback: { top: '50%', left: '50%' }})
    if (!bleedLayout) continue
    pushBattleEffect(entries, 'bleed', {
      id: entry.id,
      size: effectCanvasSize(2.6),
      width: bleedLayout.width,
      height: bleedLayout.height,
      left: bleedLayout.left,
      top: bleedLayout.top,
    })
  }

  const finalCutLayout = hasRecentTrigger('final_cut', ULTIMATE_EFFECT_WINDOW_MS)
      ? effectEntryLayout('final_cut', {anchor: 'grid', fallback: { top: '50%', left: '50%' }})
      : null
  if (finalCutLayout) {
    void effectOverlayStyle('final_cut', { anchor: 'grid', fallback: { top: '50%', left: '50%' } })
    pushBattleEffect(entries, 'final_cut', {
      id: triggerKey('final_cut', ULTIMATE_EFFECT_WINDOW_MS),
      anchor: 'grid',
      size: ultimateEffectCanvasSize(),
      width: finalCutLayout.width,
      height: finalCutLayout.height,
      left: finalCutLayout.left,
      top: finalCutLayout.top,
    })
  }

  const collapseLayout = hasRecentTrigger('collapse_trigger')
      ? effectEntryLayout('collapse_trigger', {scale: 2, fallback: { top: '50%', left: '50%' }})
      : null
  if (collapseLayout) {
    const collapseStyle = effectOverlayStyle('collapse_trigger', { scale: 2, fallback: { top: '50%', left: '50%' } })
    void collapseStyle
    pushBattleEffect(entries, 'collapse_trigger', {
      id: triggerKey('collapse_trigger'),
      size: effectCanvasSize(5),
      width: collapseLayout.width,
      height: collapseLayout.height,
      left: collapseLayout.left,
      top: collapseLayout.top,
    })
  }

  const judgmentDayLayout = hasRecentTrigger('judgment_day', JUDGMENT_DAY_EFFECT_WINDOW_MS)
      ? effectEntryLayout('judgment_day', {anchor: 'grid', fallback: { top: '50%', left: '50%' }})
      : null
  if (judgmentDayLayout) {
    const judgmentDayStyle = effectOverlayStyle('judgment_day', { anchor: 'grid', fallback: { top: '50%', left: '50%' } })
    void judgmentDayStyle
    pushBattleEffect(entries, 'judgment_day', {
      id: triggerKey('judgment_day', JUDGMENT_DAY_EFFECT_WINDOW_MS),
      anchor: 'grid',
      size: bossGridEffectSize(),
      width: judgmentDayLayout.width,
      height: judgmentDayLayout.height,
      left: judgmentDayLayout.left,
      top: judgmentDayLayout.top,
    })
  }

  const magicBurstLayout = hasRecentTrigger('magic_burst', MAGIC_BURST_EFFECT_WINDOW_MS)
      ? effectEntryLayout('magic_burst', {scale: 2.45, fallback: { top: '50%', left: '50%' }})
      : null
  if (magicBurstLayout) {
    const magicBurstStyle = effectOverlayStyle('magic_burst', { scale: 2.45, fallback: { top: '50%', left: '50%' } })
    void magicBurstStyle
    pushBattleEffect(entries, 'magic_burst', {
      id: triggerKey('magic_burst', MAGIC_BURST_EFFECT_WINDOW_MS),
      size: effectCanvasSize(2.45),
      width: magicBurstLayout.width,
      height: magicBurstLayout.height,
      left: magicBurstLayout.left,
      top: magicBurstLayout.top,
    })
  }

  const magicRuptureLayout = hasRecentTrigger('magic_rupture', MAGIC_RUPTURE_EFFECT_WINDOW_MS)
      ? effectEntryLayout('magic_rupture', {scale: 2.65, fallback: { top: '50%', left: '50%' }})
      : null
  if (magicRuptureLayout) {
    const magicRuptureStyle = effectOverlayStyle('magic_rupture', { scale: 2.65, fallback: { top: '50%', left: '50%' } })
    void magicRuptureStyle
    pushBattleEffect(entries, 'magic_rupture', {
      id: triggerKey('magic_rupture', MAGIC_RUPTURE_EFFECT_WINDOW_MS),
      size: effectCanvasSize(2.65),
      width: magicRuptureLayout.width,
      height: magicRuptureLayout.height,
      left: magicRuptureLayout.left,
      top: magicRuptureLayout.top,
    })
  }

  const magicStarfallLayout = hasRecentTrigger('magic_starfall', STARFALL_EFFECT_WINDOW_MS)
      ? effectEntryLayout('magic_starfall', {anchor: 'grid', fallback: { top: '50%', left: '50%' }})
      : null
  if (magicStarfallLayout) {
    const magicStarfallStyle = effectOverlayStyle('magic_starfall', { anchor: 'grid', fallback: { top: '50%', left: '50%' } })
    void magicStarfallStyle
    pushBattleEffect(entries, 'magic_starfall', {
      id: triggerKey('magic_starfall', STARFALL_EFFECT_WINDOW_MS),
      anchor: 'grid',
      size: ultimateEffectCanvasSize(),
      width: magicStarfallLayout.width,
      height: magicStarfallLayout.height,
      left: magicStarfallLayout.left,
      top: magicStarfallLayout.top,
    })
  }

  const doomMarkStyle = effectOverlayStyle('doom_mark', { scale: 1.25, fallback: { top: '50%', left: '50%' } })
  const doomMarkLayout = hasRecentTrigger('doom_mark') ? effectEntryLayout('doom_mark', {scale: 1.25, fallback: { top: '50%', left: '50%' }}) : null
  if (doomMarkLayout) {
    pushBattleEffect(entries, 'doom_mark', {
      id: triggerKey('doom_mark'),
      size: effectCanvasSize(1.25),
      width: doomMarkLayout.width,
      height: doomMarkLayout.height,
      left: doomMarkLayout.left,
      top: doomMarkLayout.top,
    })
  }

  void doomMarkStyle
  const silverStormLayout = hasRecentTrigger('silver_storm', ULTIMATE_EFFECT_WINDOW_MS)
      ? effectEntryLayout('silver_storm', {anchor: 'grid', fallback: { top: '50%', left: '50%' }})
      : null
  if (silverStormLayout) {
    const silverStormStyle = effectOverlayStyle('silver_storm', { anchor: 'grid', fallback: { top: '50%', left: '50%' } })
    void silverStormStyle
    pushBattleEffect(entries, 'silver_storm', {
      id: triggerKey('silver_storm', ULTIMATE_EFFECT_WINDOW_MS),
      anchor: 'grid',
      size: ultimateEffectCanvasSize(),
      width: silverStormLayout.width,
      height: silverStormLayout.height,
      left: silverStormLayout.left,
      top: silverStormLayout.top,
    })
  }

  for (const entry of clickSparkEntries.value) {
    pushBattleEffect(entries, 'click_spark', entry)
  }

  return entries
})

function findBossZoneElement(partX, partY) {
  const x = Number(partX)
  const y = Number(partY)
  if (!Number.isFinite(x) || !Number.isFinite(y)) return null
  const key = `${Math.floor(x)}-${Math.floor(y)}`
  return bossZoneElementMap.get(key) || null
}

function triggerAnchor(type, windowMs = TALENT_EFFECT_WINDOW_MS, entryOverride = null) {
  const entry = entryOverride || latestTrigger(type)
  if (!isTriggerFresh(entry, windowMs)) return null
  const overlayEl = talentEffectOverlayRef.value
  const zoneEl = findBossZoneElement(entry.partX, entry.partY)
  if (!(overlayEl instanceof HTMLElement) || !(zoneEl instanceof HTMLElement)) {
    return null
  }
  const overlayRect = overlayEl.getBoundingClientRect()
  const zoneRect = zoneEl.getBoundingClientRect()
  return {
    left: zoneRect.left - overlayRect.left + zoneRect.width / 2,
    top: zoneRect.top - overlayRect.top + zoneRect.height / 2,
  }
}

function gridOverlayAnchor() {
  const overlayEl = talentEffectOverlayRef.value
  const gridEl = bossGridRef.value
  if (!(overlayEl instanceof HTMLElement) || !(gridEl instanceof HTMLElement)) {
    return null
  }
  const overlayRect = overlayEl.getBoundingClientRect()
  const gridRect = gridEl.getBoundingClientRect()
  return {
    left: gridRect.left - overlayRect.left + gridRect.width / 2,
    top: gridRect.top - overlayRect.top + gridRect.height / 2,
  }
}

function triggerOverlayStyle(type, options = {}) {
  const {
    width = 0,
    height = 0,
    offsetX = 0,
    offsetY = 0,
    anchor: anchorOverride = null,
    fallback = {},
  } = options
  const anchor = anchorOverride || triggerAnchor(type)
  if (!anchor) {
    return fallback
  }
  return {
    width: width > 0 ? `${Math.round(width)}px` : undefined,
    height: height > 0 ? `${Math.round(height)}px` : undefined,
    left: `${Math.round(anchor.left + offsetX)}px`,
    top: `${Math.round(anchor.top + offsetY)}px`,
    marginLeft: `${Math.round(-width / 2)}px`,
    marginTop: `${Math.round(-height / 2)}px`,
  }
}

function isPartFractured(zone) {
  if (!zone) return false
  const hpRatio = zone.maxHp > 0 ? zone.currentHp / zone.maxHp : 1
  return hpRatio < 0.25 && hpRatio > 0
}

function isPartCollapsed(zone) {
  if (!zone) return false
  return talentVisualState.value.collapsePartKeys.includes(`${zone.x}-${zone.y}`)
}

function isPartDoomMarked(zone) {
  if (!zone) return false
  return talentVisualState.value.doomMarks.includes(`${zone.x}-${zone.y}`)
}

const silverStormCountdown = computed(() => {
  void nowTick.value
  const endsAt = talentVisualState.value.silverStormEndsAt
  if (!endsAt) return talentVisualState.value.silverStormRemaining || 0
  return Math.max(0, Math.ceil(endsAt - Date.now() / 1000))
})
const silverStormActive = computed(() => {
  void nowTick.value
  if (!talentVisualState.value.silverStormActive) return false
  if (!talentVisualState.value.silverStormEndsAt) return talentVisualState.value.silverStormRemaining > 0
  return talentVisualState.value.silverStormEndsAt > Date.now() / 1000
})

</script>

<template>
  <div class="battle-page-hud">
    <section class="battle-page-hud__hero" aria-label="战斗状态">
      <div class="battle-page-hud__copy">
        <p class="vote-stage__eyebrow">Hai-World Room Combat HUD</p>
        <h1 class="battle-page-hud__title">
          <span class="battle-page-hud__title-part">大海世界战斗指挥中枢</span>
          <span class="battle-page-hud__title-part">狠狠干一票</span>
        </h1>
        <div class="boss-hud__chips">
          <span>{{ currentRoomDisplay }}</span>
          <span v-if="!isHallRoom">Boss战力 {{ formatCompact(bossBattlePower) }}</span>
          <span v-else>选择房间后进入 Boss 战斗</span>
        </div>
      </div>

      <div class="battle-page-hud__status">
      <span class="live-pill" :class="{ 'live-pill--syncing': syncing }">
        <span class="live-pill__dot"></span>
        {{ syncLabel }}
      </span>
        <article class="power-glory-card power-glory-card--battle" aria-label="当前用户战斗力">
          <span class="power-glory-card__eyebrow">当前用户战斗力</span>
          <strong class="power-glory-card__value">{{ battlePowerLabel }}</strong>
        </article>
        <div class="hero__link-grid battle-page-hud__link-grid" aria-label="项目相关入口">
          <article class="hero-info-card hero-info-card--community" aria-label="游戏交流群">
            <span class="hero-info-card__eyebrow">游戏交流群</span>
            <strong class="hero-info-card__value">鼠标悬停查看微信群</strong>
            <div class="hero-info-card__preview" role="img" aria-label="微信群二维码预览">
              <img :src="wechatGroupImage" alt="Hai-World 微信群二维码"/>
            </div>
          </article>
          <a
              class="hero-info-card hero-info-card--github"
              href="https://github.com/WwhdsOne/long"
              target="_blank"
              rel="noreferrer"
              aria-label="项目地址 GitHub 仓库"
          >
            <span class="hero-info-card__eyebrow">项目地址</span>
            <strong class="hero-info-card__value">点击进入项目仓库</strong>
          </a>
        </div>
      </div>
    </section>

    <section class="stats-band stats-band--wide" aria-label="实时统计">
      <article class="stats-band__card">
        <span class="stats-band__label">我的点击</span>
        <strong>{{ isLoggedIn ? myClicks : '先登录' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">总点击</span>
        <strong>{{ totalVotes }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">我的Boss击杀数</span>
        <strong>{{ isLoggedIn ? myBossKills : '先登录' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">总Boss击杀数</span>
        <strong>{{ totalBossKills }}</strong>
      </article>
    </section>

    <section class="stage-layout stage-layout--battle">
      <section class="vote-stage">

        <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

        <RoomSelector
            v-if="isHallRoom"
            :rooms="rooms"
            :current-room-id="currentRoomId"
            :switching="roomSwitching"
            :error="roomError"
            :logged-in="isLoggedIn"
            :cooldown-remaining-seconds="roomJoinCooldownRemainingSeconds"
            @join="joinRoom"
        />

        <section
            v-if="!isHallRoom"
            class="vote-stage__boss-hud vote-stage__boss-hud--merged"
            :class="{
            'damage-stage--shake': damageStageFx.shake,
            'damage-stage--flash': damageStageFx.flash,
            'damage-stage--doom': damageStageFx.doom,
            'damage-stage--blade': damageStageFx.blade,
            'damage-stage--slowmo': damageStageFx.slowMo,
            'damage-stage--vignette': damageStageFx.vignette,
            'talent-silver-flash': showSilverFlash,
            'talent-blood-flash': damageStageFx.doom,
            [talentEdgeGlowClass]: talentEdgeGlowClass !== '',
          }"
        >
          <div class="vote-stage__boss-hud-head">
            <div class="boss-hud__title-block">
              <div class="vote-stage__head">
                <div>
                  <p class="vote-stage__eyebrow">当前世界 Boss</p>
                  <h2>{{ boss?.name || '等待 Boss 登场' }}</h2>
                </div>
              </div>
              <div class="boss-hud__chips">
                <span>{{ currentRoomDisplay }}</span>
                <span>Boss 榜 {{ bossLeaderboardCount }} 人</span>
                <span>掉落 {{ bossDropPool.length }} 件</span>
              </div>
            </div>
            <div class="boss-stage__meta">
              <span class="boss-stage__pill">{{ bossStatusLabel }}</span>
              <strong v-if="boss">HP {{ formatIntegerExact(boss.currentHp) }} / {{
                  formatIntegerExact(boss.maxHp)
                }}</strong>
              <strong v-else>我的伤害 {{ myBossDamage }}</strong>
            </div>
            <div class="boss-hud__seal" aria-hidden="true">
              <strong>{{ currentRoomSeal }}</strong>
              <span>ROOM</span>
            </div>
          </div>
          <div v-if="boss" class="boss-stage__bar boss-stage__bar--compact">
            <span class="boss-stage__bar-fill" :style="{ width: `${bossProgress}%` }"></span>
          </div>
          <div v-if="loading" class="feedback-panel feedback-panel--compact">
            <p>正在加载 Boss 战场...</p>
          </div>
          <div v-else-if="!boss" class="feedback-panel feedback-panel--compact">
            <p>当前没有活动 Boss。</p>
          </div>
          <div v-else-if="bossZones.length === 0" class="feedback-panel feedback-panel--compact">
            <p>当前 Boss 尚未配置可攻击分区。</p>
          </div>
          <div v-else class="boss-part-grid-container">
            <!-- 左侧面板列 -->
            <div class="boss-left-panels">
              <div v-for="status in globalStatusList" :key="status.key" class="status-panel"
                   :class="[`status-panel--${status.kind}`, { 'status-panel--gold': status.isGold }]"
                   :style="status.panelStyle || null">
                <template v-if="status.kind === 'combo'">
                  <div class="status-panel__row status-panel__row--combo">
                    <span class="status-panel__title">{{ status.title }}</span>
                    <span class="status-panel__count-wrap">
                    <strong class="status-panel__primary" :style="status.primaryStyle || null">{{
                        status.primary
                      }}</strong>
                    <span class="status-panel__milestone-anchor">
                      <span v-if="comboMilestoneText" :key="comboMilestoneTick"
                            class="status-panel__milestone status-panel__milestone--floating">
                        {{ comboMilestoneText }}
                      </span>
                    </span>
                  </span>
                    <span class="status-panel__meta status-panel__meta--combo">
                    <span v-if="status.secondary" class="status-panel__secondary"
                          :style="status.secondaryStyle || null">{{ status.secondary }}</span>
                    <span v-if="status.hint" class="status-panel__hint status-panel__hint--inline"
                          :style="status.hintStyle || null">{{ status.hint }}</span>
                  </span>
                  </div>
                </template>
                <template v-else>
                  <div class="status-panel__title">{{ status.title }}</div>
                  <div class="status-panel__row">
                    <strong class="status-panel__primary" :style="status.primaryStyle || null">{{
                        status.primary
                      }}</strong>
                    <span v-if="status.secondary" class="status-panel__secondary">{{ status.secondary }}</span>
                  </div>
                  <div v-if="status.hint" class="status-panel__hint" :style="status.hintStyle || null">{{
                      status.hint
                    }}
                  </div>
                </template>
                <span v-if="status.showProgress !== false" class="status-panel__bar">
                <span class="status-panel__bar-fill" :class="`status-panel__bar-fill--${status.kind}`"
                      :style="{ width: `${status.progress}%`, ...(status.barStyle || {}) }"></span>
              </span>
              </div>

              <!-- 3. 部位累计进度列表：仅当有进度时显示 -->
              <div v-if="partProgressList.length > 0" class="part-progress-panel">
                <div class="part-progress-panel__title">部位累计进度</div>
                <div v-for="p in partProgressList" :key="p.key" class="part-progress-panel__item">
                <span class="part-progress-panel__name" :class="`part-progress-panel__name--${p.type}`">{{
                    p.name
                  }}</span>
                  <span class="part-progress-panel__track part-progress-panel__track--storm">
                  追击 {{ p.storm }}/{{ stormTrigger }}
                  <span class="part-progress-panel__bar"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--storm"
                      :style="{ width: p.stormProgress + '%' }"></span></span>
                </span>
                  <span v-if="p.type === 'heavy'" class="part-progress-panel__track part-progress-panel__track--armor">
                  破甲 {{ p.armor }}/{{ armorTrigger }}
                  <span class="part-progress-panel__bar"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--armor"
                      :style="{ width: p.armorProgress + '%' }"></span></span>
                </span>
                  <span v-if="p.magic > 0" class="part-progress-panel__track part-progress-panel__track--magic">
                  回响层数 {{ p.magic }}/{{ p.magicTrigger }}
                  <span class="part-progress-panel__bar"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--magic"
                      :style="{ width: p.magicProgress + '%' }"></span></span>
                </span>
                  <span v-if="p.magicTriggered > 0" class="part-progress-panel__track part-progress-panel__track--magic-ultimate">
                  潮爆 {{ p.magicTriggered }}/{{ p.magicUltimateTrigger }}
                  <span class="part-progress-panel__bar"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--magic-ultimate"
                      :style="{ width: p.magicUltimateProgress + '%' }"></span></span>
                </span>
                  <span v-if="p.type === 'heavy' && p.judgmentDay > 0"
                        class="part-progress-panel__track part-progress-panel__track--judgment-day">
                  审判日 {{ p.judgmentDay }}/{{ judgmentDayTrigger }}
                  <span class="part-progress-panel__bar"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--judgment-day"
                      :style="{ width: p.judgmentDayProgress + '%' }"></span></span>
                </span>
                  <span v-if="p.type === 'heavy' && p.autoStrike > 0"
                        class="part-progress-panel__track part-progress-panel__track--auto-strike"
                        :style="{ position: 'relative', zIndex: 999 }">
                    碎甲重击 {{ p.autoStrike }}/{{ autoStrikeTrigger }}
                    <span class="part-progress-panel__bar"><span
                        class="part-progress-panel__bar-fill part-progress-panel__bar-fill--auto-strike"
                        :style="{ width: p.autoStrikeProgress + '%' }"></span></span>
                    <span class="part-progress-panel__countdown">{{
                        Number.isFinite(p.autoStrikeCountdown) ? Math.ceil(p.autoStrikeCountdown) : 0
                      }}s</span>
                    <span class="part-progress-panel__bar part-progress-panel__bar--timer"><span
                        class="part-progress-panel__bar-fill part-progress-panel__bar-fill--timer"
                        :style="{ width: p.autoStrikeTimeoutPercent + '%' }"></span></span>
                  </span>
                </div>
              </div>

              <div v-if="partStatusList.length > 0" class="part-status-panel">
                <div class="part-status-panel__title">部位状态</div>
                <div v-for="s in partStatusList" :key="s.key" class="part-status-panel__item">
                  <span class="part-status-panel__name" :class="`part-status-panel__name--${s.type}`">{{
                      s.name
                    }}</span>
                  <div class="part-status-panel__row">
                    <span class="part-status-panel__label">{{ s.statusLabel }}</span>
                    <span v-if="s.showCountdown !== false" class="part-status-panel__countdown">{{
                        s.remainingSec
                      }}s</span>
                  </div>
                  <div v-if="s.statusMeta" class="status-panel__hint">{{ s.statusMeta }}</div>
                  <span v-if="s.showProgress !== false" class="part-status-panel__bar">
                  <span class="part-status-panel__bar-fill" :class="`part-status-panel__bar-fill--${s.statusKey}`"
                        :style="{ width: `${s.progress}%` }"></span>
                </span>
                </div>
              </div>

            </div>


            <!-- 右侧：5×5 Boss 网格 + 连击计数 -->
            <div class="boss-part-grid-with-combo">
              <div
                  ref="bossGridRef"
                  class="boss-part-grid"
                  @pointerenter="handleBossGridPointerEnter"
                  @pointermove="handleBossGridPointerMove"
                  @pointerleave="handleBossGridPointerLeave"
                  @pointerdown="handleBossGridPointerDown"
                  @click="handleBossGridClick"
              >
                <div v-for="(row, yi) in bossZones" :key="yi" class="boss-part-grid__row">
                  <button
                      v-for="(zone, xi) in row"
                      :key="yi + '-' + xi"
                      :ref="(el) => setBossZoneElement(zone ? `${zone.x}-${zone.y}` : '', el)"
                      class="boss-part-cell boss-zone-button"
                      :data-zone-key="zone ? `${zone.x}-${zone.y}` : ''"
                      :class="{
            'boss-part-cell--alive': zone?.alive,
            'boss-part-cell--dead': zone && !zone.alive,
            'boss-part-cell--soft': zone?.type === 'soft',
            'boss-part-cell--heavy': zone?.type === 'heavy',
            'boss-part-cell--weak': zone?.type === 'weak',
            'boss-part-cell--low': zone?.alive && zone.healthPercent < 25,
            'boss-zone-button--empty': !zone,
            'boss-zone-button--pending': pendingKeys.has(getBossZoneButtonKey(zone)),
            'boss-zone-button--damage': zoneDamageBursts(zone).length > 0,
            'boss-part-cell--fracture': zone && isPartFractured(zone),
            'boss-part-cell--center': xi === 2 && yi === 2,
            'talent-hammer-strike': zone && isPartCollapsed(zone),
            [hitFlashClass(zone)]: zone && hitFlashClass(zone) !== '',
          }"
                      :style="zone ? { '--part-color': partTypeColors[zone.type] || '#64748b' } : {}"
                      type="button"
                      :disabled="isBossZoneDisabled(zone)"
                      :aria-label="bossZoneAriaLabel(zone)"
                      @click="clickBossZone(zone)"
                  >
                    <template v-if="zone">
                      <img
                          v-if="zone.imagePath"
                          class="boss-part-cell__image"
                          :src="zone.imagePath"
                          :alt="zone.displayName || partTypeLabels[zone.type] || zone.type"
                      />
                      <div class="boss-zone-button__damage-layer" aria-hidden="true">
                        <template
                            v-for="burst in zoneDamageBursts(zone)"
                            :key="burst.id"
                        >
              <span
                  class="boss-zone-button__damage-burst"
                  :class="[
                  `boss-zone-button__damage-burst--${burst.type}`,
                ]"
                  :style="{
                  '--damage-offset-x': `${burst.offsetX}px`,
                  '--damage-offset-y': `${burst.offsetY}px`,
                  '--damage-scale': burst.scale,
                  '--damage-ttl': `${burst.ttl}ms`,
                }"
              >
                <span v-if="burst.label" class="boss-zone-button__damage-label">{{ burst.label }}</span>- {{
                  burst.value
                }}
              </span>
                          <span
                              v-for="p in (burst.particles || [])"
                              :key="p.id"
                              class="boss-zone-button__damage-particle"
                              :style="{
                  '--px': `${p.x}px`,
                  '--py': `${p.y}px`,
                  width: `${p.size}px`,
                  height: `${p.size}px`,
                  background: p.color,
                }"
                          ></span>
                        </template>
                        <template
                            v-for="burst in zoneBleedBursts(zone)"
                            :key="burst.id"
                        >
              <span
                  class="boss-zone-button__damage-burst"
                  :class="[
                  `boss-zone-button__damage-burst--${burst.type}`,
                ]"
                  :style="{
                  '--damage-offset-x': `${burst.offsetX}px`,
                  '--damage-offset-y': `${burst.offsetY}px`,
                  '--damage-scale': burst.scale,
                  '--damage-ttl': `${burst.ttl}ms`,
                }"
              >
                <span v-if="burst.label" class="boss-zone-button__damage-label">{{ burst.label }}</span>- {{
                  burst.value
                }}
              </span>
                        </template>
                      </div>
                      <span
                          v-if="isPartDoomMarked(zone)"
                          class="boss-part-cell__doom-mark"
                          aria-label="末日审判标记"
                      ></span>
                      <div class="boss-part-cell__type">{{ partTypeLabels[zone.type] || zone.type }}</div>
                      <strong class="boss-zone-button__label">{{
                          zone.displayName || partTypeLabels[zone.type] || zone.type
                        }}</strong>
                      <div class="boss-part-cell__bar">
              <span
                  class="boss-part-cell__fill"
                  :style="{ width: `${zone.healthPercent}%` }"
              ></span>
                      </div>
                      <div class="boss-zone-button__meta">
                        <span>血量 : {{ formatIntegerExact(zone.currentHp) }}/{{
                            formatIntegerExact(zone.maxHp)
                          }}</span><br>
                        <span>护甲 : {{ isPartCollapsed(zone) ? '0' : formatCompact(zone.armor) }}</span>
                      </div>
                    </template>
                    <span v-else class="boss-part-cell__empty"></span>
                  </button>
                </div>
                <div
                    id="boss-sword-cursor"
                    ref="swordCursorRef"
                    :style="{
                    opacity: bossCursorVisible ? 1 : 0,
                  }"
                >
                  <img
                      ref="swordCursorImageRef"
                      class="boss-sword-cursor__image"
                      :class="{
                      'boss-sword-cursor__image--swing-a': swordAnimationPhase === 'swinging' && swordSwingTick % 2 === 1,
                      'boss-sword-cursor__image--swing-b': swordAnimationPhase === 'swinging' && swordSwingTick % 2 === 0 && swordSwingTick > 0,
                      'boss-sword-cursor__image--recover-a': swordAnimationPhase === 'recovering' && swordRecoverTick % 2 === 1,
                      'boss-sword-cursor__image--recover-b': swordAnimationPhase === 'recovering' && swordRecoverTick % 2 === 0 && swordRecoverTick > 0,
                    }"
                      :src="bossSwordCursorUrl"
                      alt=""
                      aria-hidden="true"
                  />
                </div>
              </div>
            </div>

            <!-- 右侧面板：部位系数 + 伤害类型 -->
            <div class="boss-right-panel">
              <div class="boss-part-info">
                <div class="boss-part-info__title">部位系数</div>
                <div class="boss-part-info__item boss-part-info__item--soft">
                  <span class="boss-part-info__dot"></span>
                  <span class="boss-part-info__label">软组织</span>
                  <span class="boss-part-info__value">x1.0</span>
                </div>
                <div class="boss-part-info__item boss-part-info__item--heavy">
                  <span class="boss-part-info__dot"></span>
                  <span class="boss-part-info__label">重甲</span>
                  <span class="boss-part-info__value">x0.4</span>
                </div>
                <div class="boss-part-info__item boss-part-info__item--weak">
                  <span class="boss-part-info__dot"></span>
                  <span class="boss-part-info__label">弱点</span>
                  <span class="boss-part-info__value">x2.5</span>
                </div>
                <div class="boss-part-info__divider"></div>
                <div class="boss-part-info__item boss-part-info__item--armor">
                  <span class="boss-part-info__dot"></span>
                  <span class="boss-part-info__label">护甲</span>
                  <span class="boss-part-info__value">减伤</span>
                </div>
              </div>
              <div class="boss-right-legend">
                <div class="boss-right-legend__title">伤害类型</div>
                <span class="boss-right-legend__item boss-right-legend__item--normal">普通</span>
                <span class="boss-right-legend__item boss-right-legend__item--critical">暴击 CRIT!</span>
                <span class="boss-right-legend__item boss-right-legend__item--weak">弱点暴击 WEAK!</span>
                <span class="boss-right-legend__item boss-right-legend__item--bleed">出血</span>
                <span class="boss-right-legend__item boss-right-legend__item--pursuit">追击</span>
                <span class="boss-right-legend__item boss-right-legend__item--true"><span
                    class="boss-right-legend__icon">⚡</span>真实伤害</span>
                <span class="boss-right-legend__item boss-right-legend__item--magic">🔮 魔法伤害</span>
                <span class="boss-right-legend__item boss-right-legend__item--doomsday">💀 削血</span>
                <span class="boss-right-legend__item boss-right-legend__item--judgement">K.O. 终结</span>
              </div>
              <div class="boss-right-summary">
                <div class="boss-right-summary__stats">
                  <span>我的伤害 {{ myBossDamage }}</span>
                  <span>Boss 榜 {{ bossLeaderboardCount }} 人</span>
                </div>
                <p class="boss-right-summary__rule">对 Boss 造成至少 1% 生命值的伤害，才有资格掉落装备与资源。</p>
                <div class="boss-right-summary__drop">
                  <button type="button" @click="openBossDropPool">
                    点击查看 Boss 掉落池
                  </button>
                  <span>{{ bossDropPool.length }} 件掉落物</span>
                </div>
              </div>
            </div>

          </div>
          <!-- 天赋瞬发特效覆盖层（Canvas 像素粒子） -->
          <div ref="talentEffectOverlayRef" class="talent-effect-overlay" aria-hidden="true">
            <PixelEffectCanvas
                class="talent-effect-overlay__canvas"
                :entries="battleEffectEntries"
                :canvas-width="talentOverlayCanvasSize.width"
                :canvas-height="talentOverlayCanvasSize.height"
                :loop="false"
            />
          </div>
          <div class="vote-stage__boss-note vote-stage__boss-note--rules">
            <strong>挂机规则</strong>
            <span>开启条件：<strong>离开页面 60 秒后自动开始挂机。</strong></span>
            <span>战斗效果：每秒自动攻击 1 次，无技能效果，仅仅基础伤害。</span>
            <span>奖励说明：金币和强化石获取减半，天赋点和装备正常掉落。</span>
            <span>结算方式：回到页面后自动弹出结算窗口，显示击杀数与收益。</span>
            <span>温馨提示：挂机最多持续 8 小时；<strong>关闭页面</strong>后挂机仍会继续，服务器维护不会丢失挂机状态。</span>
            <span>体力相关：挂机时伤害不会受到<strong>体力系统为0时锁定为1</strong>的限制。</span>
          </div>
          <div v-if="talentTriggerFeed.length > 0" class="vote-stage__boss-note vote-stage__boss-note--rules">
            <strong>天赋触发</strong>
            <span v-for="entry in talentTriggerFeed.slice(0, 3)" :key="entry.id">
            {{ entry.name }}：{{ entry.message }} <template v-if="entry.extraDamage > 0">（+{{
                entry.extraDamage
              }}）</template>
          </span>
          </div>
        </section>

        <section v-if="isHallRoom" class="feedback-panel feedback-panel--compact hall-leaderboard-panel">
          <div class="hall-leaderboard-panel__head">
            <div>
              <p class="vote-stage__eyebrow">大厅点击总榜</p>
              <strong>第 {{ hallLeaderboardPage + 1 }} 页 · {{ hallLeaderboardRangeLabel }}</strong>
            </div>
            <span
                class="hall-leaderboard-panel__hint">这里包含 0 点击玩家。右侧仍保留前十实时榜，当前区域只在进入大厅时回源一次。</span>
          </div>

          <div v-if="hallLeaderboardLoading" class="leaderboard-list leaderboard-list--empty">
            <p>大厅总榜加载中。</p>
          </div>
          <div v-else-if="hallLeaderboardError" class="leaderboard-list leaderboard-list--empty">
            <p>{{ hallLeaderboardError }}</p>
          </div>
          <div v-else-if="!hallLeaderboardHasPagination" class="leaderboard-list leaderboard-list--empty">
            <p>当前没有 11 名之后的排行榜数据。</p>
          </div>
          <div v-else class="hall-leaderboard-panel__body">
            <div class="hall-leaderboard-panel__grid">
              <div
                  v-for="(column, columnIndex) in hallLeaderboardColumns"
                  :key="`hall-column-${columnIndex}`"
                  class="hall-leaderboard-panel__column"
              >
                <ol v-if="column.length > 0" class="leaderboard-list">
                  <li
                      v-for="entry in column"
                      :key="entry.nickname"
                      class="leaderboard-list__item"
                      :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }"
                  >
                    <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
                    <span class="leaderboard-list__name">{{ entry.nickname }}</span>
                    <strong class="leaderboard-list__count">{{ entry.clickCount }}</strong>
                  </li>
                </ol>
              </div>
            </div>

            <div class="hall-leaderboard-panel__actions">
              <button
                  class="nickname-form__submit nickname-form__submit--ghost"
                  type="button"
                  :disabled="hallLeaderboardPage <= 0"
                  @click="showPreviousHallLeaderboardPage"
              >
                上一页
              </button>
              <span class="hall-leaderboard-panel__page-indicator">
              {{ hallLeaderboardRangeStart }}-{{ hallLeaderboardRangeEnd }}
            </span>
              <button
                  class="nickname-form__submit nickname-form__submit--ghost"
                  type="button"
                  :disabled="hallLeaderboardPage >= hallLeaderboardTotalPages - 1"
                  @click="showNextHallLeaderboardPage"
              >
                下一页
              </button>
            </div>
          </div>
        </section>

        <section v-else class="feedback-panel feedback-panel--compact">
          <p>当前战线：{{ currentRoomDisplay }}。我的战力 {{ formatCompact(myBattlePower) }}，Boss战力
            {{ formatCompact(bossBattlePower) }}。</p>
          <RoomSwitchCooldownTag :cooldown-remaining-seconds="roomJoinCooldownRemainingSeconds"/>
          <button
              class="nickname-form__submit nickname-form__submit--ghost"
              type="button"
              :disabled="roomSwitching"
              :style="{
              position: 'relative',
              minWidth: '172px',
            }"
              @click="exitCurrentRoom"
          >
            <span>退出当前房间</span>
          </button>
        </section>

      </section>

      <section
          v-if="bossDropModalOpen"
          class="boss-drop-modal"
          aria-label="Boss 掉落池"
      >
        <div class="boss-drop-modal__backdrop" @click="closeBossDropPool"></div>
        <article class="boss-drop-modal__card">
          <div class="boss-drop-modal__head">
            <div>
              <p class="vote-stage__eyebrow">Boss 掉落池</p>
              <strong>{{ boss?.name || '当前 Boss' }}</strong>
            </div>
            <button class="nickname-form__ghost" type="button" @click="closeBossDropPool">关闭</button>
          </div>

          <div v-if="bossDropPool.length === 0" class="leaderboard-list leaderboard-list--empty">
            <p>当前 Boss 还没配置掉落池。</p>
          </div>

          <section v-if="bossLoot.length > 0" class="boss-drop-modal__section">
            <div class="boss-drop-modal__section-head">
              <span>装备掉落</span>
              <strong>{{ bossLoot.length }} 件</strong>
            </div>
            <div class="boss-drop-pool__grid">
              <article
                  v-for="item in bossLoot"
                  :key="item.itemId"
                  class="boss-drop-card boss-drop-card--detail"
              >
                <span class="boss-drop-card__type">装备</span>
                <img
                    v-if="item.imagePath"
                    class="boss-drop-card__avatar"
                    :src="item.imagePath"
                    :alt="item.imageAlt || item.itemName || item.itemId"
                />
                <strong>
                  <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                  <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                </strong>
                <ul class="boss-drop-card__details">
                  <li>掉落概率：{{ formatDropRate(item.dropRatePercent) }}</li>
                  <li>稀有度：{{ formatRarityLabel(item.rarity) }}</li>
                  <li>部位：{{ item.slot || '未分类' }}</li>
                  <li v-for="line in formatItemStatLines(item)" :key="line">{{ line }}</li>
                  <li v-if="formatItemStatLines(item).length === 0">暂无词条</li>
                </ul>
              </article>
            </div>
          </section>
          <section class="boss-drop-modal__section">
            <div class="boss-drop-modal__section-head">
              <span>资源掉落</span>
            </div>
            <div class="boss-drop-pool__grid">
              <article class="boss-drop-card boss-drop-card--detail">
                <span class="boss-drop-card__type">金币</span>
                <strong>可获取金币量 : {{ bossGoldRange.min }} ~ {{ bossGoldRange.max }}</strong>
                <ul class="boss-drop-card__details">
                  <li>按击杀结算</li>
                </ul>
              </article>
              <article class="boss-drop-card boss-drop-card--detail">
                <span class="boss-drop-card__type">强化石</span>
                <strong>可获取强化石量 : {{ bossStoneRange.min }} ~ {{ bossStoneRange.max }}</strong>
                <ul class="boss-drop-card__details">
                  <li>按击杀结算</li>
                </ul>
              </article>
              <article class="boss-drop-card boss-drop-card--detail">
                <span class="boss-drop-card__type">天赋点</span>
                <strong>可获取天赋点 : {{ bossTalentPointsOnKill }}（固定）</strong>
                <ul class="boss-drop-card__details">
                  <li>按击杀结算</li>
                </ul>
              </article>
            </div>
          </section>
        </article>
      </section>

      <section v-if="rewardModal" class="boss-drop-modal" aria-label="战利品结算">
        <div class="boss-drop-modal__backdrop" @click="closeRewardModal"></div>
        <article class="boss-drop-modal__card">
          <div class="boss-drop-modal__head">
            <div>
              <p class="vote-stage__eyebrow">{{ rewardModal.mode === 'afk' ? '挂机结算' : '击杀结算' }}</p>
              <strong>{{ rewardModal.title }}</strong>
            </div>
            <button class="nickname-form__ghost" type="button" @click="closeRewardModal">关闭</button>
          </div>
          <section class="boss-drop-modal__section">
            <div class="boss-drop-modal__section-head">
              <span>资源战利品</span>
              <strong>{{ rewardModal.bossName }}</strong>
            </div>
            <div class="leaderboard-list">
              <p>击杀数：{{ rewardModal.kills }}</p>
              <p>金币：+{{ rewardModal.goldTotal }}</p>
              <p>强化石：+{{ rewardModal.stoneTotal }}</p>
            </div>
          </section>
          <section class="boss-drop-modal__section">
            <div class="boss-drop-modal__section-head">
              <span>装备战利品</span>
              <strong>{{ rewardModal.rewards.length }} 件</strong>
            </div>
            <div v-if="rewardModal.rewards.length === 0" class="leaderboard-list leaderboard-list--empty">
              <p>本次未掉落装备。</p>
            </div>
            <div v-else class="reward-grid">
              <article v-for="reward in rewardModal.rewards" :key="`${reward.itemId}-${reward.grantedAt}`"
                       class="reward-grid__item">
                <img
                    v-if="reward.imagePath"
                    class="reward-grid__icon"
                    :src="reward.imagePath"
                    :alt="reward.imageAlt || reward.itemName || reward.itemId"
                />
                <span v-else class="reward-grid__fallback">{{ reward.itemName?.slice(0, 1) || '?' }}</span>
              </article>
            </div>
          </section>
        </article>
      </section>

      <aside class="social-panel social-panel--ranking">
        <section class="social-card leaderboard-card leaderboard-card--stacked">
          <section v-if="!isHallRoom" class="leaderboard-card__section">
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

          <section v-if="isHallRoom" class="leaderboard-card__section">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">点击总榜</p>
              <strong>前 {{ leaderboard.length || 0 }} 名</strong>
            </div>
            <p class="leaderboard-card__hint">点击总榜每分钟整点更新一次</p>

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

    <aside
        class="stamina-float"
        :class="{ 'stamina-float--danger': isStaminaRiskBanned }"
        :title="staminaTooltipText"
        aria-label="体力悬浮入口"
    >
      <div class="stamina-float__bubble">
        <span class="stamina-float__value">{{ staminaFloatLabel }}</span>
        <button
            class="stamina-float__plus"
            type="button"
            :aria-label="`打开商店购买体力，${staminaTooltipText}`"
            @click="openStaminaShop"
        >
          +
        </button>
      </div>
      <div class="stamina-float__tooltip">
        <span>{{ staminaTooltipText }}</span>
        <span v-if="recoveredStamina <= 0">手点伤害固定为 1</span>
        <span>挂机伤害不受体力系统限制</span>
        <span v-if="isStaminaRiskBanned">账号异常，当前不可手点/挂机/购买体力</span>
      </div>
      <div class="stamina-float__rule-card" aria-label="体力规则说明">
        <strong>体力规则</strong>
        <span>1 点体力 = 50 次点击</span>
        <span>每 5 分钟恢复 1 点</span>
        <span>体力耗尽后，点击伤害锁定为1</span>
      </div>
    </aside>
  </div>
</template>
