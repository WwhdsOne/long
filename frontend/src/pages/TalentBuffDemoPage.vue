<script setup>
import { computed, onBeforeUnmount, ref } from 'vue'

const nowSec = ref(Date.now() / 1000)
const tickTimer = window.setInterval(() => {
  nowSec.value = Date.now() / 1000
}, 200)

const stormTrigger = 50
const armorTrigger = 50
const autoStrikeTrigger = 8
const autoStrikeWindowSec = 5
const silverStormDurationSec = 15

const silverStormEndsAt = ref(0)
const omenStacks = ref(36)
const collapseEndsAt = ref(0)
const collapseDuration = ref(12)
const collapsePartNames = ref(['胸甲核心'])

const softStorm = ref(18)
const heavyStorm = ref(34)
const heavyArmor = ref(27)
const autoStrikeCount = ref(5)
const autoStrikeEndsAt = ref(0)

const partProgressList = computed(() => {
  const autoStrikeCountdown = Math.max(0, autoStrikeEndsAt.value - nowSec.value)
  return [
    {
      key: 'soft-0-1',
      name: '左翼软组织',
      type: 'soft',
      storm: softStorm.value,
      stormProgress: clampPercent((softStorm.value / stormTrigger) * 100),
      armor: 0,
      armorProgress: 0,
      autoStrike: 0,
      autoStrikeProgress: 0,
      autoStrikeCountdown: 0,
      autoStrikeTimeoutPercent: 0,
    },
    {
      key: 'heavy-2-2',
      name: '胸甲核心',
      type: 'heavy',
      storm: heavyStorm.value,
      stormProgress: clampPercent((heavyStorm.value / stormTrigger) * 100),
      armor: heavyArmor.value,
      armorProgress: clampPercent((heavyArmor.value / armorTrigger) * 100),
      autoStrike: autoStrikeCount.value,
      autoStrikeProgress: clampPercent((autoStrikeCount.value / autoStrikeTrigger) * 100),
      autoStrikeCountdown,
      autoStrikeTimeoutPercent: clampPercent((autoStrikeCountdown / autoStrikeWindowSec) * 100),
    },
  ]
})

const silverStormCountdown = computed(() => Math.max(0, Math.ceil(silverStormEndsAt.value - nowSec.value)))
const silverStormActive = computed(() => silverStormEndsAt.value > nowSec.value)
const silverStormPercent = computed(() => clampPercent((silverStormCountdown.value / silverStormDurationSec) * 100))
const showOmenRing = computed(() => omenStacks.value > 0)
const omenRingProgress = computed(() => Math.min(1, omenStacks.value / 120))
const collapseActive = computed(() => collapseEndsAt.value > nowSec.value)
const collapseRemaining = computed(() => Math.max(0, Math.ceil(collapseEndsAt.value - nowSec.value)))
const collapsePercent = computed(() => {
  if (!collapseActive.value || collapseDuration.value <= 0) return 0
  return clampPercent((Math.max(0, collapseEndsAt.value - nowSec.value) / collapseDuration.value) * 100)
})

function clampPercent(value) {
  return Math.max(0, Math.min(100, value))
}

function pushTimer(targetRef, durationSec) {
  targetRef.value = nowSec.value + durationSec
}

function triggerSilverStorm() {
  pushTimer(silverStormEndsAt, silverStormDurationSec)
}

function addOmen(amount) {
  omenStacks.value = Math.max(0, Math.min(120, omenStacks.value + amount))
}

function triggerCollapse() {
  pushTimer(collapseEndsAt, collapseDuration.value)
}

function growStorm() {
  softStorm.value = Math.min(stormTrigger, softStorm.value + 8)
  heavyStorm.value = Math.min(stormTrigger, heavyStorm.value + 6)
}

function growArmor() {
  heavyArmor.value = Math.min(armorTrigger, heavyArmor.value + 10)
}

function growAutoStrike() {
  autoStrikeCount.value = Math.min(autoStrikeTrigger, autoStrikeCount.value + 1)
  pushTimer(autoStrikeEndsAt, autoStrikeWindowSec)
}

function applyPreset(mode) {
  if (mode === 'silver') {
    triggerSilverStorm()
    addOmen(-999)
    softStorm.value = 42
    heavyStorm.value = 18
    return
  }
  if (mode === 'doom') {
    omenStacks.value = 92
    return
  }
  if (mode === 'armor') {
    heavyArmor.value = 50
    autoStrikeCount.value = 7
    pushTimer(autoStrikeEndsAt, autoStrikeWindowSec)
    triggerCollapse()
    return
  }
  if (mode === 'all') {
    triggerSilverStorm()
    omenStacks.value = 120
    triggerCollapse()
    softStorm.value = 50
    heavyStorm.value = 43
    heavyArmor.value = 50
    autoStrikeCount.value = 8
    pushTimer(autoStrikeEndsAt, autoStrikeWindowSec)
  }
}

function resetDemo() {
  silverStormEndsAt.value = 0
  omenStacks.value = 36
  collapseEndsAt.value = 0
  softStorm.value = 18
  heavyStorm.value = 34
  heavyArmor.value = 27
  autoStrikeCount.value = 5
  autoStrikeEndsAt.value = 0
}

onBeforeUnmount(() => {
  clearInterval(tickTimer)
})
</script>

<template>
  <main class="page-shell talent-buff-demo-page">
    <section class="talent-buff-demo">
      <div class="talent-buff-demo__header">
        <p class="vote-stage__eyebrow">内部演示页</p>
        <h1>左侧 Buff 状态 Demo</h1>
        <p class="talent-buff-demo__copy">
          这里只展示战斗页左侧可见的 Buff、倒计时与累计进度，不接后端，不进入公开导航。
        </p>
      </div>

      <div class="talent-buff-demo__toolbar">
        <button type="button" class="nickname-form__submit" @click="applyPreset('silver')">白银风暴预设</button>
        <button type="button" class="nickname-form__submit" @click="applyPreset('doom')">死兆预设</button>
        <button type="button" class="nickname-form__submit" @click="applyPreset('armor')">破甲预设</button>
        <button type="button" class="nickname-form__submit" @click="applyPreset('all')">全开预设</button>
        <button type="button" class="nickname-form__submit nickname-form__submit--ghost" @click="resetDemo">重置</button>
      </div>

      <div class="talent-buff-demo__layout">
        <section class="talent-buff-demo__preview">
          <div class="talent-buff-demo__panel">
            <div class="part-progress-panel">
              <div class="part-progress-panel__title">部位累计进度</div>
              <div v-for="p in partProgressList" :key="p.key" class="part-progress-panel__item">
                <span class="part-progress-panel__name" :class="`part-progress-panel__name--${p.type}`">{{ p.name }}</span>
                <span class="part-progress-panel__track part-progress-panel__track--storm">
                  追击 {{ p.storm }}/{{ stormTrigger }}
                  <span class="part-progress-panel__bar">
                    <span class="part-progress-panel__bar-fill part-progress-panel__bar-fill--storm" :style="{ width: p.stormProgress + '%' }"></span>
                  </span>
                </span>
                <span v-if="p.type === 'heavy'" class="part-progress-panel__track part-progress-panel__track--armor">
                  破甲 {{ p.armor }}/{{ armorTrigger }}
                  <span class="part-progress-panel__bar">
                    <span class="part-progress-panel__bar-fill part-progress-panel__bar-fill--armor" :style="{ width: p.armorProgress + '%' }"></span>
                  </span>
                </span>
                <span v-if="p.type === 'heavy' && p.autoStrike > 0" class="part-progress-panel__track part-progress-panel__track--auto-strike">
                  碎甲重击 {{ p.autoStrike }}/{{ autoStrikeTrigger }}
                  <span class="part-progress-panel__bar">
                    <span class="part-progress-panel__bar-fill part-progress-panel__bar-fill--auto-strike" :style="{ width: p.autoStrikeProgress + '%' }"></span>
                  </span>
                  <span class="part-progress-panel__countdown">{{ Math.ceil(p.autoStrikeCountdown) }}s</span>
                  <span class="part-progress-panel__bar part-progress-panel__bar--timer">
                    <span class="part-progress-panel__bar-fill part-progress-panel__bar-fill--timer" :style="{ width: p.autoStrikeTimeoutPercent + '%' }"></span>
                  </span>
                </span>
              </div>
            </div>

            <div v-if="collapseActive" class="collapse-panel">
              <div class="collapse-panel__title">护甲崩塌</div>
              <div v-for="name in collapsePartNames" :key="name" class="collapse-panel__part">{{ name }}</div>
              <span class="collapse-panel__bar">
                <span class="collapse-panel__bar-fill" :style="{ width: collapsePercent + '%' }"></span>
              </span>
              <span class="collapse-panel__count">{{ collapseRemaining }}s</span>
            </div>

            <div v-if="silverStormActive || showOmenRing" class="talent-status-bar">
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
                  <circle class="talent-omen-ring__track" cx="20" cy="20" r="16" />
                  <circle
                    class="talent-omen-ring__fill"
                    cx="20"
                    cy="20"
                    r="16"
                    :style="{ strokeDasharray: `${omenRingProgress * 100.5} ${100.5 - omenRingProgress * 100.5}` }"
                  />
                </svg>
                死兆 {{ omenStacks }}
              </span>
            </div>
          </div>
        </section>

        <aside class="talent-buff-demo__controls">
          <div class="talent-buff-demo__card">
            <strong>持续 Buff</strong>
            <button type="button" class="nickname-form__submit" @click="triggerSilverStorm">触发白银风暴</button>
            <button type="button" class="nickname-form__submit" @click="triggerCollapse">触发护甲崩塌</button>
          </div>

          <div class="talent-buff-demo__card">
            <strong>叠层 Buff</strong>
            <button type="button" class="nickname-form__submit" @click="addOmen(10)">死兆 +10</button>
            <button type="button" class="nickname-form__submit" @click="addOmen(-10)">死兆 -10</button>
          </div>

          <div class="talent-buff-demo__card">
            <strong>累计进度</strong>
            <button type="button" class="nickname-form__submit" @click="growStorm">推进追击</button>
            <button type="button" class="nickname-form__submit" @click="growArmor">推进破甲</button>
            <button type="button" class="nickname-form__submit" @click="growAutoStrike">推进碎甲重击</button>
          </div>

          <div class="talent-buff-demo__card talent-buff-demo__card--note">
            <strong>本页覆盖的左侧可见状态</strong>
            <span>白银风暴</span>
            <span>死兆</span>
            <span>护甲崩塌</span>
            <span>碎甲重击累计</span>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
