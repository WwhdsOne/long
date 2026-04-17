<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const NICKNAME_STORAGE_KEY = 'vote-wall-nickname'

const buttons = ref([])
const leaderboard = ref([])
const userStats = ref(null)
const nickname = ref('')
const nicknameDraft = ref('')
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
const isLoggedIn = computed(() => nickname.value !== '')
const myClicks = computed(() => userStats.value?.clickCount ?? 0)
const myRank = computed(() => {
  if (!nickname.value) {
    return null
  }

  const matched = leaderboard.value.find((entry) => entry.nickname === nickname.value)
  return matched?.rank ?? null
})

function normalizeNickname(value) {
  return value.trim()
}

function markUpdated() {
  lastUpdatedAt.value = new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date())
}

function applyState(payload) {
  buttons.value = payload?.buttons ?? []
  leaderboard.value = payload?.leaderboard ?? []
  userStats.value = payload?.userStats ?? (nickname.value ? { nickname: nickname.value, clickCount: 0 } : null)
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

function currentNicknameQuery() {
  return nickname.value ? `?nickname=${encodeURIComponent(nickname.value)}` : ''
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

async function clickButton(key) {
  if (!nickname.value || pendingKeys.value.has(key)) {
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
      body: JSON.stringify({
        nickname: nickname.value,
      }),
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
    applyState(data)
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
  eventSource?.close()
  eventSource = new EventSource(`/api/events${currentNicknameQuery()}`)

  eventSource.onopen = () => {
    liveConnected.value = true
    errorMessage.value = ''
  }

  eventSource.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data)
      if (payload?.buttons) {
        applyState(payload)
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

async function submitNickname() {
  const nextNickname = normalizeNickname(nicknameDraft.value)
  if (!nextNickname) {
    errorMessage.value = '先给自己起个名字，再上墙。'
    return
  }

  nickname.value = nextNickname
  nicknameDraft.value = nextNickname
  window.localStorage.setItem(NICKNAME_STORAGE_KEY, nextNickname)
  userStats.value = {
    nickname: nextNickname,
    clickCount: userStats.value?.nickname === nextNickname ? userStats.value.clickCount : 0,
  }
  await loadState()
  connectEventStream()
}

async function resetNickname() {
  nickname.value = ''
  nicknameDraft.value = ''
  userStats.value = null
  window.localStorage.removeItem(NICKNAME_STORAGE_KEY)
  await loadState()
  connectEventStream()
}

onMounted(async () => {
  const savedNickname = normalizeNickname(window.localStorage.getItem(NICKNAME_STORAGE_KEY) || '')
  if (savedNickname) {
    nickname.value = savedNickname
    nicknameDraft.value = savedNickname
  }

  await loadState()
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
        <h1>报个名，再狠狠干一票。</h1>
        <p class="hero__lede">
          昵称只当现场外号用，同名直接算同一个人。点得越猛，榜单爬得越快。
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
        <span class="stats-band__label">我的点击</span>
        <strong>{{ isLoggedIn ? myClicks : '先登录' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">我的排名</span>
        <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
      </article>
    </section>

    <section class="stage-layout">
      <section class="vote-stage">
        <div class="vote-stage__head">
          <div>
            <p class="vote-stage__eyebrow">现场投票墙</p>
            <h2>看见哪个想按，就直接拍下去。</h2>
          </div>
          <p v-if="!errorMessage" class="vote-stage__hint">
            {{ isLoggedIn ? `现在上墙的是 ${nickname}` : '先报个名，再开始冲榜。' }}
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
              'vote-card--locked': !isLoggedIn,
            }"
            type="button"
            :disabled="pendingKeys.has(button.key) || !isLoggedIn"
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
                !isLoggedIn
                  ? '先报个名'
                  : pendingKeys.has(button.key)
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

      <aside class="social-panel">
        <section class="social-card login-card">
          <div class="social-card__head">
            <p class="vote-stage__eyebrow">昵称登录</p>
            <strong>{{ isLoggedIn ? '已经上墙' : '先报个名' }}</strong>
          </div>

          <p class="social-card__copy">
            {{ isLoggedIn ? `你现在用的是 ${nickname}。同名会直接并成同一个人。` : '随便起个现场外号就能开点，不设密码。' }}
          </p>

          <form class="nickname-form" @submit.prevent="submitNickname">
            <input
              v-model="nicknameDraft"
              class="nickname-form__input"
              type="text"
              maxlength="20"
              placeholder="比如：阿明"
            />
            <button class="nickname-form__submit" type="submit">
              {{ isLoggedIn ? '切换昵称' : '进入现场' }}
            </button>
          </form>

          <button
            v-if="isLoggedIn"
            class="nickname-form__ghost"
            type="button"
            @click="resetNickname"
          >
            清空昵称
          </button>
        </section>

        <section class="social-card me-card">
          <div class="social-card__head">
            <p class="vote-stage__eyebrow">个人战绩</p>
            <strong>{{ isLoggedIn ? nickname : '未登录' }}</strong>
          </div>

          <div class="me-card__stats">
            <article>
              <span>我的点击</span>
              <strong>{{ isLoggedIn ? myClicks : '--' }}</strong>
            </article>
            <article>
              <span>当前排名</span>
              <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
            </article>
          </div>
        </section>

        <section class="social-card leaderboard-card">
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
      </aside>
    </section>
  </main>
</template>
