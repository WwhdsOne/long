<script setup>
import {computed, ref} from 'vue'

function toDatetimeLocal(ts) {
  if (!ts || ts <= 0) return ''
  const d = new Date(ts * 1000)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function fromDatetimeLocal(str) {
  if (!str) return 0
  return Math.floor(new Date(str).getTime() / 1000)
}

const props = defineProps({
  taskDefinitions: {type: Array, required: true},
  taskForm: {type: Object, required: true},
  taskArchives: {type: Array, required: true},
  taskCycleResults: {type: Object, required: true},
  equipmentOptions: {type: Array, required: true},
  selectedTaskId: {type: String, required: true},
  selectedTaskCycleKey: {type: String, required: true},
  loadingTasks: {type: Boolean, required: true},
  loadingTaskArchives: {type: Boolean, required: true},
  loadingTaskResults: {type: Boolean, required: true},
  saving: {type: Boolean, required: true},
  fetchTasks: {type: Function, required: true},
  fetchTaskArchives: {type: Function, required: true},
  fetchTaskCycleResults: {type: Function, required: true},
  saveTaskDefinition: {type: Function, required: true},
  activateTaskDefinition: {type: Function, required: true},
  deactivateTaskDefinition: {type: Function, required: true},
  duplicateTaskDefinition: {type: Function, required: true},
  archiveExpiredTasks: {type: Function, required: true},
  editTaskDefinition: {type: Function, required: true},
  openNewTask: {type: Function, required: true},
  addTaskEquipmentReward: {type: Function, required: true},
  removeTaskEquipmentReward: {type: Function, required: true},
  formatTime: {type: Function, required: true},
})

const taskStatusFilter = ref('all')
const taskWindowFilter = ref('all')
const archiveStatusFilter = ref('all')

const filteredTaskDefinitions = computed(() =>
    props.taskDefinitions.filter((item) => {
      if (taskStatusFilter.value !== 'all' && item.status !== taskStatusFilter.value) {
        return false
      }
      return !(taskWindowFilter.value !== 'all' && item.windowKind !== taskWindowFilter.value);

    }),
)

const filteredTaskCycleResults = computed(() =>
    props.taskCycleResults.items.filter((item) => {
      return !(archiveStatusFilter.value !== 'all' && item.status !== archiveStatusFilter.value);

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

function eventLabel(kind) {
  switch (kind) {
    case 'boss_kill':
      return '击败 Boss'
    case 'enhance':
      return '强化次数'
    case 'click':
    default:
      return '点击次数'
  }
}

function windowLabel(kind) {
  switch (kind) {
    case 'weekly':
      return '按周累计'
    case 'fixed_range':
      return '固定时间窗'
    case 'lifetime':
      return '长期有效'
    case 'daily':
    default:
      return '按天累计'
  }
}

function taskMetaLabel(item) {
  return `${eventLabel(item?.eventKind)} · ${windowLabel(item?.windowKind)}`
}

function cycleLabel(kind) {
  switch (kind) {
    case 'weekly':
      return '周周期'
    case 'fixed_range':
      return '时间窗'
    case 'daily':
    default:
      return kind || '未设置'
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
    <div class="admin-inline-actions" style="margin-bottom: 1rem; display: flex; align-items: center; gap: 10px;">
      <button class="nickname-form__submit" type="button" :disabled="saving" @click="openNewTask">新建任务</button>
      <button class="nickname-form__ghost" type="button" :disabled="loadingTasks" @click="fetchTasks">刷新任务</button>
      <button class="nickname-form__ghost" type="button" :disabled="saving" @click="archiveExpiredTasks">归档过期任务
      </button>

      <!-- Filter Selects -->
      <select v-model="taskWindowFilter" class="nickname-form__input" style="max-width: 140px;">
        <option value="all">全部周期</option>
        <option value="daily">日常</option>
        <option value="weekly">周常</option>
        <option value="fixed_range">限时</option>
        <option value="lifetime">长期有效</option>
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
          <article v-for="item in filteredTaskDefinitions" :key="item.taskId" class="social-card"
                   style="margin-bottom: 0.75rem;">
            <div class="social-card__head">
              <div>
                <p class="vote-stage__eyebrow">{{ windowLabel(item.windowKind) }} · {{
                    taskStatusLabel(item.status)
                  }}</p>
                <strong>{{ item.title || item.taskId }}</strong>
              </div>
              <strong>{{ item.taskId }}</strong>
            </div>
            <p class="social-card__copy">{{ item.description || '未填写描述' }}</p>
            <p class="social-card__copy">{{ taskMetaLabel(item) }} ≥ {{ item.targetValue }}</p>
            <div class="admin-inline-actions">
              <button class="nickname-form__ghost" type="button" @click="editTaskDefinition(item)">编辑</button>
              <button class="nickname-form__ghost" type="button" @click="duplicateTaskDefinition(item.taskId)">复制
              </button>
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
          <div class="form-group">
            <label for="taskId">任务 ID</label>
            <input id="taskId" v-model="taskForm.taskId" class="nickname-form__input" type="text"/>
          </div>

          <div class="form-group">
            <label for="title">标题</label>
            <input id="title" v-model="taskForm.title" class="nickname-form__input" type="text"/>
          </div>

          <div class="form-group">
            <label for="description">描述</label>
            <textarea id="description" v-model="taskForm.description" class="nickname-form__input" rows="3"></textarea>
          </div>

          <div class="form-group">
            <label for="eventKind">行为类型</label>
            <select id="eventKind" v-model="taskForm.eventKind" class="nickname-form__input">
              <option value="click">点击次数</option>
              <option value="boss_kill">击败 Boss</option>
              <option value="enhance">强化次数</option>
            </select>
          </div>

          <div class="form-group">
            <label for="windowKind">累计窗口</label>
            <select id="windowKind" v-model="taskForm.windowKind" class="nickname-form__input">
              <option value="daily">按天累计</option>
              <option value="weekly">按周累计</option>
              <option value="fixed_range">固定时间窗</option>
              <option value="lifetime">长期有效</option>
            </select>
          </div>

          <div class="form-group">
            <label for="targetValue">目标值</label>
            <input id="targetValue" v-model="taskForm.targetValue" class="nickname-form__input" type="number" min="1"/>
          </div>

          <div class="form-group">
            <label for="displayOrder">展示顺序</label>
            <input id="displayOrder" v-model="taskForm.displayOrder" class="nickname-form__input" type="number"
                   min="0"/>
          </div>

          <div v-if="taskForm.windowKind === 'fixed_range'">
            <div class="form-group">
              <label for="startAt">开始时间</label>
              <input id="startAt" :value="toDatetimeLocal(taskForm.startAt)"
                     @input="taskForm.startAt = fromDatetimeLocal($event.target.value)" class="nickname-form__input"
                     type="datetime-local"/>
            </div>
            <div class="form-group">
              <label for="endAt">结束时间</label>
              <input id="endAt" :value="toDatetimeLocal(taskForm.endAt)"
                     @input="taskForm.endAt = fromDatetimeLocal($event.target.value)" class="nickname-form__input"
                     type="datetime-local"/>
            </div>
          </div>

          <div class="leaderboard-list">
            <p>奖励配置</p>
            <div class="form-group">
              <label for="gold">金币</label>
              <input id="gold" v-model="taskForm.rewards.gold" class="nickname-form__input" type="number" min="0"/>
            </div>
            <div class="form-group">
              <label for="stones">强化石</label>
              <input id="stones" v-model="taskForm.rewards.stones" class="nickname-form__input" type="number" min="0"/>
            </div>
            <div class="form-group">
              <label for="talentPoints">天赋点</label>
              <input id="talentPoints" v-model="taskForm.rewards.talentPoints" class="nickname-form__input"
                     type="number" min="0"/>
            </div>

            <div
                v-for="(entry, index) in taskForm.rewards.equipmentItems"
                :key="`${index}-${entry.itemId}`"
                class="admin-inline-actions"
            >
              <div class="form-group">
                <label for="itemId">选择装备模板</label>
                <select v-model="entry.itemId" class="nickname-form__input">
                  <option value="">选择装备模板</option>
                  <option v-for="item in equipmentOptions" :key="item.itemId" :value="item.itemId">
                    {{ equipmentOptionLabel(item) }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="quantity">数量</label>
                <input v-model="entry.quantity" class="nickname-form__input" type="number" min="1"/>
              </div>
              <button class="nickname-form__ghost" type="button" @click="removeTaskEquipmentReward(index)">移除</button>
            </div>
            <button class="nickname-form__ghost" type="button" @click="addTaskEquipmentReward">添加装备奖励</button>
          </div>

          <button class="nickname-form__submit" type="submit" :disabled="saving">{{
              saving ? '保存中...' : '保存任务'
            }}
          </button>
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
          <article v-for="archive in taskArchives" :key="archive.cycleKey" class="social-card"
                   style="margin-bottom: 0.75rem;">
            <div class="social-card__head">
              <div>
                <p class="vote-stage__eyebrow">{{ archive.cycleKey }}</p>
                <strong>{{ taskMetaLabel(archive) }}</strong>
              </div>
              <button class="nickname-form__ghost" type="button"
                      @click="fetchTaskCycleResults(archive.taskId, archive.cycleKey)">查看明细
              </button>
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
          <article v-for="item in filteredTaskCycleResults" :key="`${item.nickname}-${item.taskId}-${item.cycleKey}`"
                   class="leaderboard-list__item">
            <span class="leaderboard-list__name">{{ item.nickname }}</span>
            <span>{{ archiveStatusLabel(item.status) }}</span>
            <strong class="leaderboard-list__count">{{ item.progress }}/{{ item.targetValue }}</strong>
          </article>
        </div>
      </article>
    </section>
  </div>
</template>
