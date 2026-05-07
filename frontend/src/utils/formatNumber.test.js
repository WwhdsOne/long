import {describe, expect, it} from 'vitest'

import {formatCompact, formatIntegerExact, ratioPercent} from './formatNumber'

describe('formatNumber 大整数支持', () => {
    it('大整数会优先使用到 Qi 为止的单位缩写', () => {
        expect(formatIntegerExact('2000')).toBe('2K')
        expect(formatIntegerExact('-1200')).toBe('-1.2K')
        expect(formatIntegerExact('1500000')).toBe('1.5M')
        expect(formatIntegerExact('9223372036854775800')).toBe('9.2Qi')
    })

    it('超过 Qi 范围后才回退到科学计数法', () => {
        expect(formatIntegerExact('999')).toBe('999')
        expect(formatIntegerExact('1000000000000000000000')).toBe('1e+21')
        expect(formatIntegerExact('-1234500000000000000000')).toBe('-1.2345e+21')
    })

    it('大整数可以计算血量百分比而不依赖 Number 精度', () => {
        expect(ratioPercent('9223372036854775799', '9223372036854775800')).toBe(99.99)
        expect(ratioPercent('4611686018427387900', '9223372036854775800')).toBe(50)
    })

    it('大整数可以做紧凑显示', () => {
        expect(formatCompact('9223372036854775800')).toBe('9223372036.8B')
    })
})
