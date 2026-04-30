import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const canvasSource = readFileSync(path.resolve(currentDir, '../components/PixelEffectCanvas.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('BattlePage 战斗特效覆盖层', () => {
  it('左侧状态列宽度更充裕，并与 5x5 战斗区保留明确间距', () => {
    expect(styleSource).toContain('.boss-left-panels {')
    expect(styleSource).toContain('left: max(0px, calc(50% - 518px));')
    expect(styleSource).toContain('width: 220px;')
    expect(styleSource).toContain('@media (max-width: 960px) {')
    expect(styleSource).toContain('.boss-left-panels {\n    left: 0;\n    width: 190px;\n  }')
  })

  it('瞬发特效尺寸按格子边长倍率换算，而不是继续写死像素值', () => {
    expect(battleSource).toContain('const bossCellSizePx = ref(56)')
    expect(battleSource).toContain('function measureBossCellSize() {')
    expect(battleSource).toContain("const cell = grid?.querySelector?.('.boss-part-cell')")
    expect(battleSource).toContain('function effectCanvasSize(scale) {')
    expect(battleSource).toContain('function ultimateEffectCanvasSize() {')
    expect(battleSource).toContain('Math.round(bossCellSizePx.value * scale)')
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"storm_combo\" :size=\"effectCanvasSize(1.65)\" :loop=\"false\" />")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"auto_strike\" :size=\"effectCanvasSize(2.05)\" :loop=\"false\" />")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"bleed\" :size=\"effectCanvasSize(3.75)\" :loop=\"false\" />")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"final_cut\" :size=\"ultimateEffectCanvasSize()\" :loop=\"false\" />")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"collapse_trigger\" :size=\"effectCanvasSize(1)\" :loop=\"false\" />")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"judgment_day\" :size=\"bossGridEffectSize()\" :loop=\"false\" />")
    expect(styleSource).toContain('image-rendering: pixelated;')
  })

  it('每种特效仍保留独立锚点，但偏移改为按格子边长比例换算', () => {
    expect(battleSource).toContain('function effectOverlayStyle(type, options = {}) {')
    expect(battleSource).toContain("width: width > 0 ? `${Math.round(width)}px` : undefined")
    expect(battleSource).toContain("height: height > 0 ? `${Math.round(height)}px` : undefined")
    expect(battleSource).toContain("effectOverlayStyle('storm_combo', { scale: 1.65, fallback:")
    expect(battleSource).toContain("effectOverlayStyle('auto_strike', { scale: 2.05, fallback:")
    expect(battleSource).toContain("anchor: triggerAnchor('bleed', TALENT_EFFECT_WINDOW_MS, entry)")
    expect(battleSource).toContain("fallback: effectFallback(3.75, { top: '50%', left: '50%' })")
    expect(battleSource).toContain("effectOverlayStyle('final_cut', { anchor: 'grid', fallback:")
    expect(battleSource).toContain("effectOverlayStyle('collapse_trigger', { scale: 1, fallback:")
  })

  it('终末血斩改为直接驱动 5x5 战斗区的贯穿终结斩特效', () => {
    expect(battleSource).toContain('const ULTIMATE_EFFECT_WINDOW_MS = 3200')
    expect(battleSource).toContain("hasRecentTrigger('final_cut', ULTIMATE_EFFECT_WINDOW_MS)")
    expect(battleSource).toContain("effectOverlayStyle('final_cut', { anchor: 'grid', fallback:")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"final_cut\" :size=\"ultimateEffectCanvasSize()\" :loop=\"false\" />")
  })

  it('流血改为按事件列表并发叠加渲染，而不是只取最新一层覆盖旧层', () => {
    expect(battleSource).toContain("v-for=\"entry in recentTriggers('bleed')\"")
    expect(battleSource).not.toContain("hasRecentTrigger('bleed')")
    expect(battleSource).toContain(":key=\"entry.id\"")
    expect(battleSource).toContain('function triggerAnchor(type, windowMs = TALENT_EFFECT_WINDOW_MS, entryOverride = null) {')
  })

  it('审判日在战斗页使用独立更长的挂载窗口，避免比 demo 提前卸载', () => {
    expect(battleSource).toContain('const JUDGMENT_DAY_EFFECT_WINDOW_MS = 5000')
    expect(battleSource).toContain("hasRecentTrigger('judgment_day', JUDGMENT_DAY_EFFECT_WINDOW_MS)")
    expect(battleSource).toContain("triggerKey('judgment_day', JUDGMENT_DAY_EFFECT_WINDOW_MS)")
    expect(battleSource).toContain("effectOverlayStyle('judgment_day', { anchor: 'grid', fallback: { top: '50%', left: '50%' } })")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"judgment_day\" :size=\"bossGridEffectSize()\" :loop=\"false\" />")
  })

  it('流血以格子中心为喷发源，终末血斩的终结斩回退到更细一版的 5x5 对角重斩', () => {
    expect(canvasSource).toContain('const cx = size / 2, cy = size / 2')
    expect(canvasSource).toContain('const startX = -size * 0.32')
    expect(canvasSource).toContain('const endX = size * 1.32')
    expect(canvasSource).toContain('for (let w = -7; w <= 7; w++)')
    expect(canvasSource).toContain('sx + normalX * w * 4.8 + rnd(1.8)')
    expect(canvasSource).toContain('dist <= 1 ? 7 : dist <= 4 ? 6 : 4')
    expect(canvasSource).toContain('for (let w = -14; w <= 14; w += 3)')
    expect(canvasSource).toContain('for (const p of s.slashPixels) p.life -= 0.018')
    expect(canvasSource).toContain('sw.life -= 0.022')
    expect(canvasSource).toContain('sp.life -= 0.026')
  })

  it('崩塌只保留覆盖层瞬发特效，不再在格子内部重复渲染 PixelShatter', () => {
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"collapse_trigger\" :size=\"effectCanvasSize(1)\" :loop=\"false\" />")
    expect(battleSource).not.toContain('<PixelShatter')
  })

  it('崩塌特效缩小25%并让渲染节奏提速25%', () => {
    expect(battleSource).toContain("effectOverlayStyle('collapse_trigger', { scale: 1, fallback:")
    expect(battleSource).toContain("<PixelEffectCanvas effect=\"collapse_trigger\" :size=\"effectCanvasSize(1)\" :loop=\"false\" />")
    expect(canvasSource).toContain('const impactX = size / 2')
    expect(canvasSource).toContain('const impactY = size / 2')
    expect(canvasSource).toContain('const pivotX = impactX - handleLen')
    expect(canvasSource).toContain('const pivotY = impactY')
    expect(canvasSource).toContain('const collapseTriggerTickMs = 20')
    expect(canvasSource).toContain('const collapseTriggerSpeedScale = 1.25')
    expect(canvasSource).toContain('s.timer += collapseTriggerTickMs')
    expect(canvasSource).toContain('const speed = (0.6 + Math.random() * 2) * collapseTriggerSpeedScale')
  })

  it('特效覆盖层层级高于点击格子和伤害数字，不再被战斗格子遮挡', () => {
    expect(styleSource).toContain('.boss-zone-button--damage {')
    expect(styleSource).toContain('z-index: 30;')
    expect(styleSource).toContain('.talent-effect-overlay {')
    expect(styleSource).toContain('z-index: 40;')
    expect(styleSource).toContain('.talent-canvas-fx {')
    expect(styleSource).toContain('z-index: 41;')
  })
})
