import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = [
  './PublicPage.vue',
  './BattlePage.vue',
  './ProfilePage.vue',
  './MessagesPage.vue',
  './publicPageState.js',
]
  .map((file) => readFileSync(path.resolve(currentDir, file), 'utf8'))
  .join('\n')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('PublicPage 稀有度展示', () => {
  it('boss 装备和英雄掉落区补充成长上限文案', () => {
    expect(pageSource).toContain('可强化')
    expect(pageSource).toContain('可觉醒')
  })

  it('页面通过统一稀有度工具渲染装备名称', () => {
    expect(pageSource).toContain("from '../utils/rarity'")
    expect(pageSource).toContain('splitEquipmentName')
    expect(pageSource).toContain('getRarityClassName')
    expect(pageSource).toContain('formatRarityLabel')
  })

  it('拆分后的页面从共享状态取得稀有度与掉落展示工具', () => {
    const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
    const profileSource = readFileSync(path.resolve(currentDir, './ProfilePage.vue'), 'utf8')
    const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

    expect(stateSource).toContain('    formatDropRate,')
    expect(battleSource).toContain('  formatDropRate,')
    expect(stateSource).toContain('    formatRarityLabel,')
    expect(battleSource).toContain('  formatRarityLabel,')
    expect(profileSource).toContain('  formatRarityLabel,')
  })

  it('样式定义了六档稀有度与至臻动态文字效果', () => {
    expect(styleSource).toContain('.rarity-text--common')
    expect(styleSource).toContain('.rarity-text--supreme')
    expect(styleSource).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
