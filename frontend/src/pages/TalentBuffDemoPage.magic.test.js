import {describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './TalentBuffDemoPage.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('TalentBuffDemoPage 魔法状态演示', () => {
    it('页面包含奥术裂解和魔法进度状态块', () => {
        expect(pageSource).toContain('奥术裂解')
        expect(pageSource).toContain('魔法触发率')
        expect(pageSource).toContain('回响层数')
    })

    it('样式包含魔法状态列表专用皮肤', () => {
        expect(styleSource).toContain('.talent-status-chip--magic')
        expect(styleSource).toContain('.part-progress-panel__bar-fill--magic')
        expect(styleSource).toContain('.arcane-rupture-panel')
    })
})
