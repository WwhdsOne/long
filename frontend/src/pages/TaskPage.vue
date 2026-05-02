<script setup>
import { ref } from 'vue'

import { getRarityClassName } from '../utils/rarity'
import { usePublicPageState } from './publicPageState'

const {
  tasks,
  claimTask,
  formatItemStatLines,
  formatRarityLabel,
  equipmentNameClass,
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
    case 'daily':
    default:
      return '按天累计'
  }
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
      <div v-else class="task-card__grid">
        <article
          v-for="task in tasks"
          :key="`${task.taskId}-${task.cycleKey}`"
          class="social-card task-card__item"
        >
          <div class="social-card__head">
            <div>
              <p class="vote-stage__eyebrow">{{ eventLabel(task.eventKind) }} · {{ windowLabel(task.windowKind) }} · {{ taskStatusLabel(task.status) }}</p>
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
              <strong class="task-equipment-chip__name" :class="equipmentNameClass(detail)">{{ detail.name }}</strong>
              <span class="task-equipment-chip__rarity" :class="getRarityClassName(detail.rarity)">{{ formatRarityLabel(detail.rarity) }}</span>
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
