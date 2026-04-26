import { describe, expect, it } from 'vitest'

import { mergeClickFallbackState } from './clickResponse'

describe('clickResponse', () => {
  it('SSE 正常时只用最小响应，不覆盖现有用户态', () => {
    const current = {
      userStats: { nickname: '阿明', clickCount: 11 },
      boss: { id: 'boss-1', status: 'active', currentHp: 40, maxHp: 100 },
      bossLeaderboard: [{ rank: 1, nickname: '阿明', damage: 60 }],
      myBossStats: { nickname: '阿明', damage: 60 },
      recentRewards: [{ itemId: 'club', itemName: '木棒' }],
    }

    expect(
      mergeClickFallbackState(current, {
        button: { key: 'feel', count: 12 },
        delta: 1,
        critical: false,
      }),
    ).toEqual(current)
  })

  it('SSE 断连时会应用 HTTP 兜底字段', () => {
    const current = {
      userStats: { nickname: '阿明', clickCount: 11 },
      boss: { id: 'boss-1', status: 'active', currentHp: 40, maxHp: 100 },
      bossLeaderboard: [{ rank: 1, nickname: '阿明', damage: 60 }],
      myBossStats: { nickname: '阿明', damage: 60 },
      recentRewards: [{ itemId: 'club', itemName: '木棒' }],
    }

    expect(
      mergeClickFallbackState(current, {
        userStats: { nickname: '阿明', clickCount: 12 },
        boss: { id: 'boss-1', status: 'active', currentHp: 39, maxHp: 100 },
        bossLeaderboard: [{ rank: 1, nickname: '阿明', damage: 61 }],
        myBossStats: { nickname: '阿明', damage: 61 },
        recentRewards: [{ itemId: 'axe', itemName: '短斧' }],
      }),
    ).toEqual({
      userStats: { nickname: '阿明', clickCount: 12 },
      boss: { id: 'boss-1', status: 'active', currentHp: 39, maxHp: 100 },
      bossLeaderboard: [{ rank: 1, nickname: '阿明', damage: 61 }],
      myBossStats: { nickname: '阿明', damage: 61 },
      recentRewards: [{ itemId: 'axe', itemName: '短斧' }],
    })
  })
})
