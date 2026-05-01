<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  taskDefinitions: { type: Array, required: true },
  taskForm: { type: Object, required: true },
  taskArchives: { type: Array, required: true },
  taskCycleResults: { type: Object, required: true },
  equipmentOptions: { type: Array, required: true },
  selectedTaskId: { type: String, required: true },
  selectedTaskCycleKey: { type: String, required: true },
  loadingTasks: { type: Boolean, required: true },
  loadingTaskArchives: { type: Boolean, required: true },
  loadingTaskResults: { type: Boolean, required: true },
  saving: { type: Boolean, required: true },
  fetchTasks: { type: Function, required: true },
  fetchTaskArchives: { type: Function, required: true },
  fetchTaskCycleResults: { type: Function, required: true },
  saveTaskDefinition: { type: Function, required: true },
  activateTaskDefinition: { type: Function, required: true },
  deactivateTaskDefinition: { type: Function, required: true },
  duplicateTaskDefinition: { type: Function, required: true },
  archiveExpiredTasks: { type: Function, required: true },
  editTaskDefinition: { type: Function, required: true },
  openNewTask: { type: Function, required: true },
  addTaskEquipmentReward: { type: Function, required: true },
  removeTaskEquipmentReward: { type: Function, required: true },
  formatTime: { type: Function, required: true },
})

const taskStatusFilter = ref('all')
const taskTypeFilter = ref('all')
const archiveStatusFilter = ref('all')

const filteredTaskDefinitions = computed(() =>
  props.taskDefinitions.filter((item) => {
    if (taskStatusFilter.value !== 'all' && item.status !== taskStatusFilter.value) {
      return false
    }
    if (taskTypeFilter.value !== 'all' && item.taskType !== taskTypeFilter.value) {
      return false
    }
    return true
  }),
)

const filteredTaskCycleResults = computed(() =>
  props.taskCycleResults.items.filter((item) => {
    if (archiveStatusFilter.value !== 'all' && item.status !== archiveStatusFilter.value) {
      return false
    }
    return true
  }),
)

function taskStatusLabel(status) {
  switch (status) {
    case 'active':
      return '生效中'
    case 'inactive':
      return '已下线'
    case 'expired':
      return '已过期'
    case 'draft':
    default:
      return '草稿'
  }
}

function conditionLabel(kind) {
  switch (kind) {
    case 'daily_clicks':
      return '当天点击'
    case 'weekly_clicks':
      return '周点击'
    case 'boss_kills':
      return '击败 Boss'
    case 'enhance_count':
      return '强化次数'
    default:
      return kind || '未设置'
  }
}

function cycleLabel(kind) {
  switch (kind) {
    case 'daily':
      return '日常'
    case 'weekly':
      return '周常'
    case 'limited':
      return '限时'
    default:
      return kind || '未知'
  }
}

function archiveStatusLabel(status) {
  switch (status) {
    case 'claimed':
      return '已领取'
    case 'completed_unclaimed':
      return '完成未领'
    case 'unfinished':
      return '未完成'
    case 'not_participated':
      return '未参与'
    default:
      return status || '未知'
  }
}

function equipmentOptionLabel(item) {
  const name = String(item?.name || item?.itemId || '').trim()
  const rarity = String(item?.rarity || '').trim()
  if (!name) {
    return '未命名装备'
  }
  return rarity ? `${name} · ${rarity}` : name
}
</script>

<template>
  <div class="admin-section">
    <div class="admin-inline-actions" style="margin-bottom: 1rem;">
      <button class="nickname-form__submit" type="button" :disabled="saving" @click="openNewTask">新建任务</button>
      <button class="nickname-form__ghost" type="button" :disabled="loadingTasks" @click="fetchTasks">刷新任务</button>
      <button class="nickname-form__ghost" type="button" :disabled="saving" @click="archiveExpiredTasks">归档过期任务</button>
      <select v-model="taskTypeFilter" class="nickname-form__input" style="max-width: 140px;">
        <option value="all">全部周期</option>
        <option value="daily">日常</option>
        <option value="weekly">周常</option>
        <option value="limited">限时</option>
      </select>
      <select v-model="taskStatusFilter" class="nickname-form__input" style="max-width: 140px;">
        <option value="all">全部状态</option>
        <option value="draft">草稿</option>
        <option value="active">生效中</option>
        <option value="inactive">已下线</option>
        <option value="expired">已过期</option>
      </select>
    </div>

    <section class="admin-grid" style="grid-template-columns: minmax(0, 1.2fr) minmax(0, 0.8fr);">
      <article class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">任务列表</p>
          <strong>已配置 {{ filteredTaskDefinitions.length }} 条</strong>
        </div>
        <div v-if="loadingTasks" class="feedback-panel">
          <p>任务列表加载中...</p>
        </div>
        <div v-else-if="filteredTaskDefinitions.length === 0" class="feedback-panel">
          <p>暂无任务配置。</p>
        </div>
        <div v-else class="leaderboard-list">
          <article v-for="item in filteredTaskDefinitions" :key="item.taskId" class="social-card" style="margin-bottom: 0.75rem;">
            <div class="social-card__head">
              <div>
                <p class="vote-stage__eyebrow">{{ cycleLabel(item.taskType) }} · {{ taskStatusLabel(item.status) }}</p>
                <strong>{{ item.title || item.taskId }}</strong>
              </div>
              <strong>{{ item.taskId }}</strong>
            </div>
            <p class="social-card__copy">{{ item.description || '未填写描述' }}</p>
            <p class="social-card__copy">{{ conditionLabel(item.conditionKind) }} ≥ {{ item.targetValue }}</p>
            <div class="admin-inline-actions">
              <button class="nickname-form__ghost" type="button" @click="editTaskDefinition(item)">编辑</button>
              <button class="nickname-form__ghost" type="button" @click="duplicateTaskDefinition(item.taskId)">复制</button>
              <button
                v-if="item.status !== 'active'"
                class="nickname-form__ghost"
                type="button"
                @click="activateTaskDefinition(item.taskId)"
              >
                上线
              </button>
              <button
                v-else
                class="nickname-form__ghost"
                type="button"
                @click="deactivateTaskDefinition(item.taskId)"
              >
                下线
              </button>
              <button class="nickname-form__ghost" type="button" @click="fetchTaskArchives(item.taskId)">归档</button>
            </div>
          </article>
        </div>
      </article>

      <article class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">任务编辑器</p>
          <strong>{{ taskForm.taskId || '新任务' }}</strong>
        </div>
        <form class="admin-form" @submit.prevent="saveTaskDefinition">
          <input v-model="taskForm.taskId" class="nickname-form__input" type="text" placeholder="taskId" />
          <input v-model="taskForm.title" class="nickname-form__input" type="text" placeholder="标题" />
          <textarea v-model="taskForm.description" class="nickname-form__input" rows="3" placeholder="描述"></textarea>
          <select v-model="taskForm.taskType" class="nickname-form__input">
            <option value="daily">日常</option>
            <option value="weekly">周常</option>
            <option value="limited">限时</option>
          </select>
          <select v-model="taskForm.conditionKind" class="nickname-form__input">
            <option value="daily_clicks">当天点击</option>
            <option value="weekly_clicks">周点击</option>
            <option value="boss_kills">击败 Boss</option>
            <option value="enhance_count">强化次数</option>
          </select>
          <input v-model="taskForm.targetValue" class="nickname-form__input" type="number" min="1" placeholder="目标值" />
          <input v-model="taskForm.displayOrder" class="nickname-form__input" type="number" min="0" placeholder="展示顺序" />
          <input
            v-if="taskForm.taskType === 'limited'"
            v-model="taskForm.startAt"
            class="nickname-form__input"
            type="number"
            min="0"
            placeholder="开始时间戳，限时任务必填"
          />
          <input
            v-if="taskForm.taskType === 'limited'"
            v-model="taskForm.endAt"
            class="nickname-form__input"
            type="number"
            min="0"
            placeholder="结束时间戳，且必须大于开始时间"
          />

          <div class="leaderboard-list">
            <p>奖励配置</p>
            <input v-model="taskForm.rewards.gold" class="nickname-form__input" type="number" min="0" placeholder="金币" />
            <input v-model="taskForm.rewards.stones" class="nickname-form__input" type="number" min="0" placeholder="强化石" />
            <input v-model="taskForm.rewards.talentPoints" class="nickname-form__input" type="number" min="0" placeholder="天赋点" />
            <div
              v-for="(entry, index) in taskForm.rewards.equipmentItems"
              :key="`${index}-${entry.itemId}`"
              class="admin-inline-actions"
            >
              <select v-model="entry.itemId" class="nickname-form__input">
                <option value="">选择装备模板</option>
                <option v-for="item in equipmentOptions" :key="item.itemId" :value="item.itemId">
                  {{ equipmentOptionLabel(item) }}
                </option>
              </select>
              <input v-model="entry.quantity" class="nickname-form__input" type="number" min="1" placeholder="数量" />
              <button class="nickname-form__ghost" type="button" @click="removeTaskEquipmentReward(index)">移除</button>
            </div>
            <button class="nickname-form__ghost" type="button" @click="addTaskEquipmentReward">添加装备奖励</button>
          </div>
          <button class="nickname-form__submit" type="submit" :disabled="saving">{{ saving ? '保存中...' : '保存任务' }}</button>
        </form>
      </article>
    </section>

    <section class="admin-grid" style="margin-top: 1rem; grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);">
      <article class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">周期归档</p>
          <strong>{{ selectedTaskId || '未选择任务' }}</strong>
        </div>
        <div v-if="loadingTaskArchives" class="feedback-panel">
          <p>任务归档加载中...</p>
        </div>
        <div v-else-if="taskArchives.length === 0" class="feedback-panel">
          <p>当前任务暂无归档。</p>
        </div>
        <div v-else class="leaderboard-list">
          <article v-for="archive in taskArchives" :key="archive.cycleKey" class="social-card" style="margin-bottom: 0.75rem;">
            <div class="social-card__head">
              <div>
                <p class="vote-stage__eyebrow">{{ archive.cycleKey }}</p>
                <strong>{{ cycleLabel(archive.taskType) }} · {{ conditionLabel(archive.conditionKind) }}</strong>
              </div>
              <button class="nickname-form__ghost" type="button" @click="fetchTaskCycleResults(archive.taskId, archive.cycleKey)">查看明细</button>
            </div>
            <p class="social-card__copy">
              参与 {{ archive.participantsTotal }} · 完成 {{ archive.completedTotal }} · 已领 {{ archive.claimedTotal }}
            </p>
            <p class="social-card__copy">归档时间 {{ formatTime(archive.archivedAt) }}</p>
          </article>
        </div>
      </article>

      <article class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">周期明细</p>
          <strong>{{ selectedTaskCycleKey || '未选择周期' }}</strong>
        </div>
        <div v-if="loadingTaskResults" class="feedback-panel">
          <p>任务周期明细加载中...</p>
        </div>
        <div v-else-if="filteredTaskCycleResults.length === 0" class="feedback-panel">
          <p>当前周期暂无玩家结果。</p>
        </div>
        <div v-else class="leaderboard-list">
          <div class="admin-inline-actions" style="margin-bottom: 0.75rem;">
            <select v-model="archiveStatusFilter" class="nickname-form__input" style="max-width: 160px;">
              <option value="all">全部结果</option>
              <option value="claimed">已领取</option>
              <option value="completed_unclaimed">完成未领</option>
              <option value="unfinished">未完成</option>
              <option value="not_participated">未参与</option>
            </select>
          </div>
          <article v-for="item in filteredTaskCycleResults" :key="`${item.nickname}-${item.taskId}-${item.cycleKey}`" class="leaderboard-list__item">
            <span class="leaderboard-list__name">{{ item.nickname }}</span>
            <span>{{ archiveStatusLabel(item.status) }}</span>
            <strong class="leaderboard-list__count">{{ item.progress }}/{{ item.targetValue }}</strong>
          </article>
        </div>
      </article>
    </section>
  </div>
</template>
