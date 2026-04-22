import { describe, expect, it } from 'vitest'

import { emptyEquipmentForm, normalizeEquipmentPage, normalizeLootEntry, normalizeHeroLootEntry } from './state'

describe('admin state rarity normalization', () => {
  it('空装备表单默认使用普通稀有度', () => {
    expect(emptyEquipmentForm().rarity).toBe('普通')
  })

  it('旧装备分页数据会补默认稀有度', () => {
    const page = normalizeEquipmentPage({
      items: [{ itemId: 'wood-sword', name: '木剑', slot: 'weapon' }],
    })

    expect(page.items[0].rarity).toBe('普通')
  })

  it('boss 掉落规范化会保留强化与觉醒上限字段', () => {
    expect(normalizeLootEntry({ itemId: 'fire-ring', enhanceCap: 7 }).enhanceCap).toBe(7)
    expect(normalizeHeroLootEntry({ heroId: 'spark-cat', awakenCap: 5 }).awakenCap).toBe(5)
  })
})
