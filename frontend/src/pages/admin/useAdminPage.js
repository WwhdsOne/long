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
  emptyShopItemForm,
  emptyTaskCycleResults,
  emptyTaskForm,
  formatItemStats,
  formatTime,
  normalizeAdminState,
  normalizeAnnouncements,
  normalizeBossHistoryPage,
  normalizeButtonPage,
  normalizeEquipmentPage,
  normalizeMessagePage,
  normalizePlayerPage,
  normalizeShopItem,
  normalizeTaskArchive,
  normalizeTaskCycleResults,
  normalizeTaskDefinition,
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
  const bossForm = ref({ id: '', name: '', maxHp: '', goldOnKill: 0, stoneOnKill: 0, talentPointsOnKill: 0, layout: [] })
  const equipmentForm = ref(emptyEquipmentForm())
  const equipmentPrompt = ref('')
  const showEquipmentEditor = ref(false)
  const buttonForm = reactive(emptyButtonForm())
  const announcementForm = ref(emptyAnnouncementForm())
  const lootRows = ref(emptyLootRows())
  const selectedBossTemplateId = ref('')
  const adminRoomId = ref('1')
  const adminRooms = ref([])

  const adminState = ref(emptyAdminState())
  const buttonPage = ref(emptyButtonPage())
  const equipmentPage = ref(emptyEquipmentPage())
  const playerPage = ref(emptyPlayerPage())
  const bossHistoryPage = ref(emptyBossHistoryPage())
  const announcements = ref([])
  const messagePage = ref(emptyMessagePage())
  const taskDefinitions = ref([])
  const shopItems = ref([])
  const taskForm = ref(emptyTaskForm())
  const shopItemForm = ref(emptyShopItemForm())
  const taskArchives = ref([])
  const taskCycleResults = ref(emptyTaskCycleResults())
  const selectedTaskId = ref('')
  const selectedTaskCycleKey = ref('')

  const loadingHistory = ref(false)
  const loadingButtons = ref(false)
  const loadingEquipment = ref(false)
  const generatingEquipmentDraft = ref(false)
  const loadingAnnouncements = ref(false)
  const loadingMessages = ref(false)
  const loadingPlayers = ref(false)
  const loadingTasks = ref(false)
  const loadingShopItems = ref(false)
  const loadingTaskArchives = ref(false)
  const loadingTaskResults = ref(false)

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

  async function fetchAdminRooms() {
    try {
      const response = await fetchWithTimeout('/api/rooms')
      if (!response.ok) {
        return
      }
      const payload = await response.json()
      adminRooms.value = Array.isArray(payload?.rooms) ? payload.rooms : []
      if (!adminRoomId.value && payload?.currentRoomId) {
        adminRoomId.value = String(payload.currentRoomId)
      }
    } catch {
      adminRooms.value = []
    }
  }

  async function switchAdminRoom(roomId) {
    const nextRoomId = String(roomId || '').trim()
    if (!nextRoomId || nextRoomId === adminRoomId.value) {
      return
    }
    adminRoomId.value = nextRoomId
    await fetchAdminState()
  }



  function applyLootRows(loot) {
    lootRows.value = Array.isArray(loot) && loot.length > 0
      ? loot.map((entry) => ({ itemId: entry.itemId, dropRatePercent: entry.dropRatePercent }))
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

  async function uploadImageToOSS(event, file, applyImage, successTip, category = '') {
    uploadingImage.value = true
    try {
      const stsURL = category ? `/api/admin/oss/sts?category=${encodeURIComponent(category)}` : '/api/admin/oss/sts'
      const policyResponse = await fetchWithTimeout(stsURL, { method: 'POST' })
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
    lootRows.value.push({ itemId: '', dropRatePercent: '' })
  }



  async function fetchAdminState() {
    loading.value = true
    try {
      const query = adminRoomId.value ? `?roomId=${encodeURIComponent(adminRoomId.value)}` : ''
      const response = await fetchWithTimeout(`/api/admin/state${query}`)
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '后台状态加载失败'))
      }

      adminState.value = normalizeAdminState(await response.json())
      adminRoomId.value = adminState.value.roomId || adminRoomId.value || '1'
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

  async function fetchTasks() {
    loadingTasks.value = true
    try {
      const response = await fetchWithTimeout('/api/admin/tasks')
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '任务列表加载失败'))
      }
      const payload = await response.json()
      taskDefinitions.value = Array.isArray(payload) ? payload.map(normalizeTaskDefinition) : []
    } catch (error) {
      errorMessage.value = error.message || '任务列表加载失败'
    } finally {
      loadingTasks.value = false
    }
  }

  async function fetchShopItems() {
    loadingShopItems.value = true
    try {
      const response = await fetchWithTimeout('/api/admin/shop/items')
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '商店商品列表加载失败'))
      }
      const payload = await response.json()
      shopItems.value = Array.isArray(payload) ? payload.map(normalizeShopItem) : []
    } catch (error) {
      errorMessage.value = error.message || '商店商品列表加载失败'
    } finally {
      loadingShopItems.value = false
    }
  }

  async function fetchTaskArchives(taskId = selectedTaskId.value) {
    if (!taskId) {
      taskArchives.value = []
      return
    }
    loadingTaskArchives.value = true
    selectedTaskId.value = taskId
    selectedTaskCycleKey.value = ''
    taskCycleResults.value = emptyTaskCycleResults()
    try {
      const response = await fetchWithTimeout(`/api/admin/tasks/${encodeURIComponent(taskId)}/cycles`)
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '任务周期归档加载失败'))
      }
      const payload = await response.json()
      taskArchives.value = Array.isArray(payload) ? payload.map(normalizeTaskArchive) : []
    } catch (error) {
      errorMessage.value = error.message || '任务周期归档加载失败'
    } finally {
      loadingTaskArchives.value = false
    }
  }

  async function fetchTaskCycleResults(taskId = selectedTaskId.value, cycleKey = selectedTaskCycleKey.value) {
    if (!taskId || !cycleKey) {
      taskCycleResults.value = emptyTaskCycleResults()
      return
    }
    loadingTaskResults.value = true
    selectedTaskId.value = taskId
    selectedTaskCycleKey.value = cycleKey
    try {
      const response = await fetchWithTimeout(
        `/api/admin/tasks/${encodeURIComponent(taskId)}/cycles/${encodeURIComponent(cycleKey)}/results`,
      )
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '任务周期明细加载失败'))
      }
      taskCycleResults.value = normalizeTaskCycleResults(await response.json())
    } catch (error) {
      errorMessage.value = error.message || '任务周期明细加载失败'
    } finally {
      loadingTaskResults.value = false
    }
  }

  async function refreshAll() {
    await Promise.all([
      fetchAdminRooms(),
      fetchAdminState(),
      fetchPlayerPage(),
      fetchEquipmentPage(equipmentPage.value.page),
      fetchShopItems(),
      fetchTasks(),
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

      await fetchAdminRooms()
      await fetchAdminState()
      await Promise.all([
        fetchAnnouncements(),
        fetchMessages(),
        fetchPlayerPage(),
        fetchEquipmentPage(),
        fetchShopItems(),
        fetchTasks(),
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
      await fetchAdminRooms()
      await fetchAdminState()
      await Promise.all([fetchPlayerPage(), fetchEquipmentPage(), fetchShopItems(), fetchTasks()])
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
    adminRoomId.value = '1'
    adminRooms.value = []
    buttonPage.value = emptyButtonPage()
    equipmentPage.value = emptyEquipmentPage()
    playerPage.value = emptyPlayerPage()
    bossHistoryPage.value = emptyBossHistoryPage()
    messagePage.value = emptyMessagePage()
    taskDefinitions.value = []
    shopItems.value = []
    taskForm.value = emptyTaskForm()
    shopItemForm.value = emptyShopItemForm()
    taskArchives.value = []
    taskCycleResults.value = emptyTaskCycleResults()
    selectedTaskId.value = ''
    selectedTaskCycleKey.value = ''
    announcements.value = []
    checkingSession.value = false
    successMessage.value = ''
    bossForm.value = { id: '', name: '', maxHp: '', goldOnKill: 0, stoneOnKill: 0, talentPointsOnKill: 0, layout: [] }
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
    adminRoomId,
    adminRooms,
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
    emptyShopItemForm,
    emptyTaskForm,
    equipmentForm,
    equipmentPage,
    equipmentPrompt,
    errorMessage,
    fetchAdminState,
    fetchAdminRooms,
    fetchAnnouncements,
    fetchButtonPage,
    fetchEquipmentPage,
    fetchMessages,
    fetchShopItems,
    fetchTaskArchives,
    fetchTaskCycleResults,
    fetchTasks,
    findBossTemplate,
    generatingEquipmentDraft,
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
    shopItems,
    shopItemForm,
    taskArchives,
    taskCycleResults,
    taskDefinitions,
    taskForm,
    readErrorMessage,
    fetchWithTimeout,
    saving,
    selectedBossTemplateId,
    selectedTaskCycleKey,
    selectedTaskId,
    setSuccess,
    switchAdminRoom,
    showEquipmentEditor,
    successMessage,
    syncBossTemplateEditor,
    loadingTaskArchives,
    loadingTaskResults,
    loadingShopItems,
    loadingTasks,
    uploadImageToOSS,
  }

  const actions = createAdminPageActions(shared)

  onMounted(checkSession)

  return {
    ...actions,
    activeTab,
    addLootRow,
    adminState,
    adminRoomId,
    adminRooms,
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
    editShopItem: actions.editShopItem,
    equipmentPrompt,
    equipmentForm,
    equipmentOptions,
    equipmentPage,
    errorMessage,
    findEquipmentTemplate,
    formatItemStats,
    formatTime,
    hasBoss,
    hasEquipmentTemplates,
    generatingEquipmentDraft,
    loading,
    loadingAnnouncements,
    loadingButtons,
    loadingEquipment,
    loadingHistory,
    loadingMessages,
    loadingPlayers,
    loadingShopItems,
    loadingTaskArchives,
    loadingTaskResults,
    loadingTasks,
    showEquipmentEditor,
    loginForm,
    lootRows,
    messagePage,
    playerPage,
    shopItems,
    shopItemForm,
    selectedTaskCycleKey,
    selectedTaskId,
    checkSession,
    fetchAdminState,
    fetchAdminRooms,
    fetchAnnouncements,
    fetchBossHistory,
    fetchButtonPage,
    fetchEquipmentPage,
    fetchMessages,
    fetchShopItems,
    fetchTaskArchives,
    fetchTaskCycleResults,
    fetchTasks,
    fetchPlayerPage,
    login,
    logout,
    openNewShopItem: actions.openNewShopItem,
    refreshAll,
    resetPlayerPassword,
    saving,
    saveShopItem: actions.saveShopItem,
    selectedBossTemplate,
    selectedBossTemplateId,
    successMessage,
    switchAdminRoom,
    taskArchives,
    taskCycleResults,
    taskDefinitions,
    taskForm,
    deleteShopItem: actions.deleteShopItem,
    uploadShopCursorImage: actions.uploadShopCursorImage,
    uploadShopImage: actions.uploadShopImage,
    uploadShopPreviewImage: actions.uploadShopPreviewImage,
    uploadingImage,
  }
}
