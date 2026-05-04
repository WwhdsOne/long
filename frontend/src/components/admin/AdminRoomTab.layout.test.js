import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const source = readFileSync(path.resolve(currentDir, './AdminRoomTab.vue'), 'utf8')

describe('AdminRoomTab 布局', () => {
  it('提供房间显示名编辑表单', () => {
    expect(source).toContain('displayName')
    expect(source).toContain('保存房间名')
    expect(source).toContain('房间 ID')
    expect(source).toContain('未命名时默认显示')
    expect(source).toContain('当前 Boss')
    expect(source).toContain('循环状态')
  })
})
