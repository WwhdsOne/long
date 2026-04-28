import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './TalentsPage.vue'), 'utf8')

describe('TalentsPage 前置规则移除', () => {
  it('前端不再显示或校验单节点前置', () => {
    expect(pageSource).not.toContain('isPrerequisiteMet')
    expect(pageSource).not.toContain('prerequisiteLabel')
    expect(pageSource).not.toContain('前置：')
    expect(pageSource).not.toContain('前置未满足')
    expect(pageSource).not.toContain('需要先学习前置天赋')
  })
})
