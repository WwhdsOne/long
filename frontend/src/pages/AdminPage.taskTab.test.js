import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminPage.vue'), 'utf8')
const adminStateSource = readFileSync(path.resolve(currentDir, './admin/useAdminPage.js'), 'utf8')
const adminActionSource = readFileSync(path.resolve(currentDir, './admin/useAdminPageActions.js'), 'utf8')

describe('AdminPage 任务管理接线', () => {
  it('后台容器挂载任务 Tab 组件并提供入口按钮', () => {
    expect(pageSource).toContain("import AdminTaskTab from '../components/admin/AdminTaskTab.vue'")
    expect(pageSource).toContain("admin.activeTab === 'tasks'")
    expect(pageSource).toContain('>任务<')
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
})
