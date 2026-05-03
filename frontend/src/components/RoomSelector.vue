<script setup>
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
})

const emit = defineEmits(['join'])

function roomLabel(room) {
  return `房间 ${room?.id || '1'}`
}

function roomMeta(room) {
  if (!room) return ''
  const countText = `${Math.max(0, Number(room.onlineCount || 0))} 人攻坚`
  if (room.currentBossName) {
    return `${room.currentBossName} · ${countText}`
  }
  return room.cycleEnabled ? `循环待命 · ${countText}` : '未开启循环'
}

function canJoin(room) {
  return props.loggedIn && !props.switching && room?.joinable !== false && room?.id !== props.currentRoomId
}
</script>

<template>
  <section class="room-selector" aria-label="房间选择">
    <div class="room-selector__head">
      <span>战线</span>
      <strong>{{ roomLabel({id: currentRoomId}) }}</strong>
    </div>
    <div class="room-selector__list">
      <button
          v-for="room in rooms"
          :key="room.id"
          type="button"
          class="room-selector__button"
          :class="{ 'room-selector__button--active': room.id === currentRoomId }"
          :disabled="!canJoin(room)"
          @click="emit('join', room.id)"
      >
        <span>{{ roomLabel(room) }}</span>
        <small>{{ roomMeta(room) }}</small>
      </button>
    </div>
    <p v-if="error" class="room-selector__error">{{ error }}</p>
  </section>
</template>

<style scoped>
.room-selector {
  display: grid;
  gap: 10px;
  padding: 12px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.72);
}

.room-selector__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #cbd5e1;
  font-size: 13px;
}

.room-selector__head strong {
  color: #f8fafc;
}

.room-selector__list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: 8px;
}

.room-selector__button {
  display: grid;
  gap: 4px;
  min-height: 58px;
  padding: 10px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 8px;
  color: #e2e8f0;
  background: rgba(30, 41, 59, 0.72);
  text-align: left;
  cursor: pointer;
}

.room-selector__button small {
  overflow: hidden;
  color: #94a3b8;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.room-selector__button--active {
  border-color: rgba(34, 211, 238, 0.72);
  background: rgba(8, 145, 178, 0.22);
}

.room-selector__button:disabled {
  cursor: not-allowed;
  opacity: 0.56;
}

.room-selector__error {
  margin: 0;
  color: #fecaca;
  font-size: 13px;
}
</style>
