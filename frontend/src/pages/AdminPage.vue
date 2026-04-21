<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'

import { uploadImageWithPolicy } from '../utils/ossUpload'

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
const heroForm = ref(emptyHeroForm())
const buttonForm = reactive(emptyButtonForm())
const announcementForm = ref(emptyAnnouncementForm())
const lootRows = ref([{ itemId: '', weight: '' }])
const heroLootRows = ref([{ heroId: '', weight: '' }])
const selectedBossTemplateId = ref('')

const adminState = ref(emptyAdminState())
const playerPage = ref(emptyPlayerPage())
const bossHistory = ref([])
const loadingHistory = ref(false)
const announcements = ref([])
const loadingAnnouncements = ref(false)
const messagePage = ref(emptyMessagePage())
const loadingMessages = ref(false)
const loadingPlayers = ref(false)
const uploadingImage = ref(false)

const hasBoss = computed(() => Boolean(adminState.value.boss))
const currentBossId = computed(() => adminState.value.boss?.id || '')
const bossTemplates = computed(() => adminState.value.bossPool ?? [])
const bossCycleEnabled = computed(() => Boolean(adminState.value.bossCycleEnabled))
const selectedBossTemplate = computed(() =>
  bossTemplates.value.find((entry) => entry.id === selectedBossTemplateId.value) ?? null,
)
const equipmentOptions = computed(() => adminState.value.equipment ?? [])
const heroOptions = computed(() => adminState.value.heroes ?? [])
const hasEquipmentTemplates = computed(() => equipmentOptions.value.length > 0)
const hasHeroTemplates = computed(() => heroOptions.value.length > 0)

function emptyAdminState() {
  return {
    buttons: [],
    boss: null,
    bossLeaderboard: [],
    equipment: [],
    heroes: [],
    loot: [],
    heroLoot: [],
    bossCycleEnabled: false,
    bossPool: [],
    playerCount: 0,
    recentPlayerCount: 0,
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

function normalizeHeroLootEntry(entry) {
  return {
    heroId: entry?.heroId || '',
    heroName: entry?.heroName || '',
    imagePath: entry?.imagePath || '',
    imageAlt: entry?.imageAlt || '',
    weight: Number(entry?.weight ?? 0),
    dropRatePercent: Number(entry?.dropRatePercent ?? 0),
    bonusClicks: Number(entry?.bonusClicks ?? 0),
    bonusCriticalChancePercent: Number(entry?.bonusCriticalChancePercent ?? 0),
    bonusCriticalCount: Number(entry?.bonusCriticalCount ?? 0),
    traitType: entry?.traitType || '',
    traitValue: Number(entry?.traitValue ?? 0),
  }
}

function normalizeHeroDefinition(entry) {
  return {
    heroId: entry?.heroId || '',
    name: entry?.name || '',
    imagePath: entry?.imagePath || '',
    imageAlt: entry?.imageAlt || '',
    bonusClicks: Number(entry?.bonusClicks ?? 0),
    bonusCriticalChancePercent: Number(entry?.bonusCriticalChancePercent ?? 0),
    bonusCriticalCount: Number(entry?.bonusCriticalCount ?? 0),
    traitType: entry?.traitType || 'bonus_clicks',
    traitValue: Number(entry?.traitValue ?? 0),
  }
}

function normalizeBossTemplate(entry) {
  return {
    id: entry?.id || '',
    name: entry?.name || '',
    maxHp: Number(entry?.maxHp ?? 0),
    loot: Array.isArray(entry?.loot) ? entry.loot.map(normalizeLootEntry) : [],
    heroLoot: Array.isArray(entry?.heroLoot) ? entry.heroLoot.map(normalizeHeroLootEntry) : [],
  }
}

function normalizeAdminState(payload) {
  return {
    buttons: Array.isArray(payload?.buttons) ? payload.buttons : [],
    boss: payload?.boss ?? null,
    bossLeaderboard: Array.isArray(payload?.bossLeaderboard) ? payload.bossLeaderboard : [],
    equipment: Array.isArray(payload?.equipment) ? payload.equipment : [],
    heroes: Array.isArray(payload?.heroes) ? payload.heroes.map(normalizeHeroDefinition) : [],
    loot: Array.isArray(payload?.loot) ? payload.loot.map(normalizeLootEntry) : [],
    heroLoot: Array.isArray(payload?.heroLoot) ? payload.heroLoot.map(normalizeHeroLootEntry) : [],
    bossCycleEnabled: Boolean(payload?.bossCycleEnabled),
    bossPool: Array.isArray(payload?.bossPool) ? payload.bossPool.map(normalizeBossTemplate) : [],
    playerCount: Number(payload?.playerCount ?? 0),
    recentPlayerCount: Number(payload?.recentPlayerCount ?? 0),
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
    tagsText: '',
    starlightEligible: false,
    imagePath: '',
    imageAlt: '',
  }
}

function emptyHeroForm() {
  return {
    heroId: '',
    name: '',
    imagePath: '',
    imageAlt: '',
    bonusClicks: '',
    bonusCriticalChancePercent: '',
    bonusCriticalCount: '',
    traitType: 'bonus_clicks',
    traitValue: '',
  }
}

function emptyAnnouncementForm() {
  return {
    title: '',
    content: '',
    active: true,
  }
}

function emptyMessagePage() {
  return {
    items: [],
    nextCursor: '',
  }
}

function emptyPlayerPage() {
  return {
    items: [],
    nextCursor: '',
    total: 0,
  }
}

function emptyLootRows() {
  return [{ itemId: '', weight: '' }]
}

function emptyHeroLootRows() {
  return [{ heroId: '', weight: '' }]
}

function formatItemStats(item) {
  return `点击+${item?.bonusClicks ?? 0} 暴击率+${item?.bonusCriticalChancePercent ?? 0}% 暴击+${item?.bonusCriticalCount ?? 0}`
}

function formatHeroTrait(hero) {
  switch (hero?.traitType) {
    case 'bonus_clicks':
      return `被动：额外点击 +${hero?.traitValue ?? 0}`
    case 'critical_chance_percent':
      return `被动：暴击率 +${hero?.traitValue ?? 0}%`
    case 'critical_count_bonus':
      return `被动：暴击额外 +${hero?.traitValue ?? 0}`
    case 'final_damage_percent':
      return `被动：最终伤害 +${hero?.traitValue ?? 0}%`
    default:
      return '被动：暂无'
  }
}

function heroImageAlt(hero) {
  return hero?.imageAlt || hero?.name || hero?.heroId || '英雄头像'
}

function normalizeAnnouncements(payload) {
  return Array.isArray(payload)
    ? payload.map((item) => ({
        id: item?.id || '',
        title: item?.title || '',
        content: item?.content || '',
        publishedAt: Number(item?.publishedAt ?? 0),
        active: Boolean(item?.active),
      }))
    : []
}

function normalizeMessagePage(payload) {
  return {
    items: Array.isArray(payload?.items)
      ? payload.items.map((item) => ({
          id: item?.id || '',
          nickname: item?.nickname || '',
          content: item?.content || '',
          createdAt: Number(item?.createdAt ?? 0),
        }))
      : [],
    nextCursor: payload?.nextCursor || '',
  }
}

function normalizePlayerPage(payload) {
  return {
    items: Array.isArray(payload?.items)
      ? payload.items.map((player) => ({
          nickname: player?.nickname || '',
          clickCount: Number(player?.clickCount ?? 0),
          inventory: Array.isArray(player?.inventory) ? player.inventory : [],
          loadout: normalizeLoadout(player?.loadout),
        }))
      : [],
    nextCursor: payload?.nextCursor || '',
    total: Number(payload?.total ?? 0),
  }
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

function findEquipmentTemplate(itemId) {
  if (!itemId) {
    return null
  }

  return adminState.value.equipment.find((entry) => entry.itemId === itemId) ?? null
}

function findBossTemplate(templateId) {
  if (!templateId) {
    return null
  }

  return bossTemplates.value.find((entry) => entry.id === templateId) ?? null
}

function findHeroTemplate(heroId) {
  if (!heroId) {
    return null
  }

  return heroOptions.value.find((entry) => entry.heroId === heroId) ?? null
}

function applyLootRows(loot) {
  lootRows.value = Array.isArray(loot) && loot.length > 0
    ? loot.map((entry) => ({
        itemId: entry.itemId,
        weight: entry.weight,
      }))
    : emptyLootRows()
}

function applyHeroLootRows(loot) {
  heroLootRows.value = Array.isArray(loot) && loot.length > 0
    ? loot.map((entry) => ({
        heroId: entry.heroId,
        weight: entry.weight,
      }))
    : emptyHeroLootRows()
}

function syncBossTemplateEditor(preferredTemplateId = '') {
  const nextTemplateId = [
    preferredTemplateId,
    selectedBossTemplateId.value,
    adminState.value.boss?.templateId,
    bossTemplates.value[0]?.id,
  ].find((templateId) => findBossTemplate(templateId)) || ''

  selectedBossTemplateId.value = nextTemplateId
  applyLootRows(findBossTemplate(nextTemplateId)?.loot ?? [])
  applyHeroLootRows(findBossTemplate(nextTemplateId)?.heroLoot ?? [])
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
    syncBossTemplateEditor()
  } catch (error) {
    errorMessage.value = error.message || '后台状态加载失败'
  } finally {
    loading.value = false
    checkingSession.value = false
  }
}

async function fetchPlayerPage(cursor = '', append = false) {
  loadingPlayers.value = true
  try {
    const query = new URLSearchParams()
    if (cursor) {
      query.set('cursor', cursor)
    }
    query.set('limit', '50')

    const response = await fetch(`/api/admin/players?${query.toString()}`)
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '玩家列表加载失败'))
    }

    const nextPage = normalizePlayerPage(await response.json())
    playerPage.value = append
      ? {
          items: [...playerPage.value.items, ...nextPage.items],
          nextCursor: nextPage.nextCursor,
          total: nextPage.total,
        }
      : nextPage
  } catch (error) {
    errorMessage.value = error.message || '玩家列表加载失败'
  } finally {
    loadingPlayers.value = false
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

async function fetchAnnouncements() {
  loadingAnnouncements.value = true
  try {
    const response = await fetch('/api/admin/announcements')
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '公告列表加载失败'))
    }
    announcements.value = normalizeAnnouncements(await response.json())
  } catch (error) {
    errorMessage.value = error.message || '公告列表加载失败'
  } finally {
    loadingAnnouncements.value = false
  }
}

async function fetchMessages(cursor = '', append = false) {
  loadingMessages.value = true
  try {
    const query = cursor ? `?cursor=${encodeURIComponent(cursor)}` : ''
    const response = await fetch(`/api/admin/messages${query}`)
    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '留言列表加载失败'))
    }
    const nextPage = normalizeMessagePage(await response.json())
    messagePage.value = append
      ? {
          items: [...messagePage.value.items, ...nextPage.items],
          nextCursor: nextPage.nextCursor,
        }
      : nextPage
  } catch (error) {
    errorMessage.value = error.message || '留言列表加载失败'
  } finally {
    loadingMessages.value = false
  }
}

async function checkSession() {
  try {
    const response = await fetch('/api/admin/session')
    authenticated.value = response.ok
    if (response.ok) {
      await fetchAdminState()
      await Promise.all([fetchAnnouncements(), fetchMessages(), fetchPlayerPage()])
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
    await fetchPlayerPage()
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
  playerPage.value = emptyPlayerPage()
  bossHistory.value = []
  announcements.value = []
  messagePage.value = emptyMessagePage()
  checkingSession.value = false
  successMessage.value = ''
}

async function saveBossTemplate() {
  saving.value = true
  try {
    const method = bossTemplates.value.some((entry) => entry.id === bossForm.value.id)
      ? 'PUT'
      : 'POST'
    const targetId = encodeURIComponent(bossForm.value.id)
    const url = method === 'PUT'
      ? `/api/admin/boss/pool/${targetId}`
      : '/api/admin/boss/pool'

    const response = await fetch(url, {
      method,
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
      throw new Error(await readErrorMessage(response, '保存 Boss 模板失败'))
    }

    selectedBossTemplateId.value = bossForm.value.id
    setSuccess('Boss 模板已保存。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存 Boss 模板失败'
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

    setSuccess(bossCycleEnabled.value ? '当前 Boss 已跳过，循环会继续补位。' : '当前 Boss 已关闭。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '关闭 Boss 失败'
  } finally {
    saving.value = false
  }
}

async function enableBossCycle() {
  saving.value = true
  try {
    const response = await fetch('/api/admin/boss/cycle/enable', {
      method: 'POST',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '开启 Boss 循环失败'))
    }

    setSuccess('Boss 循环已开启。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '开启 Boss 循环失败'
  } finally {
    saving.value = false
  }
}

async function disableBossCycle() {
  saving.value = true
  try {
    const response = await fetch('/api/admin/boss/cycle/disable', {
      method: 'POST',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '停止 Boss 循环失败'))
    }

    setSuccess('Boss 循环已停止，当前 Boss 不会自动续上。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '停止 Boss 循环失败'
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

async function saveHero() {
  saving.value = true
  try {
    const method = adminState.value.heroes.some((entry) => entry.heroId === heroForm.value.heroId)
      ? 'PUT'
      : 'POST'
    const url = method === 'PUT'
      ? `/api/admin/heroes/${encodeURIComponent(heroForm.value.heroId)}`
      : '/api/admin/heroes'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...heroForm.value,
        bonusClicks: Number(heroForm.value.bonusClicks),
        bonusCriticalChancePercent: Number(heroForm.value.bonusCriticalChancePercent),
        bonusCriticalCount: Number(heroForm.value.bonusCriticalCount),
        traitValue: Number(heroForm.value.traitValue),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存英雄失败'))
    }

    setSuccess('英雄模板已保存。')
    heroForm.value = emptyHeroForm()
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存英雄失败'
  } finally {
    saving.value = false
  }
}

async function deleteHero(heroId) {
  saving.value = true
  try {
    const response = await fetch(`/api/admin/heroes/${encodeURIComponent(heroId)}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '删除英雄失败'))
    }

    setSuccess('英雄模板已删除。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '删除英雄失败'
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
    const method = adminState.value.buttons.some((entry) => entry.key === buttonForm.slug)
      ? 'PUT'
      : 'POST'
    const url = method === 'PUT'
      ? `/api/admin/buttons/${encodeURIComponent(buttonForm.slug)}`
      : '/api/admin/buttons'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...buttonForm,
        sort: Number(buttonForm.sort),
        enabled: Boolean(buttonForm.enabled),
        tags: buttonForm.tagsText
          .split(/[,，]/)
          .map((tag) => tag.trim())
          .filter(Boolean),
        starlightEligible: Boolean(buttonForm.starlightEligible),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存按钮失败'))
    }

    setSuccess('按钮配置已保存。')
    Object.assign(buttonForm, emptyButtonForm())
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存按钮失败'
  } finally {
    saving.value = false
  }
}

async function saveAnnouncement() {
  saving.value = true
  try {
    const response = await fetch('/api/admin/announcements', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(announcementForm.value),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存公告失败'))
    }

    announcementForm.value = emptyAnnouncementForm()
    setSuccess('公告已发布。')
    await fetchAnnouncements()
  } catch (error) {
    errorMessage.value = error.message || '保存公告失败'
  } finally {
    saving.value = false
  }
}

async function deleteAnnouncement(id) {
  saving.value = true
  try {
    const response = await fetch(`/api/admin/announcements/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '删除公告失败'))
    }

    setSuccess('公告已删除。')
    await fetchAnnouncements()
  } catch (error) {
    errorMessage.value = error.message || '删除公告失败'
  } finally {
    saving.value = false
  }
}

async function deleteMessage(id) {
  saving.value = true
  try {
    const response = await fetch(`/api/admin/messages/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '删除留言失败'))
    }

    setSuccess('留言已删除。')
    await fetchMessages()
  } catch (error) {
    errorMessage.value = error.message || '删除留言失败'
  } finally {
    saving.value = false
  }
}

async function uploadImageToOSS(event, applyImage, successTip) {
  const file = event.target?.files?.[0]
  if (!file) {
    return
  }

  uploadingImage.value = true
  try {
    const policyResponse = await fetch('/api/admin/oss/sts', {
      method: 'POST',
    })
    if (!policyResponse.ok) {
      throw new Error(await readErrorMessage(policyResponse, '获取 OSS 上传凭证失败'))
    }

    const policy = await policyResponse.json()
    const finalURL = await uploadImageWithPolicy(file, policy)

    applyImage(finalURL, file)
    await nextTick()
    setSuccess(successTip)
  } catch (error) {
    errorMessage.value = error.message || 'OSS 上传失败'
  } finally {
    uploadingImage.value = false
    event.target.value = ''
  }
}

async function uploadButtonImage(event) {
  await uploadImageToOSS(
    event,
    (finalURL, file) => {
      Object.assign(buttonForm, {
        imagePath: finalURL,
        imageAlt: buttonForm.imageAlt || file.name.replace(/\.[^.]+$/, ''),
      })
    },
    '按钮图片已上传到 OSS。',
  )
}

async function uploadHeroImage(event) {
  await uploadImageToOSS(
    event,
    (finalURL, file) => {
      heroForm.value = {
        ...heroForm.value,
        imagePath: finalURL,
        imageAlt: heroForm.value.imageAlt || file.name.replace(/\.[^.]+$/, ''),
      }
    },
    '英雄头像已上传到 OSS。',
  )
}

async function saveLoot() {
  if (!selectedBossTemplateId.value) {
    errorMessage.value = '先选一只 Boss 模板，再配置掉落池。'
    return
  }

  saving.value = true
  try {
    const response = await fetch(`/api/admin/boss/pool/${encodeURIComponent(selectedBossTemplateId.value)}/loot`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        loot: lootRows.value
          .filter((entry) => entry.itemId)
          .map((entry) => ({
            itemId: entry.itemId,
            weight: Number(entry.weight),
          })),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存模板掉落池失败'))
    }

    setSuccess('模板掉落池已保存。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存模板掉落池失败'
  } finally {
    saving.value = false
  }
}

async function saveHeroLoot() {
  if (!selectedBossTemplateId.value) {
    errorMessage.value = '先选一只 Boss 模板，再配置英雄掉落池。'
    return
  }

  saving.value = true
  try {
    const response = await fetch(`/api/admin/boss/pool/${encodeURIComponent(selectedBossTemplateId.value)}/hero-loot`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        loot: heroLootRows.value
          .filter((entry) => entry.heroId)
          .map((entry) => ({
            heroId: entry.heroId,
            weight: Number(entry.weight),
          })),
      }),
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '保存模板英雄池失败'))
    }

    setSuccess('模板英雄池已保存。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '保存模板英雄池失败'
  } finally {
    saving.value = false
  }
}

async function deleteBossTemplate(templateId) {
  saving.value = true
  try {
    const response = await fetch(`/api/admin/boss/pool/${encodeURIComponent(templateId)}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(await readErrorMessage(response, '删除 Boss 模板失败'))
    }

    if (selectedBossTemplateId.value === templateId) {
      selectedBossTemplateId.value = ''
    }
    if (bossForm.value.id === templateId) {
      bossForm.value = {
        id: '',
        name: '',
        maxHp: '',
      }
    }
    setSuccess('Boss 模板已删除。')
    await fetchAdminState()
  } catch (error) {
    errorMessage.value = error.message || '删除 Boss 模板失败'
  } finally {
    saving.value = false
  }
}

function editEquipment(entry) {
  equipmentForm.value = { ...entry }
  activeTab.value = 'equipment'
}

function editButton(entry) {
  Object.assign(buttonForm, {
    slug: entry.key,
    label: entry.label,
    sort: entry.sort,
    enabled: entry.enabled,
    tagsText: Array.isArray(entry.tags) ? entry.tags.join(', ') : '',
    starlightEligible: Boolean(entry.starlightEligible),
    imagePath: entry.imagePath || '',
    imageAlt: entry.imageAlt || '',
  })
  activeTab.value = 'buttons'
}

function editHero(entry) {
  heroForm.value = {
    ...entry,
    bonusClicks: entry.bonusClicks,
    bonusCriticalChancePercent: entry.bonusCriticalChancePercent,
    bonusCriticalCount: entry.bonusCriticalCount,
    traitValue: entry.traitValue,
  }
  activeTab.value = 'heroes'
}

function editBossTemplate(entry) {
  bossForm.value = {
    id: entry.id,
    name: entry.name,
    maxHp: entry.maxHp,
  }
  selectedBossTemplateId.value = entry.id
  applyLootRows(entry.loot)
  applyHeroLootRows(entry.heroLoot)
  activeTab.value = 'boss'
}

function selectBossTemplate(templateId) {
  selectedBossTemplateId.value = templateId
  applyLootRows(findBossTemplate(templateId)?.loot ?? [])
  applyHeroLootRows(findBossTemplate(templateId)?.heroLoot ?? [])
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

function addHeroLootRow() {
  heroLootRows.value.push({
    heroId: '',
    weight: '',
  })
}

function removeHeroLootRow(index) {
  heroLootRows.value.splice(index, 1)
  if (heroLootRows.value.length === 0) {
    addHeroLootRow()
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
          这里管理 Boss、装备、公告、留言和前台按钮，也能把按钮图片直传到 OSS。
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
          <button class="nickname-form__ghost" type="button" @click="Promise.all([fetchAdminState(), fetchPlayerPage()])">
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
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'heroes' }" @click="activeTab = 'heroes'">英雄</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'buttons' }" @click="activeTab = 'buttons'">按钮</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': activeTab === 'content' }" @click="activeTab = 'content'; fetchAnnouncements(); fetchMessages()">内容</button>
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
                <p class="vote-stage__eyebrow">循环状态</p>
                <strong>{{ bossCycleEnabled ? '循环已开启' : '循环未开启' }}</strong>
              </div>

              <p class="social-card__copy">
                当前 Boss：{{ adminState.boss?.name || '暂无活动 Boss' }}
              </p>
              <div class="admin-cycle-pills">
                <span class="boss-stage__pill">
                  {{ bossCycleEnabled ? '击败后会立即补下一只' : '击败后不会自动补位' }}
                </span>
                <span class="boss-stage__pill">
                  {{ adminState.boss?.templateId ? `来源模板 ${adminState.boss.templateId}` : '当前没有绑定模板' }}
                </span>
              </div>

              <div v-if="hasBoss" class="admin-boss-summary">
                <p>实例 ID：{{ adminState.boss.id }}</p>
                <p>状态：{{ adminState.boss.status }} · 血量 {{ adminState.boss.currentHp }}/{{ adminState.boss.maxHp }}</p>
              </div>
              <p v-else class="feedback" style="margin-top: 0.75rem;">
                开启循环后，如果当前没有 Boss，会立刻从 Boss 池里随机刷出一只。
              </p>

              <div class="admin-inline-actions" style="margin-top: 1rem;">
                <button
                  class="nickname-form__submit"
                  type="button"
                  :disabled="saving || bossCycleEnabled"
                  @click="enableBossCycle"
                >
                  开启循环
                </button>
                <button
                  class="nickname-form__ghost"
                  type="button"
                  :disabled="saving || !bossCycleEnabled"
                  @click="disableBossCycle"
                >
                  停止循环
                </button>
                <button
                  v-if="hasBoss"
                  class="nickname-form__ghost"
                  type="button"
                  :disabled="saving"
                  @click="deactivateBoss"
                >
                  {{ bossCycleEnabled ? '跳过当前 Boss' : '关闭当前 Boss' }}
                </button>
              </div>

              <div v-if="hasBoss && adminState.loot.length > 0" style="margin-top: 1rem;">
                <p class="vote-stage__eyebrow">当前实例掉落快照</p>
                <ul class="inventory-list">
                  <li v-for="item in adminState.loot" :key="item.itemId" class="inventory-item">
                    <div>
                      <strong>{{ item.itemName || item.itemId }}</strong>
                      <p>{{ item.itemId }} · {{ item.slot }} · 权重 {{ item.weight }}</p>
                      <p>{{ formatItemStats(item) }}</p>
                    </div>
                  </li>
                </ul>
              </div>

              <div v-if="hasBoss && adminState.heroLoot.length > 0" style="margin-top: 1rem;">
                <p class="vote-stage__eyebrow">当前实例英雄快照</p>
                <ul class="inventory-list">
                  <li v-for="hero in adminState.heroLoot" :key="hero.heroId" class="inventory-item">
                    <div class="admin-entity">
                      <img
                        v-if="hero.imagePath"
                        class="admin-entity__avatar"
                        :src="hero.imagePath"
                        :alt="heroImageAlt(hero)"
                      />
                      <div>
                        <strong>{{ hero.heroName || hero.heroId }}</strong>
                        <p>{{ hero.heroId }} · 权重 {{ hero.weight }}</p>
                        <p>{{ formatItemStats(hero) }}</p>
                        <p>{{ formatHeroTrait(hero) }}</p>
                      </div>
                    </div>
                  </li>
                </ul>
              </div>
            </section>

            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">Boss 池模板</p>
                <strong>{{ bossTemplates.length }} 只</strong>
              </div>

              <form class="admin-form" @submit.prevent="saveBossTemplate">
                <input v-model="bossForm.id" class="nickname-form__input" type="text" placeholder="模板 ID，如 dragon" />
                <input v-model="bossForm.name" class="nickname-form__input" type="text" placeholder="Boss 显示名称" />
                <input v-model="bossForm.maxHp" class="nickname-form__input" type="number" min="1" placeholder="总血量，玩家点击消耗" />
                <button class="nickname-form__submit" type="submit" :disabled="saving">
                  保存 Boss 模板
                </button>
              </form>

              <ul class="inventory-list">
                <li v-for="entry in bossTemplates" :key="entry.id" class="inventory-item inventory-item--stacked">
                  <div>
                    <strong>{{ entry.name }}</strong>
                    <p>{{ entry.id }} · 血量 {{ entry.maxHp }} · 装备 {{ entry.loot.length }} 件 · 英雄 {{ entry.heroLoot.length }} 位</p>
                  </div>
                  <div class="admin-inline-actions admin-inline-actions--stacked">
                    <button class="inventory-item__action" type="button" @click="selectBossTemplate(entry.id)">编辑掉落</button>
                    <button class="inventory-item__action" type="button" @click="editBossTemplate(entry)">编辑模板</button>
                    <button class="nickname-form__ghost" type="button" @click="deleteBossTemplate(entry.id)">删除</button>
                  </div>
                </li>
              </ul>
            </section>
          </div>

          <section class="social-card admin-section-card">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">模板掉落池</p>
              <strong>{{ selectedBossTemplate?.name || selectedBossTemplateId || '未选择模板' }}</strong>
            </div>

            <p class="feedback" style="margin-bottom: 0.75rem;">
              掉落池保存到模板上。Boss 刷出来时会复制一份到当前实例，所以你后面再改模板，不会改到场上的那只。
            </p>

            <p v-if="!hasEquipmentTemplates" class="feedback" style="margin-bottom: 0.75rem;">
              当前还没有装备模板，先去“装备”页创建装备，再回来配置掉落池。
            </p>

            <div class="admin-form admin-form--tight">
              <div v-for="(entry, index) in lootRows" :key="`${selectedBossTemplateId}-${index}-${entry.itemId}`" class="admin-inline-row">
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
                  保存模板掉落池
                </button>
              </div>
            </div>
          </section>

          <section class="social-card admin-section-card">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">模板英雄池</p>
              <strong>{{ selectedBossTemplate?.name || selectedBossTemplateId || '未选择模板' }}</strong>
            </div>

            <p class="feedback" style="margin-bottom: 0.75rem;">
              英雄池和装备池分开配置，Boss 被击败时会分别独立抽取。
            </p>

            <p v-if="!hasHeroTemplates" class="feedback" style="margin-bottom: 0.75rem;">
              当前还没有英雄模板，先去“英雄”页创建模板，再回来配置英雄池。
            </p>

            <div class="admin-form admin-form--tight">
              <div v-for="(entry, index) in heroLootRows" :key="`${selectedBossTemplateId}-hero-${index}-${entry.heroId}`" class="admin-inline-row">
                <div class="admin-loot-select">
                  <select
                    v-model="entry.heroId"
                    class="nickname-form__input"
                    :disabled="!hasHeroTemplates && !entry.heroId"
                  >
                    <option value="">选择已有英雄</option>
                    <option
                      v-if="entry.heroId && !findHeroTemplate(entry.heroId)"
                      :value="entry.heroId"
                    >
                      {{ entry.heroId }}（已删除的英雄）
                    </option>
                    <option
                      v-for="hero in heroOptions"
                      :key="hero.heroId"
                      :value="hero.heroId"
                    >
                      {{ hero.name }} · {{ hero.heroId }}
                    </option>
                  </select>
                  <p v-if="findHeroTemplate(entry.heroId)" class="admin-loot-select__meta">
                    {{ formatItemStats(findHeroTemplate(entry.heroId)) }} · {{ formatHeroTrait(findHeroTemplate(entry.heroId)) }}
                  </p>
                </div>
                <input v-model="entry.weight" class="nickname-form__input" type="number" min="1" placeholder="掉率权重，越大越容易招募" />
                <button class="nickname-form__ghost" type="button" @click="removeHeroLootRow(index)">删</button>
              </div>
              <div class="admin-inline-actions">
                <button class="nickname-form__ghost" type="button" @click="addHeroLootRow">加一行</button>
                <button class="nickname-form__submit" type="button" :disabled="saving" @click="saveHeroLoot">
                  保存模板英雄池
                </button>
              </div>
            </div>
          </section>
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

        <div v-else-if="activeTab === 'heroes'" class="admin-section">
          <div class="admin-grid">
            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">英雄模板</p>
                <strong>{{ adminState.heroes.length }} 位</strong>
              </div>

              <form class="admin-form" @submit.prevent="saveHero">
                <input v-model="heroForm.heroId" class="nickname-form__input" type="text" placeholder="唯一标识，如 spark-cat" />
                <input v-model="heroForm.name" class="nickname-form__input" type="text" placeholder="前台显示名称" />
                <input v-model="heroForm.imagePath" class="nickname-form__input" type="text" placeholder="头像 URL（可选）" />
                <input v-model="heroForm.imageAlt" class="nickname-form__input" type="text" placeholder="头像说明（可选）" />
                <label class="admin-upload">
                  <span>或上传到 OSS（支持 webp）</span>
                  <input type="file" accept="image/*" :disabled="uploadingImage" @change="uploadHeroImage" />
                </label>
                <p v-if="heroForm.imagePath" class="admin-upload__result">
                  当前头像地址：{{ heroForm.imagePath }}
                </p>
                <img
                  v-if="heroForm.imagePath"
                  class="admin-upload__preview admin-upload__preview--avatar"
                  :src="heroForm.imagePath"
                  :alt="heroForm.imageAlt || heroForm.name || heroForm.heroId || '英雄头像预览'"
                />
                <p class="feedback">
                  {{ uploadingImage ? '图片上传中...' : '如果 OSS 还没配置，也可以继续手填图片 URL。' }}
                </p>
                <input v-model="heroForm.bonusClicks" class="nickname-form__input" type="number" min="0" placeholder="点击加成" />
                <input v-model="heroForm.bonusCriticalChancePercent" class="nickname-form__input" type="number" min="0" max="100" placeholder="暴击率加成" />
                <input v-model="heroForm.bonusCriticalCount" class="nickname-form__input" type="number" min="0" placeholder="暴击额外加成" />
                <select v-model="heroForm.traitType" class="nickname-form__input">
                  <option value="bonus_clicks">额外点击</option>
                  <option value="critical_chance_percent">暴击率</option>
                  <option value="critical_count_bonus">暴击额外</option>
                  <option value="final_damage_percent">最终伤害百分比</option>
                </select>
                <input v-model="heroForm.traitValue" class="nickname-form__input" type="number" min="0" placeholder="被动数值" />
                <button class="nickname-form__submit" type="submit" :disabled="saving">
                  保存英雄
                </button>
              </form>
            </section>

            <section class="social-card">
              <ul class="inventory-list">
                <li v-for="hero in adminState.heroes" :key="hero.heroId" class="inventory-item inventory-item--stacked">
                  <div class="admin-entity">
                    <img
                      v-if="hero.imagePath"
                      class="admin-entity__avatar"
                      :src="hero.imagePath"
                      :alt="heroImageAlt(hero)"
                    />
                    <div>
                      <strong>{{ hero.name }}</strong>
                      <p>{{ hero.heroId }}</p>
                      <p>{{ formatItemStats(hero) }}</p>
                      <p>{{ formatHeroTrait(hero) }}</p>
                    </div>
                  </div>
                  <div class="admin-inline-actions">
                    <button class="inventory-item__action" type="button" @click="editHero(hero)">编辑</button>
                    <button class="nickname-form__ghost" type="button" @click="deleteHero(hero.heroId)">删除</button>
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
                <input v-model="buttonForm.tagsText" class="nickname-form__input" type="text" placeholder="标签，逗号分隔，如 日常, 活动" />
                <input v-model="buttonForm.imagePath" class="nickname-form__input" type="text" placeholder="图片 URL（可选，可直接填 OSS/CDN 地址）" />
                <input v-model="buttonForm.imageAlt" class="nickname-form__input" type="text" placeholder="图片说明（可选）" />
                <label class="admin-upload">
                  <span>或上传到 OSS</span>
                  <input type="file" accept="image/*" :disabled="uploadingImage" @change="uploadButtonImage" />
                </label>
                <p v-if="buttonForm.imagePath" class="admin-upload__result">
                  当前图片地址：{{ buttonForm.imagePath }}
                </p>
                <img
                  v-if="buttonForm.imagePath"
                  class="admin-upload__preview"
                  :src="buttonForm.imagePath"
                  :alt="buttonForm.imageAlt || buttonForm.label || '按钮预览图'"
                />
                <p class="feedback">
                  {{ uploadingImage ? '图片上传中...' : '如果 OSS 还没配置，也可以继续手填图片 URL。' }}
                </p>
                <label class="admin-check">
                  <input v-model="buttonForm.enabled" type="checkbox" />
                  启用按钮
                </label>
                <label class="admin-check">
                  <input v-model="buttonForm.starlightEligible" type="checkbox" />
                  参与星光轮换
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
                    <p>{{ Array.isArray(button.tags) && button.tags.length > 0 ? button.tags.join(' / ') : '未打标签' }} · {{ button.starlightEligible ? '参与星光' : '不参与星光' }}</p>
                  </div>
                  <button class="inventory-item__action" type="button" @click="editButton(button)">编辑</button>
                </li>
              </ul>
            </section>
          </div>
        </div>

        <div v-else-if="activeTab === 'content'" class="admin-section">
          <div class="admin-grid">
            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">更新公告</p>
                <strong>{{ announcements.length }} 条</strong>
              </div>

              <form class="admin-form" @submit.prevent="saveAnnouncement">
                <input v-model="announcementForm.title" class="nickname-form__input" type="text" placeholder="公告标题" />
                <textarea v-model="announcementForm.content" class="nickname-form__input admin-textarea" rows="5" placeholder="公告正文，首次进入前台时会弹一次提醒"></textarea>
                <label class="admin-check">
                  <input v-model="announcementForm.active" type="checkbox" />
                  设为生效公告
                </label>
                <button class="nickname-form__submit" type="submit" :disabled="saving">
                  发布公告
                </button>
              </form>

              <div v-if="loadingAnnouncements" class="feedback-panel">
                <p>公告加载中...</p>
              </div>
              <ul v-else class="inventory-list" style="margin-top: 1rem;">
                <li v-for="item in announcements" :key="item.id" class="inventory-item inventory-item--stacked">
                  <div>
                    <strong>{{ item.title }}</strong>
                    <p>{{ item.active ? '生效中' : '未生效' }} · {{ formatTime(item.publishedAt) }}</p>
                    <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
                  </div>
                  <button class="nickname-form__ghost" type="button" @click="deleteAnnouncement(item.id)">删除</button>
                </li>
              </ul>
            </section>

            <section class="social-card">
              <div class="social-card__head">
                <p class="vote-stage__eyebrow">公共留言墙</p>
                <strong>{{ messagePage.items.length }} 条</strong>
              </div>

              <div v-if="loadingMessages" class="feedback-panel">
                <p>留言加载中...</p>
              </div>
              <ul v-else class="inventory-list">
                <li v-for="item in messagePage.items" :key="item.id" class="inventory-item inventory-item--stacked">
                  <div>
                    <strong>{{ item.nickname }}</strong>
                    <p>{{ formatTime(item.createdAt) }}</p>
                    <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
                  </div>
                  <button class="nickname-form__ghost" type="button" @click="deleteMessage(item.id)">删除</button>
                </li>
              </ul>

              <button
                v-if="messagePage.nextCursor"
                class="nickname-form__ghost"
                type="button"
                :disabled="loadingMessages"
                @click="fetchMessages(messagePage.nextCursor, true)"
              >
                加载更多留言
              </button>
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
                <strong>{{ adminState.playerCount }} 人</strong>
              </div>

              <p class="social-card__copy">
                最近 24 小时活跃 {{ adminState.recentPlayerCount }} 人
              </p>

              <div v-if="loadingPlayers" class="feedback-panel">
                <p>玩家列表加载中...</p>
              </div>
              <ul v-else class="inventory-list">
                <li v-for="player in playerPage.items" :key="player.nickname" class="inventory-item inventory-item--stacked">
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

              <button
                v-if="playerPage.nextCursor"
                class="nickname-form__ghost"
                type="button"
                :disabled="loadingPlayers"
                @click="fetchPlayerPage(playerPage.nextCursor, true)"
              >
                加载更多玩家
              </button>
            </section>
          </div>
        </div>
      </article>
    </section>
  </main>
</template>
