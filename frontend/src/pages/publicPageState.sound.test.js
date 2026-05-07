import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('publicPageState 音效接入', () => {
    it('点击和战斗事件都通过统一音效工具触发', () => {
        expect(pageSource).toContain("import {playBattlePartSound, playBattleTriggerSound, playUiActionSound} from '../utils/soundEffects'")
        expect(pageSource).toContain("playBattlePartSound(part?.type || part?.displayName || '')")
        expect(pageSource).toContain("playBattleTriggerSound(event?.effectType || event?.name || '')")
    })

    it('装备相关操作成功后播放对应 UI 音效', () => {
        expect(pageSource).toContain("playUiActionSound(action === 'equip' ? 'equip' : 'unequip')")
        expect(pageSource).toContain("playUiActionSound('salvage')")
        expect(pageSource).toContain("playUiActionSound('enhance')")
    })
})
