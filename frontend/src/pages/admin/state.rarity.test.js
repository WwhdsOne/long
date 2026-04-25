import { describe, expect, it } from 'vitest'

import { EQUIPMENT_SLOTS, emptyEquipmentForm, normalizeEquipmentPage, normalizeLoadout, normalizeLootEntry } from './state'

describe('admin state rarity normalization', () => {
  it('空装备表单默认使用普通稀有度', () => {
    expect(emptyEquipmentForm().rarity).toBe('普通')
  })

  it('旧装备分页数据会补默认稀有度', () => {
    const page = normalizeEquipmentPage({
      items: [{ itemId: 'wood-sword', name: '木剑', slot: '武器' }],
    })

    expect(page.items[0].rarity).toBe('普通')
    expect(page.items[0].slot).toBe('weapon')
  })

  it('boss 掉落规范化会保留新装备属性字段', () => {
    const entry = normalizeLootEntry({ itemId: 'fire-ring', attackPower: 10, bossDamagePercent: 0.5 })
    expect(entry.attackPower).toBe(10)
    expect(entry.bossDamagePercent).toBe(0.5)
  })

  it('装备槽位使用策划案里的六部位，并把旧 armor 兼容为胸甲', () => {
    expect(EQUIPMENT_SLOTS.map((slot) => slot.value)).toEqual([
      'weapon',
      'helmet',
      'chest',
      'gloves',
      'legs',
      'accessory',
    ])

    const loadout = normalizeLoadout({
      weapon: { itemId: 'star-hammer' },
      armor: { itemId: 'old-armor' },
      gloves: { itemId: 'star-gloves' },
    })

    expect(loadout.chest?.itemId).toBe('old-armor')
    expect(loadout.gloves?.itemId).toBe('star-gloves')
    expect(loadout.armor).toBeUndefined()
  })
})
