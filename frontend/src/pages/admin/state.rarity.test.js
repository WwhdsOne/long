import { describe, expect, it } from 'vitest'

import {
  EQUIPMENT_SLOTS,
  emptyEquipmentForm,
  normalizeBossTemplate,
  normalizeEquipmentPage,
  normalizeLoadout,
  normalizeLootEntry,
} from './state'

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
    const entry = normalizeLootEntry({ itemId: 'fire-ring', dropRatePercent: 35, attackPower: 10, critRate: 0.22, bossDamagePercent: 0.5 })
    expect(entry.dropRatePercent).toBe(35)
    expect(entry.attackPower).toBe(10)
    expect(entry.critRate).toBe(0.22)
    expect(entry.bossDamagePercent).toBe(0.5)
  })

  it('boss 模板规范化会保留金币与强化石奖励字段', () => {
    const template = normalizeBossTemplate({
      id: 'dragon',
      name: '火龙',
      maxHp: 9999,
      goldOnKill: 5000,
      stoneOnKill: 120,
      talentPointsOnKill: 88,
    })

    expect(template.goldOnKill).toBe(5000)
    expect(template.stoneOnKill).toBe(120)
    expect(template.talentPointsOnKill).toBe(88)
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
