<script setup>
import {computed} from 'vue'

const props = defineProps({
  cooldownRemainingSeconds: {
    type: Number,
    default: 0,
  },
})

const active = computed(() => Number(props.cooldownRemainingSeconds || 0) > 0)

const formattedCooldown = computed(() => {
  const totalSeconds = Math.max(0, Number(props.cooldownRemainingSeconds || 0))
  const minutes = Math.floor(totalSeconds / 60)
  const remainSeconds = totalSeconds % 60
  return `${String(minutes).padStart(2, '0')}:${String(remainSeconds).padStart(2, '0')}`
})
</script>

<template>
  <span
      class="room-switch-cooldown-tag"
      :class="{ 'room-switch-cooldown-tag--active': active }"
  >
    {{ active ? `切房冷却 ${formattedCooldown}` : '可进入战线' }}
  </span>
</template>

<style scoped>
.room-switch-cooldown-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 10px;
  border: 1px solid rgba(176, 132, 58, 0.42);
  border-radius: 999px;
  color: #c9a45a;
  background: linear-gradient(180deg, rgba(62, 43, 16, 0.92), rgba(24, 17, 8, 0.9));
  box-shadow: inset 0 1px 0 rgba(255, 220, 151, 0.12),
  0 6px 18px rgba(0, 0, 0, 0.22);
  font-size: 0.74rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-shadow: 0 1px 0 rgba(26, 18, 7, 0.85);
  white-space: nowrap;
}

.room-switch-cooldown-tag--active {
  border-color: rgba(255, 198, 116, 0.42);
  color: #ffe0a3;
  background: rgba(66, 36, 8, 0.48);
  box-shadow: inset 0 0 0 1px rgba(255, 214, 143, 0.08);
}
</style>
