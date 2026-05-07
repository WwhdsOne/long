import {describe, expect, it} from 'vitest'

import {buildPityProgress} from './progressionView'

describe('progressionView', () => {
    it('能返回普通保底进度文案和百分比', () => {
        expect(buildPityProgress(7)).toEqual({
            current: 7,
            threshold: 30,
            remaining: 24,
            percent: 23,
            label: '7 / 31',
        })
    })

    it('会把超过阈值的进度钳制到满值', () => {
        expect(buildPityProgress(40)).toEqual({
            current: 30,
            threshold: 30,
            remaining: 1,
            percent: 100,
            label: '30 / 31',
        })
    })
})
