import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

describe('BattlePage 连击金色态', () => {
  it('金色连击时倒计时数字跟随切换为金色样式', () => {
    expect(battleSource).toContain('class="combo-box__timeout-text"')
    expect(battleSource).toContain("comboIsGold ? { color: '#fde047', textShadow: '0 0 10px rgba(251,191,36,0.55)' } : { color: comboColor }")
  })

  it('连击加成文案改为每25次提高10%伤害', () => {
    expect(battleSource).toContain('Math.floor(comboCount / 25) > 0')
    expect(battleSource).toContain('伤害 +{{ Math.floor(comboCount / 25) * 10 }}%')
  })

  it('进入金色态时有单独的过渡动画类', () => {
    expect(battleSource).toContain("'combo-box--gold-enter': comboGoldEntering")
    expect(battleSource).toContain('watch(comboIsGold')
  })

  it('每25连击触发一次伤害加成跳字提示', () => {
    expect(battleSource).toContain('const comboMilestoneText = ref(\'\')')
    expect(battleSource).toContain('const comboMilestoneTick = ref(0)')
    expect(battleSource).toContain('const nextMilestone = Math.floor(next / 25)')
    expect(battleSource).toContain('const prevMilestone = Math.floor((prev || 0) / 25)')
    expect(battleSource).toContain('comboMilestoneText.value = `连击加成 +${nextMilestone * 10}%`')
    expect(battleSource).toContain('combo-box__milestone combo-box__milestone--floating')
  })

  it('连击跳字从主数字右上角弹出而不是占下一行', () => {
    expect(battleSource).toContain('class="combo-box__count-wrap"')
    expect(battleSource).toContain('class="combo-box__milestone-anchor"')
    expect(battleSource).toContain('class="combo-box__milestone combo-box__milestone--floating"')
  })
})
