import { describe, expect, it } from 'vitest'

import { formatDropRate } from './buttonBoard'

describe('buttonBoard', () => {
  it('会把概率格式化成固定百分比文本', () => {
    expect(formatDropRate(25)).toBe('25%')
    expect(formatDropRate(33.3333)).toBe('33.33%')
    expect(formatDropRate(0)).toBe('0%')
  })
})
