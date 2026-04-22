import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')
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
  })

  it('样式定义了六档稀有度与至臻动态文字效果', () => {
    expect(styleSource).toContain('.rarity-text--common')
    expect(styleSource).toContain('.rarity-text--supreme')
    expect(styleSource).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
