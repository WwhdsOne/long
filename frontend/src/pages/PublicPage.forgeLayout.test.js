import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

function compact(source) {
  return source.replace(/\s+/g, ' ').trim()
}

describe('PublicPage 强化布局', () => {
  it('强化列表使用独立信息行，避免觉醒文案被右侧按钮挤窄', () => {
    expect(pageSource).toContain('class="forge-action-list__meta"')
    expect(compact(styleSource)).toContain('.forge-action-list li { display: grid; grid-template-columns: 1fr;')
  })
})
