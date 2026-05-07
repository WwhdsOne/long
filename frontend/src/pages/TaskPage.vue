<script setup>
import {computed, ref} from 'vue'

import {getRarityClassName} from '../utils/rarity'
import {usePublicPageState} from './publicPageState'

const {
  tasks,
  claimTask,
  formatItemStatLines,
  formatRarityLabel,
  equipmentNameClass,
} = usePublicPageState()

const claimingTaskId = ref('')
const claimingAllTasks = ref(false)
const selectedTaskGroup = ref('routine')

const claimableTasks = computed(() => tasks.value.filter((task) => Boolean(task?.canClaim)))
const taskGroups = [
  {key: 'routine', title: '日常 / 周常'},
  {key: 'activity', title: '活动'},
  {key: 'longTerm', title: '长期有效'},
]

const groupedTasks = computed(() => taskGroups.map((group) => ({
  ...group,
  items: tasks.value.filter((task) => taskGroupKey(task) === group.key),
})).filter((group) => group.items.length > 0))

const visibleTaskGroups = computed(() => taskGroups.filter((group) =>
    tasks.value.some((task) => taskGroupKey(task) === group.key),
))

const activeTaskGroup = computed(() => {
  const selected = visibleTaskGroups.value.find((group) => group.key === selectedTaskGroup.value)
  return selected || visibleTaskGroups.value[0] || null
})

const activeTasks = computed(() => {
  if (!activeTaskGroup.value) {
    return []
  }
  return tasks.value.filter((task) => taskGroupKey(task) === activeTaskGroup.value.key)
})

function taskStatusLabel(status) {
  switch (status) {
    case 'claimed':
      return '已领取'
    case 'completed_unclaimed':
      return '可领取'
    case 'in_progress':
      return '进行中'
    case 'unfinished':
      return '未完成'
    case 'not_participated':
      return '未参与'
    default:
      return status || '未知'
  }
}

function eventLabel(eventKind) {
  switch (eventKind) {
    case 'boss_kill':
      return '击败 Boss'
    case 'enhance':
      return '强化次数'
    case 'click':
    default:
      return '点击次数'
  }
}

function windowLabel(windowKind) {
  switch (windowKind) {
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

function taskGroupKey(task) {
  if (task?.windowKind === 'fixed_range') {
    return 'activity'
  }
  if (task?.windowKind === 'lifetime') {
    return 'longTerm'
  }
  return 'routine'
}

function selectTaskGroup(groupKey) {
  selectedTaskGroup.value = groupKey
}

function formatTaskRewards(rewards, equipDetails) {
  const parts = []
  if (Number(rewards?.gold || 0) > 0) parts.push(`金币 ${rewards.gold}`)
  if (Number(rewards?.stones || 0) > 0) parts.push(`强化石 ${rewards.stones}`)
  if (Number(rewards?.talentPoints || 0) > 0) parts.push(`天赋点 ${rewards.talentPoints}`)
  if (Array.isArray(rewards?.equipmentItems)) {
    rewards.equipmentItems.forEach((entry, idx) => {
      if (entry?.itemId) {
        const detail = Array.isArray(equipDetails) ? equipDetails[idx] : null
        const name = detail?.name || entry.itemId
        parts.push(`装备 ${name} ×${entry.quantity || 1}`)
      }
    })
  }
  return parts.join(' · ') || '无奖励'
}

function equipmentRewardChips(task) {
  if (!Array.isArray(task.equipmentRewardDetails) || task.equipmentRewardDetails.length === 0) {
    return []
  }
  return task.equipmentRewardDetails
}

async function handleClaimTask(taskId) {
  claimingTaskId.value = taskId
  try {
    await claimTask(taskId)
  } finally {
    claimingTaskId.value = ''
  }
}

async function handleClaimAllTasks() {
  if (claimingAllTasks.value || claimableTasks.value.length === 0) {
    return
  }
  claimingAllTasks.value = true
  try {
    for (const task of claimableTasks.value) {
      claimingTaskId.value = task.taskId
      const result = await claimTask(task.taskId)
      if (!result?.ok) {
        break
      }
    }
  } finally {
    claimingTaskId.value = ''
    claimingAllTasks.value = false
  }
}

function claimButtonLabel(task) {
  if (claimingTaskId.value === task.taskId) {
    return '领取中...'
  }
  if (task.canClaim) {
    return '领取奖励'
  }
  if (task.status === 'claimed') {
    return '已领取'
  }
  return taskStatusLabel(task.status)
}
</script>

<template>
  <section class="stage-layout stage-layout--single">
    <section class="task-card">
      <div class="social-card__head">
        <div>
          <p class="vote-stage__eyebrow">当前任务</p>
          <strong>{{ tasks.length > 0 ? `${tasks.length} 条进行中` : '暂无可见任务' }}</strong>
        </div>
        <button
            class="nickname-form__submit"
            type="button"
            :disabled="claimableTasks.length === 0 || claimingAllTasks"
            @click="handleClaimAllTasks"
        >
          {{ claimingAllTasks ? '领取中...' : '一键领取' }}
        </button>
      </div>
      <div v-if="tasks.length === 0" class="leaderboard-list leaderboard-list--empty">
        <p>当前没有生效任务，晚点再来看看。</p>
      </div>
      <div v-else>
        <div class="task-card__tabs">
          <button
              v-for="group in visibleTaskGroups"
              :key="group.key"
              class="task-card__tab"
              :class="{ 'task-card__tab--active': activeTaskGroup?.key === group.key }"
              type="button"
              @click="selectTaskGroup(group.key)"
          >
            {{ group.title }}
          </button>
        </div>
        <section v-if="activeTaskGroup" class="task-card__group">
          <div class="social-card__head">
            <div>
              <p class="vote-stage__eyebrow">任务分类</p>
              <strong>{{ activeTaskGroup.title }}</strong>
            </div>
            <strong>{{ activeTasks.length }} 条</strong>
          </div>
          <div class="task-card__grid">
            <article
                v-for="task in activeTasks"
                :key="`${task.taskId}-${task.cycleKey}`"
                class="social-card task-card__item"
            >
              <div class="social-card__head">
                <div>
                  <p class="vote-stage__eyebrow">{{ eventLabel(task.eventKind) }} · {{ windowLabel(task.windowKind) }} ·
                    {{ taskStatusLabel(task.status) }}</p>
                  <strong>{{ task.title }}</strong>
                </div>
                <strong>{{ task.progress }}/{{ task.targetValue }}</strong>
              </div>
              <p class="social-card__copy">{{ task.description || '未填写任务描述' }}</p>
              <p class="social-card__copy">奖励：{{ formatTaskRewards(task.rewards, task.equipmentRewardDetails) }}</p>
              <div v-if="equipmentRewardChips(task).length > 0" class="task-equipment-chips">
                <span
                    v-for="detail in equipmentRewardChips(task)"
                    :key="detail.itemId"
                    class="task-equipment-chip"
                >
                  <img
                      v-if="detail.imagePath"
                      class="task-equipment-chip__image"
                      :src="detail.imagePath"
                      :alt="detail.imageAlt || detail.name"
                  />
                  <strong class="task-equipment-chip__name" :class="equipmentNameClass(detail)">{{
                      detail.name
                    }}</strong>
                  <span class="task-equipment-chip__rarity"
                        :class="getRarityClassName(detail.rarity)">{{ formatRarityLabel(detail.rarity) }}</span>
                  <article class="armory-item-tooltip task-equipment-tooltip" aria-label="装备属性">
                    <p class="vote-stage__eyebrow">装备属性</p>
                    <strong :class="equipmentNameClass(detail)">{{ detail.name }}</strong>
                    <p>{{ formatRarityLabel(detail.rarity) }} · 装备</p>
                    <ul v-if="formatItemStatLines(detail).length > 0" class="armory-item-tooltip__stats">
                      <li v-for="line in formatItemStatLines(detail)" :key="line">{{ line }}</li>
                    </ul>
                    <p v-else>暂无词条</p>
                  </article>
                </span>
              </div>
              <div class="announcement-modal__actions" style="justify-content: flex-start;">
                <button
                    class="nickname-form__submit"
                    :class="{ 'nickname-form__submit--claimed': task.status === 'claimed' }"
                    type="button"
                    :disabled="!task.canClaim || claimingTaskId === task.taskId || claimingAllTasks"
                    @click="handleClaimTask(task.taskId)"
                >
                  {{ claimButtonLabel(task) }}
                </button>
              </div>
            </article>
          </div>
        </section>
      </div>
    </section>
  </section>
</template>
