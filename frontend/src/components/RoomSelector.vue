<script setup>
import RoomSwitchCooldownTag from './RoomSwitchCooldownTag.vue'
import {formatCompact} from '../utils/formatNumber'

const props = defineProps({
  rooms: {
    type: Array,
    default: () => [],
  },
  currentRoomId: {
    type: String,
    default: '1',
  },
  switching: {
    type: Boolean,
    default: false,
  },
  error: {
    type: String,
    default: '',
  },
  loggedIn: {
    type: Boolean,
    default: false,
  },
  cooldownRemainingSeconds: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['join'])

const ROOM_ACCENTS = [
  '105, 231, 176',
  '120, 216, 255',
  '255, 133, 119',
  '255, 212, 107',
]

function roomLabel(room) {
  const displayName = String(room?.displayName || '').trim()
  if (displayName) return displayName
  return defaultRoomLabel(room?.id)
}

function defaultRoomLabel(roomId) {
  const normalized = String(roomId || '1').trim() || '1'
  return `房间 ${normalized}`
}

function currentRoomSummary(roomId) {
  const normalized = String(roomId || '').trim()
  if (normalized === 'hall') return '当前大厅'
  if (normalized === '') return '当前大厅'
  return `当前 ${roomLabel({id: normalized})}`
}

function roomAccent(room, index) {
  const idIndex = Number.parseInt(room?.id, 10)
  const offset = Number.isFinite(idIndex) ? idIndex - 1 : index
  return ROOM_ACCENTS[Math.max(0, offset) % ROOM_ACCENTS.length]
}

function roomStyle(room, index) {
  const idIndex = Number.parseInt(room?.id, 10)
  const offset = Math.max(0, Number.isFinite(idIndex) ? idIndex - 1 : index)
  return {
    '--room-accent-rgb': roomAccent(room, index),
    '--room-orbit-delay': `${offset * -1.1}s`,
    '--room-breathe-delay': `${offset * -0.7}s`,
    '--room-orbit-duration': `${5.8 + (offset % 3) * 0.45}s`,
  }
}

function roomBossName(room) {
  if (!room) return '等待 Boss'
  else if (room.currentBossName === "" || room.currentBossName === undefined) return room.cycleEnabled ? 'Boss循环待命' : '未开启Boss循环'
  return '当前Boss : ' + room.currentBossName
}

function roomOnlineText(room) {
  return `${Math.max(0, Number(room?.onlineCount || 0))} 人`
}

function roomAvgHpText(room) {
  const value = formatCompact(Math.max(0, Number(room?.currentBossAvgHp || 0)))
  return value === '0' ? '--' : value
}

function roomStatusLabel(room) {
  if (!room) return '未知'
  if (isCurrentRoom(room)) return '当前房间'
  if (room.joinable === false) return '冷却中'
  return room.cycleEnabled ? '可切换' : '待开放'
}

function roomActionLabel(room) {
  if (!props.loggedIn) return '登录后切换'
  if (props.switching) return '切换中'
  if (isCurrentRoom(room)) return '当前所在'
  if (props.cooldownRemainingSeconds > 0) return '冷却未结束'
  if (room?.joinable === false) return '暂不可切'
  return '进入战线'
}

function isCurrentRoom(room) {
  return String(room?.id || '') === String(props.currentRoomId || '')
}

function canJoin(room) {
  return props.loggedIn && !props.switching && props.cooldownRemainingSeconds <= 0 && room?.joinable !== false && !isCurrentRoom(room)
}
</script>

<template>
  <section class="room-selector" aria-label="房间选择">
    <div class="room-selector__head">
      <div class="room-selector__title">
        <strong>战线分流</strong>
        <small>{{ currentRoomSummary(currentRoomId) }}</small>
      </div>
      <RoomSwitchCooldownTag :cooldown-remaining-seconds="cooldownRemainingSeconds"/>
    </div>
    <div class="room-selector__list">
      <button
          v-for="(room, index) in rooms"
          :key="room.id"
          type="button"
          class="room-selector__button"
          :class="{
            'room-selector__button--active': isCurrentRoom(room),
            'room-selector__button--locked': !canJoin(room) && !isCurrentRoom(room),
          }"
          :style="roomStyle(room, index)"
          :disabled="!canJoin(room)"
          @click="emit('join', room.id)"
      >
        <span class="room-selector__surface" aria-hidden="true"></span>
        <span class="room-selector__top">
          <span class="room-selector__id">{{ String(room.id || '1').padStart(2, '0') }}</span>
          <span class="room-selector__badge">{{ roomStatusLabel(room) }}</span>
        </span>
        <span class="room-selector__name">{{ roomLabel(room) }}</span>
        <span class="room-selector__bossline">
          <span class="room-selector__boss">{{ roomBossName(room) }}</span>
        </span>
        <span class="room-selector__stats">
          <span>
            <small>在线</small>
            <strong>{{ roomOnlineText(room) }}</strong>
          </span>
          <span>
            <small>Boss平均血量</small>
            <strong>{{ roomAvgHpText(room) }}</strong>
          </span>
        </span>
        <span class="room-selector__action">{{ roomActionLabel(room) }}</span>
      </button>
    </div>
    <p v-if="rooms.length === 0" class="room-selector__empty">正在同步房间战线。</p>
    <p v-if="error" class="room-selector__error">{{ error }}</p>
  </section>
</template>

<style scoped>
.room-selector {
  display: grid;
  gap: 15px 6px; /* 行间距15px，列间距6px（此处列间距其实没有用，但保留无妨） */
  padding: 8px 10px;
  border: 1px solid rgba(132, 166, 214, 0.18);
  border-radius: 18px;
  background: radial-gradient(circle at top right, rgba(120, 216, 255, 0.1), transparent 36%),
  rgba(12, 21, 35, 0.78);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.34);
  backdrop-filter: blur(18px);
  /* 改为：第一行头部20px，第二行占满剩余空间 */
  grid-template-rows: 20px 1fr;
  /* 给整体一个最大高度（按需调整），防止无限撑高 */
  max-height: 600px; /* 也可用 height: 100% 继承父级高度 */

  overflow: hidden; /* 溢出交给内部列表处理 */
}

.room-selector__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  color: #8ca1c0;
}

.room-selector__title {
  display: flex;
  align-items: center;
  gap: 2px;
  min-width: 0;
}

.room-selector__head small,
.room-selector__head strong {
  display: inline;
  line-height: 1;
}

.room-selector__head small {
  font-size: 0.62rem;
  font-weight: 700;
  line-height: 1.1;
}

.room-selector__head strong {
  color: #eef4ff;
  font-family: 'Rajdhani', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  font-size: 0.92rem;
  line-height: 1;
}

.room-selector__list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-content: start; /* 保持网格项从顶部开始，不拉伸 */
  gap: 6px;
  padding-top: 6px;
  margin-top: -6px;

  /* 关键：行高跟随内容，不要按比例平分 */
  grid-auto-rows: auto; /* 或 min-content，默认就是 auto，可显式写出 */

  /* 让这个区域可滚动 */
  overflow-y: auto;

  /* 可选：保证滚动时头部不动 */
  min-height: 0; /* 防止网格子项默认最小高度撑开 */
}

.room-selector__button {
  display: grid;
  gap: 6px;
  min-height: 74px;
  padding: 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
  color: #eef4ff;
  background: rgba(8, 14, 24, 0.72);
  text-align: left;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  isolation: isolate;
  z-index: 0;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.05);
  transition: transform 180ms ease,
  border-color 180ms ease,
  box-shadow 180ms ease,
  opacity 180ms ease;
}

.room-selector__button > * {
  position: relative;
  z-index: 2;
}

.room-selector__surface {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
  z-index: 0;
  background: linear-gradient(180deg, rgba(var(--room-accent-rgb), 0.14), rgba(var(--room-accent-rgb), 0.05) 34%, rgba(255, 255, 255, 0.03) 100%),
  radial-gradient(circle at top right, rgba(var(--room-accent-rgb), 0.34), transparent 50%),
  radial-gradient(circle at bottom left, rgba(var(--room-accent-rgb), 0.2), transparent 64%);
  opacity: 0.5;
  filter: brightness(0.98) saturate(1.12);
  transform: scale(1);
  transform-origin: center center;
  animation: room-card-surface-breathe 4.8s ease-in-out infinite;
  animation-delay: var(--room-breathe-delay);
}

.room-selector__button::before,
.room-selector__button::after {
  content: '';
  position: absolute;
  pointer-events: none;
}

.room-selector__button::before {
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(115deg,
  rgba(var(--room-accent-rgb), 0.08) 0%,
  rgba(var(--room-accent-rgb), 0.32) 28%,
  rgba(255, 255, 255, 0.18) 50%,
  rgba(var(--room-accent-rgb), 0.28) 72%,
  rgba(var(--room-accent-rgb), 0.08) 100%);
  -webkit-mask: linear-gradient(#fff 0 0) content-box,
  linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask: linear-gradient(#fff 0 0) content-box,
  linear-gradient(#fff 0 0);
  mask-composite: exclude;
  filter: drop-shadow(0 0 8px rgba(var(--room-accent-rgb), 0.18));
  opacity: 0.2;
  animation: room-card-breathe 4.8s ease-in-out infinite;
  animation-delay: var(--room-breathe-delay);
  z-index: 1;
}

.room-selector__button::after {
  top: 0;
  left: 0;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: rgba(var(--room-accent-rgb), 0.98);
  box-shadow: -2px 0 4px rgba(var(--room-accent-rgb), 0.96),
  -22px 0 14px rgba(var(--room-accent-rgb), 0.42),
  -100px 0 32px rgba(var(--room-accent-rgb), 0.08);
  opacity: 0.96;
  transform-origin: center center;
  animation: room-card-orbit var(--room-orbit-duration) linear infinite;
  animation-delay: var(--room-orbit-delay);
  z-index: 999;
}

.room-selector__button:hover:not(:disabled) {
  transform: translateY(-2px);
  z-index: 2;
  border-color: rgba(var(--room-accent-rgb), 0.42);
  box-shadow: 0 10px 18px rgba(0, 0, 0, 0.26),
  inset 0 0 0 1px rgba(var(--room-accent-rgb), 0.08);
}

.room-selector__button:hover:not(:disabled)::before {
  opacity: 0.26;
}

.room-selector__button:hover:not(:disabled) .room-selector__surface {
  opacity: 0.58;
}

.room-selector__top {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 6px;
}

.room-selector__id {
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border: 1px solid rgba(var(--room-accent-rgb), 0.32);
  border-radius: 9px;
  color: rgb(var(--room-accent-rgb));
  background: rgba(var(--room-accent-rgb), 0.12);
  font-family: 'Orbitron', 'SFMono-Regular', monospace;
  font-size: 0.68rem;
  font-weight: 800;
}

.room-selector__badge {
  display: inline-flex;
  align-items: center;
  min-height: 20px;
  padding: 0 7px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  color: #dfeaff;
  background: rgba(255, 255, 255, 0.04);
  font-size: 0.6rem;
  font-weight: 700;
  white-space: nowrap;
}

.room-selector__name {
  color: #f8fbff;
  font-family: 'Rajdhani', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1;
}

.room-selector__bossline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 2.1em;
}

.room-selector__boss {
  flex: 1;
  min-width: 0;
  color: #8ca1c0;
  font-size: 0.7rem;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.room-selector__range {
  flex: none;
  padding: 1px 6px;
  border: 1px solid rgba(159, 198, 255, 0.25);
  border-radius: 999px;
  color: #9fc6ff;
  background: rgba(22, 37, 56, 0.6);
  font-size: 0.58rem;
  font-weight: 700;
  white-space: nowrap;
}

.room-selector__stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 5px;
}

.room-selector__stats > span {
  display: grid;
  gap: 2px;
  min-width: 0;
  padding: 6px 7px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 9px;
  background: rgba(6, 11, 19, 0.48);
}

.room-selector__stats small {
  color: #65809d;
  font-size: 0.58rem;
}

.room-selector__stats strong {
  min-width: 0;
  overflow: hidden;
  color: #eef4ff;
  font-family: 'Orbitron', 'SFMono-Regular', monospace;
  font-size: 0.7rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.room-selector__action {
  font-size: 0.66rem;
  color: #9fc6ff;
  font-weight: 700;
}

.room-selector__button--active {
  border-color: rgba(255, 212, 107, 0.5);
  box-shadow: 0 18px 44px rgba(255, 154, 85, 0.16),
  inset 0 0 0 1px rgba(255, 212, 107, 0.2);
}

.room-selector__button--active::before {
  opacity: 0.24;
}

.room-selector__button--active .room-selector__badge {
  color: #fff1c0;
  border-color: rgba(255, 212, 107, 0.22);
  background: rgba(255, 212, 107, 0.08);
}

.room-selector__button:disabled {
  cursor: not-allowed;
}

.room-selector__button--locked:not(.room-selector__button--active) {
  opacity: 0.62;
}

.room-selector__button--locked:not(.room-selector__button--active)::before {
  opacity: 0.1;
}

.room-selector__button--locked:not(.room-selector__button--active)::after {
  opacity: 0.34;
}

.room-selector__empty,
.room-selector__error {
  margin: 0;
  font-size: 0.86rem;
  line-height: 1.5;
}

.room-selector__empty {
  color: #8ca1c0;
}

.room-selector__error {
  color: #fecaca;
}

@keyframes room-card-breathe {
  0%,
  100% {
    opacity: 0.14;
    filter: drop-shadow(0 0 5px rgba(var(--room-accent-rgb), 0.12));
  }

  50% {
    opacity: 0.36;
    filter: drop-shadow(0 0 12px rgba(var(--room-accent-rgb), 0.24));
  }
}

@keyframes room-card-surface-breathe {
  0%,
  100% {
    opacity: 0.5;
    filter: brightness(0.98) saturate(1.12);
    transform: scale(1);
  }

  50% {
    opacity: 0.86;
    filter: brightness(1.08) saturate(1.32);
    transform: scale(1.018);
  }
}

@keyframes room-card-orbit {
  0% {
    top: 1px;
    left: 8px;
    opacity: 0;
    transform: rotate(0deg) scale(0.72);
  }

  6% {
    opacity: 0.92;
    transform: rotate(0deg) scale(1);
  }

  24% {
    top: 1px;
    left: calc(100% - 12px);
    opacity: 0.82;
    transform: rotate(0deg) scale(0.94);
  }

  49% {
    top: calc(100% - 12px);
    left: calc(100% - 12px);
    opacity: 0.88;
    transform: rotate(90deg) scale(1);
  }

  74% {
    top: calc(100% - 12px);
    left: 8px;
    opacity: 0.78;
    transform: rotate(180deg) scale(0.92);
  }

  94% {
    top: 1px;
    left: 8px;
    opacity: 0.86;
    transform: rotate(270deg) scale(0.96);
  }

  100% {
    top: 1px;
    left: 8px;
    opacity: 0;
    transform: rotate(360deg) scale(0.72);
  }
}

@media (max-width: 720px) {
  .room-selector {
    padding: 10px;
    border-radius: 16px;
  }

  .room-selector__head {
    justify-content: start;
  }

  .room-selector__list {
    grid-template-columns: 1fr;
  }
}
</style>
