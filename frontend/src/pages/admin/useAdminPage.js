import { computed, nextTick, onMounted, reactive, ref } from 'vue'

import {
  emptyAdminState,
  emptyAnnouncementForm,
  emptyBossHistoryPage,
  emptyButtonForm,
  emptyButtonPage,
  emptyEquipmentForm,
  emptyEquipmentPage,
  emptyLootRows,
  emptyMessagePage,
  emptyPlayerPage,
  formatItemStats,
  formatTime,
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
  const bossForm = ref({ id: '', name: '', maxHp: '', layout: [] })
  const equipmentForm = ref(emptyEquipmentForm())
  const buttonForm = reactive(emptyButtonForm())
  const announcementForm = ref(emptyAnnouncementForm())
  const lootRows = ref(emptyLootRows())
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
  const hasEquipmentTemplates = computed(() => equipmentPage.value.total > 0)
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



  function applyLootRows(loot) {
    lootRows.value = Array.isArray(loot) && loot.length > 0
      ? loot.map((entry) => ({ itemId: entry.itemId, weight: entry.weight }))
      : emptyLootRows()
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
    bossForm.value = { id: '', name: '', maxHp: '', layout: [] }
  }

  async function resetPlayerPassword(nickname) {
    const nextPassword = window.prompt(`给 ${nickname} 设置一个新密码`, '')
    if (!nextPassword) {
      return
    }

    saving.value = true
    try {
      const response = await fetchWithTimeout(`/api/admin/players/${encodeURIComponent(nickname)}/password/reset`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password: nextPassword }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '玩家密码重置失败'))
      }

      setSuccess(`已重置 ${nickname} 的密码。`)
    } catch (error) {
      errorMessage.value = error.message || '玩家密码重置失败'
    } finally {
      saving.value = false
    }
  }

  const shared = {
    activeTab,
    addLootRow,
    adminState,
    announcementForm,
    announcements,
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
    equipmentForm,
    equipmentOptions,
    equipmentPage,
    errorMessage,
    findEquipmentTemplate,
    formatItemStats,
    formatTime,
    hasBoss,
    hasEquipmentTemplates,
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
    resetPlayerPassword,
    saving,
    selectedBossTemplate,
    selectedBossTemplateId,
    successMessage,
    uploadingImage,
  }
}
