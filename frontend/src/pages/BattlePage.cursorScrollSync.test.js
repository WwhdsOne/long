import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

describe('BattlePage Boss 光标滚动同步', () => {
  it('在滚动或视口变化后使 boss 网格 rect 缓存失效，避免旧坐标继续参与计算', () => {
    expect(battleSource).toContain('function invalidateBossGridRect() {')
    expect(battleSource).toContain('bossGridRect = null')
    expect(battleSource).toContain("window.addEventListener('scroll', invalidateBossGridRect, {passive: true})")
    expect(battleSource).toContain("window.addEventListener('resize', invalidateBossGridRect)")
    expect(battleSource).toContain("window.removeEventListener('scroll', invalidateBossGridRect)")
    expect(battleSource).toContain("window.removeEventListener('resize', invalidateBossGridRect)")
  })

  it('每次更新 Boss 光标位置前都重测当前网格 rect，确保 client 坐标基准与页面滚动同步', () => {
    expect(battleSource).toContain('const rect = measureBossGridRect()')
    expect(battleSource).not.toContain('const rect = bossGridRect || measureBossGridRect()')
  })
})
