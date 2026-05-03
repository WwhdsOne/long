import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const source = readFileSync(path.resolve(currentDir, './RoomSelector.vue'), 'utf8')

describe('RoomSelector 紧凑卡片', () => {
  it('房间卡补充 Boss 均血并精简布局', () => {
    expect(source).toContain('function roomAvgHpText(room) {')
    expect(source).toContain('<small>均血</small>')
    expect(source).toContain('<span class="room-selector__action">{{ roomActionLabel(room) }}</span>')
    expect(source).toContain('grid-template-columns: repeat(3, minmax(0, 1fr));')
    expect(source).toContain('min-height: 74px;')
    expect(source).toContain('padding: 8px;')
  })
})
