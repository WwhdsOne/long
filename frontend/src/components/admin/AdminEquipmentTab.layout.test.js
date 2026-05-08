import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const componentSource = readFileSync(path.resolve(currentDir, './AdminEquipmentTab.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../../style.css'), 'utf8')

describe('AdminEquipmentTab 布局与草稿生成入口', () => {
    it('右上角提供新增装备按钮并按状态显示编辑区', () => {
        expect(componentSource).toContain('新增装备')
        expect(componentSource).toContain('openNewEquipment')
        expect(componentSource).toContain('showEquipmentEditor')
    })

    it('提供自然语言生成草稿入口且不会绑定保存动作', () => {
        expect(componentSource).toContain('equipmentPrompt')
        expect(componentSource).toContain('generateEquipmentDraft')
        expect(componentSource).toContain('生成草稿')
        expect(componentSource).toContain('textarea')
    })

    it('表单使用左右标签式字段并列表使用 5 列响应式网格', () => {
        expect(componentSource).toContain('admin-labeled-field')
        expect(componentSource).toContain('inventory-list--equipment-grid')
        expect(styleSource).toContain('grid-template-columns: repeat(5, minmax(0, 1fr))')
        expect(styleSource).toContain('@media (max-width: 900px)')
    })

    it('新增装备表单移除描述和天赋，改为魔法词条', () => {
        expect(componentSource).not.toContain('装备描述')
        expect(componentSource).not.toContain('talentAffinity')
        expect(componentSource).toContain('magicProcRateBonus')
        expect(componentSource).toContain('magicDamageBonus')
        expect(componentSource).toContain('魔法触发')
        expect(componentSource).toContain('魔法增伤')
    })
})
