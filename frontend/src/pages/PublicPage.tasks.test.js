import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const taskPageSource = readFileSync(path.resolve(currentDir, './TaskPage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('PublicPage 任务面板', () => {
  it('任务作为独立资料页签展示，并在可领取时显示红点提示', () => {
    const publicSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')
    expect(publicSource).toContain('hasClaimableTasks')
    expect(publicSource).toContain('public-nav__task-dot')
    expect(stateSource).toContain("id: 'tasks'")
    expect(stateSource).toContain("path: '/profile/tasks'")
  })

  it('任务页展示任务列表和领取按钮', () => {
    expect(taskPageSource).toContain('当前任务')
    expect(taskPageSource).toContain('claimTask')
    expect(taskPageSource).toContain('task-card')
  })

  it('个人资料态会承接任务列表并提供领取动作', () => {
    expect(stateSource).toContain('const tasks = ref([])')
    expect(stateSource).toContain('const hasClaimableTasks = computed')
    expect(stateSource).toContain("if ('tasks' in payload)")
    expect(stateSource).toContain("fetch(`/api/tasks/${encodeURIComponent(taskId)}/claim`")
  })

  it('登录后会拉取资料，并以 10 秒频率轮询任务列表更新红点', () => {
    expect(stateSource).toContain('const TASK_POLL_INTERVAL_MS = 10000')
    expect(stateSource).toContain('async function loadTasks()')
    expect(stateSource).toContain("const response = await fetch('/api/tasks')")
    expect(stateSource).toContain('startTaskPolling()')
    expect(stateSource).toContain('stopTaskPolling()')
    expect(stateSource).toContain('await loadPlayerProfile()')
  })
})
