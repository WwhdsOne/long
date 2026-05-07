import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const componentSource = readFileSync(path.resolve(currentDir, './AdminBossTab.vue'), 'utf8')
const actionSource = readFileSync(path.resolve(currentDir, '../../pages/admin/useAdminPageActions.js'), 'utf8')

describe('AdminBossTab 部位血量口径', () => {
    it('Boss 总血量只读展示并由部位最大血量合计决定', () => {
        expect(componentSource).toContain('bossPartTotalHp')
        expect(componentSource).toContain('sumBossPartMaxHp(props.bossForm.layout)')
        expect(componentSource).toContain(':value="bossPartTotalHp"')
        expect(componentSource).toContain('readonly')
        expect(actionSource).toContain('sumBossPartMaxHp')
        expect(actionSource).toContain('maxHp: sumBossPartMaxHp(bossForm.value.layout)')
        expect(componentSource).not.toContain('Number(part?.maxHp ?? 0)')
        expect(actionSource).not.toContain('Number(part?.maxHp ?? 0)')
    })

    it('Boss 部位编辑器支持部位名称和小图路径', () => {
        expect(componentSource).toContain('selectedCell.displayName')
        expect(componentSource).toContain('selectedCell.imagePath')
        expect(componentSource).toContain('<span>名称</span>')
        expect(componentSource).toContain('<span>图片</span>')
        expect(componentSource).toContain('normalizeBossPartCell')
        expect(componentSource).toContain('inputmode="numeric"')
        expect(componentSource).not.toContain('v-model="selectedCell.maxHp" class="nickname-form__input" type="number"')
    })

    it('模板掉落池直接填写掉落几率，不再填写权重', () => {
        expect(componentSource).toContain('entry.dropRatePercent')
        expect(componentSource).toContain('placeholder="掉落几率 %"')
        expect(componentSource).not.toContain('entry.weight')
        expect(componentSource).not.toContain('权重')
        expect(actionSource).toContain('dropRatePercent: Number(entry.dropRatePercent)')
        expect(actionSource).not.toContain('weight: Number(entry.weight)')
    })

    it('保存模板时先快照掉落行，避免刷新状态覆盖后导致掉落保存丢失', () => {
        expect(componentSource).toContain('const lootSnapshot = props.lootRows.map')
        expect(componentSource).toContain('await props.saveLoot(lootSnapshot)')
        expect(actionSource).toContain('async function saveLoot(lootRowsOverride = null)')
        expect(actionSource).toContain('const rowsToSave = Array.isArray(lootRowsOverride) ? lootRowsOverride : lootRows.value')
    })
})
