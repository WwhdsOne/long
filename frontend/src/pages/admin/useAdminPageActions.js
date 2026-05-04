export function createAdminPageActions(state) {
  const {
    activeTab,
    addLootRow,
    adminState,
    adminRoomId,
    announcementForm,
    adminRoomSettings,
    applyLootRows,
    bossCycleEnabled,
    bossForm,
    bossTemplates,
    buttonForm,
    buttonPage,
    emptyAnnouncementForm,
    emptyButtonForm,
    emptyEquipmentForm,
    emptyShopItemForm,
    emptyTaskForm,
    equipmentForm,
    equipmentPage,
    equipmentPrompt,
    errorMessage,
    fetchAdminState,
    fetchAdminRooms,
    fetchAnnouncements,
    fetchBlacklist,
    fetchAdminRoomSettings,
    fetchButtonPage,
    fetchEquipmentPage,
    fetchMessages,
    fetchShopItems,
    fetchTaskArchives,
    fetchTaskCycleResults,
    fetchTasks,
    findBossTemplate,
    generatingEquipmentDraft,
    lootRows,
    readErrorMessage,
    saving,
    selectedBossTemplateId,
    selectedTaskCycleKey,
    selectedTaskId,
    setSuccess,
    showEquipmentEditor,
    shopItemForm,
    shopItems,
    uploadImageToOSS,
    taskDefinitions,
    taskForm,
  } = state

  function currentAdminRoomId() {
    return String(adminRoomId?.value || adminState.value?.roomId || '1').trim() || '1'
  }

  function withAdminRoom(url) {
    const joiner = url.includes('?') ? '&' : '?'
    return `${url}${joiner}roomId=${encodeURIComponent(currentAdminRoomId())}`
  }

  async function postAction(url, successTip, fallback) {
    saving.value = true
    try {
      const response = await fetch(url, { method: 'POST' })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, fallback))
      }
      setSuccess(successTip)
      await fetchAdminState()
    } catch (error) {
      errorMessage.value = error.message || fallback
    } finally {
      saving.value = false
    }
  }

  async function saveRoomDisplayName(roomId, displayName) {
    saving.value = true
    try {
      const response = await fetch(`/api/admin/rooms/${encodeURIComponent(roomId)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ displayName }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '保存房间名失败'))
      }
      const payload = await response.json()
      const nextList = Array.isArray(adminRoomSettings.value) ? [...adminRoomSettings.value] : []
      const nextRoom = payload && typeof payload === 'object' ? payload : { id: roomId, displayName }
      const matchIndex = nextList.findIndex((item) => String(item?.id || '') === String(roomId))
      if (matchIndex >= 0) {
        nextList.splice(matchIndex, 1, { ...nextList[matchIndex], ...nextRoom })
      } else {
        nextList.push(nextRoom)
      }
      adminRoomSettings.value = nextList
      setSuccess('房间名已保存。')
      await fetchAdminRooms()
      await fetchAdminRoomSettings()
    } catch (error) {
      errorMessage.value = error.message || '保存房间名失败'
    } finally {
      saving.value = false
    }
  }

  function normalizeTaskRewardItems(items) {
    return Array.isArray(items)
      ? items
          .filter((entry) => entry?.itemId)
          .map((entry) => ({
            itemId: String(entry.itemId).trim(),
            quantity: Number(entry.quantity || 1),
          }))
          .filter((entry) => entry.itemId && entry.quantity > 0)
      : []
  }

  function legacyTaskTypeForWindow(windowKind) {
    switch (windowKind) {
      case 'weekly':
        return 'weekly'
      case 'fixed_range':
        return 'limited'
      default:
        return 'daily'
    }
  }

  function legacyConditionKindForTask(eventKind, windowKind) {
    switch (eventKind) {
      case 'boss_kill':
        return 'boss_kills'
      case 'enhance':
        return 'enhance_count'
      default:
        return windowKind === 'weekly' ? 'weekly_clicks' : 'daily_clicks'
    }
  }

  function eventKindFromLegacyCondition(conditionKind) {
    switch (conditionKind) {
      case 'boss_kills':
        return 'boss_kill'
      case 'enhance_count':
        return 'enhance'
      default:
        return 'click'
    }
  }

  function windowKindFromLegacyTask(taskType, conditionKind) {
    if (conditionKind === 'weekly_clicks' && taskType !== 'limited') {
      return 'weekly'
    }
    if (conditionKind === 'daily_clicks' && taskType !== 'limited') {
      return 'daily'
    }
    switch (taskType) {
      case 'weekly':
        return 'weekly'
      case 'limited':
        return 'fixed_range'
      default:
        return 'daily'
    }
  }

  function normalizeTaskFormModel() {
    const legacyTaskType = String(taskForm.value.taskType || '').trim()
    const legacyConditionKind = String(taskForm.value.conditionKind || '').trim()
    const derivedEventKind = eventKindFromLegacyCondition(legacyConditionKind)
    const derivedWindowKind = windowKindFromLegacyTask(legacyTaskType, legacyConditionKind)
    const rawEventKind = String(taskForm.value.eventKind || '').trim()
    const rawWindowKind = String(taskForm.value.windowKind || '').trim()
    const eventKind = !rawEventKind || (rawEventKind === 'click' && derivedEventKind !== 'click')
      ? derivedEventKind
      : rawEventKind
    const windowKind = !rawWindowKind || (rawWindowKind === 'daily' && derivedWindowKind !== 'daily')
      ? derivedWindowKind
      : rawWindowKind
    return {
      eventKind,
      windowKind,
      taskType: legacyTaskType || legacyTaskTypeForWindow(windowKind),
      conditionKind: legacyConditionKind || legacyConditionKindForTask(eventKind, windowKind),
    }
  }

  function validateTaskDefinitionForm() {
    const model = normalizeTaskFormModel()
    const taskID = String(taskForm.value.taskId || '').trim()
    if (!taskID) {
      return '先填写 taskId。'
    }
    if (!String(taskForm.value.title || '').trim()) {
      return '先填写任务标题。'
    }
    if (Number(taskForm.value.targetValue || 0) <= 0) {
      return '目标值必须大于 0。'
    }
    if (model.windowKind === 'fixed_range') {
      const startAt = Number(taskForm.value.startAt || 0)
      const endAt = Number(taskForm.value.endAt || 0)
      if (startAt <= 0 || endAt <= 0 || endAt <= startAt) {
        return '固定时间窗任务需要填写合法的开始时间和结束时间。'
      }
    }
    const rewardItems = normalizeTaskRewardItems(taskForm.value.rewards?.equipmentItems)
    const hasRewards = Number(taskForm.value.rewards?.gold || 0) > 0 ||
      Number(taskForm.value.rewards?.stones || 0) > 0 ||
      Number(taskForm.value.rewards?.talentPoints || 0) > 0 ||
      rewardItems.length > 0
    if (!hasRewards) {
      return '任务奖励至少填写一项。'
    }
    return ''
  }

  async function deleteByID(url, successTip, fallback, afterDelete) {
    saving.value = true
    try {
      const response = await fetch(url, { method: 'DELETE' })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, fallback))
      }
      setSuccess(successTip)
      await afterDelete()
    } catch (error) {
      errorMessage.value = error.message || fallback
    } finally {
      saving.value = false
    }
  }

  function normalizeBossIntegerString(value, min = 0n) {
    const raw = String(value ?? '').trim()
    if (!/^\d+$/.test(raw)) {
      return String(min)
    }
    const normalized = raw.replace(/^0+(?=\d)/, '') || '0'
    const parsed = BigInt(normalized)
    return parsed < min ? String(min) : normalized
  }

  async function saveBossTemplate() {
    saving.value = true
    try {
      const method = bossTemplates.value.some((entry) => entry.id === bossForm.value.id) ? 'PUT' : 'POST'
      const targetID = encodeURIComponent(bossForm.value.id)
      const url = method === 'PUT' ? `/api/admin/boss/pool/${targetID}` : '/api/admin/boss/pool'
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: bossForm.value.id,
          name: bossForm.value.name,
          maxHp: sumBossPartMaxHp(bossForm.value.layout),
          goldOnKill: Number(bossForm.value.goldOnKill || 0),
          stoneOnKill: Number(bossForm.value.stoneOnKill || 0),
          talentPointsOnKill: Number(bossForm.value.talentPointsOnKill || 0),
          layout: bossForm.value.layout || [],
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
    await postAction(withAdminRoom('/api/admin/boss/deactivate'), bossCycleEnabled.value ? '当前 Boss 已跳过，循环会继续补位。' : '当前 Boss 已关闭。', '关闭 Boss 失败')
  }

  async function enableBossCycle() {
    await postAction(withAdminRoom('/api/admin/boss/cycle/enable'), 'Boss 循环已开启。', '开启 Boss 循环失败')
  }

  async function disableBossCycle() {
    await postAction(withAdminRoom('/api/admin/boss/cycle/disable'), 'Boss 循环已停止，当前 Boss 不会自动续上。', '停止 Boss 循环失败')
  }

  async function saveBossCycleQueue(templateIds) {
    saving.value = true
    try {
      const response = await fetch(withAdminRoom('/api/admin/boss/cycle/queue'), {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ roomId: currentAdminRoomId(), templateIds }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '保存 Boss 循环队列失败'))
      }
      setSuccess('Boss 循环队列已保存。')
      await fetchAdminState()
    } catch (error) {
      errorMessage.value = error.message || '保存 Boss 循环队列失败'
    } finally {
      saving.value = false
    }
  }

  async function saveEquipment() {
    saving.value = true
    try {
      const method = equipmentPage.value.items.some((entry) => entry.itemId === equipmentForm.value.itemId) ? 'PUT' : 'POST'
      const url = method === 'PUT' ? `/api/admin/equipment/${encodeURIComponent(equipmentForm.value.itemId)}` : '/api/admin/equipment'
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...equipmentForm.value,
          rarity: equipmentForm.value.rarity,
          attackPower: Number(equipmentForm.value.attackPower),
          armorPenPercent: Number(equipmentForm.value.armorPenPercent),
          critRate: Number(equipmentForm.value.critRate),
          critDamageMultiplier: Number(equipmentForm.value.critDamageMultiplier),
          bossDamagePercent: Number(equipmentForm.value.bossDamagePercent),
          partTypeDamageSoft: Number(equipmentForm.value.partTypeDamageSoft),
          partTypeDamageHeavy: Number(equipmentForm.value.partTypeDamageHeavy),
          partTypeDamageWeak: Number(equipmentForm.value.partTypeDamageWeak),
          talentAffinity: equipmentForm.value.talentAffinity,
        }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '保存装备失败'))
      }
      setSuccess('装备模板已保存。')
      equipmentForm.value = emptyEquipmentForm()
      showEquipmentEditor.value = false
      await fetchEquipmentPage(equipmentPage.value.page)
    } catch (error) {
      errorMessage.value = error.message || '保存装备失败'
    } finally {
      saving.value = false
    }
  }

  function openNewEquipment() {
    equipmentForm.value = emptyEquipmentForm()
    showEquipmentEditor.value = true
    activeTab.value = 'equipment'
  }

  function updateEquipmentPrompt(value) {
    equipmentPrompt.value = value
  }

  async function generateEquipmentDraft() {
    const prompt = equipmentPrompt.value.trim()
    if (!prompt) {
      errorMessage.value = '先输入装备描述。'
      return
    }

    generatingEquipmentDraft.value = true
    try {
      const response = await fetch('/api/admin/equipment/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '生成装备草稿失败'))
      }

      const payload = await response.json()
      equipmentForm.value = {
        ...emptyEquipmentForm(),
        ...(payload.draft || {}),
      }
      showEquipmentEditor.value = true
      setSuccess('装备草稿已生成，请检查后保存。')
    } catch (error) {
      errorMessage.value = error.message || '生成装备草稿失败'
    } finally {
      generatingEquipmentDraft.value = false
    }
  }



  async function deleteEquipment(itemId) {
    await deleteByID(
      `/api/admin/equipment/${encodeURIComponent(itemId)}`,
      '装备模板已删除。',
      '删除装备失败',
      () => fetchEquipmentPage(equipmentPage.value.page),
    )
  }

  async function saveButton() {
    saving.value = true
    try {
      const method = buttonPage.value.items.some((entry) => entry.key === buttonForm.slug) ? 'PUT' : 'POST'
      const url = method === 'PUT' ? `/api/admin/buttons/${encodeURIComponent(buttonForm.slug)}` : '/api/admin/buttons'
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...buttonForm,
          sort: Number(buttonForm.sort),
          enabled: Boolean(buttonForm.enabled),
          tags: buttonForm.tagsText.split(/[,，]/).map((tag) => tag.trim()).filter(Boolean),
        }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '保存按钮失败'))
      }
      setSuccess('按钮配置已保存。')
      Object.assign(buttonForm, emptyButtonForm())
      await fetchButtonPage(buttonPage.value.page)
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
        headers: { 'Content-Type': 'application/json' },
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
    await deleteByID(`/api/admin/announcements/${encodeURIComponent(id)}`, '公告已删除。', '删除公告失败', fetchAnnouncements)
  }

  async function deleteMessage(id) {
    await deleteByID(`/api/admin/messages/${encodeURIComponent(id)}`, '留言已删除。', '删除留言失败', fetchMessages)
  }

  async function unblockBlacklistEntry(clientId, nickname) {
    saving.value = true
    try {
      const response = await fetch(`/api/admin/blacklist/${encodeURIComponent(clientId)}/unblock`, {
        method: 'POST',
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '手动解封失败'))
      }
      setSuccess(`已解封 ${nickname || clientId}。`)
      await fetchBlacklist()
    } catch (error) {
      errorMessage.value = error.message || '手动解封失败'
    } finally {
      saving.value = false
    }
  }

  async function uploadImageInner(event, applyImage, successTip, category = '') {
    const file = event.target?.files?.[0]
    if (!file) {
      return
    }

    await uploadImageToOSS(event, file, applyImage, successTip, category)
  }

  async function uploadButtonImage(event) {
    await uploadImageInner(event, (finalURL, file) => {
      Object.assign(buttonForm, {
        imagePath: finalURL,
        imageAlt: buttonForm.imageAlt || file.name.replace(/\.[^.]+$/, ''),
      })
    }, '按钮图片已上传到 OSS。')
  }

  async function uploadEquipmentImage(event) {
    await uploadImageInner(event, (finalURL, file) => {
      equipmentForm.value.imagePath = finalURL
      if (!equipmentForm.value.imageAlt) {
        equipmentForm.value.imageAlt = file.name.replace(/\.[^.]+$/, '')
      }
    }, '装备图片已上传到 OSS。')
  }

  async function uploadShopImage(event) {
    await uploadImageInner(event, (finalURL, file) => {
      shopItemForm.value.imagePath = finalURL
      if (!shopItemForm.value.imageAlt) {
        shopItemForm.value.imageAlt = file.name.replace(/\.[^.]+$/, '')
      }
    }, '商店主图已上传到 OSS。', 'shop')
  }

  async function uploadShopPreviewImage(event) {
    await uploadImageInner(event, (finalURL) => {
      shopItemForm.value.previewImagePath = finalURL
    }, '商店预览图已上传到 OSS。', 'shop')
  }

  async function uploadShopCursorImage(event) {
    await uploadImageInner(event, (finalURL) => {
      shopItemForm.value.battleClickCursorImagePath = finalURL
    }, '战斗点击图标已上传到 OSS。', 'shop')
  }

  async function saveLoot(lootRowsOverride = null) {
    if (!selectedBossTemplateId.value) {
      errorMessage.value = '先选一只 Boss 模板，再配置掉落池。'
      return
    }

    const rowsToSave = Array.isArray(lootRowsOverride) ? lootRowsOverride : lootRows.value

    saving.value = true
    try {
      const response = await fetch(`/api/admin/boss/pool/${encodeURIComponent(selectedBossTemplateId.value)}/loot`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          loot: rowsToSave.filter((entry) => entry.itemId).map((entry) => ({
            itemId: entry.itemId,
            dropRatePercent: Number(entry.dropRatePercent),
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

  async function saveTaskDefinition() {
    const validationMessage = validateTaskDefinitionForm()
    if (validationMessage) {
      errorMessage.value = validationMessage
      return
    }
    saving.value = true
    try {
      const model = normalizeTaskFormModel()
      const taskID = String(taskForm.value.taskId || '').trim()
      const rewardItems = normalizeTaskRewardItems(taskForm.value.rewards?.equipmentItems)
      const exists = taskDefinitions.value.some((entry) => entry.taskId === taskID)
      const method = exists ? 'PUT' : 'POST'
      const url = exists ? `/api/admin/tasks/${encodeURIComponent(taskID)}` : '/api/admin/tasks'
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...taskForm.value,
          taskId: taskID,
          eventKind: model.eventKind,
          windowKind: model.windowKind,
          taskType: legacyTaskTypeForWindow(model.windowKind),
          conditionKind: legacyConditionKindForTask(model.eventKind, model.windowKind),
          targetValue: Number(taskForm.value.targetValue || 0),
          displayOrder: Number(taskForm.value.displayOrder || 0),
          startAt: model.windowKind === 'fixed_range' ? Number(taskForm.value.startAt || 0) : 0,
          endAt: model.windowKind === 'fixed_range' ? Number(taskForm.value.endAt || 0) : 0,
          rewards: {
            gold: Number(taskForm.value.rewards?.gold || 0),
            stones: Number(taskForm.value.rewards?.stones || 0),
            talentPoints: Number(taskForm.value.rewards?.talentPoints || 0),
            equipmentItems: rewardItems,
          },
        }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '保存任务失败'))
      }
      setSuccess('任务已保存。')
      taskForm.value = emptyTaskForm()
      await fetchTasks()
    } catch (error) {
      errorMessage.value = error.message || '保存任务失败'
    } finally {
      saving.value = false
    }
  }

  async function saveShopItem() {
    const itemId = String(shopItemForm.value.itemId || '').trim()
    const title = String(shopItemForm.value.title || '').trim()
    if (!itemId) {
      errorMessage.value = '先填写商品 ID。'
      return
    }
    if (!title) {
      errorMessage.value = '先填写商品标题。'
      return
    }

    saving.value = true
    try {
      const exists = shopItems.value.some((entry) => entry.itemId === itemId)
      const method = exists ? 'PUT' : 'POST'
      const url = exists ? `/api/admin/shop/items/${encodeURIComponent(itemId)}` : '/api/admin/shop/items'
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...shopItemForm.value,
          itemId,
          title,
          priceGold: Number(shopItemForm.value.priceGold || 0),
          sortOrder: Number(shopItemForm.value.sortOrder || 0),
          active: Boolean(shopItemForm.value.active),
          autoEquipOnPurchase: shopItemForm.value.autoEquipOnPurchase !== false,
        }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '保存商店商品失败'))
      }
      setSuccess('商店商品已保存。')
      shopItemForm.value = emptyShopItemForm()
      await fetchShopItems()
    } catch (error) {
      errorMessage.value = error.message || '保存商店商品失败'
    } finally {
      saving.value = false
    }
  }

  async function deleteShopItem(itemId) {
    await deleteByID(
      `/api/admin/shop/items/${encodeURIComponent(itemId)}`,
      '商店商品已删除。',
      '删除商店商品失败',
      fetchShopItems,
    )
  }

  function editShopItem(entry) {
    shopItemForm.value = { ...entry }
    activeTab.value = 'shop'
  }

  function openNewShopItem() {
    shopItemForm.value = emptyShopItemForm()
    activeTab.value = 'shop'
  }

  async function activateTaskDefinition(taskId) {
    await postAction(`/api/admin/tasks/${encodeURIComponent(taskId)}/activate`, '任务已上线。', '任务上线失败')
    await fetchTasks()
  }

  async function deactivateTaskDefinition(taskId) {
    await postAction(`/api/admin/tasks/${encodeURIComponent(taskId)}/deactivate`, '任务已下线。', '任务下线失败')
    await fetchTasks()
  }

  async function duplicateTaskDefinition(taskId) {
    const nextTaskId = window.prompt(`给复制出的任务输入新的 taskId`, `${taskId}-copy`)
    if (!nextTaskId) {
      return
    }
    saving.value = true
    try {
      const response = await fetch(`/api/admin/tasks/${encodeURIComponent(taskId)}/duplicate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ taskId: nextTaskId }),
      })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '复制任务失败'))
      }
      setSuccess('任务已复制。')
      await fetchTasks()
    } catch (error) {
      errorMessage.value = error.message || '复制任务失败'
    } finally {
      saving.value = false
    }
  }

  async function archiveExpiredTasks() {
    saving.value = true
    try {
      const response = await fetch('/api/admin/tasks/archive-expired', { method: 'POST' })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '归档过期任务失败'))
      }
      setSuccess('过期任务归档已执行。')
      if (selectedTaskId.value) {
        await fetchTaskArchives(selectedTaskId.value)
        if (selectedTaskCycleKey.value) {
          await fetchTaskCycleResults(selectedTaskId.value, selectedTaskCycleKey.value)
        }
      }
    } catch (error) {
      errorMessage.value = error.message || '归档过期任务失败'
    } finally {
      saving.value = false
    }
  }

  function editTaskDefinition(entry) {
    taskForm.value = {
      ...emptyTaskForm(),
      ...entry,
      rewards: {
        gold: Number(entry?.rewards?.gold || 0),
        stones: Number(entry?.rewards?.stones || 0),
        talentPoints: Number(entry?.rewards?.talentPoints || 0),
        equipmentItems: Array.isArray(entry?.rewards?.equipmentItems)
          ? entry.rewards.equipmentItems.map((item) => ({
              itemId: item.itemId || '',
              quantity: Number(item.quantity || 1),
            }))
          : [],
      },
    }
    activeTab.value = 'tasks'
  }

  function openNewTask() {
    taskForm.value = emptyTaskForm()
    activeTab.value = 'tasks'
  }

  function addTaskEquipmentReward() {
    const items = Array.isArray(taskForm.value.rewards?.equipmentItems) ? taskForm.value.rewards.equipmentItems : []
    taskForm.value.rewards = {
      ...taskForm.value.rewards,
      equipmentItems: [...items, { itemId: '', quantity: 1 }],
    }
  }

  function removeTaskEquipmentReward(index) {
    const items = Array.isArray(taskForm.value.rewards?.equipmentItems) ? [...taskForm.value.rewards.equipmentItems] : []
    items.splice(index, 1)
    taskForm.value.rewards = {
      ...taskForm.value.rewards,
      equipmentItems: items,
    }
  }



  async function deleteBossTemplate(templateId) {
    saving.value = true
    try {
      const response = await fetch(`/api/admin/boss/pool/${encodeURIComponent(templateId)}`, { method: 'DELETE' })
      if (!response.ok) {
        throw new Error(await readErrorMessage(response, '删除 Boss 模板失败'))
      }
      if (selectedBossTemplateId.value === templateId) {
        selectedBossTemplateId.value = ''
      }
      if (bossForm.value.id === templateId) {
        bossForm.value = { id: '', name: '', maxHp: '', goldOnKill: 0, stoneOnKill: 0, talentPointsOnKill: 0, layout: [] }
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
    showEquipmentEditor.value = true
    activeTab.value = 'equipment'
  }

  function editButton(entry) {
    Object.assign(buttonForm, {
      slug: entry.key,
      label: entry.label,
      sort: entry.sort,
      enabled: entry.enabled,
      tagsText: Array.isArray(entry.tags) ? entry.tags.join(', ') : '',
      imagePath: entry.imagePath || '',
      imageAlt: entry.imageAlt || '',
    })
    activeTab.value = 'buttons'
  }



  function editBossTemplate(entry) {
    bossForm.value = {
      id: entry.id,
      name: entry.name,
      maxHp: String(entry.maxHp ?? ''),
      goldOnKill: Number(entry.goldOnKill || 0),
      stoneOnKill: Number(entry.stoneOnKill || 0),
      talentPointsOnKill: Number(entry.talentPointsOnKill || 0),
      layout: Array.isArray(entry.layout) ? entry.layout.map((part) => ({
        ...part,
        maxHp: String(part?.maxHp ?? ''),
        currentHp: String(part?.currentHp ?? part?.maxHp ?? ''),
      })) : [],
    }
    selectedBossTemplateId.value = entry.id
    applyLootRows(entry.loot)
    activeTab.value = 'boss'
  }

  function selectBossTemplate(templateId) {
    selectedBossTemplateId.value = templateId
    applyLootRows(findBossTemplate(templateId)?.loot ?? [])
  }

  function sumBossPartMaxHp(layout) {
    if (!Array.isArray(layout) || layout.length === 0) {
      return '1'
    }
    return layout.reduce((total, part) => {
      const maxHp = BigInt(normalizeBossIntegerString(part?.maxHp, 1n))
      return total + maxHp
    }, 0n).toString()
  }

  function removeLootRow(index) {
    lootRows.value.splice(index, 1)
    if (lootRows.value.length === 0) {
      addLootRow()
    }
  }



  return {
    deactivateBoss,
    deleteAnnouncement,
    deleteBossTemplate,
    deleteEquipment,
    deleteMessage,
    unblockBlacklistEntry,
    disableBossCycle,
    editBossTemplate,
    editTaskDefinition,
    editButton,
    editEquipment,
    editShopItem,
    enableBossCycle,
    archiveExpiredTasks,
    activateTaskDefinition,
    addTaskEquipmentReward,
    deactivateTaskDefinition,
    duplicateTaskDefinition,
    openNewTask,
    openNewShopItem,
    generateEquipmentDraft,
    openNewEquipment,
    removeLootRow,
    removeTaskEquipmentReward,
    saveAnnouncement,
    saveBossTemplate,
    saveBossCycleQueue,
    saveButton,
    saveEquipment,
    saveLoot,
    saveRoomDisplayName,
    saveShopItem,
    saveTaskDefinition,
    selectBossTemplate,
    updateEquipmentPrompt,
    uploadButtonImage,
    uploadEquipmentImage,
    uploadShopCursorImage,
    uploadShopImage,
    uploadShopPreviewImage,
    deleteShopItem,
  }
}
