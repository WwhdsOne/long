import { describe, expect, it } from 'vitest'

import { emptyEquipmentForm, normalizeEquipmentPage, normalizeLootEntry } from './state'

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

  it('boss 掉落规范化会保留新装备属性字段', () => {
    const entry = normalizeLootEntry({ itemId: 'fire-ring', attackPower: 10, bossDamagePercent: 0.5 })
    expect(entry.attackPower).toBe(10)
    expect(entry.bossDamagePercent).toBe(0.5)
  })
})
