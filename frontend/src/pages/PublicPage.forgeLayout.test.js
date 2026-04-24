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

function compact(source) {
  return source.replace(/\s+/g, ' ').trim()
}

describe('PublicPage 强化布局', () => {
  it('强化列表使用独立信息行，避免觉醒文案被右侧按钮挤窄', () => {
    expect(pageSource).toContain('class="forge-action-list__meta"')
    expect(compact(styleSource)).toContain('.forge-action-list li { display: grid; grid-template-columns: 1fr;')
  })

  it('移除了旧 3 合 1 与 reforge/pity 文案，改为 enhance 强化接口', () => {
    expect(pageSource).not.toContain('3 合 1 升星')
    expect(pageSource).not.toContain("/reforge'")
    expect(pageSource).not.toContain('reforgePityCounter')
    expect(pageSource).toContain("'enhance'")
  })

  it('强化面板补充了规则说明与按钮旁状态提示', () => {
    expect(pageSource).toContain('强化规则')
    expect(pageSource).toContain('三项基础属性等概率命中')
    expect(pageSource).toContain('暴击率每次固定 +0.20%')
    expect(pageSource).toContain('仅提升点击 / 暴击 / 暴击率中的一项')
    expect(pageSource).toContain('点击 / 暴击单次成长 = ceil((当前点击 + 当前暴击 + 当前暴击率) / 4)，至少 +1')
    expect(pageSource).toContain('原石不足')
    expect(pageSource).toContain('已达模板上限')
  })
})
