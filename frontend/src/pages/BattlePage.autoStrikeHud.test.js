import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('碎甲重击 HUD', () => {
  it('从 talentCombatState 读取自动打击阈值与倒计时', () => {
    expect(stateSource).toContain('autoStrikeTriggerCount')
    expect(stateSource).toContain('autoStrikeExpiresAt')
    expect(stateSource).toContain('autoStrikeTimeoutPercent')
  })

  it('战斗页显示锁定重击累计与倒计时', () => {
    expect(battleSource).toContain('碎甲重击 {{ p.autoStrike }}/{{ autoStrikeTrigger }}')
    expect(battleSource).toContain('Math.ceil(p.autoStrikeCountdown)')
  })
})
