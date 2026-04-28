<template>
  <canvas
    ref="canvasRef"
    class="pixel-shatter"
    :width="GRID * PX"
    :height="GRID * PX"
    aria-hidden="true"
  />
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const GRID = 15
const PX = 6

// 银闪闪金属盾牌 — 对角渐变左上亮→右下暗
// 0=透明 1=暗钢边 2=深钢十字 3=高光边 4=亮银 5=浅银 6=中银 7=暗银 8=闪光白
const PALETTE = [
  '#00000000', // 0
  '#1a1a1e',   // 1 暗钢边
  '#3a3a44',   // 2 深钢十字
  '#e8e8f4',   // 3 高光边框
  '#e0e0ec',   // 4 亮银 (左上)
  '#c0c0d0',   // 5 浅银 (右上)
  '#9098a8',   // 6 中银 (左下)
  '#787884',   // 7 暗银 (右下)
  '#e8e8f4',   // 8 (同高光色)
]

const SHIELD = [
  [0,0,0,0,0,0,0,1,0,0,0,0,0,0,0],
  [0,0,0,0,0,0,1,3,1,0,0,0,0,0,0],
  [0,0,0,0,0,1,4,2,5,1,0,0,0,0,0],
  [0,0,0,0,1,4,4,2,5,5,1,0,0,0,0],
  [0,0,0,1,4,4,4,2,5,5,5,1,0,0,0],
  [0,0,1,4,4,4,4,2,5,5,5,5,1,0,0],
  [0,1,4,4,4,4,4,2,5,5,5,5,5,1,0],
  [1,3,2,2,2,2,2,2,2,2,2,2,2,3,1],
  [0,1,6,6,6,6,6,2,7,7,7,7,7,1,0],
  [0,0,1,6,6,6,6,2,7,7,7,7,1,0,0],
  [0,0,0,1,6,6,6,2,7,7,7,1,0,0,0],
  [0,0,0,0,1,6,6,2,7,7,1,0,0,0,0],
  [0,0,0,0,0,1,6,2,7,1,0,0,0,0,0],
  [0,0,0,0,0,0,1,3,1,0,0,0,0,0,0],
  [0,0,0,0,0,0,0,1,0,0,0,0,0,0,0],
]

const canvasRef = ref(null)
let raf = 0
let particles = []
let phase = 'idle'
let phaseTimer = 0
function createParticles() {
  const list = []
  for (let y = 0; y < GRID; y++) {
    for (let x = 0; x < GRID; x++) {
      const ci = SHIELD[y][x]
      if (ci === 0) continue
      list.push({
        x: x * PX + PX / 2,
        y: y * PX + PX / 2,
        vx: 0,
        vy: 0,
        opacity: 1,
        color: PALETTE[ci],
      })
    }
  }
  return list
}

function explode() {
  const W = GRID * PX
  const H = GRID * PX
  const cx = W / 2
  const cy = H / 2
  for (const p of particles) {
    const dx = p.x - cx
    const dy = p.y - cy
    const dist = Math.sqrt(dx * dx + dy * dy) || 1
    const speed = 1.2 + Math.random() * 2.8
    p.vx = (dx / dist) * speed + (Math.random() - 0.5) * 1.5
    p.vy = (dy / dist) * speed + (Math.random() - 0.5) * 1.5
  }
}

function draw(ctx) {
  ctx.clearRect(0, 0, GRID * PX, GRID * PX)
  for (const p of particles) {
    if (p.opacity <= 0) continue
    ctx.globalAlpha = p.opacity
    ctx.fillStyle = p.color
    ctx.fillRect(Math.round(p.x - PX / 2), Math.round(p.y - PX / 2), PX, PX)
  }
  ctx.globalAlpha = 1
}

function frame() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  const W = GRID * PX
  const H = GRID * PX

  phaseTimer += 16

  if (phase === 'idle' && phaseTimer > 500) {
    phase = 'explode'
    phaseTimer = 0
    explode()
  }

  if (phase === 'explode') {
    let alive = 0
    for (const p of particles) {
      if (p.opacity <= 0) continue
      p.x += p.vx
      p.y += p.vy
      p.vx *= 0.985
      p.vy *= 0.985
      if (p.x < -6 || p.x > W + 6 || p.y < -6 || p.y > H + 6) {
        p.opacity = Math.max(0, p.opacity - 0.08)
      } else if (phaseTimer > 300) {
        p.opacity = Math.max(0, p.opacity - 0.015)
      }
      if (p.opacity > 0) alive++
    }
    if (alive === 0) {
      phase = 'done'
    }
  }

  draw(ctx)

  if (phase !== 'done') {
    raf = requestAnimationFrame(frame)
  }
}

onMounted(() => {
  particles = createParticles()
  const canvas = canvasRef.value
  if (canvas) {
    draw(canvas.getContext('2d'))
  }
  raf = requestAnimationFrame(frame)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(raf)
})
</script>

<style scoped>
.pixel-shatter {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 3;
  image-rendering: pixelated;
}
</style>
