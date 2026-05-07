import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './ArmoryPage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('ArmoryPage 战斗属性与装备栏', () => {
    it('强化上限按当前稀有度规则展示', () => {
        expect(pageSource).toContain("case '优秀':\n      return 10")
        expect(pageSource).toContain("case '稀有':\n      return 15")
        expect(pageSource).toContain("case '史诗':\n      return 20")
        expect(pageSource).toContain("case '传说':\n      return 25")
        expect(pageSource).toContain("case '至臻':\n      return 35")
        expect(pageSource).toContain("default:\n      return 5")
    })

    it('战斗属性改为紧凑两列摘要，不再拆分基础属性和强化属性', () => {
        expect(pageSource).not.toContain('基础属性')
        expect(pageSource).not.toContain('强化属性')
        expect(pageSource).not.toContain('criticalCount')
        expect(pageSource).toContain('const combatStatSummaryItems = computed(() => [')
        expect(pageSource).toContain('class="armory-combat-summary"')
        expect(pageSource).toContain('class="armory-combat-summary__item"')
        expect(pageSource).toContain('formatArmorPenPercent')
        expect(pageSource).toContain('formatCritDamageBonus')
    })

    it('装备栏改为左右 3 + 3 的正方形槽位，并复用装备详情浮层', () => {
        expect(pageSource).toContain('const loadoutColumns = computed(() => [')
        expect(pageSource).toContain('class="loadout-grid loadout-grid--paired"')
        expect(pageSource).toContain('class="loadout-slot__visual"')
        expect(pageSource).toContain('class="loadout-slot__placeholder"')
        expect(pageSource).toContain('class="armory-item-tooltip"')
        expect(pageSource).toContain('item.imagePath')
    })

    it('强化弹窗支持滑条预览与批量提交', () => {
        expect(pageSource).toContain('type="range"')
        expect(pageSource).toContain('enhancePreviewStatRows')
        expect(pageSource).toContain('enhanceAffordableLevelsByStone')
        expect(pageSource).toContain('enhanceSelectedLevels')
        expect(pageSource).toContain('->')
        expect(stateSource).toContain('async function enhanceItem(instanceId, levels = 1)')
        expect(stateSource).toContain('JSON.stringify({nickname: nickname.value, levels})')
    })
})
