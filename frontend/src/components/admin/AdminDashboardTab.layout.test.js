import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const componentSource = readFileSync(path.resolve(currentDir, './AdminDashboardTab.vue'), 'utf8')

describe('AdminDashboardTab 玩家详情发装布局', () => {
    it('玩家详情区包含装备模板下拉和数量输入', () => {
        expect(componentSource).toContain('player-detail-card')
        expect(componentSource).toContain('grant-form')
        expect(componentSource).toContain('grantDraft.itemId')
        expect(componentSource).toContain('grantDraft.quantity')
        expect(componentSource).toContain('equipmentOptions')
        expect(componentSource).toContain('发装备')
    })
})
