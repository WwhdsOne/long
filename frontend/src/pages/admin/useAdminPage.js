import { computed, nextTick, onMounted, reactive, ref } from 'vue'

import {
  emptyAdminState,
  emptyAnnouncementForm,
  emptyBossHistoryPage,
  emptyButtonForm,
  emptyButtonPage,
  emptyEquipmentForm,
  emptyEquipmentPage,
  emptyHeroForm,
  emptyHeroLootRows,
  emptyLootRows,
  emptyMessagePage,
  emptyPlayerPage,
  formatHeroTrait,
  formatItemStats,
  formatTime,
  heroImageAlt,
  normalizeAdminState,
  normalizeAnnouncements,
  normalizeBossHistoryPage,
  normalizeButtonPage,
  normalizeEquipmentPage,
  normalizeMessagePage,
  normalizePlayerPage,
} from './state'
import { fetchWithTimeout, readErrorMessage } from './request'
import { createAdminPageActions } from './useAdminPageActions'
import { uploadImageWithPolicy } from '../../utils/ossUpload'

export function useAdminPage() {
  const checkingSession = ref(true)
  const authenticated = ref(false)
  const loading = ref(false)
  const saving = ref(false)
  const errorMessage = ref('')
  const successMessage = ref('')
  const activeTab = ref('boss')
  const uploadingImage = ref(false)

  const loginForm = ref({ username: 'admin', password: '' })
  const bossForm = ref({ id: '', name: '', maxHp: '' })
  const equipmentForm = ref(emptyEquipmentForm())
  const heroForm = ref(emptyHeroForm())
  const buttonForm = reactive(emptyButtonForm())
  const announcementForm = ref(emptyAnnouncementForm())
  const lootRows = ref(emptyLootRows())
  const heroLootRows = ref(emptyHeroLootRows())
  const selectedBossTemplateId = ref('')

  const adminState = ref(emptyAdminState())
  const buttonPage = ref(emptyButtonPage())
  const equipmentPage = ref(emptyEquipmentPage())
  const playerPage = ref(emptyPlayerPage())
  const bossHistoryPage = ref(emptyBossHistoryPage())
  const announcements = ref([])
  const messagePage = ref(emptyMessagePage())

  const loadingHistory = ref(false)
  const loadingButtons = ref(false)
  const loadingEquipment = ref(false)
  const loadingAnnouncements = ref(false)
  const loadingMessages = ref(false)
  const loadingPlayers = ref(false)

  const hasBoss = computed(() => Boolean(adminState.value.boss))
  const bossTemplates = computed(() => adminState.value.bossPool ?? [])
  const bossCycleEnabled = computed(() => Boolean(adminState.value.bossCycleEnabled))
  const selectedBossTemplate = computed(() =>
    bossTemplates.value.find((entry) => entry.id === selectedBossTemplateId.value) ?? null,
  )
  const equipmentOptions = computed(() => equipmentPage.value.items ?? [])
  const heroOptions = computed(() => adminState.value.heroes ?? [])
  const hasEquipmentTemplates = computed(() => equipmentPage.value.total > 0)
  const hasHeroTemplates = computed(() => heroOptions.value.length > 0)

  function setSuccess(message) {
    successMessage.value = message
    errorMessage.value = ''
  }

  function findEquipmentTemplate(itemId) {
    if (!itemId) {
      return null
    }
    return equipmentPage.value.items.find((entry) => entry.itemId === itemId) ?? null
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
      ? loot.map((entry) => ({ itemId: entry.itemId, weight: entry.weight }))
      : emptyLootRows()
  }

  function applyHeroLootRows(loot) {
    heroLootRows.value = Array.isArray(loot) && loot.length > 0
      ? loot.map((entry) => ({ heroId: entry.heroId, weight: entry.weight }))
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

  async function uploadImageToOSS(event, file, applyImage, successTip) {
    uploadingImage.value = true
    try {
      const policyResponse = await fetchWithTimeout('/api/admin/oss/sts', { method: 'POST' })
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

  function addLootRow() {
    lootRows.value.push({ itemId: '', weight: '' })
  }

  function addHeroLootRow() {
    heroLootRows.value.push({ heroId: '', weight: '' })
  }

  async function fetchAdminState() {
    loading.value = true
    try {
      const response = await fetchWithTimeout('/api/admin/state')
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '后台状态加载失败'))
      }

      adminState.value = normalizeAdminState(await response.json())
      syncBossTemplateEditor()
    } catch (error) {
      errorMessage.value = error.message || '后台状态加载失败'
    } finally {
      loading.value = false
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

      const response = await fetchWithTimeout(`/api/admin/players?${query.toString()}`)
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

  async function fetchButtonPage(page = 1) {
    loadingButtons.value = true
    try {
      const response = await fetchWithTimeout(`/api/admin/buttons?page=${page}&pageSize=20`)
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '按钮列表加载失败'))
      }

      buttonPage.value = normalizeButtonPage(await response.json())
    } catch (error) {
      errorMessage.value = error.message || '按钮列表加载失败'
    } finally {
      loadingButtons.value = false
    }
  }

  async function fetchEquipmentPage(page = 1) {
    loadingEquipment.value = true
    try {
      const response = await fetchWithTimeout(`/api/admin/equipment?page=${page}&pageSize=20`)
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '装备列表加载失败'))
      }

      equipmentPage.value = normalizeEquipmentPage(await response.json())
    } catch (error) {
      errorMessage.value = error.message || '装备列表加载失败'
    } finally {
      loadingEquipment.value = false
    }
  }

  async function fetchBossHistory(page = 1) {
    loadingHistory.value = true
    try {
      const response = await fetchWithTimeout(`/api/admin/boss/history?page=${page}&pageSize=20`)
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '历史 Boss 加载失败'))
      }
      bossHistoryPage.value = normalizeBossHistoryPage(await response.json())
    } catch (error) {
      errorMessage.value = error.message || '历史 Boss 加载失败'
    } finally {
      loadingHistory.value = false
    }
  }

  async function fetchAnnouncements() {
    loadingAnnouncements.value = true
    try {
      const response = await fetchWithTimeout('/api/admin/announcements')
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
      const response = await fetchWithTimeout(`/api/admin/messages${query}`)
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

  async function refreshAll() {
    await Promise.all([
      fetchAdminState(),
      fetchPlayerPage(),
      fetchButtonPage(buttonPage.value.page),
      fetchEquipmentPage(equipmentPage.value.page),
    ])
  }

  async function checkSession() {
    try {
      const response = await fetchWithTimeout('/api/admin/session')
      authenticated.value = response.ok
      checkingSession.value = false
      if (!response.ok) {
        return
      }

      await fetchAdminState()
      await Promise.all([
        fetchAnnouncements(),
        fetchMessages(),
        fetchPlayerPage(),
        fetchButtonPage(),
        fetchEquipmentPage(),
      ])
    } catch {
      checkingSession.value = false
      authenticated.value = false
    }
  }

  async function login() {
    saving.value = true
    try {
      const response = await fetchWithTimeout('/api/admin/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(loginForm.value),
      })

      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '登录失败'))
      }

      authenticated.value = true
      checkingSession.value = false
      setSuccess('后台已解锁。')
      await fetchAdminState()
      await Promise.all([fetchPlayerPage(), fetchButtonPage(), fetchEquipmentPage()])
    } catch (error) {
      errorMessage.value = error.message || '登录失败'
    } finally {
      saving.value = false
    }
  }

  async function logout() {
    await fetchWithTimeout('/api/admin/logout', { method: 'POST' })
    authenticated.value = false
    adminState.value = emptyAdminState()
    buttonPage.value = emptyButtonPage()
    equipmentPage.value = emptyEquipmentPage()
    playerPage.value = emptyPlayerPage()
    bossHistoryPage.value = emptyBossHistoryPage()
    messagePage.value = emptyMessagePage()
    announcements.value = []
    checkingSession.value = false
    successMessage.value = ''
    bossForm.value = { id: '', name: '', maxHp: '' }
  }

  const shared = {
    activeTab,
    addHeroLootRow,
    addLootRow,
    adminState,
    announcementForm,
    announcements,
    applyHeroLootRows,
    applyLootRows,
    authenticated,
    bossCycleEnabled,
    bossForm,
    bossHistoryPage,
    bossTemplates,
    buttonForm,
    buttonPage,
    checkingSession,
    emptyAdminState,
    emptyAnnouncementForm,
    emptyBossHistoryPage,
    emptyButtonForm,
    emptyButtonPage,
    emptyEquipmentForm,
    emptyEquipmentPage,
    emptyHeroForm,
    emptyMessagePage,
    emptyPlayerPage,
    equipmentForm,
    equipmentPage,
    errorMessage,
    fetchAdminState,
    fetchAnnouncements,
    fetchButtonPage,
    fetchEquipmentPage,
    fetchMessages,
    findBossTemplate,
    heroForm,
    heroLootRows,
    loading,
    loadingAnnouncements,
    loadingButtons,
    loadingEquipment,
    loadingHistory,
    loadingMessages,
    loadingPlayers,
    loginForm,
    lootRows,
    messagePage,
    normalizeAdminState,
    normalizeAnnouncements,
    normalizeBossHistoryPage,
    normalizeButtonPage,
    normalizeEquipmentPage,
    normalizeMessagePage,
    normalizePlayerPage,
    playerPage,
    readErrorMessage,
    fetchWithTimeout,
    saving,
    selectedBossTemplateId,
    setSuccess,
    successMessage,
    syncBossTemplateEditor,
    uploadImageToOSS,
  }

  const actions = createAdminPageActions(shared)

  onMounted(checkSession)

  return {
    ...actions,
    activeTab,
    addHeroLootRow,
    addLootRow,
    adminState,
    announcementForm,
    announcements,
    authenticated,
    bossCycleEnabled,
    bossForm,
    bossHistoryPage,
    bossTemplates,
    buttonForm,
    buttonPage,
    checkingSession,
    editBossTemplate: actions.editBossTemplate,
    editButton: actions.editButton,
    editEquipment: actions.editEquipment,
    editHero: actions.editHero,
    equipmentForm,
    equipmentOptions,
    equipmentPage,
    errorMessage,
    findEquipmentTemplate,
    findHeroTemplate,
    formatHeroTrait,
    formatItemStats,
    formatTime,
    hasBoss,
    hasEquipmentTemplates,
    hasHeroTemplates,
    heroForm,
    heroImageAlt,
    heroLootRows,
    heroOptions,
    loading,
    loadingAnnouncements,
    loadingButtons,
    loadingEquipment,
    loadingHistory,
    loadingMessages,
    loadingPlayers,
    loginForm,
    lootRows,
    messagePage,
    playerPage,
    checkSession,
    fetchAdminState,
    fetchAnnouncements,
    fetchBossHistory,
    fetchButtonPage,
    fetchEquipmentPage,
    fetchMessages,
    fetchPlayerPage,
    login,
    logout,
    refreshAll,
    saving,
    selectedBossTemplate,
    selectedBossTemplateId,
    successMessage,
    uploadingImage,
  }
}
