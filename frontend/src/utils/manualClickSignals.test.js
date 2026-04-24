import { describe, expect, it } from 'vitest'

import { buildFingerprintProof, summarizePointerTrajectory } from './manualClickSignals'

describe('manualClickSignals', () => {
  it('会生成稳定的指纹挑战证明', async () => {
    const proof = await buildFingerprintProof({
      fingerprintHash: 'fp-1',
      ticket: 'ticket-1',
      challengeNonce: 'nonce-1',
    })

    expect(proof).toHaveLength(64)
    expect(proof).toMatch(/^[0-9a-f]+$/)
  })

  it('会计算轨迹距离 曲率和速度方差', () => {
    const summary = summarizePointerTrajectory([
      { x: 10, y: 10, t: 0 },
      { x: 13, y: 12, t: 30 },
      { x: 17, y: 18, t: 70 },
      { x: 22, y: 21, t: 120 },
    ])

    expect(summary.pointCount).toBe(4)
    expect(summary.pathDistance).toBeGreaterThan(10)
    expect(summary.displacement).toBeGreaterThan(2)
    expect(summary.curvature).toBeGreaterThan(0)
    expect(summary.speedVariance).toBeGreaterThan(0)
  })
})
