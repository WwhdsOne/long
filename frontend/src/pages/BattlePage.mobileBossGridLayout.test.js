import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('BattlePage 移动端 Boss 网格尺寸', () => {
  it('手机端进一步压缩 5x5 Boss 网格和格内信息，避免右侧区域过满难点', () => {
    expect(styleSource).toContain('@media (max-width: 640px) {')
    expect(styleSource).toContain('max-width: min(100%, 330px);')
    expect(styleSource).toContain('gap: 2px;')
    expect(styleSource).toContain('padding: 3px;')
    expect(styleSource).toContain('font-size: 0.4rem;')
    expect(styleSource).toContain('font-size: 0.54rem;')
    expect(styleSource).toContain('font-size: 0.36rem;')
    expect(styleSource).toContain('height: 4px;')
    expect(styleSource).toContain('width: clamp(16px, 30%, 28px);')
  })
})
