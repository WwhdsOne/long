import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('BattlePage 大整数 Boss 血量', () => {
    it('Boss 和部位血量百分比不再直接使用 Number 除法', () => {
        expect(stateSource).toContain('ratioPercent(')
        expect(stateSource).not.toContain('(boss.value.currentHp / boss.value.maxHp) * 100')
        expect(battleSource).toContain('ratioPercent(')
        expect(battleSource).not.toContain('(part.currentHp / part.maxHp) * 100')
    })
})
