import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('战利品弹窗图片链路', () => {
  it('保留奖励记录自带的图片字段，并优先用于弹窗渲染', () => {
    expect(stateSource).toContain("imagePath: String(item.imagePath || '').trim()")
    expect(stateSource).toContain("imageAlt: String(item.imageAlt || '').trim()")
    expect(stateSource).toContain("imagePath: reward.imagePath || rewardIconForItem(reward.itemId)")
    expect(stateSource).toContain("imageAlt: reward.imageAlt || reward.itemName || reward.itemId || '装备图标'")
  })
})
