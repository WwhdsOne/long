<script setup>
import {reactive, watch} from 'vue'

const props = defineProps({
  rooms: {
    type: Array,
    default: () => [],
  },
  saving: {
    type: Boolean,
    default: false,
  },
  saveRoomDisplayName: {
    type: Function,
    required: true,
  },
})

const draftNames = reactive({})

watch(
    () => props.rooms,
    (nextRooms) => {
      for (const key of Object.keys(draftNames)) {
        delete draftNames[key]
      }
      for (const room of nextRooms) {
        draftNames[String(room?.id || '')] = String(room?.displayName || '')
      }
    },
    {immediate: true, deep: true},
)
</script>

<template>
  <section class="admin-stack">
    <div class="social-card__head">
      <p class="vote-stage__eyebrow">房间管理</p>
      <strong>房间显示名</strong>
    </div>
    <p class="social-card__copy">未命名时默认显示 `房间 N`，这里只改显示名，不改真实房间 ID。</p>

    <div class="admin-room-list">
      <article v-for="room in rooms" :key="room.id" class="admin-room-card">
        <div class="admin-room-card__head">
          <strong>房间 ID {{ room.id }}</strong>
          <span>{{ room.displayName || `房间 ${room.id}` }}</span>
        </div>
        <div class="admin-room-card__meta">
          <span>当前 Boss：{{ room.currentBossName || '暂无 Boss' }}</span>
          <span>循环状态：{{ room.cycleEnabled ? '开启' : '关闭' }}</span>
        </div>
        <div class="admin-room-card__body">
          <input
              v-model="draftNames[room.id]"
              class="nickname-form__input"
              type="text"
              placeholder="输入房间显示名"
          />
          <button
              class="nickname-form__submit"
              type="button"
              :disabled="saving"
              @click="saveRoomDisplayName(room.id, draftNames[room.id])"
          >
            {{ saving ? '保存中...' : '保存房间名' }}
          </button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.admin-stack {
  display: grid;
  gap: 16px;
}

.admin-room-list {
  display: grid;
  gap: 12px;
}

.admin-room-card {
  display: grid;
  gap: 10px;
  padding: 14px;
  border: 1px solid rgba(132, 166, 214, 0.2);
  border-radius: 14px;
  background: rgba(10, 18, 31, 0.62);
}

.admin-room-card__head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.admin-room-card__body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
}

.admin-room-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  color: #8ca1c0;
  font-size: 0.92rem;
}
</style>
