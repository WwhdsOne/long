<script setup>
import { ref } from 'vue'

import { usePublicPageState } from './publicPageState'

const {
  tasks,
  claimTask,
} = usePublicPageState()

const claimingTaskId = ref('')

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

function taskTypeLabel(taskType) {
  switch (taskType) {
    case 'daily':
      return '日常'
    case 'weekly':
      return '周常'
    case 'limited':
      return '限时'
    default:
      return taskType || '任务'
  }
}

function formatTaskRewards(rewards) {
  const parts = []
  if (Number(rewards?.gold || 0) > 0) parts.push(`金币 ${rewards.gold}`)
  if (Number(rewards?.stones || 0) > 0) parts.push(`强化石 ${rewards.stones}`)
  if (Number(rewards?.talentPoints || 0) > 0) parts.push(`天赋点 ${rewards.talentPoints}`)
  if (Array.isArray(rewards?.equipmentItems)) {
    rewards.equipmentItems.forEach((entry) => {
      if (entry?.itemId) {
        parts.push(`装备 ${entry.itemId} ×${entry.quantity || 1}`)
      }
    })
  }
  return parts.join(' · ') || '无奖励'
}

async function handleClaimTask(taskId) {
  claimingTaskId.value = taskId
  try {
    await claimTask(taskId)
  } finally {
    claimingTaskId.value = ''
  }
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
      </div>
      <div v-if="tasks.length === 0" class="leaderboard-list leaderboard-list--empty">
        <p>当前没有生效任务，晚点再来看看。</p>
      </div>
      <div v-else class="leaderboard-list">
        <article
          v-for="task in tasks"
          :key="`${task.taskId}-${task.cycleKey}`"
          class="social-card task-card__item"
          style="margin-bottom: 0.75rem;"
        >
          <div class="social-card__head">
            <div>
              <p class="vote-stage__eyebrow">{{ taskTypeLabel(task.taskType) }} · {{ taskStatusLabel(task.status) }}</p>
              <strong>{{ task.title }}</strong>
            </div>
            <strong>{{ task.progress }}/{{ task.targetValue }}</strong>
          </div>
          <p class="social-card__copy">{{ task.description || '未填写任务描述' }}</p>
          <p class="social-card__copy">奖励：{{ formatTaskRewards(task.rewards) }}</p>
          <div class="announcement-modal__actions" style="justify-content: flex-start;">
            <button
              class="nickname-form__submit"
              type="button"
              :disabled="!task.canClaim || claimingTaskId === task.taskId"
              @click="handleClaimTask(task.taskId)"
            >
              {{ claimingTaskId === task.taskId ? '领取中...' : (task.canClaim ? '领取奖励' : '暂不可领') }}
            </button>
          </div>
        </article>
      </div>
    </section>
  </section>
</template>
