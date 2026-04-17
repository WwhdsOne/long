<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const buttons = ref([])
const loading = ref(true)
const syncing = ref(false)
const errorMessage = ref('')
const pendingKeys = ref(new Set())
const lastUpdatedAt = ref('')
const liveConnected = ref(false)
const criticalBursts = ref({})

let eventSource
const burstTimers = new Map()

const buttonCount = computed(() => buttons.value.length)
const totalVotes = computed(() =>
  buttons.value.reduce((total, button) => total + button.count, 0),
)
const syncLabel = computed(() => {
  if (syncing.value) {
    return '同步中'
  }

  return liveConnected.value ? '全员在线' : '正在重连'
})

function markUpdated() {
  lastUpdatedAt.value = new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date())
}

function applySnapshot(nextButtons) {
  buttons.value = nextButtons
  pendingKeys.value = new Set()
  syncing.value = false
  markUpdated()
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

  const nextBursts = { ...criticalBursts.value }
  delete nextBursts[key]
  criticalBursts.value = nextBursts
}

function triggerCriticalBurst(key, delta) {
  clearCriticalBurst(key)

  criticalBursts.value = {
    ...criticalBursts.value,
    [key]: {
      label: `暴击 +${delta}`,
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

async function loadButtons() {
  loading.value = true
  syncing.value = true

  try {
    const response = await fetch('/api/buttons')
    if (!response.ok) {
      throw new Error('按钮列表加载失败')
    }

    const data = await response.json()
    applySnapshot(data.buttons)
  } catch (error) {
    errorMessage.value = error.message || '加载失败，请稍后重试。'
  } finally {
    loading.value = false
    syncing.value = false
  }
}

async function clickButton(key) {
  if (pendingKeys.value.has(key)) {
    return
  }

  const nextPending = new Set(pendingKeys.value)
  nextPending.add(key)
  pendingKeys.value = nextPending
  errorMessage.value = ''

  try {
    const response = await fetch(`/api/buttons/${encodeURIComponent(key)}/click`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      let message = '点击失败，请稍后重试。'

      try {
        const payload = await response.json()
        if (payload?.message) {
          message = payload.message
        }
      } catch {
        // Ignore malformed error payloads and keep the fallback copy.
      }

      throw new Error(message)
    }

    const data = await response.json()
    if (data.critical) {
      triggerCriticalBurst(key, data.delta)
    }
    applySnapshot(data.buttons)
    liveConnected.value = true
    errorMessage.value = ''
  } catch (error) {
    const restored = new Set(pendingKeys.value)
    restored.delete(key)
    pendingKeys.value = restored
    errorMessage.value = error.message || '点击失败，请稍后重试。'
  }
}

function connectEventStream() {
  eventSource = new EventSource('/api/events')

  eventSource.onopen = () => {
    liveConnected.value = true
    errorMessage.value = ''
  }

  eventSource.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data)
      if (payload?.buttons) {
        applySnapshot(payload.buttons)
        liveConnected.value = true
        errorMessage.value = ''
      }
    } catch {
      errorMessage.value = '实时消息解析失败，请稍后刷新页面。'
    }
  }

  eventSource.onerror = () => {
    liveConnected.value = false
    errorMessage.value = '实时连接暂时不可用，页面会自动重连。'
  }
}

onMounted(async () => {
  await loadButtons()
  connectEventStream()
})

onBeforeUnmount(() => {
  eventSource?.close()
  burstTimers.forEach((timer) => window.clearTimeout(timer))
  burstTimers.clear()
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
        <h1>来都来了，按一下再走。</h1>
        <p class="hero__lede">
          谁都能点，数字会立刻弹上去。这个页面就该像现场一样热闹一点。
        </p>
      </div>

      <div class="hero__status">
        <span class="live-pill" :class="{ 'live-pill--syncing': syncing }">
          <span class="live-pill__dot"></span>
          {{ syncLabel }}
        </span>
        <span class="hero__time">最近刷新 {{ lastUpdatedAt || '--:--:--' }}</span>
      </div>
    </section>

    <section class="stats-band" aria-label="实时统计">
      <article class="stats-band__card">
        <span class="stats-band__label">当前按钮</span>
        <strong>{{ buttonCount }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">累计点击</span>
        <strong>{{ totalVotes }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">现在状态</span>
        <strong>{{ syncing ? '正在拉新数据' : '可以开点' }}</strong>
      </article>
    </section>

    <section class="vote-stage">
      <div class="vote-stage__head">
        <div>
          <p class="vote-stage__eyebrow">现场投票墙</p>
          <h2>看见哪个想按，就直接拍下去。</h2>
        </div>
        <p v-if="!errorMessage" class="vote-stage__hint">
          所有人看到的是同一份实时总数。
        </p>
      </div>

      <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

      <div v-if="loading" class="feedback-panel">
        <p>正在把现场按钮搬上来...</p>
      </div>

      <div v-else-if="buttons.length === 0" class="feedback-panel">
        <p>还没有按钮上墙，先加一个再回来看看。</p>
      </div>

      <div v-else class="button-grid">
        <button
          v-for="button in buttons"
          :key="button.key"
          class="vote-card"
          :class="{
            'vote-card--image': button.imagePath,
            'vote-card--pending': pendingKeys.has(button.key),
            'vote-card--critical': Boolean(criticalBursts[button.key]),
          }"
          type="button"
          :disabled="pendingKeys.has(button.key)"
          :aria-label="`${button.label}，当前 ${button.count} 票`"
          @click="clickButton(button.key)"
        >
          <span class="vote-card__shine"></span>
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
              pendingKeys.has(button.key)
                ? '正在记票'
                : criticalBursts[button.key]
                  ? '这下真暴击了'
                  : '拍一下 +1'
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
    </section>
  </main>
</template>
