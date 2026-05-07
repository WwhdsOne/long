import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('白银风暴秒级状态', () => {
    it('实时状态从 talentCombatState 读取剩余时间，不在事件里硬编码 15', () => {
        expect(stateSource).toContain('vs.silverStormEndsAt = Number(state.silverStormEndsAt) || 0')
        expect(stateSource).not.toContain('vs.silverStormRemaining = 15')
    })

    it('白银风暴活跃期间重复同步战斗状态时，总时长不会被当前剩余时间重置成满进度', () => {
        expect(stateSource).toContain('const prevSilverStormEndsAt = Number(vs.silverStormEndsAt) || 0')
        expect(stateSource).toContain('const shouldResetSilverStormDuration = vs.silverStormActive && (!prevSilverStormActive || vs.silverStormEndsAt > prevSilverStormEndsAt)')
        expect(stateSource).toContain('vs.silverStormDuration = vs.silverStormActive')
        expect(stateSource).toContain('? (shouldResetSilverStormDuration')
        expect(stateSource).toContain(': Math.max(Number(vs.silverStormDuration) || 0, Number(state.silverStormRemaining) || 0, vs.silverStormRemaining))')
    })

    it('白银风暴进度条按连续时间差平滑缩减，而不是跟随整数秒跳变', () => {
        expect(stateSource).toContain('const silverStormRemainingMs = silverStormEndsAt')
        expect(stateSource).toContain('? Math.max(0, silverStormEndsAt * 1000 - Date.now())')
        expect(stateSource).toContain('(silverStormRemainingMs / (silverStormDuration * 1000)) * 100')
        expect(stateSource).not.toContain('(silverStormRemaining / silverStormDuration) * 100')
    })

    it('白银风暴并入左侧统一全局状态面板，而不是继续保留独立状态条', () => {
        expect(stateSource).toContain('globalStatusList')
        expect(stateSource).toContain("key: 'silver-storm'")
        expect(stateSource).toContain("kind: 'silver_storm'")
        expect(stateSource).toContain("title: '白银风暴'")
        expect(stateSource).toContain("secondary: '最终伤害额外追加白银风暴'")
        expect(battleSource).toContain('globalStatusList')
        expect(battleSource).not.toContain('talent-status-chip talent-status-chip--silver')
        expect(battleSource).not.toContain('talent-status-chip__bar-fill talent-status-chip__bar-fill--silver')
    })

    it('白银风暴特效覆盖整个 5x5 战斗区，而不是继续按单格倍率绘制', () => {
        expect(battleSource).toContain("hasRecentTrigger('silver_storm', ULTIMATE_EFFECT_WINDOW_MS)")
        expect(battleSource).toContain("effectOverlayStyle('silver_storm', { anchor: 'grid', fallback: { top: '50%', left: '50%' } })")
        expect(battleSource).toContain('function ultimateEffectCanvasSize() {')
        expect(battleSource).toContain('<PixelEffectCanvas effect="silver_storm" :size="ultimateEffectCanvasSize()" :loop="false"/>')
    })

    it('战斗页不再保留旧的暴伤x3和死亡狂喜持续 Buff 占位', () => {
        expect(battleSource).not.toContain('talentVisualState.doomCritBuff')
        expect(battleSource).not.toContain('death-ecstasy-timer')
    })
})
