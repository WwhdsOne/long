<script setup>
import RoomSwitchCooldownTag from './RoomSwitchCooldownTag.vue'
import { formatIntegerExact } from '../utils/formatNumber'

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
  return `房间 ${room?.id || '1'}`
}

function currentRoomSummary(roomId) {
  const normalized = String(roomId || '').trim()
  if (normalized === 'hall') return '当前大厅'
  if (normalized === '') return '当前大厅'
  return `当前 ${roomLabel({ id: normalized })}`
}

function roomAccent(room, index) {
  const idIndex = Number.parseInt(room?.id, 10)
  const offset = Number.isFinite(idIndex) ? idIndex - 1 : index
  return ROOM_ACCENTS[Math.max(0, offset) % ROOM_ACCENTS.length]
}

function roomStyle(room, index) {
  return {
    '--room-accent-rgb': roomAccent(room, index),
  }
}

function roomBossName(room) {
  if (!room) return '等待 Boss'
  else if (room.currentBossName === "" || room.currentBossName === undefined)return room.cycleEnabled ? 'Boss循环待命' : '未开启Boss循环'
  return '当前Boss : ' + room.currentBossName
}

function roomOnlineText(room) {
  return `${Math.max(0, Number(room?.onlineCount || 0))} 人`
}

function roomAvgHpText(room) {
  const value = formatIntegerExact(room?.currentBossAvgHp)
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
      <RoomSwitchCooldownTag :cooldown-remaining-seconds="cooldownRemainingSeconds" />
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
        <span class="room-selector__top">
          <span class="room-selector__id">{{ String(room.id || '1').padStart(2, '0') }}</span>
          <span class="room-selector__badge">{{ roomStatusLabel(room) }}</span>
        </span>
        <span class="room-selector__name">{{ roomLabel(room) }}</span>
        <span class="room-selector__boss">{{ roomBossName(room) }}</span>
        <span class="room-selector__stats">
          <span>
            <small>在线</small>
            <strong>{{ roomOnlineText(room) }}</strong>
          </span>
          <span>
            <small>均血</small>
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
  gap: 15px 6px;   /* 行间距15px，列间距6px（此处列间距其实没有用，但保留无妨） */
  padding: 8px 10px;
  border: 1px solid rgba(132, 166, 214, 0.18);
  border-radius: 18px;
  background:
    radial-gradient(circle at top right, rgba(120, 216, 255, 0.1), transparent 36%),
    rgba(12, 21, 35, 0.78);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.34);
  backdrop-filter: blur(18px);
  /* 改为：第一行头部20px，第二行占满剩余空间 */
  grid-template-rows: 20px 1fr;
  /* 给整体一个最大高度（按需调整），防止无限撑高 */
  max-height: 600px;       /* 也可用 height: 100% 继承父级高度 */

  overflow: hidden;        /* 溢出交给内部列表处理 */}

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
  align-content: start;          /* 保持网格项从顶部开始，不拉伸 */
  gap: 6px;
  padding-top: 6px;
  margin-top: -6px;

  /* 关键：行高跟随内容，不要按比例平分 */
  grid-auto-rows: auto;          /* 或 min-content，默认就是 auto，可显式写出 */

  /* 让这个区域可滚动 */
  overflow-y: auto;

  /* 可选：保证滚动时头部不动 */
  min-height: 0;                 /* 防止网格子项默认最小高度撑开 */
}

.room-selector__button {
  display: grid;
  gap: 6px;
  min-height: 74px;
  padding: 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
  color: #eef4ff;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.04), rgba(255, 255, 255, 0.02)),
    radial-gradient(circle at top right, rgba(var(--room-accent-rgb), 0.12), transparent 42%),
    rgba(8, 14, 24, 0.72);
  text-align: left;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  z-index: 0;
  transition:
    transform 180ms ease,
    border-color 180ms ease,
    box-shadow 180ms ease,
    opacity 180ms ease;
}

.room-selector__button::after {
  content: '';
  position: absolute;
  right: -26px;
  bottom: -38px;
  width: 76px;
  height: 76px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(var(--room-accent-rgb), 0.28), transparent 70%);
  pointer-events: none;
}

.room-selector__button:hover:not(:disabled) {
  transform: translateY(-2px);
  z-index: 2;
  border-color: rgba(var(--room-accent-rgb), 0.42);
  box-shadow:
    0 10px 18px rgba(0, 0, 0, 0.26),
    inset 0 0 0 1px rgba(var(--room-accent-rgb), 0.08);
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

.room-selector__boss {
  min-height: 2.1em;
  color: #8ca1c0;
  font-size: 0.7rem;
  line-height: 1.3;
  overflow-wrap: anywhere;
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
  box-shadow:
    0 18px 44px rgba(255, 154, 85, 0.16),
    inset 0 0 0 1px rgba(255, 212, 107, 0.2);
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
