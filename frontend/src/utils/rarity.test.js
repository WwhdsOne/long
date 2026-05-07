import {describe, expect, it} from 'vitest'

import {DEFAULT_RARITY, formatRarityLabel, getRarityClassName, normalizeRarity, splitEquipmentName,} from './rarity'

describe('rarity utils', () => {
    it('会把未知稀有度兜底到普通，并保留六档有效值', () => {
        expect(DEFAULT_RARITY).toBe('普通')
        expect(normalizeRarity('')).toBe('普通')
        expect(normalizeRarity('神话')).toBe('普通')
        expect(normalizeRarity('至臻')).toBe('至臻')
    })

    it('会把 emoji 前缀和可着色文字拆开', () => {
        expect(splitEquipmentName('🗡 木剑')).toEqual({
            prefix: '🗡 ',
            text: '木剑',
        })
        expect(splitEquipmentName('烈焰戒')).toEqual({
            prefix: '',
            text: '烈焰戒',
        })
    })

    it('至臻会返回动态文字 class，其他稀有度返回静态 class', () => {
        expect(getRarityClassName('普通')).toContain('rarity-text--common')
        expect(getRarityClassName('传说')).toContain('rarity-text--legendary')
        expect(getRarityClassName('至臻')).toContain('rarity-text--supreme')
        expect(getRarityClassName('至臻')).toContain('rarity-text--animated')
    })

    it('只输出稀有度档位名，字段名由调用处负责', () => {
        expect(formatRarityLabel('传说')).toBe('传说')
        expect(formatRarityLabel('神话')).toBe('普通')
    })
})
