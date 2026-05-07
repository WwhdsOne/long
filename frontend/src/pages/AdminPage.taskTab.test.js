import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminPage.vue'), 'utf8')
const adminStateSource = readFileSync(path.resolve(currentDir, './admin/useAdminPage.js'), 'utf8')
const adminActionSource = readFileSync(path.resolve(currentDir, './admin/useAdminPageActions.js'), 'utf8')
const compactPageSource = pageSource.replace(/\s+/g, ' ')

describe('AdminPage 任务管理接线', () => {
    it('后台容器挂载任务 Tab 组件并提供入口按钮', () => {
        expect(pageSource).toContain("import AdminTaskTab from '../components/admin/AdminTaskTab.vue'")
        expect(pageSource).toContain("admin.activeTab === 'tasks'")
        expect(compactPageSource).toContain('>任务 </button>')
    })

    it('后台状态层提供任务列表和周期查询动作', () => {
        expect(adminStateSource).toContain('const taskDefinitions = ref([])')
        expect(adminStateSource).toContain('const taskArchives = ref([])')
        expect(adminStateSource).toContain('fetchTasks')
        expect(adminStateSource).toContain('fetchTaskArchives')
    })

    it('后台动作层支持保存、上下线、复制和归档任务', () => {
        expect(adminActionSource).toContain('async function saveTaskDefinition()')
        expect(adminActionSource).toContain('async function activateTaskDefinition(taskId)')
        expect(adminActionSource).toContain('async function deactivateTaskDefinition(taskId)')
        expect(adminActionSource).toContain('async function duplicateTaskDefinition(taskId)')
        expect(adminActionSource).toContain('async function archiveExpiredTasks()')
    })

    it('任务面板支持筛选和装备奖励下拉选择', () => {
        const tabSource = readFileSync(path.resolve(currentDir, '../components/admin/AdminTaskTab.vue'), 'utf8')
        expect(tabSource).toContain('taskStatusFilter')
        expect(tabSource).toContain('archiveStatusFilter')
        expect(tabSource).toContain('equipmentOptions')
        expect(tabSource).toContain('<select v-model="entry.itemId"')
    })

    it('任务面板使用行为类型和累计窗口模型', () => {
        const tabSource = readFileSync(path.resolve(currentDir, '../components/admin/AdminTaskTab.vue'), 'utf8')
        expect(tabSource).toContain('eventKind')
        expect(tabSource).toContain('windowKind')
        expect(tabSource).toContain("taskForm.windowKind === 'fixed_range'")
        expect(tabSource).toContain("value=\"lifetime\"")
        expect(tabSource).toContain('长期有效')
    })
})
