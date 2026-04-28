<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {usePublicPageState} from './publicPageState'
import {effectAssetUrl} from '../utils/effectAssets'
import PixelShatter from '../components/PixelShatter.vue'

const {
  boss,
  bossLeaderboard,
  bossLoot,
  bossGoldRange,
  bossStoneRange,
  bossTalentPointsOnKill,
  leaderboard,
  nickname,
  loading,
  errorMessage,
  pendingKeys,
  damageBursts,
  talentTriggerFeed,
  talentVisualState,
  comboCount,
  stormCombo,
  armorCombo,
  stormProgress,
  armorProgress,
  stormTrigger,
  armorTrigger,
  autoStrikeTrigger,
  comboTriggerFlash,
  comboTimeoutPercent,
  damageStageFx,
  totalVotes,
  isLoggedIn,
  myClicks,
  myRank,
  myBossDamage,
  talentPoints,
  myBossRank,
  effectiveIncrement,
  bossStatusLabel,
  bossProgress,
  formatDropRate,
  formatRarityLabel,
  formatItemStats,
  equipmentNameParts,
  equipmentNameClass,
  rewardModal,
  closeRewardModal,
  clickButton,
  partProgressList,
} = usePublicPageState()

const bossDropModalOpen = ref(false)

const comboColor = computed(() => {
  const t = Math.min(comboCount.value / 200, 1)
  const hue = 120 - t * 120 // 绿 → 黄 → 红
  return `hsl(${hue}, 90%, ${55 - t * 15}%)`
})

const comboIsGold = computed(() => comboCount.value >= 200)

const comboBounce = ref(false)
watch(() => comboCount.value, () => {
  if (comboCount.value > 0) {
    comboBounce.value = true
    setTimeout(() => { comboBounce.value = false }, 200)
  }
})
const talentEffectOverlayRef = ref(null)

// 每秒 tick 驱动倒计时刷新
const nowTick = ref(0)
let tickTimer = 0

// 像素剑光标（全站）
let swordCursor = null
let recoverTimer = 0
let lastAttackTime = 0
let lastPointerDown = 0

function updateCursorPos(e) {
  if (!swordCursor) return
  swordCursor.style.left = e.clientX + 'px'
  swordCursor.style.top = e.clientY + 'px'
}

// 剑挥火花（按部位类型变色，参照 DAMAGE_VARIANTS）
const sparkPresets = {
  weak:     { colors: ['#facc15','#ef4444','#f87171','#fbbf24','#f59e0b','#dc2626'], count: 12, gravity: 0.04 },
  heavy:    { colors: ['#9ca3af','#787888','#64748b','#94a3b8'], count: 5, gravity: 0.18 },
  soft:     { colors: ['#f8fafc','#e2e8f0','#cbd5e1','#fafaff'], count: 6, gravity: 0.08 },
}

function detectCellType(el) {
  const cell = el?.closest?.('.boss-part-cell')
  if (!cell) return 'soft'
  if (cell.classList.contains('boss-part-cell--weak')) return 'weak'
  if (cell.classList.contains('boss-part-cell--heavy')) return 'heavy'
  return 'soft'
}

const sparks = []
let sparksRunning = false
let particleRaf = 0

function spawnSparks(cx, cy, cellType) {
  const preset = sparkPresets[cellType] || sparkPresets.soft
  for (let i = 0; i < preset.count; i++) {
    const el = document.createElement('div')
    el.className = 'sword-spark'
    const angle = Math.PI * 0.1 + Math.random() * Math.PI * 0.5
    const speed = 4 + Math.random() * 10
    const sz = 4 + Math.floor(Math.random() * 8)
    el.style.width = el.style.height = sz + 'px'
    el.style.background = preset.colors[Math.floor(Math.random() * preset.colors.length)]
    el.style.left = cx + 'px'; el.style.top = cy + 'px'
    document.body.appendChild(el)
    sparks.push({ el, x: cx, y: cy, vx: Math.cos(angle) * speed, vy: Math.sin(angle) * speed, life: 1, gravity: preset.gravity, decay: 0.03 + Math.random() * 0.04 })
  }
  if (!sparksRunning) { sparksRunning = true; particleRaf = requestAnimationFrame(updateSparks) }
}

function updateSparks() {
  for (let i = sparks.length - 1; i >= 0; i--) {
    const s = sparks[i]; s.x += s.vx; s.y += s.vy; s.vy += s.gravity; s.life -= s.decay
    s.el.style.left = s.x + 'px'; s.el.style.top = s.y + 'px'; s.el.style.opacity = s.life
    if (s.life <= 0) { s.el.remove(); sparks.splice(i, 1) }
  }
  if (sparks.length > 0) { particleRaf = requestAnimationFrame(updateSparks) }
  else { sparksRunning = false; particleRaf = 0 }
}

function doCursorAttack(e) {
  updateCursorPos(e)
  const now = Date.now()
  if (now - lastAttackTime < 16) return
  lastAttackTime = now
  e.preventDefault()
  if (!swordCursor) return

  clearTimeout(recoverTimer)
  swordCursor.classList.remove('swinging', 'recovering')
  void swordCursor.offsetWidth // 强制重排，连续点击动画重新触发
  swordCursor.classList.add('swinging')

  recoverTimer = setTimeout(() => {
    if (!swordCursor) return
    swordCursor.classList.remove('swinging')
    swordCursor.classList.add('recovering')
  }, 50)

  spawnSparks(e.clientX, e.clientY, detectCellType(e.target))
}

onMounted(() => {
  tickTimer = setInterval(() => {
    nowTick.value++
  }, 250)

  swordCursor = document.getElementById('boss-sword-cursor')

  // 仅战斗区内显示和追踪
  document.addEventListener('pointermove', (e) => {
    const grid = document.querySelector('.boss-part-grid-with-combo')
    if (!grid) return
    if (grid.contains(e.target)) {
      if (swordCursor.style.display === 'none') swordCursor.style.display = 'block'
      updateCursorPos(e)
    } else {
      swordCursor.style.display = 'none'
    }
  })

  document.addEventListener('pointerdown', (e) => {
    const grid = document.querySelector('.boss-part-grid-with-combo')
    if (!grid || !grid.contains(e.target)) return
    lastPointerDown = Date.now()
    doCursorAttack(e)
  })
  document.addEventListener('click', (e) => {
    const grid = document.querySelector('.boss-part-grid-with-combo')
    if (!grid || !grid.contains(e.target)) return
    if (Date.now() - lastPointerDown < 48) return
    doCursorAttack(e)
  })
})

onBeforeUnmount(() => {
  clearInterval(tickTimer)
  cancelAnimationFrame(particleRaf)
  clearTimeout(recoverTimer)
  sparks.forEach(s => s.el.remove())
  sparks.length = 0
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

const bossPartCount = computed(() => {
  if (!boss.value?.parts || !Array.isArray(boss.value.parts)) return 0
  return boss.value.parts.length
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
  return Math.max(0, Math.min(100, (part.currentHp / part.maxHp) * 100))
}

function getBossZoneButtonKey(zone) {
  if (!zone) return ''
  return `boss-part:${zone.x}-${zone.y}`
}

// 纯点击
function clickBossZone(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return
  const el = findBossZoneElement(zone.x, zone.y)
  if (el) {
    el.classList.remove('hit-flash')
    void el.offsetWidth
    el.classList.add('hit-flash')
  }
  clickButton(key)
}

function isBossZoneDisabled(zone) {
  const key = getBossZoneButtonKey(zone)
  return !key || !isLoggedIn.value || !zone?.alive || pendingKeys.value.has(key)
}

function bossZoneAriaLabel(zone) {
  if (!zone) return '空 Boss 分区'
  const label = zone.displayName || partTypeLabels[zone.type] || zone.type
  return `${label} 分区，血量 ${zone.currentHp}/${zone.maxHp}`
}

function zoneDamageBursts(zone) {
  const key = getBossZoneButtonKey(zone)
  if (!key) return []
  return damageBursts.value?.[key] || []
}

// 天赋视觉状态
const talentEdgeGlowClass = computed(() => {
  const vs = talentVisualState.value
  if (vs.deathEcstasyActive) return 'talent-edge-glow--crit'
  if (silverStormActive.value) return ''
  const recent = talentTriggerFeed.value[0]
  if (recent && recent.name === '暴风连击') {
    return 'talent-edge-glow--normal'
  }
  return ''
})

const showSilverFlash = computed(() => silverStormActive.value && silverStormCountdown.value >= 14)
const TALENT_EFFECT_WINDOW_MS = 1350

function slashOverlayStyle() {
  const recent = talentTriggerFeed.value[0]
  if (recent && (recent.name === '暴风连击' || recent.talentId === 'normal_core')) {
    return 'active'
  }
  return ''
}

function latestTrigger(type) {
  return talentTriggerFeed.value.find((e) => e.effectType === type) || null
}

function isTriggerFresh(entry, windowMs = TALENT_EFFECT_WINDOW_MS) {
  if (!entry) return false
  const at = Number(entry.triggeredAt || 0)
  if (!Number.isFinite(at) || at <= 0) return false
  return Date.now() - at <= windowMs
}

function hasRecentTrigger(type, windowMs = TALENT_EFFECT_WINDOW_MS) {
  return isTriggerFresh(latestTrigger(type), windowMs)
}

function triggerKey(type, windowMs = TALENT_EFFECT_WINDOW_MS) {
  const entry = latestTrigger(type)
  if (!isTriggerFresh(entry, windowMs)) return ''
  return `${type}-${entry.id}`
}

function findBossZoneElement(partX, partY) {
  if (typeof window === 'undefined') return null
  const x = Number(partX)
  const y = Number(partY)
  if (!Number.isFinite(x) || !Number.isFinite(y)) return null
  const key = `${Math.floor(x)}-${Math.floor(y)}`
  return document.querySelector(`.boss-part-cell[data-zone-key="${key}"]`)
}

function triggerAnchor(type, windowMs = TALENT_EFFECT_WINDOW_MS) {
  const entry = latestTrigger(type)
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

function triggerOverlayStyle(type, options = {}) {
  const {
    width = 0,
    height = 0,
    offsetX = 0,
    offsetY = 0,
    fallback = {},
  } = options
  const anchor = triggerAnchor(type)
  if (!anchor) {
    return fallback
  }
  return {
    left: `${Math.round(anchor.left + offsetX)}px`,
    top: `${Math.round(anchor.top + offsetY)}px`,
    marginLeft: `${Math.round(-width / 2)}px`,
    marginTop: `${Math.round(-height / 2)}px`,
  }
}

function effectSrc(filename) {
  return effectAssetUrl(filename)
}

function spriteStyle(filename, style = {}) {
  return {
    ...style,
    backgroundImage: `url("${effectSrc(filename)}")`,
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

const showOmenRing = computed(() => talentVisualState.value.omenStacks > 0)
const omenRingProgress = computed(() => {
  const stacks = talentVisualState.value.omenStacks
  return Math.min(1, stacks / 100)
})

const silverStormCountdown = computed(() => {
  void nowTick.value
  const endsAt = talentVisualState.value.silverStormEndsAt
  if (!endsAt) return talentVisualState.value.silverStormRemaining || 0
  return Math.max(0, Math.ceil(endsAt - Date.now() / 1000))
})
const silverStormRemainingSec = computed(() => {
  void nowTick.value
  const endsAt = talentVisualState.value.silverStormEndsAt
  if (!endsAt) return Number(talentVisualState.value.silverStormRemaining) || 0
  return Math.max(0, endsAt - Date.now() / 1000)
})

const silverStormActive = computed(() => {
  void nowTick.value
  if (!talentVisualState.value.silverStormActive) return false
  if (!talentVisualState.value.silverStormEndsAt) return talentVisualState.value.silverStormRemaining > 0
  return talentVisualState.value.silverStormEndsAt > Date.now() / 1000
})
const silverStormTotalSec = ref(0)
watch([silverStormActive, silverStormCountdown], ([active, countdown]) => {
  if (!active) {
    silverStormTotalSec.value = 0
    return
  }
  silverStormTotalSec.value = Math.max(silverStormTotalSec.value, countdown)
}, { immediate: true })
const silverStormPercent = computed(() => {
  const total = silverStormTotalSec.value
  if (!silverStormActive.value || total <= 0) return 0
  return Math.min(100, Math.max(0, (silverStormRemainingSec.value / total) * 100))
})

const collapsePartNames = computed(() => {
  const parts = boss.value?.parts
  if (!Array.isArray(parts)) return []
  return talentVisualState.value.collapsePartKeys
      .map(key => {
        const [x, y] = key.split('-').map(Number)
        const part = parts.find(p => p.x === x && p.y === y)
        if (!part || !part.alive) return null
        return part.displayName || partTypeLabels[part?.type] || key
      })
      .filter(Boolean)
})
const collapseActive = computed(() => collapsePartNames.value.length > 0)

const collapseTick = ref(0)
let collapseTickTimer = 0
watch(collapseActive, (val) => {
  if (val) {
    collapseTickTimer = setInterval(() => { collapseTick.value++ }, 200)
  } else {
    clearInterval(collapseTickTimer)
    collapseTick.value = 0
  }
})
onBeforeUnmount(() => { clearInterval(collapseTickTimer) })

const collapsePercent = computed(() => {
  void collapseTick.value
  const endsAt = talentVisualState.value.collapseEndsAt
  const duration = talentVisualState.value.collapseDuration || 8
  if (!endsAt || !duration) return 0
  const remaining = Math.max(0, endsAt - Date.now() / 1000)
  return Math.min(100, Math.round((remaining / duration) * 100))
})

const collapseRemaining = computed(() => {
  void collapseTick.value
  const endsAt = talentVisualState.value.collapseEndsAt
  if (!endsAt) return 0
  return Math.max(0, Math.ceil(endsAt - Date.now() / 1000))
})
</script>

<template>
  <section class="stats-band stats-band--wide" aria-label="实时统计">
    <article class="stats-band__card">
      <span class="stats-band__label">Boss 部位</span>
      <strong>{{ bossPartCount }}</strong>
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
      <span class="stats-band__label">Boss 排名</span>
      <strong>{{ isLoggedIn ? (myBossRank ? `#${myBossRank}` : '未上榜') : '先登录' }}</strong>
    </article>
    <article class="stats-band__card">
      <span class="stats-band__label">天赋点</span>
      <strong>{{ isLoggedIn ? talentPoints : '--' }}</strong>
    </article>
  </section>

  <section class="stage-layout stage-layout--battle">
    <section class="vote-stage">

      <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

      <section
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
          <div>
            <div class="vote-stage__head">
              <div>
                <h1 class="vote-stage__worldBoss">世界 Boss 战场</h1>
                <p class="vote-stage__eyebrow">当前 Boss</p>
                <h2>{{ boss?.name || '等待 Boss 登场' }}</h2>
              </div>
            </div>
          </div>
          <div class="boss-stage__meta">
            <span class="boss-stage__pill">{{ bossStatusLabel }}</span>
            <strong v-if="boss">HP {{ boss.currentHp }} / {{ boss.maxHp }}</strong>
            <strong v-else>我的伤害 {{ myBossDamage }}</strong>
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
            <!-- 1. 连击框：始终可见 -->
            <div class="combo-box" :class="{ 'combo-box--gold': comboIsGold }">
              <template v-if="comboCount > 0">
                <div class="combo-box__row">
                  <span class="combo-box__count" :class="{ 'combo-box__count--bounce': comboBounce, 'combo-box__count--gold': comboIsGold }" :style="comboIsGold ? {} : { color: comboColor, textShadow: `0 0 14px ${comboColor}80` }">连击 x{{ comboCount }}</span>
                  <span v-if="Math.floor(comboCount / 50) > 0" class="combo-box__bonus">
                    伤害 +{{ Math.floor(comboCount / 50) * 5 }}%
                  </span>
                </div>
                <span class="combo-box__timeout-bar" :style="comboIsGold ? { borderColor: 'rgba(251,191,36,0.5)' } : { borderColor: comboColor + '40' }">
                  <span
                      class="combo-box__timeout-fill"
                      :style="comboIsGold ? { width: comboTimeoutPercent + '%', background: 'linear-gradient(90deg, #fde047, #fbbf24, #f59e0b)', boxShadow: '0 0 10px rgba(251,191,36,0.6)' } : { width: comboTimeoutPercent + '%', background: `linear-gradient(90deg, ${comboColor}, ${comboColor}cc)`, boxShadow: `0 0 8px ${comboColor}80` }"
                  ></span>
                </span>
                <span class="combo-box__timeout-text" :style="{ color: comboColor }">
                  {{ Math.ceil(comboTimeoutPercent / 20) }}s
                </span>
              </template>
              <template v-else>
                <span class="combo-box__count combo-box__count--idle">
                  连击 x 0
                </span>
                <span class="combo-box__timeout-bar combo-box__timeout-bar--empty">
                  <span class="combo-box__timeout-fill"></span>
                </span>
              </template>
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
                <span v-if="p.type === 'heavy' && p.autoStrike > 0" class="part-progress-panel__track part-progress-panel__track--auto-strike">
                  碎甲重击 {{ p.autoStrike }}/{{ autoStrikeTrigger }}
                  <span class="part-progress-panel__bar"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--auto-strike"
                      :style="{ width: p.autoStrikeProgress + '%' }"></span></span>
                  <span class="part-progress-panel__countdown">{{ Math.ceil(p.autoStrikeCountdown) }}s</span>
                  <span class="part-progress-panel__bar part-progress-panel__bar--timer"><span
                      class="part-progress-panel__bar-fill part-progress-panel__bar-fill--timer"
                      :style="{ width: p.autoStrikeTimeoutPercent + '%' }"></span></span>
                </span>
              </div>
            </div>

            <!-- 4. 崩塌倒计时 -->
            <div v-if="collapseActive" class="collapse-panel">
              <div class="collapse-panel__title">护甲崩塌</div>
              <div v-for="name in collapsePartNames" :key="name" class="collapse-panel__part">{{ name }}</div>
              <span class="collapse-panel__bar">
                <span class="collapse-panel__bar-fill"
                      :style="{ width: collapsePercent + '%' }"></span>
              </span>
              <span class="collapse-panel__count">{{ collapseRemaining }}s</span>
            </div>
          </div>



          <!-- 右侧：5×5 Boss 网格 + 连击计数 -->
          <div class="boss-part-grid-with-combo">
            <div class="boss-part-grid">
              <div v-for="(row, yi) in bossZones" :key="yi" class="boss-part-grid__row">
                <button
                    v-for="(zone, xi) in row"
                    :key="yi + '-' + xi"
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
                <span v-if="burst.label" class="boss-zone-button__damage-label">{{ burst.label }}</span>- {{ burst.value }}
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
                    </div>
                    <PixelShatter
                        v-if="isPartCollapsed(zone)"
                    />
                    <img
                        v-if="isPartDoomMarked(zone)"
                        class="boss-part-cell__crosshair"
                        :src="effectSrc('crosshair-mark.png')"
                        alt="末日审判标记"
                    />
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
                      <span>血量 : {{ zone.currentHp }}/{{ zone.maxHp }}</span><br>
                      <span>护甲 : {{ isPartCollapsed(zone) ? 0 : zone.armor }}</span>
                    </div>
                  </template>
                  <span v-else class="boss-part-cell__empty"></span>
                </button>
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
              <span class="boss-right-legend__item boss-right-legend__item--pursuit">追击</span>
              <span class="boss-right-legend__item boss-right-legend__item--true">重甲</span>
              <span class="boss-right-legend__item boss-right-legend__item--doomsday">💀 削血</span>
              <span class="boss-right-legend__item boss-right-legend__item--judgement">K.O. 终结</span>
            </div>
          </div>

        </div>
        <div id="boss-sword-cursor"></div>
        <!-- 天赋瞬发特效覆盖层 -->
        <div ref="talentEffectOverlayRef" class="talent-effect-overlay" aria-hidden="true">
          <div v-if="hasRecentTrigger('storm_combo')"
               :key="triggerKey('storm_combo')"
               class="talent-spritesheet talent-spritesheet--slash talent-spritesheet--playing"
               :style="spriteStyle('slash-green-45.png', triggerOverlayStyle('storm_combo', {
                 width: 480,
                 height: 480,
                 fallback: { top: '50%', left: '50%', marginLeft: '-240px', marginTop: '-240px' },
               }))"></div>
          <img v-if="hasRecentTrigger('auto_strike')"
               :key="triggerKey('auto_strike')"
               class="talent-effect-overlay__hammer"
               :src="effectSrc('hammer-impact.png')"
               :style="triggerOverlayStyle('auto_strike', {
                 width: 240,
                 height: 480,
                 offsetY: -80,
                 fallback: { top: '10%', left: '50%', marginLeft: '-120px' },
               })"
               alt="碎甲重击"/>
          <img v-if="hasRecentTrigger('bleed')"
               :key="triggerKey('bleed')"
               class="talent-effect-overlay__bleed"
               :src="effectSrc('blood-drop-large.png')"
               :style="triggerOverlayStyle('bleed', {
                 width: 120,
                 height: 160,
                 offsetY: -30,
                 fallback: { top: '20%', right: '30%' },
               })"
               alt="致命出血"/>
          <div v-if="hasRecentTrigger('omen_harvest')"
               :key="triggerKey('omen_harvest')"
               class="talent-spritesheet talent-spritesheet--scythe talent-spritesheet--playing"
               :style="spriteStyle('scythe-sweep-spritesheet.png', triggerOverlayStyle('omen_harvest', {
                 width: 480,
                 height: 480,
                 fallback: { top: '50%', left: '50%', marginLeft: '-240px', marginTop: '-240px' },
               }))"></div>
          <div v-if="hasRecentTrigger('final_cut')"
               :key="triggerKey('final_cut')"
               class="talent-spritesheet talent-spritesheet--blade talent-spritesheet--playing"
               :style="spriteStyle('blood-blade-spritesheet.png', triggerOverlayStyle('final_cut', {
                 width: 480,
                 height: 480,
                 fallback: { top: '50%', left: '50%', marginLeft: '-240px', marginTop: '-240px' },
               }))"></div>
          <img v-if="hasRecentTrigger('collapse_trigger')"
               :key="triggerKey('collapse_trigger')"
               class="talent-effect-overlay__crack-burst"
               :src="effectSrc('crack-pattern-3.png')"
               :style="triggerOverlayStyle('collapse_trigger', {
                 width: 320,
                 height: 320,
                 fallback: { top: '50%', left: '50%', marginLeft: '-160px', marginTop: '-160px' },
               })"
               alt="结构崩塌"/>
          <div v-if="hasRecentTrigger('collapse_trigger')"
               :key="`armor-break-${triggerKey('collapse_trigger')}`"
               class="talent-spritesheet talent-spritesheet--armor-break talent-spritesheet--playing"
               :style="spriteStyle('armor-break-spritesheet.png', triggerOverlayStyle('collapse_trigger', {
                 width: 320,
                 height: 320,
                 fallback: { top: '50%', left: '50%', marginLeft: '-160px', marginTop: '-160px' },
               }))"></div>
          <div v-if="hasRecentTrigger('judgment_day')"
               :key="triggerKey('judgment_day')"
               class="talent-spritesheet talent-spritesheet--hammer-crack talent-spritesheet--playing"
               :style="spriteStyle('hammer-crack-spritesheet.png', triggerOverlayStyle('judgment_day', {
                 width: 320,
                 height: 320,
                 fallback: { top: '50%', left: '50%', marginLeft: '-160px', marginTop: '-160px' },
               }))"></div>
          <div v-if="hasRecentTrigger('doom_judgment')"
               :key="triggerKey('doom_judgment')"
               class="talent-spritesheet talent-spritesheet--ring talent-spritesheet--playing"
               :style="spriteStyle('green-ring-spritesheet.png', triggerOverlayStyle('doom_judgment', {
                 width: 480,
                 height: 480,
                 fallback: { top: '50%', left: '50%', marginLeft: '-240px', marginTop: '-240px' },
               }))"></div>
          <div v-if="hasRecentTrigger('silver_storm')"
               :key="triggerKey('silver_storm')"
               class="talent-spritesheet talent-spritesheet--scythe talent-spritesheet--playing"
               :style="spriteStyle('scythe-sweep-spritesheet.png', triggerOverlayStyle('silver_storm', {
                 width: 480,
                 height: 480,
                 fallback: { top: '40%', left: '30%', marginLeft: '-240px', marginTop: '-240px' },
               }))"></div>
        </div>
        <div
            v-if="silverStormActive || talentVisualState.omenStacks > 0"
            class="talent-status-bar">
          <div v-if="silverStormActive" class="talent-status-chip talent-status-chip--silver">
            <span class="talent-status-chip__head">
              <span class="talent-status-chip__label">白银风暴</span>
              <span class="talent-status-chip__count">{{ silverStormCountdown }}s</span>
            </span>
            <span class="talent-status-chip__bar">
              <span
                  class="talent-status-chip__bar-fill talent-status-chip__bar-fill--silver"
                  :style="{ width: silverStormPercent + '%' }"
              ></span>
            </span>
          </div>
          <span v-if="showOmenRing" class="talent-status-bar__item talent-status-bar__item--danger talent-omen-ring">
            <svg class="talent-omen-ring__svg" viewBox="0 0 40 40">
              <circle class="talent-omen-ring__track" cx="20" cy="20" r="16"/>
              <circle class="talent-omen-ring__fill" cx="20" cy="20" r="16"
                      :style="{ strokeDasharray: `${omenRingProgress * 100.5} ${100.5 - omenRingProgress * 100.5}` }"/>
            </svg>
            死兆 {{ talentVisualState.omenStacks }}
          </span>
        </div>
        <div class="vote-stage__boss-hud-stats">
          <span>我的伤害 {{ myBossDamage }}</span>
          <span>Boss 榜 {{ bossLeaderboard.length }} 人</span>
          <span>掉落池 {{ bossDropPool.length }} 件</span>
        </div>
        <div class="vote-stage__boss-note">
          <span>只有对 Boss 造成至少 1% 生命值的伤害，才有资格掉落装备与资源。</span>
          <span v-if="boss" class="boss-drop-link">
              <button type="button" @click="openBossDropPool">
                点击查看 Boss 掉落池
              </button>
              <span>{{ bossDropPool.length }} 件掉落物</span>
            </span>
        </div>
        <div class="vote-stage__boss-note vote-stage__boss-note--rules">
          <strong>挂机规则</strong>
          <span>开启条件：离开页面 60 秒后自动开始挂机。</span>
          <span>战斗效果：每秒自动攻击 1 次，挂机效率约为手动操作四分之一。</span>
          <span>奖励说明：金币和强化石获取减半，天赋点和装备正常掉落。</span>
          <span>结算方式：回到页面后自动弹出结算窗口，显示击杀数与收益。</span>
          <span>温馨提示：挂机最多持续 8 小时；关闭页面后挂机仍会继续，服务器重启后不会丢失挂机状态。</span>
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
                <li>{{ formatItemStats(item) }}</li>
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
</template>
