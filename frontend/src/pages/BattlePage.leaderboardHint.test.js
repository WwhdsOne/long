import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const source = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

describe('BattlePage 点击总榜说明', () => {
  it('展示每分钟整点更新一次的提示', () => {
    expect(source).toContain('点击总榜每分钟整点更新一次')
  })
})
