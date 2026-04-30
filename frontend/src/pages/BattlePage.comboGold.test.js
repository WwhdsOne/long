import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

describe('BattlePage 连击金色态', () => {
  it('连击并入统一全局状态面板，但每25次提高10%伤害的口径保留', () => {
    expect(battleSource).toContain('globalStatusList')
    expect(battleSource).not.toContain('class="combo-box"')
    expect(battleSource).toContain('status-panel--${status.kind}')
    expect(battleSource).toContain("v-if=\"status.kind === 'combo'\"")
    expect(battleSource).toContain('class="status-panel__row status-panel__row--combo"')
  })

  it('连击状态恢复按连击数变化的渐变色，而不是固定绿色样式', () => {
    const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
    expect(stateSource).toContain('const comboHue = 120 - Math.min(comboValue / 200, 1) * 120')
    expect(stateSource).toContain("const comboColor = `hsl(${comboHue}, 90%, ${55 - Math.min(comboValue / 200, 1) * 15}%)`")
    expect(stateSource).toContain('const comboIsGold = comboValue >= 200')
    expect(stateSource).toContain('const comboGoldGradient =')
    expect(stateSource).toContain('panelStyle:')
    expect(stateSource).toContain('barStyle:')
    expect(stateSource).toContain('primaryStyle:')
    expect(stateSource).toContain('hintStyle:')
    expect(battleSource).toContain(':style="status.panelStyle || null"')
    expect(battleSource).toContain(':style="status.primaryStyle || null"')
    expect(battleSource).toContain(':style="status.hintStyle || null"')
    expect(battleSource).toContain('...(status.barStyle || {})')
  })

  it('连击加成文案仍按每25次提高10%伤害计算', () => {
    expect(readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')).toContain('const comboBonus = Math.floor(comboValue / 25) * 10')
    expect(readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')).toContain("secondary: comboBonus > 0 ? `伤害 +${comboBonus}%` : ''")
  })

  it('连击状态保留倒计时与待命显示', () => {
    const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
    expect(stateSource).toContain("hint: comboValue > 0 ? `${Math.ceil(comboTimeoutPercent.value / 20)}s` : '待命'")
    expect(stateSource).toContain('comboTimeoutPercent.value = Math.max(0, 100 - (elapsed / COMBO_TIMEOUT_MS) * 100)')
    expect(stateSource).not.toContain('comboTimeoutPercent.value = Math.round(Math.max(0, 100 - (elapsed / COMBO_TIMEOUT_MS) * 100))')
    expect(stateSource).toContain("kind: 'combo'")
    expect(battleSource).toContain('class="status-panel__meta status-panel__meta--combo"')
    expect(battleSource).toContain("class=\"status-panel__secondary\"\n                          :style=\"status.secondaryStyle || null\"")
    expect(battleSource).toContain("class=\"status-panel__hint status-panel__hint--inline\"\n                          :style=\"status.hintStyle || null\"")
  })

  it('200连击以上恢复金黄色主题，数字、倒计时、加伤与进度条都切到金色渐变', () => {
    const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
    const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')
    expect(stateSource).toContain('const comboIsGold = comboValue >= 200')
    expect(stateSource).toContain("background: comboIsGold ? comboGoldGradient : undefined")
    expect(stateSource).toContain("WebkitBackgroundClip: comboIsGold ? 'text' : undefined")
    expect(stateSource).toContain("color: comboIsGold ? 'transparent' : comboColor")
    expect(stateSource).toContain("hintStyle: comboHintStyle")
    expect(stateSource).toContain("secondaryStyle: comboSecondaryStyle")
    expect(stateSource).toContain("background: comboIsGold ? `linear-gradient(90deg, ${comboGoldGradientStops.join(', ')})`")
    expect(stateSource).toContain("backgroundSize: '300% 300%'")
    expect(stateSource).toContain("animation: 'combo-gold-shimmer 2s linear infinite'")
    expect(battleSource).toContain(':style="status.secondaryStyle || null"')
    expect(styleSource).toContain('@keyframes combo-gold-shimmer')
  })

  it('每25连击仍会弹出一次类似街机效果的 +x% 提示', () => {
    expect(battleSource).toContain('const comboMilestoneText = ref(\'\')')
    expect(battleSource).toContain('const comboMilestoneTick = ref(0)')
    expect(battleSource).toContain('watch(() => comboCount.value, (next, prev) => {')
    expect(battleSource).toContain('comboMilestoneText.value = `+${nextMilestone * 10}%`')
    expect(battleSource).toContain('class="status-panel__milestone-anchor"')
    expect(battleSource).toContain('class="status-panel__milestone status-panel__milestone--floating"')
  })
})
