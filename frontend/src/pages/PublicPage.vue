<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const NICKNAME_STORAGE_KEY = 'vote-wall-nickname'

const buttons = ref([])
const leaderboard = ref([])
const boss = ref(null)
const bossLeaderboard = ref([])
const bossLoot = ref([])
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
const activeHudTab = ref('inventory')
const lastUpdatedAt = ref('')
const liveConnected = ref(false)
const criticalBursts = ref({})
const bossHistory = ref([])
const bossHistoryQuery = ref('')
const loadingBossHistory = ref(false)
const bossHistoryLoaded = ref(false)
const bossHistoryError = ref('')

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
const normalDamage = computed(() => combatStats.value?.normalDamage ?? effectiveIncrement.value)
const criticalDamage = computed(() => combatStats.value?.criticalDamage ?? normalDamage.value)
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
  return `点击+${item?.bonusClicks ?? 0} 暴击率+${item?.bonusCriticalChancePercent ?? 0}% 暴击+${item?.bonusCriticalCount ?? 0}`
}

function formatItemStatLines(item) {
  return [
    `点击 +${item?.bonusClicks ?? 0}`,
    `暴击率 +${item?.bonusCriticalChancePercent ?? 0}%`,
    `暴击 +${item?.bonusCriticalCount ?? 0}`,
  ]
}

function normalizeNickname(value) {
  return value.trim()
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
  }
}

function applyState(payload) {
  buttons.value = payload?.buttons ?? []
  leaderboard.value = payload?.leaderboard ?? []
  userStats.value = payload?.userStats ?? null
  boss.value = payload?.boss ?? null
  bossLeaderboard.value = payload?.bossLeaderboard ?? []
  bossLoot.value = payload?.bossLoot ?? []
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
  bossLoot.value = []
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
          <span>掉落池 {{ bossLoot.length }} 件</span>
          <span v-if="lastReward">最近掉落 {{ lastReward.itemName }}</span>
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
              <strong>{{ item.itemName || item.itemId }}</strong>
              <p>{{ item.slot || '未分类' }} · 权重 {{ item.weight }}</p>
              <p>{{ formatItemStats(item) }}</p>
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
            {{ isLoggedIn ? `你现在用的是 ${nickname}。同名会直接并成同一个人。` : '先报个名，HUD 里的背包、属性和装备栏就都会跟你走。' }}
          </p>

          <form class="nickname-form player-hud__form" @submit.prevent="submitNickname">
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
            class="nickname-form__ghost player-hud__reset"
            type="button"
            @click="resetNickname"
          >
            清空昵称
          </button>

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
              :class="{ 'player-hud__tab--active': activeHudTab === 'info' }"
              type="button"
              @click="selectHudTab('info')"
            >
              信息
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
                <li v-for="item in inventory" :key="item.itemId" class="inventory-item">
                  <div>
                    <strong>{{ item.name }}</strong>
                    <p>{{ item.slot || '未分类' }} · 库存 {{ item.quantity }}</p>
                    <p>{{ formatItemStats(item) }}</p>
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
                  <strong>{{ combatStats.criticalChancePercent }}%</strong>
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
                    <strong>{{ loadout.weapon?.name || '未穿戴' }}</strong>
                  </div>
                  <ul v-if="loadout.weapon" class="loadout-slot__attrs">
                    <li v-for="line in formatItemStatLines(loadout.weapon)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>护甲</span>
                    <strong>{{ loadout.armor?.name || '未穿戴' }}</strong>
                  </div>
                  <ul v-if="loadout.armor" class="loadout-slot__attrs">
                    <li v-for="line in formatItemStatLines(loadout.armor)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>饰品</span>
                    <strong>{{ loadout.accessory?.name || '未穿戴' }}</strong>
                  </div>
                  <ul v-if="loadout.accessory" class="loadout-slot__attrs">
                    <li v-for="line in formatItemStatLines(loadout.accessory)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
              </div>
            </section>

            <section v-else class="player-hud__panel player-hud__panel--info">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">信息</p>
                <strong>{{ bossHistory.length }} 条战报</strong>
              </div>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>最近掉落</span>
                  <strong>{{ lastReward?.itemName || '暂无' }}</strong>
                </div>
                <p class="player-hud__note">
                  {{
                    lastReward
                      ? `来自 ${lastReward.bossName || lastReward.bossId || '当前 Boss'}，已经放进你的背包。`
                      : '还没有新的掉落记录。'
                  }}
                </p>
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
                      <strong>{{ item.itemName || item.itemId }}</strong>
                      <p>{{ item.slot || '未分类' }} · 权重 {{ item.weight }}</p>
                      <p>{{ formatItemStats(item) }}</p>
                    </div>
                  </li>
                </ul>
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
            {{ isLoggedIn ? `现在上墙的是 ${nickname}` : '先报个名，再开始冲榜。' }}
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
            <span v-if="lastReward">最近掉落 {{ lastReward.itemName }}</span>
          </div>
        </section>

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

      <aside class="social-panel social-panel--ranking">
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
      </aside>
    </section>
  </main>
</template>
