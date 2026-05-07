import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = [
    './PublicPage.vue',
    './BattlePage.vue',
    './ArmoryPage.vue',
    './MessagesPage.vue',
    './publicPageState.js',
]
    .map((file) => readFileSync(path.resolve(currentDir, file), 'utf8'))
    .join('\n')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

function compact(source) {
    return source.replace(/\s+/g, ' ').trim()
}

describe('PublicPage 装备布局', () => {
    it('装备使用新属性格式显示', () => {
        expect(pageSource).toContain('function formatItemStats')
        expect(pageSource).toContain('attackPower')
        expect(pageSource).toContain('armorPenPercent')
    })

    it('装备名称通过统一稀有度工具渲染', () => {
        expect(pageSource).toContain('splitEquipmentName')
        expect(pageSource).toContain('getRarityClassName')
    })
})
