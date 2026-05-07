import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {playBattlePartSound, playBattleTriggerSound, resolveSoundEffectUrl,} from './soundEffects'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const soundEffectsSource = readFileSync(path.resolve(currentDir, './soundEffects.js'), 'utf8')
const originalAudio = globalThis.Audio

describe('soundEffects', () => {
    beforeEach(() => {
        vi.useFakeTimers()
        const play = vi.fn().mockResolvedValue(undefined)
        globalThis.Audio = vi.fn(() => ({
            preload: '',
            volume: 1,
            play,
        }))
    })

    afterEach(() => {
        vi.runOnlyPendingTimers()
        vi.useRealTimers()
        globalThis.Audio = originalAudio
    })

    it('集中管理战斗音效映射，后续可继续扩展注册表', () => {
        expect(soundEffectsSource).toContain('registerSoundEffect')
        expect(soundEffectsSource).toContain('battle.trigger.final-cut')
        expect(soundEffectsSource).toContain('battle.click.soft')
        expect(soundEffectsSource).toContain("'battle.trigger.pursuit'")
        expect(soundEffectsSource).toContain('cooldownMs: 3200')
    })

    it('能按名称解析战斗音效地址', () => {
        expect(resolveSoundEffectUrl('白银风暴')).toBe('/sfx/battle/trigger/silver-storm.wav')
        expect(resolveSoundEffectUrl('storm_combo')).toBe('/sfx/battle/trigger/pursuit.wav')
        expect(resolveSoundEffectUrl('重甲')).toBe('/sfx/battle/click/heavy.wav')
    })

    it('能按映射名称播放点击音效', () => {
        expect(playBattlePartSound('弱点')).toBe(true)
        expect(globalThis.Audio).toHaveBeenCalledWith('/sfx/battle/click/weak.wav')
    })

    it('追击和审判日音效按配置延迟触发', () => {
        expect(playBattleTriggerSound('storm_combo')).toBe(true)
        expect(playBattleTriggerSound('审判日')).toBe(true)
        expect(globalThis.Audio).not.toHaveBeenCalled()

        vi.advanceTimersByTime(499)
        expect(globalThis.Audio).not.toHaveBeenCalled()

        vi.advanceTimersByTime(1)
        expect(globalThis.Audio).toHaveBeenCalledWith('/sfx/battle/trigger/pursuit.wav')

        vi.advanceTimersByTime(500)
        expect(globalThis.Audio).toHaveBeenCalledWith('/sfx/battle/trigger/judgment-day.wav')
    })

    it('终末血斩音效按配置延迟触发', () => {
        expect(playBattleTriggerSound('终末血斩')).toBe(true)
        expect(globalThis.Audio).not.toHaveBeenCalled()

        vi.advanceTimersByTime(699)
        expect(globalThis.Audio).not.toHaveBeenCalled()

        vi.advanceTimersByTime(1)
        expect(globalThis.Audio).toHaveBeenCalledWith('/sfx/battle/trigger/final-cut.wav')
    })

    it('白银风暴在持续窗口内不重复播触发音效', () => {
        expect(playBattleTriggerSound('白银风暴')).toBe(true)
        expect(globalThis.Audio).toHaveBeenCalledTimes(1)
        expect(playBattleTriggerSound('silver_storm')).toBe(false)
        expect(globalThis.Audio).toHaveBeenCalledTimes(1)

        vi.advanceTimersByTime(3200)
        expect(playBattleTriggerSound('silver_storm')).toBe(true)
        expect(globalThis.Audio).toHaveBeenCalledTimes(2)
    })

    it('自动打击会叠一层补强音量', () => {
        expect(playBattleTriggerSound('auto_strike')).toBe(true)
        expect(globalThis.Audio).toHaveBeenCalledTimes(0)

        vi.advanceTimersByTime(999)
        expect(globalThis.Audio).toHaveBeenCalledTimes(0)

        vi.advanceTimersByTime(1)
        expect(globalThis.Audio).toHaveBeenCalledTimes(1)

        vi.advanceTimersByTime(35)
        expect(globalThis.Audio).toHaveBeenCalledTimes(2)
        expect(globalThis.Audio.mock.results[1].value.volume).toBe(0.72)
    })
})
