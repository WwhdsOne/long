import {describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './TalentsPage.vue'), 'utf8')

describe('TalentsPage 魔法树入口', () => {
    it('前端树配置包含第四系奥术潮汐', () => {
        expect(pageSource).toContain("magic: {")
        expect(pageSource).toContain("name: '奥术潮汐'")
        expect(pageSource).toContain("const trees = ['normal', 'armor', 'crit', 'magic']")
    })

    it('奥术潮汐按钮带独立动态样式', () => {
        expect(pageSource).toContain("talent-select__btn--magic")
        expect(pageSource).toContain("@keyframes magic-button-flow")
        expect(pageSource).toContain("@keyframes magic-button-sheen")
    })
})
