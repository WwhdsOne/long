import { describe, expect, it } from 'vitest'

import { resolveStarlightRefreshPlan } from './starlightRefresh'

describe('resolveStarlightRefreshPlan', () => {
  it('对已过期且重复的 endsAt 只安排一次兜底刷新', () => {
    const firstPlan = resolveStarlightRefreshPlan({
      endsAtSeconds: 100,
      nowMs: 100_500,
      lastExpiredEndsAtSeconds: 0,
    })

    expect(firstPlan).toEqual({
      delayMs: 1000,
      expiredRetryEndsAtSeconds: 100,
    })

    const repeatedPlan = resolveStarlightRefreshPlan({
      endsAtSeconds: 100,
      nowMs: 100_800,
      lastExpiredEndsAtSeconds: firstPlan.expiredRetryEndsAtSeconds,
    })

    expect(repeatedPlan).toEqual({
      delayMs: null,
      expiredRetryEndsAtSeconds: 100,
    })
  })

  it('对未过期的 endsAt 按窗口结束时间正常调度', () => {
    expect(
      resolveStarlightRefreshPlan({
        endsAtSeconds: 200,
        nowMs: 150_000,
      }),
    ).toEqual({
      delayMs: 50_150,
      expiredRetryEndsAtSeconds: 0,
    })
  })
})
