<script setup>
import { computed, onMounted, ref } from 'vue'

const checkingSession = ref(true)
const authenticated = ref(false)
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const activeTab = ref('boss')

const loginForm = ref({
  username: 'admin',
  password: '',
})

const bossForm = ref({
  id: '',
  name: '',
  maxHp: '',
})

const equipmentForm = ref(emptyEquipmentForm())
const buttonForm = ref(emptyButtonForm())
const lootRows = ref([{ itemId: '', weight: '' }])

const adminState = ref(emptyAdminState())
const bossHistory = ref([])
const loadingHistory = ref(false)

const hasBoss = computed(() => Boolean(adminState.value.boss))
const currentBossId = computed(() => adminState.value.boss?.id || '')
const equipmentOptions = computed(() => adminState.value.equipment ?? [])
const hasEquipmentTemplates = computed(() => equipmentOptions.value.length > 0)

function emptyAdminState() {
  return {
    buttons: [],
    boss: null,
    bossLeaderboard: [],
    equipment: [],
    loot: [],
    players: [],
  }
}

function normalizeLoadout(loadout) {
  return {
    weapon: loadout?.weapon ?? null,
    armor: loadout?.armor ?? null,
    accessory: loadout?.accessory ?? null,
  }
}

function normalizeLootEntry(entry) {
  return {
    itemId: entry?.itemId || '',
    itemName: entry?.itemName || '',
    slot: entry?.slot || '',
    weight: Number(entry?.weight ?? 0),
    bonusClicks: Number(entry?.bonusClicks ?? 0),
    bonusCriticalChancePercent: Number(entry?.bonusCriticalChancePercent ?? 0),
    bonusCriticalCount: Number(entry?.bonusCriticalCount ?? 0),
  }
}

function normalizeAdminState(payload) {
  return {
    buttons: Array.isArray(payload?.buttons) ? payload.buttons : [],
    boss: payload?.boss ?? null,
    bossLeaderboard: Array.isArray(payload?.bossLeaderboard) ? payload.bossLeaderboard : [],
    equipment: Array.isArray(payload?.equipment) ? payload.equipment : [],
    loot: Array.isArray(payload?.loot) ? payload.loot.map(normalizeLootEntry) : [],
    players: Array.isArray(payload?.players)
      ? payload.players.map((player) => ({
          nickname: player?.nickname || '',
          clickCount: Number(player?.clickCount ?? 0),
          inventory: Array.isArray(player?.inventory) ? player.inventory : [],
          loadout: normalizeLoadout(player?.loadout),
        }))
      : [],
  }
}

function normalizeBossHistory(payload) {
  if (!Array.isArray(payload)) {
    return []
  }

  return payload.map((entry) => ({
    ...entry,
    loot: Array.isArray(entry?.loot) ? entry.loot.map(normalizeLootEntry) : [],
    damage: Array.isArray(entry?.damage) ? entry.damage : [],
  }))
}

function emptyEquipmentForm() {
  return {
    itemId: '',
    name: '',
    slot: 'weapon',
    bonusClicks: '',
    bonusCriticalChancePercent: '',
    bonusCriticalCount: '',
  }
}

function emptyButtonForm() {
  return {
    slug: '',
    label: '',
    sort: '',
    enabled: true,
    imagePath: '',
    imageAlt: '',
  }
}

function formatItemStats(item) {
  return `点击+${item?.bonusClicks ?? 0} 暴击率+${item?.bonusCriticalChancePercent ?? 0}% 暴击+${item?.bonusCriticalCount ?? 0}`
}

function findEquipmentTemplate(itemId) {
  if (!itemId) {
    return null
  }

  return adminState.value.equipment.find((entry) => entry.itemId === itemId) ?? null
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

function setSuccess(message) {
  successMessage.value = message
  errorMessage.value = ''
}

async function fetchAdminState() {
  loading.value = true

  try {
    const response = await fetch('/api/admin/state')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '后台状态加载失败'))
    }

    adminState.value = normalizeAdminState(await response.json())
    lootRows.value = adminState.value.loot.length > 0
      ? adminState.value.loot.map((entry) => ({
          itemId: entry.itemId,
          weight: entry.weight,
        }))
      : [{ itemId: '', weight: '' }]
  } catch (error) {
    errorMessage.value = error.message || '后台状态加载失败'
  } finally {
    loading.value = false
    checkingSession.value = false
  }
}

async function fetchBossHistory() {
  loadingHistory.value = true
  try {
    const response = await fetch('/api/admin/boss/history')
    if (!response.ok) {
      throw new Error('历史 Boss 加载失败')
    }
    bossHistory.value = normalizeBossHistory(await response.json())
  } catch (error) {
    errorMessage.value = error.message || '历史 Boss 加载失败'
  } finally {
    loadingHistory.value = false
  }
}

async function checkSession() {
  try {
    const response = await fetch('/api/admin/session')
    authenticated.value = response.ok
    if (response.ok) {
      await fetchAdminState()
    } else {
      checkingSession.value = false
    }
  } catch {
    checkingSession.value = false
    authenticated.value = false
  }
}

async function login() {
  saving.value = true

  try {
    const response = await fetch('/api/admin/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(loginForm.value),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '登录失败'))
    }

    authenticated.value = true
    setSuccess('后台已解锁。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '登录失败'
  } finally {
    saving.value = false
  }
}

async function logout() {
  await fetch('/api/admin/logout', { method: 'POST' })
  authenticated.value = false
  adminState.value = emptyAdminState()
  bossHistory.value = []
  checkingSession.value = false
  successMessage.value = ''
}

async function activateBoss() {
  saving.value = true
  try {
    const response = await fetch('/api/admin/boss/activate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: bossForm.value.id,
        name: bossForm.value.name,
        maxHp: Number(bossForm.value.maxHp),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '开启 Boss 失败'))
    }

    setSuccess('Boss 已开启。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '开启 Boss 失败'
  } finally {
    saving.value = false
  }
}

async function deactivateBoss() {
  saving.value = true
  try {
    const response = await fetch('/api/admin/boss/deactivate', {
      method: 'POST',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '关闭 Boss 失败'))
    }

    setSuccess('当前 Boss 已关闭。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '关闭 Boss 失败'
  } finally {
    saving.value = false
  }
}

async function saveEquipment() {
  saving.value = true
  try {
    const method = adminState.value.equipment.some((entry) => entry.itemId === equipmentForm.value.itemId)
      ? 'PUT'
      : 'POST'
    const url = method === 'PUT'
      ? `/api/admin/equipment/${encodeURIComponent(equipmentForm.value.itemId)}`
      : '/api/admin/equipment'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...equipmentForm.value,
        bonusClicks: Number(equipmentForm.value.bonusClicks),
        bonusCriticalChancePercent: Number(equipmentForm.value.bonusCriticalChancePercent),
        bonusCriticalCount: Number(equipmentForm.value.bonusCriticalCount),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存装备失败'))
    }

    setSuccess('装备模板已保存。')
    equipmentForm.value = emptyEquipmentForm()
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存装备失败'
  } finally {
    saving.value = false
  }
}

async function deleteEquipment(itemId) {
  saving.value = true
  try {
    const response = await fetch(`/api/admin/equipment/${encodeURIComponent(itemId)}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '删除装备失败'))
    }

    setSuccess('装备模板已删除。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '删除装备失败'
  } finally {
    saving.value = false
  }
}

async function saveButton() {
  saving.value = true
  try {
    const method = adminState.value.buttons.some((entry) => entry.key === buttonForm.value.slug)
      ? 'PUT'
      : 'POST'
    const url = method === 'PUT'
      ? `/api/admin/buttons/${encodeURIComponent(buttonForm.value.slug)}`
      : '/api/admin/buttons'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...buttonForm.value,
        sort: Number(buttonForm.value.sort),
        enabled: Boolean(buttonForm.value.enabled),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存按钮失败'))
    }

    setSuccess('按钮配置已保存。')
    buttonForm.value = emptyButtonForm()
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存按钮失败'
  } finally {
    saving.value = false
  }
}

async function saveLoot() {
  if (!currentBossId.value) {
    errorMessage.value = '先开启一只 Boss，再配置掉落池。'
    return
  }

  saving.value = true
  try {
    const response = await fetch('/api/admin/boss/loot', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        bossId: currentBossId.value,
        loot: lootRows.value
          .filter((entry) => entry.itemId)
          .map((entry) => ({
            itemId: entry.itemId,
            weight: Number(entry.weight),
          })),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存掉落池失败'))
    }

    setSuccess('掉落池已保存。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存掉落池失败'
  } finally {
    saving.value = false
  }
}

function editEquipment(entry) {
  equipmentForm.value = { ...entry }
  activeTab.value = 'equipment'
}

function editButton(entry) {
  buttonForm.value = {
    slug: entry.key,
    label: entry.label,
    sort: entry.sort,
    enabled: entry.enabled,
    imagePath: entry.imagePath || '',
    imageAlt: entry.imageAlt || '',
  }
  activeTab.value = 'buttons'
}

function addLootRow() {
  lootRows.value.push({
    itemId: '',
    weight: '',
  })
}

function removeLootRow(index) {
  lootRows.value.splice(index, 1)
  if (lootRows.value.length === 0) {
    addLootRow()
  }
}

onMounted(() => {
  checkSession()
})
</script>

<template>
  <main class="page-shell admin-shell">
    <div class="page-shell__glow page-shell__glow--pink"></div>
    <div class="page-shell__glow page-shell__glow--blue"></div>
    <div class="page-shell__glow page-shell__glow--yellow"></div>

    <section class="hero">
      <div class="hero__copy">
        <p class="hero__eyebrow">Long Control Room</p>
        <h1>管理现场、Boss 与掉落。</h1>
        <p class="hero__lede">
          这里先准备装备模板，再开启当前 Boss、配置它的掉落池，也能直接维护前台按钮内容。
        </p>
      </div>

      <div class="hero__status">
        <span class="live-pill">
          <span class="live-pill__dot"></span>
          {{ authenticated ? '后台已解锁' : '等待登录' }}
        </span>
        <a class="hero__admin-link" href="/">返回前台</a>
      </div>
    </section>

    <section v-if="checkingSession" class="admin-card admin-card--single">
      <p class="feedback-panel">正在确认后台会话...</p>
    </section>

    <section v-else-if="!authenticated" class="admin-card admin-card--single">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">后台登录</p>
        <strong>固定口令</strong>
      </div>

      <p class="social-card__copy">先输入后台账号口令，解锁 Boss、装备和按钮配置。</p>

      <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

      <form class="admin-form" @submit.prevent="login">
        <input v-model="loginForm.username" class="nickname-form__input" type="text" placeholder="账号" />
        <input v-model="loginForm.password" class="nickname-form__input" type="password" placeholder="口令" />
        <button class="nickname-form__submit" type="submit" :disabled="saving">
          {{ saving ? '正在解锁...' : '进入后台' }}
        </button>
      </form>
    </section>

    <section v-else class="admin-layout">
      <article class="admin-card admin-card--toolbar">
        <div>
          <p class="vote-stage__eyebrow">控制台</p>
          <strong>{{ adminState.boss?.name || '暂无活动 Boss' }}</strong>
        </div>

        <div class="admin-toolbar__actions">
          <button class="nickname-form__ghost" type="button" @click="fetchAdminState">
            刷新数据
          </button>
          <button class="nickname-form__ghost" type="button" @click="logout">
            退出后台
          </button>
        </div>

        <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>
        <p v-else-if="successMessage" class="feedback">{{ successMessage }}</p>
      </article>

      <article class="admin-card">
        <div class="admin-tabs">
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'boss' }" @click="activeTab = 'boss'">Boss</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'equipment' }" @click="activeTab = 'equipment'">装备</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'buttons' }" @click="activeTab = 'buttons'">按钮</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'history' }" @click="activeTab = 'history'; fetchBossHistory()">历史</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'dashboard' }" @click="activeTab = 'dashboard'">看板</button>
        </div>

        <div v-if="loading" class="feedback-panel">
          <p>后台数据加载中...</p>
        </div>

        <div v-else-if="activeTab === 'boss'" class="admin-section">
          <div class="admin-grid">
            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">开启 / 切换 Boss</p>
                <strong>{{ hasBoss ? adminState.boss.status : '无活动 Boss' }}</strong>
              </div>

              <form class="admin-form" @submit.prevent="activateBoss">
                <input v-model="bossForm.id" class="nickname-form__input" type="text" placeholder="Boss ID（留空自动生成）" />
                <input v-model="bossForm.name" class="nickname-form__input" type="text" placeholder="Boss 显示名称" />
                <input v-model="bossForm.maxHp" class="nickname-form__input" type="number" min="1" placeholder="总血量，玩家点击消耗" />
                <button class="nickname-form__submit" type="submit" :disabled="saving">
                  开启 Boss
                </button>
              </form>

              <button v-if="hasBoss" class="nickname-form__ghost" type="button" :disabled="saving" @click="deactivateBoss">
                关闭当前 Boss
              </button>
            </section>

            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">掉落池</p>
                <strong>{{ currentBossId || '未绑定 Boss' }}</strong>
              </div>

              <ul v-if="hasBoss && adminState.loot.length > 0" class="inventory-list" style="margin-bottom: 0.75rem;">
                <li v-for="item in adminState.loot" :key="item.itemId" class="inventory-item">
                  <div>
                    <strong>{{ item.itemName || item.itemId }}</strong>
                    <p>{{ item.itemId }} · {{ item.slot }} · 权重 {{ item.weight }}</p>
                    <p>{{ formatItemStats(item) }}</p>
                  </div>
                </li>
              </ul>

              <p class="feedback" style="margin-bottom: 0.75rem;">
                当前流程是先开启 Boss，再给这只 Boss 保存掉落池。掉落只会从已有装备模板里选；Boss 被击败后再补配不会补发。
              </p>

              <p v-if="!hasEquipmentTemplates" class="feedback" style="margin-bottom: 0.75rem;">
                当前还没有装备模板，先去“装备”页创建装备，再回来配置掉落池。
              </p>

              <div class="admin-form admin-form--tight">
                <div v-for="(entry, index) in lootRows" :key="`${index}-${entry.itemId}`" class="admin-inline-row">
                  <div class="admin-loot-select">
                    <select
                      v-model="entry.itemId"
                      class="nickname-form__input"
                      :disabled="!hasEquipmentTemplates && !entry.itemId"
                    >
                      <option value="">选择已有装备</option>
                      <option
                        v-if="entry.itemId && !findEquipmentTemplate(entry.itemId)"
                        :value="entry.itemId"
                      >
                        {{ entry.itemId }}（已删除的装备）
                      </option>
                      <option
                        v-for="item in equipmentOptions"
                        :key="item.itemId"
                        :value="item.itemId"
                      >
                        {{ item.name }} · {{ item.itemId }} · {{ item.slot }}
                      </option>
                    </select>
                    <p v-if="findEquipmentTemplate(entry.itemId)" class="admin-loot-select__meta">
                      {{ formatItemStats(findEquipmentTemplate(entry.itemId)) }}
                    </p>
                  </div>
                  <input v-model="entry.weight" class="nickname-form__input" type="number" min="1" placeholder="掉率权重，越大越容易掉落" />
                  <button class="nickname-form__ghost" type="button" @click="removeLootRow(index)">删</button>
                </div>
                <div class="admin-inline-actions">
                  <button class="nickname-form__ghost" type="button" @click="addLootRow">加一行</button>
                  <button class="nickname-form__submit" type="button" :disabled="saving" @click="saveLoot">
                    保存掉落池
                  </button>
                </div>
              </div>
            </section>
          </div>
        </div>

        <div v-else-if="activeTab === 'equipment'" class="admin-section">
          <div class="admin-grid">
            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">装备模板</p>
                <strong>{{ adminState.equipment.length }} 件</strong>
              </div>

              <form class="admin-form" @submit.prevent="saveEquipment">
                <input v-model="equipmentForm.itemId" class="nickname-form__input" type="text" placeholder="唯一标识，如 wood-sword" />
                <input v-model="equipmentForm.name" class="nickname-form__input" type="text" placeholder="前台显示的名称" />
                <select v-model="equipmentForm.slot" class="nickname-form__input">
                  <option value="weapon">weapon</option>
                  <option value="armor">armor</option>
                  <option value="accessory">accessory</option>
                </select>
                <input v-model="equipmentForm.bonusClicks" class="nickname-form__input" type="number" min="0" placeholder="每次点击额外加几票" />
                <input v-model="equipmentForm.bonusCriticalChancePercent" class="nickname-form__input" type="number" min="0" max="100" placeholder="暴击概率 +N%" />
                <input v-model="equipmentForm.bonusCriticalCount" class="nickname-form__input" type="number" min="0" placeholder="暴击时额外加几票" />
                <button class="nickname-form__submit" type="submit" :disabled="saving">
                  保存装备
                </button>
              </form>
            </section>

            <section class="social-card">
              <ul class="inventory-list">
                <li v-for="item in adminState.equipment" :key="item.itemId" class="inventory-item">
                  <div>
                    <strong>{{ item.name }}</strong>
                    <p>{{ item.itemId }} · {{ item.slot }}</p>
                    <p>{{ formatItemStats(item) }}</p>
                  </div>
                  <div class="admin-inline-actions">
                    <button class="inventory-item__action" type="button" @click="editEquipment(item)">编辑</button>
                    <button class="nickname-form__ghost" type="button" @click="deleteEquipment(item.itemId)">删除</button>
                  </div>
                </li>
              </ul>
            </section>
          </div>
        </div>

        <div v-else-if="activeTab === 'buttons'" class="admin-section">
          <div class="admin-grid">
            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">按钮配置</p>
                <strong>{{ adminState.buttons.length }} 个</strong>
              </div>

              <form class="admin-form" @submit.prevent="saveButton">
                <input v-model="buttonForm.slug" class="nickname-form__input" type="text" placeholder="唯一标识，如 feel" />
                <input v-model="buttonForm.label" class="nickname-form__input" type="text" placeholder="前台显示的文字" />
                <input v-model="buttonForm.sort" class="nickname-form__input" type="number" placeholder="排序，数字小的排前面" />
                <input v-model="buttonForm.imagePath" class="nickname-form__input" type="text" placeholder="图片路径（可选）" />
                <input v-model="buttonForm.imageAlt" class="nickname-form__input" type="text" placeholder="图片说明（可选）" />
                <label class="admin-check">
                  <input v-model="buttonForm.enabled" type="checkbox" />
                  启用按钮
                </label>
                <button class="nickname-form__submit" type="submit" :disabled="saving">
                  保存按钮
                </button>
              </form>
            </section>

            <section class="social-card">
              <ul class="inventory-list">
                <li v-for="button in adminState.buttons" :key="button.key" class="inventory-item">
                  <div>
                    <strong>{{ button.label }}</strong>
                    <p>{{ button.key }} · sort {{ button.sort }} · {{ button.enabled ? '启用' : '停用' }}</p>
                  </div>
                  <button class="inventory-item__action" type="button" @click="editButton(button)">编辑</button>
                </li>
              </ul>
            </section>
          </div>
        </div>

        <div v-else-if="activeTab === 'history'" class="admin-section">
          <div v-if="loadingHistory" class="feedback-panel">
            <p>加载历史 Boss...</p>
          </div>
          <div v-else-if="bossHistory.length === 0" class="feedback-panel">
            <p>暂无历史 Boss 记录。</p>
          </div>
          <div v-else class="admin-grid">
            <section v-for="entry in bossHistory" :key="entry.id" class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">{{ entry.status === 'defeated' ? '已击败' : entry.status }}</p>
                <strong>{{ entry.name }}</strong>
              </div>
              <p class="social-card__copy">
                ID: {{ entry.id }} · 血量 {{ entry.currentHp }}/{{ entry.maxHp }}
              </p>

              <div v-if="entry.loot.length > 0" style="margin-top: 0.5rem;">
                <p class="vote-stage__eyebrow">掉落池</p>
                <ul class="inventory-list">
                  <li v-for="item in entry.loot" :key="item.itemId" class="inventory-item">
                    <div>
                      <strong>{{ item.itemName || item.itemId }}</strong>
                      <p>{{ item.itemId }} · {{ item.slot }} · 权重 {{ item.weight }}</p>
                      <p>{{ formatItemStats(item) }}</p>
                    </div>
                  </li>
                </ul>
              </div>

              <div v-if="entry.damage.length > 0" style="margin-top: 0.5rem;">
                <p class="vote-stage__eyebrow">伤害榜</p>
                <ol class="leaderboard-list">
                  <li v-for="d in entry.damage" :key="d.nickname" class="leaderboard-list__item">
                    <span class="leaderboard-list__rank">#{{ d.rank }}</span>
                    <span class="leaderboard-list__name">{{ d.nickname }}</span>
                    <strong class="leaderboard-list__count">{{ d.damage }}</strong>
                  </li>
                </ol>
              </div>
            </section>
          </div>
        </div>

        <div v-else class="admin-section">
          <div class="admin-grid">
            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">Boss 伤害榜</p>
                <strong>{{ adminState.bossLeaderboard.length }} 人</strong>
              </div>

              <ol class="leaderboard-list">
                <li v-for="entry in adminState.bossLeaderboard" :key="entry.nickname" class="leaderboard-list__item">
                  <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
                  <span class="leaderboard-list__name">{{ entry.nickname }}</span>
                  <strong class="leaderboard-list__count">{{ entry.damage }}</strong>
                </li>
              </ol>
            </section>

            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">玩家概览</p>
                <strong>{{ adminState.players.length }} 人</strong>
              </div>

              <ul class="inventory-list">
                <li v-for="player in adminState.players" :key="player.nickname" class="inventory-item inventory-item--stacked">
                  <div>
                    <strong>{{ player.nickname }}</strong>
                    <p>累计点击 {{ player.clickCount }} · 背包 {{ player.inventory.length }} 件</p>
                    <p>
                      穿戴：
                      {{ player.loadout.weapon?.name || '空武器' }} /
                      {{ player.loadout.armor?.name || '空护甲' }} /
                      {{ player.loadout.accessory?.name || '空饰品' }}
                    </p>
                  </div>
                </li>
              </ul>
            </section>
          </div>
        </div>
      </article>
    </section>
  </main>
</template>
