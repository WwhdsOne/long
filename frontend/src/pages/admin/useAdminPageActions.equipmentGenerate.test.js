import { ref } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import { createAdminPageActions } from './useAdminPageActions'
import { emptyEquipmentForm } from './state'

function makeState() {
  const equipmentForm = ref(emptyEquipmentForm())
  const equipmentPage = ref({ items: [], page: 1 })
  const errorMessage = ref('')
  const successMessage = ref('')

  return {
    activeTab: ref('equipment'),
    addLootRow: vi.fn(),
    adminState: ref({}),
    announcementForm: ref({}),
    applyLootRows: vi.fn(),
    bossCycleEnabled: ref(false),
    bossForm: ref({}),
    bossTemplates: ref([]),
    buttonForm: {},
    buttonPage: ref({ items: [], page: 1 }),
    emptyAnnouncementForm: vi.fn(),
    emptyButtonForm: vi.fn(),
    emptyEquipmentForm,
    equipmentForm,
    equipmentPage,
    equipmentPrompt: ref('做一把偏普攻流的软组织武器'),
    errorMessage,
    fetchAdminState: vi.fn(),
    fetchAnnouncements: vi.fn(),
    fetchButtonPage: vi.fn(),
    fetchEquipmentPage: vi.fn(),
    fetchMessages: vi.fn(),
    findBossTemplate: vi.fn(),
    generatingEquipmentDraft: ref(false),
    lootRows: ref([]),
    readErrorMessage: vi.fn(async () => '生成失败'),
    saving: ref(false),
    selectedBossTemplateId: ref(''),
    setSuccess(message) {
      successMessage.value = message
      errorMessage.value = ''
    },
    showEquipmentEditor: ref(false),
    uploadImageToOSS: vi.fn(),
  }
}

describe('装备草稿生成动作', () => {
  it('生成后只填充表单，不刷新列表也不保存', async () => {
    const state = makeState()
    global.fetch = vi.fn(async () => ({
      ok: true,
      json: async () => ({
        draft: {
          itemId: 'soft-blade',
          name: '软组织切割刃',
          slot: 'weapon',
          rarity: '史诗',
          imagePath: '',
          imageAlt: '',
          attackPower: 12,
          armorPenPercent: 0.2,
          critDamageMultiplier: 1.5,
          bossDamagePercent: 0.1,
          partTypeDamageSoft: 0.35,
          partTypeDamageHeavy: 0,
          partTypeDamageWeak: 0.15,
          talentAffinity: 'normal',
        },
      }),
    }))

    const actions = createAdminPageActions(state)
    await actions.generateEquipmentDraft()

    expect(global.fetch).toHaveBeenCalledWith('/api/admin/equipment/generate', expect.objectContaining({ method: 'POST' }))
    expect(state.equipmentForm.value.itemId).toBe('soft-blade')
    expect(state.equipmentForm.value.attackPower).toBe(12)
    expect(state.fetchEquipmentPage).not.toHaveBeenCalled()
    expect(state.showEquipmentEditor.value).toBe(true)
  })
})
