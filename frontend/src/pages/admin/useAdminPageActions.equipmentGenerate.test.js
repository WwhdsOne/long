import { ref } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import { createAdminPageActions } from './useAdminPageActions'
import { emptyEquipmentForm, emptyTaskForm } from './state'

function makeState() {
  const equipmentForm = ref(emptyEquipmentForm())
  const equipmentPage = ref({ items: [], page: 1 })
  const errorMessage = ref('')
  const successMessage = ref('')
  const taskForm = ref(emptyTaskForm())

  return {
    activeTab: ref('equipment'),
    addLootRow: vi.fn(),
    adminRoomId: ref('2'),
    adminState: ref({ roomId: '2' }),
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
    emptyTaskForm,
    equipmentForm,
    equipmentPage,
    equipmentPrompt: ref('做一把偏普攻流的软组织武器'),
    errorMessage,
    fetchAdminState: vi.fn(),
    fetchAnnouncements: vi.fn(),
    fetchButtonPage: vi.fn(),
    fetchEquipmentPage: vi.fn(),
    fetchMessages: vi.fn(),
    fetchTaskArchives: vi.fn(),
    fetchTaskCycleResults: vi.fn(),
    fetchTasks: vi.fn(),
    findBossTemplate: vi.fn(),
    generatingEquipmentDraft: ref(false),
    loadingTaskArchives: ref(false),
    loadingTaskResults: ref(false),
    loadingTasks: ref(false),
    lootRows: ref([]),
    readErrorMessage: vi.fn(async () => '生成失败'),
    saving: ref(false),
    selectedBossTemplateId: ref(''),
    selectedTaskCycleKey: ref(''),
    selectedTaskId: ref(''),
    setSuccess(message) {
      successMessage.value = message
      errorMessage.value = ''
    },
    showEquipmentEditor: ref(false),
    taskDefinitions: ref([]),
    taskForm,
    uploadImageToOSS: vi.fn(),
  }
}

describe('装备草稿生成动作', () => {
  it('Boss 房间操作会提交当前后台房间', async () => {
    const state = makeState()
    global.fetch = vi.fn(async () => ({
      ok: true,
      json: async () => ({}),
    }))

    const actions = createAdminPageActions(state)
    await actions.enableBossCycle()
    await actions.saveBossCycleQueue(['dragon'])

    expect(global.fetch).toHaveBeenNthCalledWith(1, '/api/admin/boss/cycle/enable?roomId=2', { method: 'POST' })
    expect(global.fetch).toHaveBeenNthCalledWith(2, '/api/admin/boss/cycle/queue?roomId=2', expect.objectContaining({
      method: 'PUT',
      body: expect.any(String),
    }))
    expect(JSON.parse(global.fetch.mock.calls[1][1].body)).toEqual({
      roomId: '2',
      templateIds: ['dragon'],
    })
  })

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

  it('保存装备时会提交 critRate 数值字段', async () => {
    const state = makeState()
    state.equipmentForm.value = {
      ...emptyEquipmentForm(),
      itemId: 'soft-blade',
      name: '软组织切割刃',
      slot: 'weapon',
      rarity: '史诗',
      attackPower: '12',
      armorPenPercent: '0.2',
      critRate: '0.22',
      critDamageMultiplier: '1.5',
    }

    global.fetch = vi.fn(async () => ({
      ok: true,
      json: async () => ({}),
    }))

    const actions = createAdminPageActions(state)
    await actions.saveEquipment()

    expect(global.fetch).toHaveBeenCalledWith('/api/admin/equipment', expect.objectContaining({
      method: 'POST',
      body: expect.any(String),
    }))

    const body = JSON.parse(global.fetch.mock.calls[0][1].body)
    expect(body.critRate).toBe(0.22)
  })

  it('保存任务时会提交奖励和时间窗口数值字段', async () => {
    const state = makeState()
    state.taskForm.value = {
      ...emptyTaskForm(),
      taskId: 'limited-enhance',
      title: '限时强化',
      taskType: 'limited',
      conditionKind: 'enhance_count',
      targetValue: '5',
      displayOrder: '7',
      startAt: '100',
      endAt: '200',
      rewards: {
        gold: '300',
        stones: '6',
        talentPoints: '2',
        equipmentItems: [
          { itemId: 'blade-01', quantity: '2' },
          { itemId: '', quantity: '4' },
        ],
      },
    }

    global.fetch = vi.fn(async () => ({
      ok: true,
      json: async () => ({}),
    }))

    const actions = createAdminPageActions(state)
    await actions.saveTaskDefinition()

    expect(global.fetch).toHaveBeenCalledWith('/api/admin/tasks', expect.objectContaining({
      method: 'POST',
      body: expect.any(String),
    }))

    const body = JSON.parse(global.fetch.mock.calls[0][1].body)
    expect(body.targetValue).toBe(5)
    expect(body.displayOrder).toBe(7)
    expect(body.startAt).toBe(100)
    expect(body.endAt).toBe(200)
    expect(body.rewards.gold).toBe(300)
    expect(body.rewards.stones).toBe(6)
    expect(body.rewards.talentPoints).toBe(2)
    expect(body.rewards.equipmentItems).toEqual([{ itemId: 'blade-01', quantity: 2 }])
  })

  it('限时任务缺少合法时间窗口时不会发请求', async () => {
    const state = makeState()
    state.taskForm.value = {
      ...emptyTaskForm(),
      taskId: 'limited-enhance',
      title: '限时强化',
      taskType: 'limited',
      conditionKind: 'enhance_count',
      targetValue: 5,
      rewards: {
        gold: 300,
        stones: 0,
        talentPoints: 0,
        equipmentItems: [],
      },
      startAt: 200,
      endAt: 100,
    }

    global.fetch = vi.fn()

    const actions = createAdminPageActions(state)
    await actions.saveTaskDefinition()

    expect(global.fetch).not.toHaveBeenCalled()
    expect(state.errorMessage.value).toContain('开始时间')
  })

  it('任务奖励全空时不会发请求', async () => {
    const state = makeState()
    state.taskForm.value = {
      ...emptyTaskForm(),
      taskId: 'daily-click',
      title: '今日点击',
      taskType: 'daily',
      conditionKind: 'daily_clicks',
      targetValue: 5,
      rewards: {
        gold: 0,
        stones: 0,
        talentPoints: 0,
        equipmentItems: [],
      },
    }

    global.fetch = vi.fn()

    const actions = createAdminPageActions(state)
    await actions.saveTaskDefinition()

    expect(global.fetch).not.toHaveBeenCalled()
    expect(state.errorMessage.value).toContain('奖励')
  })
})
