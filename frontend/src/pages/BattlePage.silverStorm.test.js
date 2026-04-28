import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('白银风暴秒级状态', () => {
  it('实时状态从 talentCombatState 读取剩余时间，不在事件里硬编码 15', () => {
    expect(stateSource).toContain('vs.silverStormEndsAt = Number(state.silverStormEndsAt) || 0')
    expect(stateSource).not.toContain('vs.silverStormRemaining = 15')
  })

  it('战斗页显示白银风暴倒计时而不是攻击轮次计数', () => {
    expect(battleSource).toContain('silverStormCountdown')
    expect(battleSource).not.toContain('白银风暴 {{ talentVisualState.silverStormRemaining }}')
  })
})
