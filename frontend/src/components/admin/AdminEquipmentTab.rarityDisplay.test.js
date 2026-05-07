import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminEquipmentTab.vue'), 'utf8')

describe('AdminEquipmentTab 稀有度展示', () => {
    it('装备列表明确展示稀有度字段文案', () => {
        expect(pageSource).toContain("from '../../utils/rarity'")
        expect(pageSource).toContain('formatRarityLabel')
    })
})
