import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {
  playBattlePartSound,
  playBattleTriggerSound,
  resolveSoundEffectUrl,
} from './soundEffects'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const soundEffectsSource = readFileSync(path.resolve(currentDir, './soundEffects.js'), 'utf8')
const originalAudio = globalThis.Audio

describe('soundEffects', () => {
  beforeEach(() => {
    const play = vi.fn().mockResolvedValue(undefined)
    globalThis.Audio = vi.fn(() => ({
      preload: '',
      volume: 1,
      play,
    }))
  })

  afterEach(() => {
    globalThis.Audio = originalAudio
  })

  it('集中管理战斗音效映射，后续可继续扩展注册表', () => {
    expect(soundEffectsSource).toContain('registerSoundEffect')
    expect(soundEffectsSource).toContain('battle.trigger.final-cut')
    expect(soundEffectsSource).toContain('battle.click.soft')
  })

  it('能按名称解析战斗音效地址', () => {
    expect(resolveSoundEffectUrl('白银风暴')).toBe('/sfx/battle/trigger/silver-storm.wav')
    expect(resolveSoundEffectUrl('storm_combo')).toBe('/sfx/battle/trigger/pursuit.wav')
    expect(resolveSoundEffectUrl('重甲')).toBe('/sfx/battle/click/heavy.wav')
  })

  it('能按映射名称播放点击和触发音效', () => {
    expect(playBattlePartSound('弱点')).toBe(true)
    expect(playBattleTriggerSound('审判日')).toBe(true)
    expect(globalThis.Audio).toHaveBeenCalledWith('/sfx/battle/click/weak.wav')
    expect(globalThis.Audio).toHaveBeenCalledWith('/sfx/battle/trigger/judgment-day.wav')
  })
})
