<template>
  <canvas
      ref="canvasRef"
      class="pixel-canvas"
      :width="canvasWidthValue"
      :height="canvasHeightValue"
  />
</template>

<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue'

const props = defineProps({
  effect: {type: String, default: ''},
  size: {type: Number, default: 90},
  loop: {type: Boolean, default: true},
  entries: {type: Array, default: null},
  canvasWidth: {type: Number, default: 0},
  canvasHeight: {type: Number, default: 0},
})

const canvasRef = ref(null)
let raf = 0
let state = null
let renderer = null
const managerMode = computed(() => Array.isArray(props.entries))
const canvasWidthValue = computed(() => managerMode.value ? Math.max(1, Math.round(props.canvasWidth || props.size || 1)) : props.size)
const canvasHeightValue = computed(() => managerMode.value ? Math.max(1, Math.round(props.canvasHeight || props.size || 1)) : props.size)

// ======= 盾牌图案 (collapse_trigger 用) =======
const SHIELD = [
  [0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0],
  [0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0],
  [0, 0, 0, 0, 0, 1, 4, 2, 4, 1, 0, 0, 0, 0, 0],
  [0, 0, 0, 0, 1, 4, 4, 2, 4, 4, 1, 0, 0, 0, 0],
  [0, 0, 0, 1, 4, 4, 4, 2, 4, 4, 4, 1, 0, 0, 0],
  [0, 0, 1, 4, 4, 4, 4, 2, 4, 4, 4, 4, 1, 0, 0],
  [0, 1, 4, 4, 4, 4, 4, 2, 4, 4, 4, 4, 4, 1, 0],
  [1, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3, 1],
  [0, 1, 4, 4, 4, 4, 4, 2, 4, 4, 4, 4, 4, 1, 0],
  [0, 0, 1, 4, 4, 4, 4, 2, 4, 4, 4, 4, 1, 0, 0],
  [0, 0, 0, 1, 4, 4, 4, 2, 4, 4, 4, 1, 0, 0, 0],
  [0, 0, 0, 0, 1, 4, 4, 2, 4, 4, 1, 0, 0, 0, 0],
  [0, 0, 0, 0, 0, 1, 4, 2, 4, 1, 0, 0, 0, 0, 0],
  [0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0],
  [0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0],
]
const SHIELD_COLORS = ['#0000', '#1a1a1e', '#3a3a44', '#e8e8f4', '#e0e0ec']
const GRID = 15

function shieldParticles(size) {
  const px = size / GRID
  const parts = []
  for (let y = 0; y < GRID; y++) {
    for (let x = 0; x < GRID; x++) {
      const ci = SHIELD[y][x]
      if (ci === 0) continue
      parts.push({
        x: x * px + px / 2, y: y * px + px / 2,
        vx: 0, vy: 0, life: 1, maxLife: 1,
        size: px, color: SHIELD_COLORS[ci],
      })
    }
  }
  return parts
}

function makeParticle(x, y, vx, vy, life, size, color) {
  return {x, y, vx, vy, life, maxLife: life, size, color}
}

function drawPixelRing(ctx, cx, cy, radius, thickness, color) {
  const outer = radius + thickness / 2
  const inner = Math.max(0, radius - thickness / 2)
  const minX = Math.floor(cx - outer - 1)
  const maxX = Math.ceil(cx + outer + 1)
  const minY = Math.floor(cy - outer - 1)
  const maxY = Math.ceil(cy + outer + 1)
  for (let y = minY; y <= maxY; y++) {
    for (let x = minX; x <= maxX; x++) {
      const dx = x - cx
      const dy = y - cy
      const dist = Math.hypot(dx, dy)
      if (dist >= inner && dist <= outer) {
        ctx.fillStyle = color
        ctx.fillRect(Math.round(x), Math.round(y), 1, 1)
      }
    }
  }
}

// ======= 各特效渲染器 =======
// 每个返回 { init, update, draw, isDone }

function rnd(n) {
  return (Math.random() - 0.5) * n
}

// ---- 1. storm_combo: 绿色像素斩击（加粗像素群） ----
const stormComboRenderer = {
  init(_size) {
    return {phase: 'idle', timer: 0, parts: [], trails: [], sparks: []}
  },
  update(s, size) {
    s.timer += 16
    if (s.phase === 'idle' && s.timer > 600) {
      s.phase = 'slash';
      s.timer = 0
      // 粗斩击：3条平行像素线，每条 16 个粒子
      for (let row = -3; row <= 3; row += 3) {
        for (let i = 0; i < 16; i++) {
          const t = i / 15
          const baseX = 4 + t * (size - 8)
          const baseY = size - 4 - t * (size - 8)
          const greenShades = ['#22c55e', '#4ade80', '#16a34a', '#86efac', '#15803d', '#166534']
          s.parts.push(makeParticle(
              baseX + rnd(2), baseY + row + rnd(2),
              2.5 + rnd(1.8), -2.5 + rnd(1.8),
              0.5 + Math.random() * 0.5, 3 + Math.floor(Math.random() * 3),
              greenShades[Math.floor(Math.random() * greenShades.length)]
          ))
        }
      }
      // 火花
      for (let i = 0; i < 20; i++) {
        s.sparks.push(makeParticle(size * 0.1 + rnd(10), size * 0.9 + rnd(10), 1 + rnd(3), -1 + rnd(3), 0.3 + Math.random() * 0.3, 2, i % 2 ? '#fef08a' : '#86efac'))
      }
    }
    if (s.phase === 'slash') {
      for (const p of s.parts) {
        p.x += p.vx;
        p.y += p.vy
        p.life -= 0.014
        if (p.life > 0.2 && Math.random() < 0.5) s.trails.push(makeParticle(p.x, p.y, rnd(0.3), rnd(0.3), 0.3, 2, p.color))
      }
      for (const sp of s.sparks) {
        sp.x += sp.vx;
        sp.y += sp.vy;
        sp.life -= 0.02
      }
      for (const t of s.trails) {
        t.life -= 0.04
      }
      s.trails = s.trails.filter(t => t.life > 0)
      s.sparks = s.sparks.filter(sp => sp.life > 0)
      if (s.parts.every(p => p.life <= 0) && s.trails.length === 0 && s.sparks.length === 0) {
        s.phase = 'done';
        s.timer = 0
      }
    }
    if (s.phase === 'done') return s
    return s
  },
  draw(ctx, s, _size) {
    for (const t of s.trails) {
      ctx.globalAlpha = t.life;
      ctx.fillStyle = t.color;
      ctx.fillRect(Math.round(t.x), Math.round(t.y), t.size, t.size)
    }
    for (const p of s.parts) {
      if (p.life <= 0) continue;
      ctx.globalAlpha = p.life;
      ctx.fillStyle = p.color;
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const sp of s.sparks) {
      if (sp.life <= 0) continue;
      ctx.globalAlpha = sp.life;
      ctx.fillStyle = sp.color;
      ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 900
  },
}

// ---- 2. auto_strike: T形锤，锤柄底端在格子左侧边缘为圆心，绕90度砸向中心 ----
const autoStrikeRenderer = {
  init(_size) {
    return {
      phase: 'idle',
      timer: 0,
      swingAngle: -Math.PI / 2,
      fragments: [],
      dent: [],
      flash: [],
    }
  },

  update(s, size) {
    s.timer += 16

    const handleLen = size * 0.5
    const impactX = size / 2
    const impactY = size / 2
    const pivotX = impactX - handleLen
    const pivotY = impactY

    if (s.phase === 'idle' && s.timer > 500) {
      s.phase = 'swing'
      s.timer = 0
    }

    if (s.phase === 'swing') {
      const progress = Math.min(1, s.timer / 450)
      const eased = progress < 0.5
          ? 2 * progress * progress
          : 1 - Math.pow(-2 * progress + 2, 2) / 2

      s.swingAngle = -Math.PI / 2 + eased * (Math.PI / 2)

      if (progress >= 1) {
        s.phase = 'impact'
        s.timer = 0

        const hitX = pivotX + Math.cos(s.swingAngle) * handleLen
        const hitY = pivotY + Math.sin(s.swingAngle) * handleLen

        for (let i = 0; i < 24; i++) {
          const ang = Math.random() * Math.PI * 2
          const spd = 1.5 + Math.random() * 3.5
          s.fragments.push(
              makeParticle(
                  hitX,
                  hitY,
                  Math.cos(ang) * spd,
                  Math.sin(ang) * spd,
                  1,
                  3 + Math.floor(Math.random() * 3),
                  ['#9ca3af', '#6b7280', '#d1d5db', '#4b5563'][Math.floor(Math.random() * 4)],
              ),
          )
        }

        for (let i = 0; i < 10; i++) {
          s.flash.push(makeParticle(hitX + rnd(6), hitY + rnd(6), rnd(1), rnd(1), 0.4, 2, '#fef3c7'))
        }

        for (let i = 0; i < 10; i++) {
          s.dent.push(makeParticle(hitX + rnd(18), hitY + rnd(14), 0, 0, 2.5, 3, '#374151'))
        }
      }
    }

    if (s.phase === 'impact') {
      for (const f of s.fragments) {
        f.x += f.vx
        f.y += f.vy
        f.vx *= 0.93
        f.vy *= 0.93
        f.life -= 0.022
      }

      for (const fl of s.flash) {
        fl.life -= 0.04
      }

      for (const d of s.dent) {
        d.life -= 0.01
      }

      if (
          s.fragments.every(f => f.life <= 0)
          && s.flash.every(fl => fl.life <= 0)
          && s.timer > 1400
      ) {
        s.phase = 'done'
        s.timer = 0
      }
    }

    if (s.phase === 'done') return s

    return s
  },

  draw(ctx, s, size) {
    const handleLen = size * 0.5
    const impactX = size / 2
    const impactY = size / 2
    const pivotX = impactX - handleLen
    const pivotY = impactY

    if (s.phase === 'idle' || s.phase === 'swing') {
      const ang = s.swingAngle

      const headX = pivotX + Math.cos(ang) * handleLen
      const headY = pivotY + Math.sin(ang) * handleLen

      // 加粗锤柄
      const handleBase = handleLen * 0.15
      const steps = Math.ceil((handleLen - handleBase) / 4)

      for (let i = 0; i <= steps; i++) {
        const t = i / steps
        const px = Math.round(pivotX + (headX - pivotX) * t)
        const py = Math.round(pivotY + (headY - pivotY) * t)

        ctx.fillStyle = '#6b7280'
        ctx.fillRect(px - 2, py - 2, 7, 6)

        ctx.fillStyle = '#374151'
        ctx.fillRect(px + 4, py - 2, 3, 6)
      }

      // 放大 T 形锤头
      const perpX = -Math.sin(ang)
      const perpY = Math.cos(ang)

      for (let s = -7; s <= 7; s++) {
        for (let t = -3; t <= 4; t++) {
          const px = Math.round(
              headX + perpX * s * 4 + Math.cos(ang) * t * 4,
          )
          const py = Math.round(
              headY + perpY * s * 4 + Math.sin(ang) * t * 4,
          )

          // 中心连接处少挖一点，避免锤头显小
          if (Math.abs(s) <= 1 && t < 0) continue

          const dark = Math.abs(s) >= 6 || Math.abs(t) >= 3

          ctx.fillStyle = dark
              ? '#4b5563'
              : Math.abs(s) <= 2
                  ? '#d1d5db'
                  : '#9ca3af'

          ctx.fillRect(px - 1, py - 1, 6, 6)
        }
      }
    }

    for (const f of s.fragments) {
      if (f.life <= 0) continue
      ctx.globalAlpha = f.life
      ctx.fillStyle = f.color
      ctx.fillRect(Math.round(f.x), Math.round(f.y), f.size, f.size)
    }

    for (const fl of s.flash) {
      if (fl.life <= 0) continue
      ctx.globalAlpha = fl.life
      ctx.fillStyle = fl.color
      ctx.fillRect(Math.round(fl.x), Math.round(fl.y), fl.size, fl.size)
    }

    for (const d of s.dent) {
      if (d.life <= 0) continue
      ctx.globalAlpha = Math.min(1, d.life)
      ctx.fillStyle = d.color
      ctx.fillRect(Math.round(d.x), Math.round(d.y), d.size, d.size)
    }

    ctx.globalAlpha = 1
  },

  isDone(s) {
    return s.phase === 'done' && s.timer > 800
  },
}

// ---- 3. bleed: 血迹扩散 ----
const bleedRenderer = {
  init(size) {
    const cx = size / 2, cy = size / 2
    const parts = []
    for (let i = 0; i < 24; i++) {
      const ang = Math.random() * Math.PI * 2, dist = Math.random() * 20
      parts.push(makeParticle(cx + Math.cos(ang) * dist, cy + Math.sin(ang) * dist, Math.cos(ang) * 0.24 + rnd(0.2), Math.sin(ang) * 0.24 + rnd(0.2), 0.5 + Math.random() * 0.25, 2 + Math.floor(Math.random() * 3), ['#7f1d1d', '#991b1b', '#450a0a', '#b91c1c'][Math.floor(Math.random() * 4)]))
    }
    return {phase: 'spread', timer: 0, parts}
  },
  update(s, _size) {
    s.timer += 16
    for (const p of s.parts) {
      p.x += p.vx;
      p.y += p.vy;
      p.vx *= 0.94;
      p.vy *= 0.94;
      p.life -= 0.006
    }
    if (s.timer > 2200) s.phase = 'done'
    return s
  },
  draw(ctx, s, _size) {
    for (const p of s.parts) {
      if (p.life <= 0) continue;
      ctx.globalAlpha = Math.min(1, p.life);
      ctx.fillStyle = p.color;
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 2200
  },
}

// ---- 5. final_cut: 贯穿全战斗区的超长对角终结斩 ----
const finalCutRenderer = {
  init(_size) {
    return {phase: 'idle', timer: 0, slashPixels: [], shockwave: [], sparks: [], screenFlash: 0}
  },
  update(s, size) {
    s.timer += 16
    if (s.phase === 'idle' && s.timer > 440) {
      s.phase = 'anticipate';
      s.timer = 0
    }
    if (s.phase === 'anticipate' && s.timer > 90) {
      s.phase = 'slash';
      s.timer = 0;
      s.screenFlash = 1
      const startX = -size * 0.32
      const startY = -size * 0.32
      const endX = size * 1.32
      const endY = size * 1.32
      const dx = endX - startX
      const dy = endY - startY
      const lineLen = Math.sqrt(dx * dx + dy * dy) || 1
      const normalX = -dy / lineLen
      const normalY = dx / lineLen
      for (let i = 0; i < 84; i++) {
        const t = i / 83
        const sx = startX + dx * t
        const sy = startY + dy * t
        for (let w = -7; w <= 7; w++) {
          const dist = Math.abs(w)
          const intensity = dist === 0 ? 1 : dist <= 2 ? 0.92 : dist <= 4 ? 0.68 : dist <= 6 ? 0.44 : 0.24
          const color = dist === 0 ? '#fff' : dist <= 2 ? '#fee2e2' : dist <= 4 ? '#f87171' : dist <= 6 ? '#dc2626' : '#7f1d1d'
          s.slashPixels.push(makeParticle(
              sx + normalX * w * 4.8 + rnd(1.8),
              sy + normalY * w * 4.8 + rnd(1.8),
              0,
              0,
              0.5 + intensity * 0.5,
              dist <= 1 ? 7 : dist <= 4 ? 6 : 4,
              color,
          ))
        }
        if (i % 2 === 0) {
          for (let w = -14; w <= 14; w += 3) {
            if (Math.abs(w) <= 8) continue
            s.slashPixels.push(makeParticle(
                sx + normalX * w * 3.8 + rnd(2.4),
                sy + normalY * w * 3.8 + rnd(2.4),
                rnd(1.4),
                rnd(1.4),
                0.36 + Math.random() * 0.34,
                3,
                ['#7f1d1d', '#991b1b', '#b91c1c', '#ef4444'][Math.floor(Math.random() * 4)],
            ))
          }
        }
      }
      for (let i = 0; i < 72; i++) {
        const t = i / 71
        const cx = startX + dx * t + rnd(size * 0.07)
        const cy = startY + dy * t + rnd(size * 0.07)
        const spd = 1 + Math.random() * 3.6
        const ang = Math.atan2(dy, dx) + (Math.random() < 0.5 ? -1 : 1) * (Math.PI / 2) * (0.15 + Math.random() * 0.35)
        s.shockwave.push(makeParticle(cx, cy, Math.cos(ang) * spd, Math.sin(ang) * spd, 0.7 + Math.random() * 0.22, 5 + Math.floor(Math.random() * 3), i % 2 ? '#fca5a5' : '#e11d48'))
      }
      for (let i = 0; i < 64; i++) {
        const t = i / 63
        const cx = startX + dx * t + rnd(size * 0.1)
        const cy = startY + dy * t + rnd(size * 0.1)
        s.sparks.push(makeParticle(cx, cy, rnd(4.2), rnd(4.2), 0.34 + Math.random() * 0.28, 4, ['#fef08a', '#fff', '#fca5a5', '#fde68a'][Math.floor(Math.random() * 4)]))
      }
    }
    if (s.phase === 'slash') {
      s.screenFlash = Math.max(0, s.screenFlash - 0.06)
      for (const p of s.slashPixels) p.life -= 0.018
      for (const sw of s.shockwave) {
        sw.x += sw.vx;
        sw.y += sw.vy;
        sw.vx *= 0.95;
        sw.vy *= 0.95;
        sw.life -= 0.022
      }
      for (const sp of s.sparks) {
        sp.x += sp.vx;
        sp.y += sp.vy;
        sp.vx *= 0.93;
        sp.vy *= 0.93;
        sp.life -= 0.026
      }
      if (s.slashPixels.every(p => p.life <= 0) && s.shockwave.every(sw => sw.life <= 0) && s.sparks.every(sp => sp.life <= 0)) {
        s.phase = 'done';
        s.timer = 0
      }
    }
    if (s.phase === 'done') return s
    return s
  },
  draw(ctx, s, size) {
    // 屏幕闪白
    if (s.screenFlash > 0) {
      ctx.globalAlpha = s.screenFlash * 0.15;
      ctx.fillStyle = '#fff';
      ctx.fillRect(0, 0, size, size);
      ctx.globalAlpha = 1
    }
    for (const p of s.slashPixels) {
      if (p.life <= 0) continue;
      ctx.globalAlpha = p.life;
      ctx.fillStyle = p.color;
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const sw of s.shockwave) {
      if (sw.life <= 0) continue;
      ctx.globalAlpha = sw.life;
      ctx.fillStyle = sw.color;
      ctx.fillRect(Math.round(sw.x), Math.round(sw.y), sw.size, sw.size)
    }
    for (const sp of s.sparks) {
      if (sp.life <= 0) continue;
      ctx.globalAlpha = sp.life;
      ctx.fillStyle = sp.color;
      ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 1400
  },
}

// ---- 6. collapse_trigger: 盾牌爆裂（方案9，删除裂纹残留） ----
const collapseTriggerTickMs = 20
const collapseTriggerSpeedScale = 1.25
const collapseTriggerSpreadRatio = 0.85

const collapseTriggerRenderer = {
  init(size) {
    return {phase: 'idle', timer: 0, parts: shieldParticles(size)}
  },
  update(s, size) {
    s.timer += collapseTriggerTickMs
    const cx = size / 2, cy = size / 2
    if (s.phase === 'idle' && s.timer > 800) {
      s.phase = 'explode';
      s.timer = 0
      for (const p of s.parts) {
        const dx = p.x - cx, dy = p.y - cy
        const dist = Math.sqrt(dx * dx + dy * dy) || 1
        const speed = (0.6 + Math.random() * 2) * collapseTriggerSpeedScale
        p.vx = (dx / dist) * speed + rnd(1.2)
        p.vy = (dy / dist) * speed + rnd(1.2)
      }
    }
    if (s.phase === 'explode') {
      const spreadRadius = size * collapseTriggerSpreadRatio
      let alive = 0
      for (const p of s.parts) {
        if (p.life <= 0) continue
        p.x += p.vx;
        p.y += p.vy;
        p.vx *= 0.992;
        p.vy *= 0.992
        const distFromCenter = Math.hypot(p.x - cx, p.y - cy)
        if (p.x < -8 || p.x > size + 8 || p.y < -8 || p.y > size + 8) p.life -= 0.08
        else if (distFromCenter >= spreadRadius) p.life -= 0.02
        else if (s.timer > 520) p.life -= 0.006
        if (p.life > 0) alive++
      }
      if (alive === 0) {
        s.phase = 'done';
        s.timer = 0
      }
    }
    if (s.phase === 'done') return s
    return s
  },
  draw(ctx, s, _size) {
    for (const p of s.parts) {
      if (p.life <= 0) continue;
      ctx.globalAlpha = p.life;
      ctx.fillStyle = p.color;
      ctx.fillRect(Math.round(p.x), Math.round(p.y), Math.round(p.size), Math.round(p.size))
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 1000
  },
}

// ---- 7. judgment_day: 圣洁黄金十字裁决（像素十字斩） ----
const judgmentDayRenderer = {
  init(size) {
    const cx = size / 2
    const cell = size / 5
    const barHw = Math.round(cell * 0.5)   // 臂宽 = 一整格
    const spacing = cell * 0.055            // 采样间距（密度翻 4 倍）

    // 预生成十字所有像素块
    const px = []
    // 横臂：覆盖中心整行
    for (let x = 0; x < size; x += spacing) {
      const dx = Math.abs(x - cx)
      for (let wy = -barHw; wy <= barHw; wy += spacing) {
        const y = cx + wy
        if (y < 0 || y >= size) continue
        const d = Math.max(dx, Math.abs(wy) * 1.6)
        px.push({bx: x, by: y, d})
      }
    }
    // 竖臂：覆盖中心整列（跳过已覆盖的中心块）
    for (let y = 0; y < size; y += spacing) {
      const dy = Math.abs(y - cx)
      for (let wx = -barHw; wx <= barHw; wx += spacing) {
        const x = cx + wx
        if (x < 0 || x >= size) continue
        if (Math.abs(x - cx) < barHw - spacing && Math.abs(y - cx) < barHw - spacing) continue
        const d = Math.max(Math.abs(wx) * 1.6, dy)
        px.push({bx: x, by: y, d})
      }
    }
    return {
      phase: 'expand',
      timer: 0,
      flashA: 0.30,
      maxD: 0,
      px,
      shardPx: [],
      edgeSparks: [],
      holdTimer: 0,
    }
  },
  update(s, size) {
    s.timer += 16
    const cx = size / 2

    if (s.phase === 'expand') {
      const elapsed = s.timer
      s.maxD = size * 0.5 * Math.min(1, elapsed / 280)
      s.flashA = Math.max(0, 0.30 * (1 - elapsed / 280))

      if (elapsed >= 280) {
        s.phase = 'hold'
        s.timer = 0
        s.maxD = size * 0.5
        s.flashA = 0
        // 四端正交方向的冲击火花
        const ends = [[0, cx], [size - 1, cx], [cx, 0], [cx, size - 1]]
        for (const [ex, ey] of ends) {
          const baseAng = Math.atan2(ey - cx, ex - cx)
          for (let i = 0; i < 12; i++) {
            const ang = baseAng + (Math.random() - 0.5) * 0.9
            const spd = 1.5 + Math.random() * 2.8
            s.edgeSparks.push(makeParticle(ex, ey, Math.cos(ang) * spd, Math.sin(ang) * spd, 0.28 + Math.random() * 0.45, 2 + Math.floor(Math.random() * 3), i % 3 === 0 ? '#fef08a' : i % 3 === 1 ? '#fcd34d' : '#f59e0b'))
          }
        }
      }
    }

    if (s.phase === 'hold') {
      s.holdTimer = s.timer
      for (const sp of s.edgeSparks) {
        sp.x += sp.vx;
        sp.y += sp.vy;
        sp.vx *= 0.93;
        sp.vy *= 0.93;
        sp.life -= 0.017
      }
      if (s.timer > 2200) {
        s.phase = 'shatter';
        s.timer = 0;
        s.flashA = 0.20
      }
    }

    if (s.phase === 'shatter') {
      // 首帧：将所有可见十字像素转为碎片粒子
      if (s.shardPx.length === 0) {
        for (const p of s.px) {
          if (p.d > s.maxD + 1) continue
          const ang = Math.atan2(p.by - cx, p.bx - cx) + (Math.random() - 0.5) * 1.6
          const spd = 0.7 + Math.random() * 3.0
          let color
          const dNorm = p.d / (size * 0.5)
          if (p.d < 3) color = '#ffffff'
          else if (dNorm < 0.18) color = '#fef08a'
          else if (dNorm < 0.40) color = '#fcd34d'
          else if (dNorm < 0.70) color = '#f59e0b'
          else color = '#d97706'
          s.shardPx.push(makeParticle(p.bx, p.by, Math.cos(ang) * spd, Math.sin(ang) * spd, 0.35 + Math.random() * 0.5, p.d < 3 ? 10 : p.d < size * 0.15 ? 8 : 6, color))
        }
        s.px = []
        // 中心爆发额外火花
        for (let i = 0; i < 20; i++) {
          const ang = Math.random() * Math.PI * 2
          const spd = 1.2 + Math.random() * 3.5
          s.shardPx.push(makeParticle(cx, cx, Math.cos(ang) * spd, Math.sin(ang) * spd, 0.3 + Math.random() * 0.35, 6 + Math.floor(Math.random() * 4), i % 3 === 0 ? '#ffffff' : i % 3 === 1 ? '#fef08a' : '#fcd34d'))
        }
      }
      // 碎片飞行 + 衰减
      for (const sh of s.shardPx) {
        sh.x += sh.vx;
        sh.y += sh.vy;
        sh.vx *= 0.96;
        sh.vy *= 0.96;
        sh.life -= 0.016
      }
      s.flashA = Math.max(0, s.flashA - 0.008)
      if (s.shardPx.every(sh => sh.life <= 0)) {
        s.phase = 'done';
        s.timer = 0
      }
    }

    return s
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 600
  },
  draw(ctx, s, size) {
    const cx = size / 2

    if (s.flashA > 0) {
      ctx.fillStyle = `rgba(253,224,71,${s.flashA.toFixed(3)})`
      ctx.fillRect(0, 0, size, size)
    }

    // 十字像素（展开/保持阶段）
    if (s.phase !== 'shatter') {
      for (const p of s.px) {
        if (p.d > s.maxD + 1) continue
        const fadeIn = p.d < s.maxD - 6 ? 1 : Math.max(0, (s.maxD - p.d) / 6)
        if (fadeIn <= 0) continue

        let color, sz
        const dNorm = p.d / (size * 0.5)
        if (p.d < 3) {
          color = '#ffffff';
          sz = 10
        } else if (dNorm < 0.18) {
          color = '#fef08a';
          sz = 8
        } else if (dNorm < 0.40) {
          color = '#fcd34d';
          sz = 6
        } else if (dNorm < 0.70) {
          color = '#f59e0b';
          sz = 4
        } else {
          color = '#d97706';
          sz = 4
        }

        // 中心交叉区域额外白色高亮
        const cell = size / 5
        const hw = cell * 0.5
        if (Math.abs(p.bx - cx) < hw && Math.abs(p.by - cx) < hw && s.phase !== 'expand') {
          color = p.d < 2 ? '#ffffff' : '#fef08a'
          sz = Math.max(sz, p.d < 1.5 ? 12 : 8)
        }

        ctx.globalAlpha = Math.min(1, fadeIn)
        ctx.fillStyle = color

        let dx = p.bx, dy = p.by
        if (s.phase === 'hold') {
          dx += (Math.sin(s.holdTimer * 0.018 + p.by * 0.12) * 0.4)
          dy += (Math.cos(s.holdTimer * 0.018 + p.bx * 0.12) * 0.4)
        }
        ctx.fillRect(Math.round(dx), Math.round(dy), sz, sz)
      }
    }

    // 碎片粒子（破碎阶段）
    for (const sh of s.shardPx) {
      if (sh.life <= 0) continue
      ctx.globalAlpha = sh.life
      ctx.fillStyle = sh.color
      ctx.fillRect(Math.round(sh.x), Math.round(sh.y), sh.size, sh.size)
    }

    for (const sp of s.edgeSparks) {
      if (sp.life <= 0) continue
      ctx.globalAlpha = sp.life
      ctx.fillStyle = sp.color
      ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size)
    }

    ctx.globalAlpha = 1
  },
}

// ---- 8. doom_judgment: 断裂像素环 ----
const doomJudgmentRenderer = {
  init(size) {
    return {phase: 'build', timer: 0, ringPixels: [], cx: size / 2, cy: size / 2, r: size * 0.44}
  },
  update(s, _size) {
    s.timer += 16
    const totalSegments = 28
    const gapIndices = new Set([3, 9, 16, 22, 26]) // 断裂位置
    if (s.phase === 'build') {
      const built = Math.floor(s.timer / 40)
      s.ringPixels = []
      for (let i = 0; i < Math.min(built, totalSegments); i++) {
        if (gapIndices.has(i)) continue
        const ang = (i / totalSegments) * Math.PI * 2
        s.ringPixels.push({
          x: s.cx + Math.cos(ang) * s.r,
          y: s.cy + Math.sin(ang) * s.r,
          life: 1,
          size: 3,
          color: i % 4 === 0 ? '#f87171' : '#7f1d1d'
        })
      }
      if (built >= totalSegments) {
        s.phase = 'contract';
        s.timer = 0
      }
    }
    if (s.phase === 'contract') {
      const shrink = s.timer < 300 ? s.timer / 300 * 4 : 4
      for (const p of s.ringPixels) {
        const dx = p.x - s.cx, dy = p.y - s.cy, dist = Math.sqrt(dx * dx + dy * dy) || 1
        p.x -= (dx / dist) * shrink * 0.06;
        p.y -= (dy / dist) * shrink * 0.06
      }
      if (s.timer > 3500) {
        s.phase = 'done'
      }
    }
    return s
  },
  draw(ctx, s, _size) {
    for (const p of s.ringPixels) {
      ctx.fillStyle = p.color;
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 3500
  },
}

// ---- 8.5 magic_burst: 蓝色闪电霹雳直击 ----
const magicBurstRenderer = {
  init(size) {
    return {
      phase: 'charge',
      timer: 0,
      flash: 0,
      linePixels: [],
      impactPixels: [],
      emberPixels: [],
      arcPixels: [],
      ringPixels: [],
      cx: size / 2,
      cy: size / 2,
    }
  },
  update(s, size) {
    s.timer += 16
    if (s.phase === 'charge') {
      s.flash = Math.min(0.22, s.timer / 900)
      if (s.timer > 320) {
        s.phase = 'strike'
        s.timer = 0
        s.flash = 0.38
        s.linePixels = []
        s.impactPixels = []
        s.emberPixels = []
        s.arcPixels = []
        s.ringPixels = []
        const path = []
        let x = s.cx + rnd(size * 0.06)
        let y = -6
        path.push({x, y})
        while (y < s.cy) {
          x += rnd(size * 0.12)
          y += size * 0.11 + Math.random() * size * 0.08
          path.push({x, y: Math.min(y, s.cy)})
        }
        path[path.length - 1] = {x: s.cx + rnd(4), y: s.cy}
        for (let i = 1; i < path.length; i++) {
          const prev = path[i - 1]
          const next = path[i]
          const dx = next.x - prev.x
          const dy = next.y - prev.y
          const dist = Math.max(1, Math.hypot(dx, dy))
          const nx = -dy / dist
          const ny = dx / dist
          const steps = Math.max(1, Math.ceil(dist / 3))
          for (let step = 0; step <= steps; step++) {
            const t = step / steps
            const px = prev.x + dx * t
            const py = prev.y + dy * t
            for (let w = -1; w <= 1; w++) {
              const weight = Math.abs(w)
              s.linePixels.push({
                x: px + nx * w * 2 + rnd(0.8),
                y: py + ny * w * 2 + rnd(0.8),
                life: 0.85 - weight * 0.15 + Math.random() * 0.15,
                size: weight === 0 ? 4 : 3,
                color: weight === 0 ? '#f8fdff' : weight === 1 ? '#60a5fa' : '#312e81',
              })
            }
          }
        }
        for (let i = 0; i < 24; i++) {
          const ang = Math.random() * Math.PI * 2
          const speed = 1.1 + Math.random() * 3.2
          s.impactPixels.push(makeParticle(
              s.cx + rnd(4),
              s.cy + rnd(4),
              Math.cos(ang) * speed,
              Math.sin(ang) * speed * 0.75,
              0.35 + Math.random() * 0.35,
              i % 5 === 0 ? 4 : 3,
              i % 4 === 0 ? '#dbeafe' : i % 4 === 1 ? '#93c5fd' : i % 4 === 2 ? '#60a5fa' : '#312e81',
          ))
        }
        for (let i = 0; i < 12; i++) {
          const ang = Math.random() * Math.PI * 2
          const speed = 0.8 + Math.random() * 1.8
          s.emberPixels.push(makeParticle(
              s.cx + rnd(6),
              s.cy + rnd(6),
              Math.cos(ang) * speed,
              Math.sin(ang) * speed,
              0.45 + Math.random() * 0.2,
              2,
              i % 2 === 0 ? '#38bdf8' : '#8b5cf6',
          ))
        }
        for (let i = 0; i < 18; i++) {
          const angle = (Math.PI * 2 * i) / 18
          const radius = size * 0.08
          s.ringPixels.push({
            angle,
            radius,
            life: 0.75 + Math.random() * 0.18,
            size: i % 3 === 0 ? 3 : 2,
            color: i % 4 === 0 ? '#dbeafe' : i % 2 === 0 ? '#60a5fa' : '#38bdf8',
          })
        }
        for (let branch = 0; branch < 3; branch++) {
          const baseAngle = -Math.PI / 2 + branch * (Math.PI / 2)
          for (let i = 0; i < 10; i++) {
            const arcAngle = baseAngle + (Math.random() - 0.5) * 0.85
            const radius = size * (0.06 + i * 0.012)
            s.arcPixels.push({
              angle: arcAngle,
              radius,
              twist: (Math.random() - 0.5) * 0.7,
              life: 0.72 + Math.random() * 0.18,
              size: i % 4 === 0 ? 3 : 2,
              color: i % 3 === 0 ? '#dbeafe' : i % 2 === 0 ? '#60a5fa' : '#8b5cf6',
            })
          }
        }
      }
    }
    if (s.phase === 'strike') {
      for (const p of s.linePixels) p.life -= 0.045
      for (const p of s.impactPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.92
        p.vy *= 0.9
        p.life -= 0.03
      }
      for (const p of s.emberPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.95
        p.vy *= 0.95
        p.life -= 0.022
      }
      for (const p of s.ringPixels) {
        p.radius += size * 0.01
        p.life -= 0.018
      }
      for (const p of s.arcPixels) {
        p.angle += p.twist * 0.06
        p.radius -= size * 0.002
        p.life -= 0.016
      }
      s.flash = Math.max(0, s.flash - 0.028)
      if (s.timer > 220) {
        s.phase = 'afterglow'
        s.timer = 0
      }
    }
    if (s.phase === 'afterglow') {
      for (const p of s.linePixels) p.life -= 0.025
      for (const p of s.impactPixels) {
        p.x += p.vx * 0.6
        p.y += p.vy * 0.6
        p.vx *= 0.9
        p.vy *= 0.88
        p.life -= 0.018
      }
      for (const p of s.emberPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.97
        p.vy *= 0.97
        p.life -= 0.014
      }
      for (const p of s.ringPixels) {
        p.radius += size * 0.008
        p.life -= 0.014
      }
      for (const p of s.arcPixels) {
        p.angle += p.twist * 0.05
        p.radius = Math.max(0, p.radius - size * 0.0015)
        p.life -= 0.012
      }
      s.flash = Math.max(0, s.flash - 0.012)
      if (
          s.linePixels.every((p) => p.life <= 0) &&
          s.impactPixels.every((p) => p.life <= 0) &&
          s.emberPixels.every((p) => p.life <= 0) &&
          s.ringPixels.every((p) => p.life <= 0) &&
          s.arcPixels.every((p) => p.life <= 0)
      ) {
        s.phase = 'done'
        s.timer = 0
      }
    }
    return s
  },
  draw(ctx, s, size) {
    if (s.flash > 0) {
      ctx.globalAlpha = s.flash
      const flashSize = Math.max(12, Math.round(size * 0.34))
      ctx.fillStyle = '#dbeafe'
      ctx.fillRect(
          Math.round(s.cx - flashSize / 2),
          Math.round(s.cy - flashSize / 2),
          flashSize,
          flashSize,
      )
      ctx.globalAlpha = Math.max(0, s.flash * 0.58)
      ctx.fillStyle = '#93c5fd'
      ctx.fillRect(
          Math.round(s.cx - flashSize * 0.32),
          Math.round(s.cy - flashSize * 0.32),
          Math.round(flashSize * 0.64),
          Math.round(flashSize * 0.64),
      )
    }
    for (const p of s.linePixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const p of s.impactPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const p of s.emberPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const p of s.ringPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      const x = s.cx + Math.cos(p.angle) * p.radius
      const y = s.cy + Math.sin(p.angle) * p.radius
      ctx.fillRect(Math.round(x), Math.round(y), p.size, p.size)
    }
    for (const p of s.arcPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      const x = s.cx + Math.cos(p.angle) * p.radius
      const y = s.cy + Math.sin(p.angle) * p.radius
      ctx.fillRect(Math.round(x), Math.round(y), p.size, p.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 700
  },
}

// ---- 8.6 magic_rupture: 法球显现后迅速碎裂 ----
const magicRuptureRenderer = {
  init(size) {
    const cx = size / 2
    const cy = size / 2
    const orbs = []
    for (let i = 0; i < 4; i++) {
      const angle = (Math.PI * 2 * i) / 4 + Math.PI / 4
      orbs.push({
        angle,
        radius: size * 0.14,
        size: i % 2 === 0 ? 6 : 5,
        life: 1,
        color: i % 2 === 0 ? '#60a5fa' : '#8b5cf6',
      })
    }
    return {phase: 'form', timer: 0, cx, cy, orbs, runePixels: [], shardPixels: [], sparkPixels: [], flash: 0}
  },
  update(s, size) {
    s.timer += 16
    if (s.phase === 'form') {
      s.flash = 0.08 + Math.sin(s.timer * 0.05) * 0.04
      s.runePixels = []
      const runeRadius = size * 0.24
      for (let i = 0; i < 8; i++) {
        if (i === 1 || i === 5) continue
        const angle = (Math.PI * 2 * i) / 8
        s.runePixels.push({
          x: s.cx + Math.cos(angle) * runeRadius,
          y: s.cy + Math.sin(angle) * runeRadius,
          life: 0.6 + Math.sin(s.timer * 0.03 + i) * 0.18,
          size: i % 2 === 0 ? 3 : 2,
          color: i % 2 === 0 ? '#c4b5fd' : '#38bdf8',
        })
      }
      for (const orb of s.orbs) {
        orb.radius = size * (0.14 - Math.min(0.065, s.timer / 3400))
      }
      if (s.timer > 220) {
        s.phase = 'shatter'
        s.timer = 0
        s.flash = 0.3
        s.shardPixels = []
        s.sparkPixels = []
        for (const orb of s.orbs) {
          const ox = s.cx + Math.cos(orb.angle) * orb.radius
          const oy = s.cy + Math.sin(orb.angle) * orb.radius
          for (let i = 0; i < 10; i++) {
            const ang = orb.angle + (Math.random() - 0.5) * 1.6
            const speed = 0.9 + Math.random() * 2.4
            s.shardPixels.push(makeParticle(
                ox + rnd(2),
                oy + rnd(2),
                Math.cos(ang) * speed,
                Math.sin(ang) * speed,
                0.42 + Math.random() * 0.28,
                i % 3 === 0 ? 4 : 3,
                i % 4 === 0 ? '#dbeafe' : i % 4 === 1 ? '#60a5fa' : i % 4 === 2 ? '#8b5cf6' : '#312e81',
            ))
          }
        }
        for (let i = 0; i < 16; i++) {
          const ang = Math.random() * Math.PI * 2
          const speed = 1 + Math.random() * 2
          s.sparkPixels.push(makeParticle(
              s.cx + rnd(4),
              s.cy + rnd(4),
              Math.cos(ang) * speed,
              Math.sin(ang) * speed,
              0.3 + Math.random() * 0.2,
              2,
              i % 2 === 0 ? '#67e8f9' : '#c4b5fd',
          ))
        }
      }
    }
    if (s.phase === 'shatter') {
      for (const p of s.shardPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.93
        p.vy *= 0.91
        p.life -= 0.025
      }
      for (const p of s.sparkPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.95
        p.vy *= 0.94
        p.life -= 0.035
      }
      s.flash = Math.max(0, s.flash - 0.03)
      if (s.shardPixels.every((p) => p.life <= 0) && s.sparkPixels.every((p) => p.life <= 0)) {
        s.phase = 'done'
        s.timer = 0
      }
    }
    return s
  },
  draw(ctx, s, size) {
    if (s.flash > 0) {
      ctx.globalAlpha = s.flash
      ctx.fillStyle = '#a78bfa'
      ctx.fillRect(0, 0, size, size)
    }
    for (const rune of s.runePixels) {
      ctx.globalAlpha = Math.max(0, rune.life)
      ctx.fillStyle = rune.color
      ctx.fillRect(Math.round(rune.x), Math.round(rune.y), rune.size, rune.size)
    }
    if (s.phase === 'form') {
      for (const orb of s.orbs) {
        const ox = s.cx + Math.cos(orb.angle) * orb.radius
        const oy = s.cy + Math.sin(orb.angle) * orb.radius
        ctx.globalAlpha = orb.life
        ctx.fillStyle = orb.color
        ctx.fillRect(Math.round(ox - orb.size / 2), Math.round(oy - orb.size / 2), orb.size, orb.size)
        ctx.fillStyle = '#eff6ff'
        ctx.fillRect(Math.round(ox - 1), Math.round(oy - 1), 2, 2)
      }
    }
    for (const p of s.shardPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const p of s.sparkPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 450
  },
}

// ---- 8.7 magic_starfall: 中心扩散到 5x5 的高阶星陨潮爆 ----
const magicStarfallRenderer = {
  init(size) {
    return {
      phase: 'meteor',
      timer: 0,
      waveRadius: 0,
      flash: 0,
      corePulse: 0,
      wavePixels: [],
      dustPixels: [],
      emberPixels: [],
      meteorTrailPixels: [],
      meteorShards: [],
      meteorStartX: size * 0.12,
      meteorStartY: size * 0.14,
      meteorX: size * 0.12,
      meteorY: size * 0.14,
    }
  },
  update(s, size) {
    s.timer += 16
    const cx = size / 2
    const cy = size / 2
    const maxRadius = size * 0.7
    if (s.phase === 'meteor') {
      const progress = Math.min(1, s.timer / 460)
      const eased = 1 - Math.pow(1 - progress, 3)
      s.meteorX = s.meteorStartX + (cx - s.meteorStartX) * eased
      s.meteorY = s.meteorStartY + (cy - s.meteorStartY) * eased
      s.corePulse = progress * 0.25
      s.flash = progress * 0.08
      s.meteorTrailPixels = []
      const dx = s.meteorX - s.meteorStartX
      const dy = s.meteorY - s.meteorStartY
      const trailSteps = 18
      for (let i = 0; i < trailSteps; i++) {
        const t = i / Math.max(1, trailSteps - 1)
        const px = s.meteorX - dx * t * 0.62 + rnd(1.2)
        const py = s.meteorY - dy * t * 0.62 + rnd(1.2)
        s.meteorTrailPixels.push({
          x: px,
          y: py,
          life: 0.94 - t * 0.6,
          size: i < 3 ? 6 : i < 8 ? 4 : 3,
          color: i < 2 ? '#f8fdff' : i < 6 ? '#7dd3fc' : i < 12 ? '#60a5fa' : '#8b5cf6',
        })
      }
      if (progress >= 1) {
        s.phase = 'impact'
        s.timer = 0
        s.flash = 0.26
        s.corePulse = 0.45
        s.meteorShards = []
        for (let i = 0; i < 22; i++) {
          const angle = Math.random() * Math.PI * 2
          const speed = 0.9 + Math.random() * 2.2
          s.meteorShards.push(makeParticle(
              cx + rnd(5),
              cy + rnd(5),
              Math.cos(angle) * speed,
              Math.sin(angle) * speed,
              0.34 + Math.random() * 0.24,
              i % 6 === 0 ? 4 : 3,
              i % 7 === 0 ? '#fcd34d' : i % 2 === 0 ? '#7dd3fc' : '#8b5cf6',
          ))
        }
      }
    }
    if (s.phase === 'impact') {
      s.corePulse = Math.min(1, 0.45 + s.timer / 260)
      s.flash = Math.max(0, s.flash - 0.02)
      for (const p of s.meteorShards) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.94
        p.vy *= 0.94
        p.life -= 0.026
      }
      s.meteorTrailPixels = s.meteorTrailPixels.filter((p) => p.life > 0).map((p) => ({
        ...p,
        life: p.life - 0.04,
      }))
      if (s.timer > 420) {
        s.phase = 'expand'
        s.timer = 0
        s.waveRadius = size * 0.08
      }
    }
    if (s.phase === 'expand') {
      const progress = Math.min(1, s.timer / 2600)
      s.waveRadius = size * 0.08 + (maxRadius - size * 0.08) * progress
      s.flash = 0.05 + (1 - progress) * 0.08
      s.wavePixels = []
      const ringCount = 84
      const innerBandOffset = Math.max(3, size * 0.022)
      for (let i = 0; i < ringCount; i++) {
        const angle = (Math.PI * 2 * i) / ringCount
        const wobble = Math.sin(progress * 4 + i * 0.72) * size * 0.006
        const outerRadius = s.waveRadius + wobble
        const innerRadius = Math.max(0, outerRadius - innerBandOffset)
        const outerX = cx + Math.cos(angle) * outerRadius
        const outerY = cy + Math.sin(angle) * outerRadius
        const innerX = cx + Math.cos(angle) * innerRadius
        const innerY = cy + Math.sin(angle) * innerRadius
        s.wavePixels.push({
          x: outerX,
          y: outerY,
          life: 0.9,
          size: i % 10 === 0 ? 6 : i % 2 === 0 ? 4 : 3,
          color: i % 18 === 0 ? '#fcd34d' : i % 6 === 0 ? '#c4b5fd' : '#7dd3fc',
        })
        s.wavePixels.push({
          x: innerX,
          y: innerY,
          life: 0.88,
          size: i % 10 === 0 ? 5 : 3,
          color: i % 18 === 0 ? '#f59e0b' : i % 6 === 0 ? '#60a5fa' : '#1e3a8a',
        })
      }
      if (Math.random() < 0.95) {
        for (let i = 0; i < 5; i++) {
          const angle = Math.random() * Math.PI * 2
          const radius = Math.max(0, s.waveRadius - Math.random() * size * 0.1)
          s.dustPixels.push(makeParticle(
              cx + Math.cos(angle) * radius,
              cy + Math.sin(angle) * radius,
              Math.cos(angle) * (0.6 + Math.random() * 1.6),
              Math.sin(angle) * (0.6 + Math.random() * 1.6),
              0.45 + Math.random() * 0.3,
              Math.random() < 0.22 ? 4 : 3,
              Math.random() < 0.12 ? '#fcd34d' : Math.random() < 0.22 ? '#fde68a' : Math.random() < 0.62 ? '#60a5fa' : '#8b5cf6',
          ))
        }
      }
      for (const p of s.dustPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.96
        p.vy *= 0.96
        p.life -= 0.009
      }
      s.dustPixels = s.dustPixels.filter((p) => p.life > 0)
      if (progress >= 1) {
        s.phase = 'fade'
        s.timer = 0
        s.emberPixels = []
        for (let i = 0; i < 42; i++) {
          const angle = Math.random() * Math.PI * 2
          const radius = Math.random() * maxRadius
          const speed = 0.5 + Math.random() * 1.8
          s.emberPixels.push(makeParticle(
              cx + Math.cos(angle) * radius,
              cy + Math.sin(angle) * radius,
              Math.cos(angle) * speed,
              Math.sin(angle) * speed,
              0.38 + Math.random() * 0.28,
              i % 6 === 0 ? 5 : i % 3 === 0 ? 4 : 3,
              i % 5 === 0 ? '#fcd34d' : i % 2 === 0 ? '#a78bfa' : '#60a5fa',
          ))
        }
      }
    }
    if (s.phase === 'fade') {
      s.flash = Math.max(0, s.flash - 0.01)
      for (const p of s.dustPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.96
        p.vy *= 0.96
        p.life -= 0.009
      }
      for (const p of s.emberPixels) {
        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.95
        p.vy *= 0.95
        p.life -= 0.01
      }
      s.dustPixels = s.dustPixels.filter((p) => p.life > 0)
      if (s.emberPixels.every((p) => p.life <= 0) && s.dustPixels.length === 0) {
        s.phase = 'done'
        s.timer = 0
      }
    }
    return s
  },
  draw(ctx, s, size) {
    const cx = size / 2
    const cy = size / 2
    if (s.flash > 0) {
      ctx.globalAlpha = s.flash
      ctx.fillStyle = '#312e81'
      ctx.fillRect(0, 0, size, size)
    }
    for (const p of s.meteorTrailPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    if (s.phase === 'meteor' || s.phase === 'impact') {
      ctx.globalAlpha = 0.95
      ctx.fillStyle = '#eef6ff'
      ctx.fillRect(Math.round(s.meteorX - 5), Math.round(s.meteorY - 5), 10, 10)
      ctx.fillStyle = '#7dd3fc'
      ctx.fillRect(Math.round(s.meteorX - 3), Math.round(s.meteorY - 3), 6, 6)
      ctx.fillStyle = '#4f46e5'
      ctx.fillRect(Math.round(s.meteorX - 1), Math.round(s.meteorY - 1), 3, 3)
    }
    for (const p of s.meteorShards) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    if (s.phase === 'impact') {
      const craterRadius = 6 + s.corePulse * 6
      const debrisKick = Math.max(0, 1 - s.timer / 120) * 4
      ctx.globalAlpha = 0.92
      drawPixelRing(ctx, cx, cy, craterRadius, 2.8, '#7dd3fc')
      ctx.globalAlpha = 0.82
      ctx.fillStyle = '#312e81'
      ctx.fillRect(Math.round(cx - craterRadius * 1.02 - debrisKick), Math.round(cy - craterRadius * 0.9 - debrisKick * 0.65), Math.round(craterRadius * 0.34), Math.round(craterRadius * 0.2))
      ctx.fillRect(Math.round(cx + craterRadius * 0.66 + debrisKick), Math.round(cy - craterRadius * 0.62 - debrisKick * 0.48), Math.round(craterRadius * 0.3), Math.round(craterRadius * 0.18))
      ctx.fillRect(Math.round(cx - craterRadius * 0.84 - debrisKick * 0.82), Math.round(cy + craterRadius * 0.44 + debrisKick * 0.28), Math.round(craterRadius * 0.28), Math.round(craterRadius * 0.18))
      ctx.fillRect(Math.round(cx + craterRadius * 0.52 + debrisKick * 0.74), Math.round(cy + craterRadius * 0.58 + debrisKick * 0.34), Math.round(craterRadius * 0.36), Math.round(craterRadius * 0.2))
      ctx.globalAlpha = 0.94
      ctx.fillStyle = '#60a5fa'
      ctx.fillRect(Math.round(cx - craterRadius * 0.98 - debrisKick), Math.round(cy - craterRadius * 0.94 - debrisKick * 0.65), Math.round(craterRadius * 0.28), Math.round(craterRadius * 0.12))
      ctx.fillRect(Math.round(cx + craterRadius * 0.68 + debrisKick), Math.round(cy - craterRadius * 0.66 - debrisKick * 0.48), Math.round(craterRadius * 0.24), Math.round(craterRadius * 0.1))
      ctx.fillRect(Math.round(cx - craterRadius * 0.8 - debrisKick * 0.82), Math.round(cy + craterRadius * 0.42 + debrisKick * 0.28), Math.round(craterRadius * 0.22), Math.round(craterRadius * 0.1))
      ctx.fillRect(Math.round(cx + craterRadius * 0.56 + debrisKick * 0.74), Math.round(cy + craterRadius * 0.56 + debrisKick * 0.34), Math.round(craterRadius * 0.3), Math.round(craterRadius * 0.12))
      ctx.globalAlpha = 0.95
      ctx.fillStyle = '#1e1b4b'
      ctx.fillRect(Math.round(cx - craterRadius * 0.58), Math.round(cy - craterRadius * 0.42), Math.round(craterRadius * 1.16), Math.round(craterRadius * 0.84))
      ctx.fillStyle = '#0f172a'
      ctx.fillRect(Math.round(cx - craterRadius * 0.42), Math.round(cy - craterRadius * 0.28), Math.round(craterRadius * 0.84), Math.round(craterRadius * 0.56))
      ctx.globalAlpha = 0.46
      ctx.fillStyle = '#f8fdff'
      ctx.fillRect(Math.round(cx - craterRadius * 0.48), Math.round(cy - craterRadius * 0.52), Math.round(craterRadius * 0.62), 2)
    }
    if (s.phase === 'expand' || s.phase === 'fade' || s.phase === 'done') {
      const outerRingRadius = 5 + s.corePulse * 5
      const innerRingRadius = Math.max(2, outerRingRadius - 3)
      ctx.globalAlpha = 0.82
      drawPixelRing(ctx, cx, cy, outerRingRadius, 2.4, '#93c5fd')
      ctx.globalAlpha = 0.96
      drawPixelRing(ctx, cx, cy, innerRingRadius, 2.2, '#1d4ed8')
    }
    for (const p of s.wavePixels) {
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const p of s.dustPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const p of s.emberPixels) {
      if (p.life <= 0) continue
      ctx.globalAlpha = p.life
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 300
  },
}

// ---- 9. silver_storm: 顶部到底部的多道随机银色刀光 ----
const silverStormRenderer = {
  init(size) {
    const slashes = []
    for (let i = 0; i < 10; i++) {
      const topX = -10 + Math.random() * (size + 20)
      const botX = -6 + Math.random() * (size + 12)
      const dx = botX - topX
      const dy = size + 18
      const len = Math.sqrt(dx * dx + dy * dy)
      slashes.push({
        startMs: 30 + i * 90,
        topX,
        botX,
        dx,
        dy,
        len,
        headDist: 0,
        speed: 5.4 + Math.random() * 0.9,
        width: 1 + Math.floor(Math.random() * 2),
        arrived: false,
      })
    }
    return {
      phase: 'slash',
      timer: 0,
      slashes,
      linePixels: [],
      shardPixels: [],
      edgeSparks: [],
      screenFlash: 0,
    }
  },
  update(s, size) {
    s.timer += 16
    s.linePixels = []
    if (s.phase === 'slash' || s.phase === 'hold') {
      let arrivedCount = 0
      for (const def of s.slashes) {
        if (s.phase === 'slash' && s.timer >= def.startMs && !def.arrived) {
          const elapsed = s.timer - def.startMs
          def.headDist = elapsed * def.speed / 16 * 5.6
          if (def.headDist >= def.len) {
            def.headDist = def.len
            def.arrived = true
            s.screenFlash = Math.max(s.screenFlash, 0.12)
            const hitX = def.botX + rnd(4)
            const hitY = size - 4 + rnd(4)
            for (let i = 0; i < 8; i++) {
              s.edgeSparks.push(makeParticle(hitX, hitY, rnd(1.8), -Math.random() * 2.2, 0.26 + Math.random() * 0.18, 2, i % 2 === 0 ? '#e2e8f0' : '#94a3b8'))
            }
          }
        }
        if (s.phase === 'slash' && s.timer < def.startMs) {
          continue
        }
        if (def.arrived) arrivedCount++
        const visibleLen = def.arrived ? def.len : Math.max(0, def.headDist)
        if (visibleLen <= 0) continue
        const steps = Math.ceil(def.len / 2)
        for (let i = 0; i <= steps; i++) {
          const t = i / steps
          const dist = visibleLen - t * def.len
          if (dist < -14) continue
          const frac = dist / def.len
          const px = def.topX + def.dx * frac
          const py = frac * def.dy - 10
          if (py < -10 || py > size + 10) continue
          for (let w = -def.width; w <= def.width; w++) {
            const intensity = Math.abs(w) === 0 ? 1 : Math.abs(w) <= 1 ? 0.82 : Math.abs(w) <= 2 ? 0.56 : 0.3
            const color = Math.abs(w) === 0 ? '#ffffff' : Math.abs(w) <= 1 ? '#e2e8f0' : Math.abs(w) <= 2 ? '#cbd5e1' : '#94a3b8'
            s.linePixels.push({
              x: px + w * 1.1 + rnd(0.6),
              y: py + rnd(0.9),
              life: 0.42 + intensity * 0.46,
              size: Math.abs(w) <= 1 ? 3 : 2,
              color,
            })
          }
          if (i % 5 === 0) {
            s.linePixels.push({
              x: px + def.width * 1.2 + rnd(0.8),
              y: py + rnd(1.2),
              life: 0.28,
              size: 2,
              color: '#94a3b8',
            })
            s.linePixels.push({
              x: px - def.width * 1.2 + rnd(0.8),
              y: py + rnd(1.2),
              life: 0.22,
              size: 2,
              color: '#cbd5e1',
            })
          }
        }
      }
      if (s.phase === 'slash' && arrivedCount === s.slashes.length) {
        s.phase = 'hold'
        s.timer = 0
      }
      if (s.phase === 'hold' && s.timer >= 500) {
        s.phase = 'shatter'
        s.timer = 0
        s.screenFlash = Math.max(s.screenFlash, 0.18)
        for (const def of s.slashes) {
          const steps = Math.ceil(def.len / 8)
          for (let i = 0; i <= steps; i++) {
            const t = i / steps
            const px = def.topX + def.dx * t
            const py = def.dy * t - 10
            for (let w = -def.width; w <= def.width; w++) {
              const ang = Math.random() * Math.PI * 2
              const spd = 0.8 + Math.random() * 2.2
              s.shardPixels.push(makeParticle(
                  px + w * 1.1 + rnd(0.6),
                  py + rnd(0.8),
                  Math.cos(ang) * spd,
                  Math.sin(ang) * spd,
                  0.34 + Math.random() * 0.28,
                  Math.abs(w) <= 1 ? 3 : 2,
                  Math.abs(w) === 0 ? '#ffffff' : Math.abs(w) <= 1 ? '#e2e8f0' : '#94a3b8',
              ))
            }
          }
        }
        s.linePixels = []
      }
    }
    if (s.phase === 'shatter') {
      for (const shard of s.shardPixels) {
        shard.x += shard.vx
        shard.y += shard.vy
        shard.vx *= 0.96
        shard.vy *= 0.96
        shard.life -= 0.016
      }
      s.shardPixels = s.shardPixels.filter((shard) => shard.life > 0)
    }
    if (s.phase === 'slash' || s.phase === 'hold' || s.phase === 'shatter') {
      for (const sp of s.edgeSparks) {
        sp.x += sp.vx
        sp.y += sp.vy
        sp.vx *= 0.94
        sp.vy *= 0.92
        sp.life -= 0.028
      }
      s.edgeSparks = s.edgeSparks.filter((sp) => sp.life > 0)
      s.screenFlash = Math.max(0, s.screenFlash - 0.02)
      if (s.phase === 'shatter' && s.shardPixels.length === 0 && s.edgeSparks.length === 0 && s.timer > 700) {
        s.phase = 'done'
        s.timer = 0
      }
    }

    if (s.phase === 'done') return s
    return s
  },
  draw(ctx, s, size) {
    if (s.screenFlash > 0) {
      ctx.globalAlpha = s.screenFlash
      ctx.fillStyle = '#e2e8f0'
      ctx.fillRect(0, 0, size, size)
      ctx.globalAlpha = 1
    }
    for (const p of s.linePixels) {
      if (p.life <= 0) continue;
      ctx.globalAlpha = p.life;
      ctx.fillStyle = p.color;
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    for (const shard of s.shardPixels) {
      if (shard.life <= 0) continue;
      ctx.globalAlpha = shard.life;
      ctx.fillStyle = shard.color;
      ctx.fillRect(Math.round(shard.x), Math.round(shard.y), shard.size, shard.size)
    }
    for (const sp of s.edgeSparks) {
      if (sp.life <= 0) continue;
      ctx.globalAlpha = sp.life;
      ctx.fillStyle = sp.color;
      ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.phase === 'done' && s.timer > 900
  },
}

const clickSparkRenderer = {
  init(size, entry = {}) {
    const cx = size / 2
    const cy = size / 2
    const variant = String(entry.cellType || 'soft')
    const paletteByType = {
      weak: ['#facc15', '#ef4444', '#f87171', '#fbbf24', '#f59e0b', '#dc2626'],
      heavy: ['#9ca3af', '#787888', '#64748b', '#94a3b8'],
      soft: ['#f8fafc', '#e2e8f0', '#cbd5e1', '#fafaff'],
    }
    const gravityByType = {
      weak: 0.04,
      heavy: 0.18,
      soft: 0.08,
    }
    const palette = paletteByType[variant] || paletteByType.soft
    const gravity = gravityByType[variant] || gravityByType.soft
    const count = Math.max(4, Number(entry.count || 6))
    const particles = []
    for (let i = 0; i < count; i++) {
      const angle = Math.PI * 0.1 + Math.random() * Math.PI * 0.5
      const speed = 4 + Math.random() * 10
      const sz = 4 + Math.floor(Math.random() * 8)
      particles.push({
        x: cx,
        y: cy,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed,
        gravity,
        decay: 0.03 + Math.random() * 0.04,
        life: 1,
        size: sz,
        color: palette[Math.floor(Math.random() * palette.length)],
      })
    }
    return {timer: 0, particles}
  },
  update(s) {
    s.timer += 16
    for (const p of s.particles) {
      p.x += p.vx
      p.y += p.vy
      p.vy += p.gravity
      p.life -= p.decay
    }
    return s
  },
  draw(ctx, s) {
    for (const p of s.particles) {
      if (p.life <= 0) continue
      ctx.globalAlpha = Math.max(0, p.life)
      ctx.fillStyle = p.color
      ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size)
    }
    ctx.globalAlpha = 1
  },
  isDone(s) {
    return s.particles.every((p) => p.life <= 0)
  },
}

// ======= 注册表（与后端 EffectType 对齐） =======
const renderers = {
  storm_combo: stormComboRenderer,
  auto_strike: autoStrikeRenderer,
  bleed: bleedRenderer,
  final_cut: finalCutRenderer,
  collapse_trigger: collapseTriggerRenderer,
  judgment_day: judgmentDayRenderer,
  doom_mark: doomJudgmentRenderer,
  silver_storm: silverStormRenderer,
  magic_burst: magicBurstRenderer,
  magic_rupture: magicRuptureRenderer,
  magic_starfall: magicStarfallRenderer,
  click_spark: clickSparkRenderer,
}

// ======= 组件生命周期 =======
const pooledEffects = new Map()
const activeManagedEffects = new Map()
const retiredManagedEffectIds = new Set()

function acquireManagedEffect(effect) {
  const effectPool = pooledEffects.get(effect)
  if (Array.isArray(effectPool) && effectPool.length > 0) {
    return effectPool.pop()
  }
  return {
    id: '',
    effect,
    renderer: null,
    state: null,
    size: 0,
    width: 0,
    height: 0,
    left: 0,
    top: 0,
  }
}

function releaseManagedEffect(instance) {
  if (!instance?.effect) return
  const effectPool = pooledEffects.get(instance.effect) || []
  instance.id = ''
  instance.state = null
  effectPool.push(instance)
  pooledEffects.set(instance.effect, effectPool)
}

function syncManagedEffects() {
  const nextEntries = Array.isArray(props.entries) ? props.entries : []
  const seen = new Set()
  for (const entry of nextEntries) {
    const effect = String(entry?.effect || '')
    const id = String(entry?.id || '')
    const currentRenderer = renderers[effect]
    if (!currentRenderer || !id) continue
    seen.add(id)
    if (retiredManagedEffectIds.has(id)) continue
    let instance = activeManagedEffects.get(id)
    if (!instance) {
      instance = acquireManagedEffect(effect)
      instance.id = id
      instance.effect = effect
      instance.renderer = currentRenderer
      instance.state = currentRenderer.init(Math.max(1, Number(entry?.size || 1)), entry)
      activeManagedEffects.set(id, instance)
    }
    instance.size = Math.max(1, Number(entry?.size || 1))
    instance.width = Math.max(1, Number(entry?.width || instance.size))
    instance.height = Math.max(1, Number(entry?.height || instance.size))
    instance.left = Math.round(Number(entry?.left || 0))
    instance.top = Math.round(Number(entry?.top || 0))
    instance.entry = entry
  }
  for (const id of retiredManagedEffectIds) {
    if (seen.has(id)) continue
    retiredManagedEffectIds.delete(id)
  }
}

function drawManagedFrame(ctx) {
  syncManagedEffects()
  ctx.clearRect(0, 0, canvasWidthValue.value, canvasHeightValue.value)
  for (const [id, instance] of activeManagedEffects.entries()) {
    instance.state = instance.renderer.update(instance.state, instance.size, instance.entry)
    ctx.save()
    ctx.translate(instance.left, instance.top)
    ctx.scale(instance.width / instance.size, instance.height / instance.size)
    instance.renderer.draw(ctx, instance.state, instance.size, instance.entry)
    ctx.restore()
    if (instance.renderer.isDone(instance.state)) {
      activeManagedEffects.delete(id)
      retiredManagedEffectIds.add(id)
      releaseManagedEffect(instance)
    }
  }
}

function start() {
  stop()
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!ctx) return
  ctx.imageSmoothingEnabled = false

  if (managerMode.value) {
    function frame() {
      if (!canvasRef.value || !managerMode.value) return
      drawManagedFrame(ctx)
      raf = requestAnimationFrame(frame)
    }
    drawManagedFrame(ctx)
    raf = requestAnimationFrame(frame)
    return
  }

  const r = renderers[props.effect]
  if (!r) return
  renderer = r
  state = r.init(props.size)

  function frame() {
    if (!canvasRef.value || renderer !== r) return
    state = r.update(state, props.size)
    if (r.isDone(state)) {
      if (props.loop) state = r.init(props.size)
      else {
        draw();
        return
      }
    }
    draw()
    raf = requestAnimationFrame(frame)
  }

  function draw() {
    const c = canvasRef.value
    const cx = c?.getContext('2d')
    if (!cx) return
    cx.clearRect(0, 0, props.size, props.size)
    r.draw(cx, state, props.size)
  }

  draw()
  raf = requestAnimationFrame(frame)
}

function stop() {
  cancelAnimationFrame(raf)
  raf = 0
  renderer = null
  state = null
  retiredManagedEffectIds.clear()
  for (const instance of activeManagedEffects.values()) {
    releaseManagedEffect(instance)
  }
  activeManagedEffects.clear()
}

onMounted(start)
onBeforeUnmount(stop)
watch(
    () => managerMode.value,
    () => {
      start()
    },
)
watch(
    () => [props.effect, props.size, props.loop],
    () => {
      if (managerMode.value) return
      start()
    },
)
</script>

<style scoped>
.pixel-canvas {
  display: block;
  image-rendering: pixelated;
  image-rendering: crisp-edges;
  background: transparent;
}
</style>
