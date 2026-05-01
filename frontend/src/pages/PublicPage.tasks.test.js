import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const armorySource = readFileSync(path.resolve(currentDir, './ArmoryPage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('PublicPage 任务面板', () => {
  it('资料页资源区展示任务列表和领取按钮', () => {
    expect(armorySource).toContain('当前任务')
    expect(armorySource).toContain('claimTask')
    expect(armorySource).toContain('task-card')
  })

  it('个人资料态会承接任务列表并提供领取动作', () => {
    expect(stateSource).toContain('const tasks = ref([])')
    expect(stateSource).toContain("if ('tasks' in payload)")
    expect(stateSource).toContain("fetch(`/api/tasks/${encodeURIComponent(taskId)}/claim`")
  })
})
