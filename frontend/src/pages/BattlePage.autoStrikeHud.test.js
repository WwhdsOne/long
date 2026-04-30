import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('碎甲重击 HUD', () => {
  it('终末血斩 HUD 改为读取死兆资源态与最近触发状态', () => {
    expect(stateSource).toContain('const globalStatusList = computed(() => {')
    expect(stateSource).toContain("kind: 'final_cut'")
    expect(stateSource).toContain('const finalCutLastTriggerAt = Math.max(0, Number(talentCombatState.value?.lastFinalCutAt) || 0)')
    expect(stateSource).toContain('const finalCutRecentWindowSec = 3')
    expect(stateSource).toContain('const finalCutRecentlyTriggered = finalCutLastTriggerAt > 0')
    expect(stateSource).toContain("secondary: finalCutRecentlyTriggered > 0 ? '刚触发' : '距自动触发剩余层数'")
    expect(stateSource).toContain("hint: finalCutRecentlyTriggered > 0 ? '' : `${omenCap} 层自动引爆`")
    expect(battleSource).toContain('globalStatusList')
    expect(battleSource).toContain('class="status-panel"')
    expect(battleSource).toContain('status-panel--${status.kind}')
  })

  it('从 talentCombatState 读取自动打击阈值与倒计时', () => {
    expect(stateSource).toContain('autoStrikeTriggerCount')
    expect(stateSource).toContain('autoStrikeExpiresAt')
    expect(stateSource).toContain('autoStrikeTimeoutPercent')
  })

  it('战斗页显示锁定重击累计与倒计时', () => {
    expect(battleSource).toContain('碎甲重击 {{ p.autoStrike }}/{{ autoStrikeTrigger }}')
    expect(battleSource).toContain('Number.isFinite(p.autoStrikeCountdown) ? Math.ceil(p.autoStrikeCountdown) : 0')
  })

  it('自动打击倒计时与百分比在非有限数值时回退为 0，避免出现 NaN', () => {
    expect(stateSource).toContain('const nowTick = ref(0)')
    expect(stateSource).toContain('let talentTickTimer = 0')
    expect(stateSource).toContain('talentTickTimer = window.setInterval(() => {')
    expect(stateSource).toContain('nowTick.value++')
    expect(stateSource).toContain('window.clearInterval(talentTickTimer)')
    expect(stateSource).toContain('const safeAutoStrikeCountdown = computed(() => Number.isFinite(autoStrikeCountdown.value) ? autoStrikeCountdown.value : 0)')
    expect(stateSource).toContain('const ratio = safeAutoStrikeCountdown.value / windowSec')
    expect(stateSource).toContain('return Number.isFinite(ratio)')
    expect(stateSource).not.toContain('Math.round(ratio * 100)')
    expect(stateSource).toContain('autoStrikeCountdown: autoStrike > 0 ? safeAutoStrikeCountdown.value : 0')
    expect(stateSource).toContain('autoStrikeTimeoutPercent: autoStrike > 0 ? safeAutoStrikeTimeoutPercent.value : 0')
  })
})
