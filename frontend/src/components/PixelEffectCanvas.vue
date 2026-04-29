<template>
  <canvas
    ref="canvasRef"
    class="pixel-canvas"
    :width="size"
    :height="size"
  />
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'

const props = defineProps({
  effect: { type: String, required: true },
  size: { type: Number, default: 90 },
  loop: { type: Boolean, default: true },
})

const canvasRef = ref(null)
let raf = 0
let state = null
let renderer = null

// ======= 盾牌图案 (collapse_trigger 用) =======
const SHIELD = [
  [0,0,0,0,0,0,0,1,0,0,0,0,0,0,0],
  [0,0,0,0,0,0,1,3,1,0,0,0,0,0,0],
  [0,0,0,0,0,1,4,2,4,1,0,0,0,0,0],
  [0,0,0,0,1,4,4,2,4,4,1,0,0,0,0],
  [0,0,0,1,4,4,4,2,4,4,4,1,0,0,0],
  [0,0,1,4,4,4,4,2,4,4,4,4,1,0,0],
  [0,1,4,4,4,4,4,2,4,4,4,4,4,1,0],
  [1,3,2,2,2,2,2,2,2,2,2,2,2,3,1],
  [0,1,4,4,4,4,4,2,4,4,4,4,4,1,0],
  [0,0,1,4,4,4,4,2,4,4,4,4,1,0,0],
  [0,0,0,1,4,4,4,2,4,4,4,1,0,0,0],
  [0,0,0,0,1,4,4,2,4,4,1,0,0,0,0],
  [0,0,0,0,0,1,4,2,4,1,0,0,0,0,0],
  [0,0,0,0,0,0,1,3,1,0,0,0,0,0,0],
  [0,0,0,0,0,0,0,1,0,0,0,0,0,0,0],
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
  return { x, y, vx, vy, life, maxLife: life, size, color }
}

// ======= 各特效渲染器 =======
// 每个返回 { init, update, draw, isDone }

function rnd(n) { return (Math.random() - 0.5) * n }

// ---- 1. storm_combo: 绿色像素斩击（加粗像素群） ----
const stormComboRenderer = {
  init(size) {
    return { phase: 'idle', timer: 0, parts: [], trails: [], sparks: [] }
  },
  update(s, size) {
    s.timer += 16
    if (s.phase === 'idle' && s.timer > 600) {
      s.phase = 'slash'; s.timer = 0
      // 粗斩击：3条平行像素线，每条 16 个粒子
      for (let row = -3; row <= 3; row += 3) {
        for (let i = 0; i < 16; i++) {
          const t = i / 15
          const baseX = 4 + t * (size - 8)
          const baseY = size - 4 - t * (size - 8)
          const greenShades = ['#22c55e','#4ade80','#16a34a','#86efac','#15803d','#166534']
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
        p.x += p.vx; p.y += p.vy
        p.life -= 0.014
        if (p.life > 0.2 && Math.random() < 0.5) s.trails.push(makeParticle(p.x, p.y, rnd(0.3), rnd(0.3), 0.3, 2, p.color))
      }
      for (const sp of s.sparks) { sp.x += sp.vx; sp.y += sp.vy; sp.life -= 0.02 }
      for (const t of s.trails) { t.life -= 0.04 }
      s.trails = s.trails.filter(t => t.life > 0)
      s.sparks = s.sparks.filter(sp => sp.life > 0)
      if (s.parts.every(p => p.life <= 0) && s.trails.length === 0 && s.sparks.length === 0) { s.phase = 'done'; s.timer = 0 }
    }
    if (s.phase === 'done' && s.timer > 900) return this.init(size)
    return s
  },
  draw(ctx, s, size) {
    for (const t of s.trails) { ctx.globalAlpha = t.life; ctx.fillStyle = t.color; ctx.fillRect(Math.round(t.x), Math.round(t.y), t.size, t.size) }
    for (const p of s.parts) { if (p.life <= 0) continue; ctx.globalAlpha = p.life; ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size) }
    for (const sp of s.sparks) { if (sp.life <= 0) continue; ctx.globalAlpha = sp.life; ctx.fillStyle = sp.color; ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 2. auto_strike: T形锤，锤柄底端在格子左侧边缘为圆心，绕90度砸向中心 ----
const autoStrikeRenderer = {
  init(size) {
    return { phase: 'idle', timer: 0, swingAngle: -Math.PI / 2, fragments: [], dent: [], flash: [] }
  },
  update(s, size) {
    s.timer += 16
    // 锤击落点固定在格子中心
    const handleLen = size * 0.5
    const impactX = size / 2
    const impactY = size / 2
    const pivotX = impactX - handleLen
    const pivotY = impactY
    if (s.phase === 'idle' && s.timer > 500) { s.phase = 'swing'; s.timer = 0 }
    if (s.phase === 'swing') {
      const progress = Math.min(1, s.timer / 450)
      const eased = progress < 0.5 ? 2 * progress * progress : 1 - Math.pow(-2 * progress + 2, 2) / 2
      s.swingAngle = -Math.PI / 2 + eased * (Math.PI / 2)
      if (progress >= 1) {
        s.phase = 'impact'; s.timer = 0
        const hitX = pivotX + Math.cos(s.swingAngle) * handleLen
        const hitY = pivotY + Math.sin(s.swingAngle) * handleLen
        for (let i = 0; i < 24; i++) {
          const ang = Math.random() * Math.PI * 2, spd = 1.5 + Math.random() * 3.5
          s.fragments.push(makeParticle(hitX, hitY, Math.cos(ang) * spd, Math.sin(ang) * spd, 1, 3 + Math.floor(Math.random() * 3), ['#9ca3af','#6b7280','#d1d5db','#4b5563'][Math.floor(Math.random() * 4)]))
        }
        for (let i = 0; i < 10; i++) s.flash.push(makeParticle(hitX + rnd(6), hitY + rnd(6), rnd(1), rnd(1), 0.4, 2, '#fef3c7'))
        for (let i = 0; i < 10; i++) s.dent.push(makeParticle(hitX + rnd(18), hitY + rnd(14), 0, 0, 2.5, 3, '#374151'))
      }
    }
    if (s.phase === 'impact') {
      for (const f of s.fragments) { f.x += f.vx; f.y += f.vy; f.vx *= 0.93; f.vy *= 0.93; f.life -= 0.022 }
      for (const fl of s.flash) { fl.life -= 0.04 }
      for (const d of s.dent) d.life -= 0.01
      if (s.fragments.every(f => f.life <= 0) && s.flash.every(fl => fl.life <= 0) && s.timer > 1400) { s.phase = 'done'; s.timer = 0 }
    }
    if (s.phase === 'done' && s.timer > 800) return this.init(size)
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
      // T 形锤：柄从 pivot 向 ang 延伸，锤头在柄顶端呈 T 形横梁
      const headX = pivotX + Math.cos(ang) * handleLen
      const headY = pivotY + Math.sin(ang) * handleLen
      // 锤柄（从 pivot 到 head 下方）
      const handleBase = handleLen * 0.15
      const steps = Math.ceil((handleLen - handleBase) / 3)
      for (let i = 0; i <= steps; i++) {
        const t = i / steps
        const px = Math.round(pivotX + (headX - pivotX) * t)
        const py = Math.round(pivotY + (headY - pivotY) * t)
        ctx.fillStyle = '#6b7280'; ctx.fillRect(px - 1, py, 4, 3)
        ctx.fillStyle = '#374151'; ctx.fillRect(px + 3, py, 2, 3)
      }
      // T 形锤头横梁（垂直于柄）
      const perpX = -Math.sin(ang), perpY = Math.cos(ang)
      // 横梁跨柄顶端
      for (let s = -5; s <= 5; s++) {
        for (let t = -2; t <= 3; t++) {
          const px = Math.round(headX + perpX * s * 3 + Math.cos(ang) * t * 3)
          const py = Math.round(headY + perpY * s * 3 + Math.sin(ang) * t * 3)
          if (Math.abs(s) <= 1 && t < 1) continue // 空出柄连接处
          const dark = Math.abs(s) >= 4 || Math.abs(t) >= 2
          ctx.fillStyle = dark ? '#4b5563' : Math.abs(s) <= 2 ? '#d1d5db' : '#9ca3af'
          ctx.fillRect(px, py, 4, 4)
        }
      }
    }
    for (const f of s.fragments) { if (f.life <= 0) continue; ctx.globalAlpha = f.life; ctx.fillStyle = f.color; ctx.fillRect(Math.round(f.x), Math.round(f.y), f.size, f.size) }
    for (const fl of s.flash) { if (fl.life <= 0) continue; ctx.globalAlpha = fl.life; ctx.fillStyle = fl.color; ctx.fillRect(Math.round(fl.x), Math.round(fl.y), fl.size, fl.size) }
    for (const d of s.dent) { if (d.life <= 0) continue; ctx.globalAlpha = Math.min(1, d.life); ctx.fillStyle = d.color; ctx.fillRect(Math.round(d.x), Math.round(d.y), d.size, d.size) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 3. bleed: 血迹扩散 ----
const bleedRenderer = {
  init(size) {
    const cx = size / 2, cy = size / 2
    const parts = []
    for (let i = 0; i < 70; i++) {
      const ang = Math.random() * Math.PI * 2, dist = Math.random() * 36
      parts.push(makeParticle(cx + Math.cos(ang) * dist, cy + Math.sin(ang) * dist, Math.cos(ang) * 0.45 + rnd(0.45), Math.sin(ang) * 0.45 + rnd(0.45), 0.6 + Math.random() * 0.4, 3 + Math.floor(Math.random() * 4), ['#7f1d1d','#991b1b','#450a0a','#b91c1c'][Math.floor(Math.random() * 4)]))
    }
    return { phase: 'spread', timer: 0, parts }
  },
  update(s, size) {
    s.timer += 16
    for (const p of s.parts) { p.x += p.vx; p.y += p.vy; p.vx *= 0.96; p.vy *= 0.96; p.life -= 0.003 }
    if (s.timer > 4000) return this.init(size)
    return s
  },
  draw(ctx, s, size) {
    for (const p of s.parts) { if (p.life <= 0) continue; ctx.globalAlpha = Math.min(1, p.life); ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 4. omen_harvest: 镰刀逐步划过（像素量增大） ----
const omenHarvestRenderer = {
  init(size) {
    return { phase: 'idle', timer: 0, bladePixels: [], trailPixels: [], residue: [] }
  },
  update(s, size) {
    s.timer += 16
    const cx = size * 0.38, cy = size * 0.4
    const rOuter = size * 0.58, rInner = size * 0.42
    if (s.phase === 'idle' && s.timer > 600) { s.phase = 'sweep'; s.timer = 0 }
    if (s.phase === 'sweep') {
      const progress = Math.min(1, s.timer / 400)
      const sweepAngle = -0.5 + progress * 2.8  // 从左上扫到右下
      s.bladePixels = []
      s.trailPixels = []
      // 镰刀主体：双层弧形，像素密集
      for (let a = -0.5; a < Math.min(sweepAngle + 0.15, 2.3); a += 0.04) {
        // 外弧
        const ox = cx + Math.cos(a) * rOuter, oy = cy + Math.sin(a) * rOuter
        const purpleShades = ['#c084fc','#a855f7','#9333ea','#7e22ce','#d8b4fe']
        const isEdge = a > sweepAngle - 0.1
        s.bladePixels.push({ x: ox, y: oy, life: isEdge ? 0.7 : 1, size: isEdge ? 2 : 3, color: purpleShades[Math.floor(Math.random() * purpleShades.length)] })
        // 内弧
        const ix = cx + Math.cos(a) * rInner, iy = cy + Math.sin(a) * rInner
        s.bladePixels.push({ x: ix, y: iy, life: 1, size: 2, color: a < sweepAngle - 0.2 ? '#581c87' : '#a855f7' })
        // 中间填充
        for (let f = 0; f < 3; f++) {
          const frac = Math.random(), rx = ox + (ix - ox) * frac + rnd(2), ry = oy + (iy - oy) * frac + rnd(2)
          s.bladePixels.push({ x: rx, y: ry, life: 0.85, size: 2, color: ['#7e22ce','#9333ea','#581c87'][Math.floor(Math.random() * 3)] })
        }
        // 尾迹像素
        if (a < sweepAngle && Math.random() < 0.4) {
          s.trailPixels.push({ x: ox + rnd(6), y: oy + rnd(6), life: 0.5, size: 2, color: '#c084fc' })
        }
      }
      // 刀柄
      const handleAngle = -0.5
      for (let h = 0; h < 10; h++) {
        const hx = cx + Math.cos(handleAngle) * (rOuter + h * 3), hy = cy + Math.sin(handleAngle) * (rOuter + h * 3)
        s.bladePixels.push({ x: hx, y: hy, life: 1, size: 3, color: h < 5 ? '#4a1942' : '#2d0a22' })
      }
      if (progress >= 1) {
        // 完成后留大量紫色残粒逐步淡出（留在原位不动）
        s.residue = []
        for (const p of s.bladePixels) {
          if (Math.random() < 0.55) {
            s.residue.push({
              x: p.x, y: p.y,
              life: 0.5 + Math.random() * 0.5,
              size: p.size,
              color: ['#c084fc','#a855f7','#d8b4fe','#7e22ce'][Math.floor(Math.random() * 4)],
            })
          }
        }
        // 额外散落粒子（也不动，只淡出）
        for (let i = 0; i < 25; i++) {
          const a = -0.5 + Math.random() * 2.8, r = rOuter + Math.random() * 14
          s.residue.push({
            x: cx + Math.cos(a) * r, y: cy + Math.sin(a) * r,
            life: 0.25 + Math.random() * 0.45,
            size: 2 + Math.floor(Math.random() * 2),
            color: ['#c084fc','#d8b4fe','#a855f7'][Math.floor(Math.random() * 3)],
          })
        }
        s.phase = 'done'; s.timer = 0
      }
    }
    if (s.phase === 'done') {
      for (const rp of s.residue) rp.life -= 0.005
      s.residue = s.residue.filter(rp => rp.life > 0)
      if (s.timer > 3500) return this.init(size)
    }
    return s
  },
  draw(ctx, s, size) {
    for (const tp of s.trailPixels) { if (tp.life <= 0) continue; ctx.globalAlpha = Math.min(1, tp.life); ctx.fillStyle = tp.color; ctx.fillRect(Math.round(tp.x), Math.round(tp.y), tp.size, tp.size) }
    for (const p of s.bladePixels) { ctx.globalAlpha = p.life; ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size) }
    for (const rp of s.residue) { if (rp.life <= 0) continue; ctx.globalAlpha = rp.life; ctx.fillStyle = rp.color; ctx.fillRect(Math.round(rp.x), Math.round(rp.y), rp.size, rp.size) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 5. final_cut: 贯穿全战斗区的超长对角终结斩 ----
const finalCutRenderer = {
  init(size) {
    return { phase: 'idle', timer: 0, slashPixels: [], shockwave: [], sparks: [], screenFlash: 0 }
  },
  update(s, size) {
    s.timer += 16
    if (s.phase === 'idle' && s.timer > 440) { s.phase = 'anticipate'; s.timer = 0 }
    if (s.phase === 'anticipate' && s.timer > 90) {
      s.phase = 'slash'; s.timer = 0; s.screenFlash = 1
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
      for (const sw of s.shockwave) { sw.x += sw.vx; sw.y += sw.vy; sw.vx *= 0.95; sw.vy *= 0.95; sw.life -= 0.022 }
      for (const sp of s.sparks) { sp.x += sp.vx; sp.y += sp.vy; sp.vx *= 0.93; sp.vy *= 0.93; sp.life -= 0.026 }
      if (s.slashPixels.every(p => p.life <= 0) && s.shockwave.every(sw => sw.life <= 0) && s.sparks.every(sp => sp.life <= 0)) { s.phase = 'done'; s.timer = 0 }
    }
    if (s.phase === 'done' && s.timer > 1400) return this.init(size)
    return s
  },
  draw(ctx, s, size) {
    // 屏幕闪白
    if (s.screenFlash > 0) { ctx.globalAlpha = s.screenFlash * 0.15; ctx.fillStyle = '#fff'; ctx.fillRect(0, 0, size, size); ctx.globalAlpha = 1 }
    for (const p of s.slashPixels) { if (p.life <= 0) continue; ctx.globalAlpha = p.life; ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size) }
    for (const sw of s.shockwave) { if (sw.life <= 0) continue; ctx.globalAlpha = sw.life; ctx.fillStyle = sw.color; ctx.fillRect(Math.round(sw.x), Math.round(sw.y), sw.size, sw.size) }
    for (const sp of s.sparks) { if (sp.life <= 0) continue; ctx.globalAlpha = sp.life; ctx.fillStyle = sp.color; ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 6. collapse_trigger: 盾牌爆裂（方案9，删除裂纹残留） ----
const collapseTriggerTickMs = 20
const collapseTriggerSpeedScale = 1.25

const collapseTriggerRenderer = {
  init(size) {
    return { phase: 'idle', timer: 0, parts: shieldParticles(size) }
  },
  update(s, size) {
    s.timer += collapseTriggerTickMs
    const cx = size / 2, cy = size / 2
    if (s.phase === 'idle' && s.timer > 800) {
      s.phase = 'explode'; s.timer = 0
      for (const p of s.parts) {
        const dx = p.x - cx, dy = p.y - cy
        const dist = Math.sqrt(dx * dx + dy * dy) || 1
        const speed = (0.6 + Math.random() * 2) * collapseTriggerSpeedScale
        p.vx = (dx / dist) * speed + rnd(1.2)
        p.vy = (dy / dist) * speed + rnd(1.2)
      }
    }
    if (s.phase === 'explode') {
      let alive = 0
      for (const p of s.parts) {
        if (p.life <= 0) continue
        p.x += p.vx; p.y += p.vy; p.vx *= 0.985; p.vy *= 0.985
        if (p.x < -8 || p.x > size + 8 || p.y < -8 || p.y > size + 8) p.life -= 0.08
        else if (s.timer > 300) p.life -= 0.015
        if (p.life > 0) alive++
      }
      if (alive === 0) { s.phase = 'done'; s.timer = 0 }
    }
    if (s.phase === 'done' && s.timer > 1000) return this.init(size)
    return s
  },
  draw(ctx, s, size) {
    for (const p of s.parts) { if (p.life <= 0) continue; ctx.globalAlpha = p.life; ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), Math.round(p.size), Math.round(p.size)) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 7. judgment_day: 圣洁黄金十字裁决（像素十字斩） ----
const judgmentDayRenderer = {
  init(size) {
    const cx = size / 2
    const cell = size / 5
    const barHw = Math.round(cell * 0.5)   // 臂宽 = 一整格
    const spacing = cell * 0.055            // 采样间距（密度翻 4 倍）
    const half = size * 0.5

    // 预生成十字所有像素块
    const px = []
    // 横臂：覆盖中心整行
    for (let x = 0; x < size; x += spacing) {
      const dx = Math.abs(x - cx)
      for (let wy = -barHw; wy <= barHw; wy += spacing) {
        const y = cx + wy
        if (y < 0 || y >= size) continue
        const d = Math.max(dx, Math.abs(wy) * 1.6)
        px.push({ bx: x, by: y, d })
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
        px.push({ bx: x, by: y, d })
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
      for (const sp of s.edgeSparks) { sp.x += sp.vx; sp.y += sp.vy; sp.vx *= 0.93; sp.vy *= 0.93; sp.life -= 0.017 }
      if (s.timer > 2200) { s.phase = 'shatter'; s.timer = 0; s.flashA = 0.20 }
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
      for (const sh of s.shardPx) { sh.x += sh.vx; sh.y += sh.vy; sh.vx *= 0.96; sh.vy *= 0.96; sh.life -= 0.016 }
      s.flashA = Math.max(0, s.flashA - 0.008)
      if (s.shardPx.every(sh => sh.life <= 0)) { s.phase = 'done'; s.timer = 0 }
    }

    return s
  },
  isDone(s) { return s.phase === 'done' && s.timer > 600 },
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
        if (p.d < 3) { color = '#ffffff'; sz = 10 }
        else if (dNorm < 0.18) { color = '#fef08a'; sz = 8 }
        else if (dNorm < 0.40) { color = '#fcd34d'; sz = 6 }
        else if (dNorm < 0.70) { color = '#f59e0b'; sz = 4 }
        else { color = '#d97706'; sz = 4 }

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
    return { phase: 'build', timer: 0, ringPixels: [], cx: size / 2, cy: size / 2, r: size * 0.44 }
  },
  update(s, size) {
    s.timer += 16
    const totalSegments = 28, gapCount = 5
    const gapIndices = new Set([3, 9, 16, 22, 26]) // 断裂位置
    if (s.phase === 'build') {
      const built = Math.floor(s.timer / 40)
      s.ringPixels = []
      for (let i = 0; i < Math.min(built, totalSegments); i++) {
        if (gapIndices.has(i)) continue
        const ang = (i / totalSegments) * Math.PI * 2
        s.ringPixels.push({ x: s.cx + Math.cos(ang) * s.r, y: s.cy + Math.sin(ang) * s.r, life: 1, size: 3, color: i % 4 === 0 ? '#f87171' : '#7f1d1d' })
      }
      if (built >= totalSegments) { s.phase = 'contract'; s.timer = 0 }
    }
    if (s.phase === 'contract') {
      const shrink = s.timer < 300 ? s.timer / 300 * 4 : 4
      for (const p of s.ringPixels) {
        const dx = p.x - s.cx, dy = p.y - s.cy, dist = Math.sqrt(dx*dx+dy*dy) || 1
        p.x -= (dx/dist) * shrink * 0.06; p.y -= (dy/dist) * shrink * 0.06
      }
      if (s.timer > 3500) return this.init(size)
    }
    return s
  },
  draw(ctx, s, size) {
    for (const p of s.ringPixels) { ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size) }
  },
  isDone(s) { return false },
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

    if (s.phase === 'done' && s.timer > 900) return this.init(size)
    return s
  },
  draw(ctx, s, size) {
    if (s.screenFlash > 0) {
      ctx.globalAlpha = s.screenFlash
      ctx.fillStyle = '#e2e8f0'
      ctx.fillRect(0, 0, size, size)
      ctx.globalAlpha = 1
    }
    for (const p of s.linePixels) { if (p.life <= 0) continue; ctx.globalAlpha = p.life; ctx.fillStyle = p.color; ctx.fillRect(Math.round(p.x), Math.round(p.y), p.size, p.size) }
    for (const shard of s.shardPixels) { if (shard.life <= 0) continue; ctx.globalAlpha = shard.life; ctx.fillStyle = shard.color; ctx.fillRect(Math.round(shard.x), Math.round(shard.y), shard.size, shard.size) }
    for (const sp of s.edgeSparks) { if (sp.life <= 0) continue; ctx.globalAlpha = sp.life; ctx.fillStyle = sp.color; ctx.fillRect(Math.round(sp.x), Math.round(sp.y), sp.size, sp.size) }
    ctx.globalAlpha = 1
  },
  isDone(s) { return false },
}

// ---- 10. crosshair-mark: 准星标记 ----
const crosshairMarkRenderer = {
  init(size) {
    return { phase: 'idle', timer: 0, lines: [], centerDot: null, cx: size * 0.72, cy: size * 0.28 }
  },
  update(s, size) {
    s.timer += 16
    const { cx, cy } = s
    if (s.phase === 'idle' && s.timer > 500) { s.phase = 'converge'; s.timer = 0 }
    if (s.phase === 'converge') {
      const progress = Math.min(1, s.timer / 400)
      s.lines = []
      const directions = [[0,-1],[0,1],[-1,0],[1,0]]
      for (const [dx, dy] of directions) {
        const len = 16, endX = cx + dx * len, endY = cy + dy * len
        const startX = cx + dx * len * 3, startY = cy + dy * len * 3
        const x = startX + (endX - startX) * progress
        const y = startY + (endY - startY) * progress
        for (let i = 0; i <= len; i += 2) {
          s.lines.push({ x: Math.round(startX + (x - startX) * i / len), y: Math.round(startY + (y - startY) * i / len), life: 1, size: 2, color: '#f87171' })
        }
      }
      if (progress >= 1) {
        s.centerDot = makeParticle(cx, cy, 0, 0, 999, 3, '#ef4444')
        s.phase = 'freeze'; s.timer = 0
      }
    }
    if (s.phase === 'freeze' && s.timer > 3500) return this.init(size)
    return s
  },
  draw(ctx, s, size) {
    for (const l of s.lines) { ctx.fillStyle = l.color; ctx.fillRect(l.x, l.y, l.size, l.size) }
    if (s.centerDot) { ctx.fillStyle = s.centerDot.color; ctx.fillRect(Math.round(s.centerDot.x - 1), Math.round(s.centerDot.y - 1), 3, 3) }
  },
  isDone(s) { return false },
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
}

// ======= 组件生命周期 =======
function start() {
  const r = renderers[props.effect]
  if (!r) return
  renderer = r
  state = r.init(props.size)
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!ctx) return
  ctx.imageSmoothingEnabled = false

  function frame() {
    if (!canvasRef.value || renderer !== r) return
    state = r.update(state, props.size)
    if (r.isDone(state)) {
      if (props.loop) state = r.init(props.size)
      else { draw(); return }
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
}

onMounted(start)
onBeforeUnmount(stop)
watch(() => props.effect, () => { stop(); start() })
</script>

<style scoped>
.pixel-canvas {
  display: block;
  image-rendering: pixelated;
  image-rendering: crisp-edges;
  background: transparent;
}
</style>
