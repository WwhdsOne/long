import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminPage.vue'), 'utf8')
const adminStateSource = readFileSync(path.resolve(currentDir, './admin/useAdminPage.js'), 'utf8')
const adminActionSource = readFileSync(path.resolve(currentDir, './admin/useAdminPageActions.js'), 'utf8')

describe('AdminPage 玩家详情发装接线', () => {
    it('后台容器把装备模板和发装动作传给看板页', () => {
        expect(pageSource).toContain("import AdminDashboardTab from '../components/admin/AdminDashboardTab.vue'")
        expect(pageSource).toContain(':equipment-options="admin.equipmentOptions"')
        expect(pageSource).toContain(':grant-player-equipment="admin.grantPlayerEquipment"')
    })

    it('后台状态层暴露装备模板选项给玩家详情页', () => {
        expect(adminStateSource).toContain('const equipmentOptions = computed(() => equipmentPage.value.items ?? [])')
        expect(adminStateSource).toContain('equipmentOptions,')
    })

    it('后台动作层提供给玩家发装备动作', () => {
        expect(adminActionSource).toContain('async function grantPlayerEquipment(')
        expect(adminActionSource).toContain('/api/admin/players/${encodeURIComponent(playerNickname)}/equipment')
    })
})
