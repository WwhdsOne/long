export function createAdminPageActions(state) {
  const {
    activeTab,
    addLootRow,
    adminState,
    announcementForm,
    applyLootRows,
    bossCycleEnabled,
    bossForm,
    bossTemplates,
    buttonForm,
    buttonPage,
    emptyAnnouncementForm,
    emptyButtonForm,
    emptyEquipmentForm,
    equipmentForm,
    equipmentPage,
    equipmentPrompt,
    errorMessage,
    fetchAdminState,
    fetchAnnouncements,
    fetchButtonPage,
    fetchEquipmentPage,
    fetchMessages,
    findBossTemplate,
    generatingEquipmentDraft,
    lootRows,
    readErrorMessage,
    saving,
    selectedBossTemplateId,
    setSuccess,
    showEquipmentEditor,
    uploadImageToOSS,
  } = state

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
    await postAction('/api/admin/boss/deactivate', bossCycleEnabled.value ? '当前 Boss 已跳过，循环会继续补位。' : '当前 Boss 已关闭。', '关闭 Boss 失败')
  }

  async function enableBossCycle() {
    await postAction('/api/admin/boss/cycle/enable', 'Boss 循环已开启。', '开启 Boss 循环失败')
  }

  async function disableBossCycle() {
    await postAction('/api/admin/boss/cycle/disable', 'Boss 循环已停止，当前 Boss 不会自动续上。', '停止 Boss 循环失败')
  }

  async function saveBossCycleQueue(templateIds) {
    saving.value = true
    try {
      const response = await fetch('/api/admin/boss/cycle/queue', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ templateIds }),
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

  async function uploadImageInner(event, applyImage, successTip) {
    const file = event.target?.files?.[0]
    if (!file) {
      return
    }

    await uploadImageToOSS(event, file, applyImage, successTip)
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
        bossForm.value = { id: '', name: '', maxHp: '', goldOnKill: 0, stoneOnKill: 0, layout: [] }
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
      maxHp: entry.maxHp,
      goldOnKill: Number(entry.goldOnKill || 0),
      stoneOnKill: Number(entry.stoneOnKill || 0),
      layout: entry.layout || [],
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
      return 1
    }
    return layout.reduce((total, part) => total + Math.max(1, Number(part?.maxHp ?? 0)), 0)
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
    disableBossCycle,
    editBossTemplate,
    editButton,
    editEquipment,
    enableBossCycle,
    generateEquipmentDraft,
    openNewEquipment,
    removeLootRow,
    saveAnnouncement,
    saveBossTemplate,
    saveBossCycleQueue,
    saveButton,
    saveEquipment,
    saveLoot,
    selectBossTemplate,
    updateEquipmentPrompt,
    uploadButtonImage,
    uploadEquipmentImage,
  }
}
