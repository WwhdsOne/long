<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const NICKNAME_STORAGE_KEY = 'vote-wall-nickname'

const buttons = ref([])
const leaderboard = ref([])
const boss = ref(null)
const bossLeaderboard = ref([])
const myBossStats = ref(null)
const inventory = ref([])
const loadout = ref(emptyLoadout())
const combatStats = ref(defaultCombatStats())
const lastReward = ref(null)
const userStats = ref(null)
const nickname = ref('')
const nicknameDraft = ref('')
const loading = ref(true)
const syncing = ref(false)
const errorMessage = ref('')
const pendingKeys = ref(new Set())
const actioningItemId = ref('')
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
const myBossDamage = computed(() => myBossStats.value?.damage ?? 0)
const effectiveIncrement = computed(() => combatStats.value?.effectiveIncrement ?? 1)
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
    criticalChancePercent: 0,
    criticalCount: 1,
  }
}

function normalizeNickname(value) {
  return value.trim()
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
  userStats.value = payload?.userStats ?? null
  boss.value = payload?.boss ?? null
  bossLeaderboard.value = payload?.bossLeaderboard ?? []
  myBossStats.value = payload?.myBossStats ?? null
  inventory.value = payload?.inventory ?? []
  loadout.value = payload?.loadout ?? emptyLoadout()
  combatStats.value = payload?.combatStats ?? defaultCombatStats()
  lastReward.value = payload?.lastReward ?? null
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
      throw new Error(await readErrorMessage(response, '点击失败，请稍后重试。'))
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

async function postEquipmentAction(itemId, action) {
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

  errorMessage.value = ''

  try {
    await validateNicknameWithServer(nextNickname)

    nickname.value = nextNickname
    nicknameDraft.value = nextNickname
    window.localStorage.setItem(NICKNAME_STORAGE_KEY, nextNickname)
    await loadState()
    connectEventStream()
  } catch (error) {
    errorMessage.value = error.message || '昵称校验失败，请稍后重试。'
  }
}

async function resetNickname() {
  nickname.value = ''
  nicknameDraft.value = ''
  userStats.value = null
  inventory.value = []
  loadout.value = emptyLoadout()
  combatStats.value = defaultCombatStats()
  myBossStats.value = null
  window.localStorage.removeItem(NICKNAME_STORAGE_KEY)
  await loadState()
  connectEventStream()
}

onMounted(async () => {
  const savedNickname = normalizeNickname(window.localStorage.getItem(NICKNAME_STORAGE_KEY) || '')
  if (savedNickname) {
    try {
      await validateNicknameWithServer(savedNickname)
      nickname.value = savedNickname
      nicknameDraft.value = savedNickname
    } catch (error) {
      window.localStorage.removeItem(NICKNAME_STORAGE_KEY)
      errorMessage.value = error.message || '已保存昵称不可用，请换一个试试。'
    }
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
          <span v-if="lastReward">最近掉落 {{ lastReward.itemName }}</span>
        </div>
      </div>
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
            <p class="vote-stage__eyebrow">战斗属性</p>
            <strong>{{ isLoggedIn ? nickname : '未登录' }}</strong>
          </div>

          <div class="me-card__stats">
            <article>
              <span>单击增量</span>
              <strong>+{{ effectiveIncrement }}</strong>
            </article>
            <article>
              <span>暴击率</span>
              <strong>{{ combatStats.criticalChancePercent }}%</strong>
            </article>
            <article>
              <span>暴击总增量</span>
              <strong>+{{ combatStats.criticalCount }}</strong>
            </article>
            <article>
              <span>我的 Boss 伤害</span>
              <strong>{{ myBossDamage }}</strong>
            </article>
          </div>
        </section>

        <section class="social-card loadout-card">
          <div class="social-card__head">
            <p class="vote-stage__eyebrow">装备栏</p>
            <strong>{{ equippedItems.length }} / 3</strong>
          </div>

          <div class="loadout-grid">
            <article class="loadout-slot">
              <span>武器</span>
              <strong>{{ loadout.weapon?.name || '未穿戴' }}</strong>
              <p v-if="loadout.weapon" class="loadout-slot__attrs">
                点击+{{ loadout.weapon.bonusClicks }} 暴击率+{{ loadout.weapon.bonusCriticalChancePercent }}% 暴击+{{ loadout.weapon.bonusCriticalCount }}
              </p>
            </article>
            <article class="loadout-slot">
              <span>护甲</span>
              <strong>{{ loadout.armor?.name || '未穿戴' }}</strong>
              <p v-if="loadout.armor" class="loadout-slot__attrs">
                点击+{{ loadout.armor.bonusClicks }} 暴击率+{{ loadout.armor.bonusCriticalChancePercent }}% 暴击+{{ loadout.armor.bonusCriticalCount }}
              </p>
            </article>
            <article class="loadout-slot">
              <span>饰品</span>
              <strong>{{ loadout.accessory?.name || '未穿戴' }}</strong>
              <p v-if="loadout.accessory" class="loadout-slot__attrs">
                点击+{{ loadout.accessory.bonusClicks }} 暴击率+{{ loadout.accessory.bonusCriticalChancePercent }}% 暴击+{{ loadout.accessory.bonusCriticalCount }}
              </p>
            </article>
          </div>
        </section>

        <section class="social-card inventory-card">
          <div class="social-card__head">
            <p class="vote-stage__eyebrow">背包</p>
            <strong>{{ inventory.length }} 件</strong>
          </div>

          <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
            <p>先去打 Boss 或等后台发装备，背包就会慢慢满起来。</p>
          </div>

          <ul v-else class="inventory-list">
            <li v-for="item in inventory" :key="item.itemId" class="inventory-item">
              <div>
                <strong>{{ item.name }}</strong>
                <p>
                  {{ item.slot || '未分类' }} · 库存 {{ item.quantity }}
                </p>
                <p>
                  点击+{{ item.bonusClicks }} 暴击率+{{ item.bonusCriticalChancePercent }}% 暴击+{{ item.bonusCriticalCount }}
                </p>
              </div>
              <button
                class="inventory-item__action"
                type="button"
                :disabled="!isLoggedIn || actioningItemId === item.itemId"
                @click="item.equipped ? postEquipmentAction(item.itemId, 'unequip') : postEquipmentAction(item.itemId, 'equip')"
              >
                {{ item.equipped ? '卸下' : '穿戴' }}
              </button>
            </li>
          </ul>
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

        <section class="social-card leaderboard-card">
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

        <section v-if="lastReward" class="social-card reward-card">
          <div class="social-card__head">
            <p class="vote-stage__eyebrow">最近掉落</p>
            <strong>{{ lastReward.itemName }}</strong>
          </div>
          <p class="social-card__copy">
            来自 {{ lastReward.bossId || '当前 Boss' }}，已经放进你的背包。
          </p>
        </section>
      </aside>
    </section>
  </main>
</template>
