import { describe, expect, it } from 'vitest'

import { formatCompact, formatIntegerExact, ratioPercent } from './formatNumber'

describe('formatNumber 大整数支持', () => {
  it('大整数可以按原值带千分位显示', () => {
    expect(formatIntegerExact('9223372036854775800')).toBe('9,223,372,036,854,775,800')
  })

  it('大整数可以计算血量百分比而不依赖 Number 精度', () => {
    expect(ratioPercent('9223372036854775799', '9223372036854775800')).toBe(99.99)
    expect(ratioPercent('4611686018427387900', '9223372036854775800')).toBe(50)
  })

  it('大整数可以做紧凑显示', () => {
    expect(formatCompact('9223372036854775800')).toBe('9223372036.8B')
  })
})
