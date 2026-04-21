import { describe, expect, it } from 'vitest'

import {
  buildCosmeticCollections,
  canEquipCosmeticSelection,
  resolveCosmeticEffectConfig,
  summarizeEquippedCosmetics,
} from './cosmetics'

const shopCatalog = [
  {
    cosmeticId: 'trail-ribbon',
    name: '流星彩带轨迹',
    type: 'trail',
    price: 30,
    owned: true,
    equipped: true,
    preview: {
      theme: 'ribbon',
    },
  },
  {
    cosmeticId: 'impact-ribbon',
    name: '流星彩带点击特效',
    type: 'impact',
    price: 30,
    owned: false,
    equipped: false,
    preview: {
      theme: 'ribbon',
    },
  },
  {
    cosmeticId: 'trail-confetti',
    name: '纸片庆典轨迹',
    type: 'trail',
    price: 30,
    owned: false,
    equipped: false,
    preview: {
      theme: 'confetti',
    },
  },
  {
    cosmeticId: 'impact-confetti',
    name: '纸片庆典点击特效',
    type: 'impact',
    price: 30,
    owned: false,
    equipped: false,
    preview: {
      theme: 'confetti',
    },
  },
  {
    cosmeticId: 'trail-stamp',
    name: '印章敲击轨迹',
    type: 'trail',
    price: 30,
    owned: false,
    equipped: false,
    preview: {
      theme: 'stamp',
    },
  },
  {
    cosmeticId: 'impact-stamp',
    name: '印章敲击点击特效',
    type: 'impact',
    price: 30,
    owned: false,
    equipped: false,
    preview: {
      theme: 'stamp',
    },
  },
  {
    cosmeticId: 'trail-firefly',
    name: '流萤追光轨迹',
    type: 'trail',
    price: 30,
    owned: true,
    equipped: false,
    preview: {
      theme: 'firefly',
    },
  },
  {
    cosmeticId: 'impact-firefly',
    name: '流萤追光点击特效',
    type: 'impact',
    price: 30,
    owned: true,
    equipped: false,
    preview: {
      theme: 'firefly',
    },
  },
]

describe('cosmetics utils', () => {
  it('按槽位拆分一期 8 个外观商品', () => {
    const collections = buildCosmeticCollections(shopCatalog)

    expect(collections.trails).toHaveLength(4)
    expect(collections.impacts).toHaveLength(4)
    expect(collections.trails[0].cosmeticId).toBe('trail-ribbon')
    expect(collections.impacts[3].cosmeticId).toBe('impact-firefly')
  })

  it('允许已拥有外观自由混搭，未拥有外观不能装备', () => {
    expect(
      canEquipCosmeticSelection(shopCatalog, {
        trailId: 'trail-ribbon',
        impactId: 'impact-firefly',
      }),
    ).toBe(true)

    expect(
      canEquipCosmeticSelection(shopCatalog, {
        trailId: 'trail-confetti',
        impactId: 'impact-firefly',
      }),
    ).toBe(false)
  })

  it('自动挂机与星光按钮会降级外观特效', () => {
    const autoConfig = resolveCosmeticEffectConfig(
      shopCatalog,
      { trailId: 'trail-ribbon', impactId: 'impact-firefly' },
      { mode: 'auto', starlight: false },
    )
    const starlightConfig = resolveCosmeticEffectConfig(
      shopCatalog,
      { trailId: 'trail-ribbon', impactId: 'impact-firefly' },
      { mode: 'normal', starlight: true },
    )

    expect(autoConfig.particleCount).toBeLessThan(starlightConfig.particleCount)
    expect(starlightConfig.suppressed).toBe(true)
    expect(starlightConfig.durationMs).toBeLessThan(autoConfig.durationMs)
  })

  it('能返回当前搭配的轨迹和点击特效名称', () => {
    const summary = summarizeEquippedCosmetics(shopCatalog, {
      trailId: 'trail-ribbon',
      impactId: 'impact-firefly',
    })

    expect(summary.trailName).toBe('流星彩带轨迹')
    expect(summary.impactName).toBe('流萤追光点击特效')
  })
})
