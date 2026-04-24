import { describe, expect, it } from 'vitest'

import { buildFingerprintProof, createClickBehaviorTracker } from './manualClickSignals'

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

  it('会优先使用事件时间戳计算按压时长', () => {
    const tracker = createClickBehaviorTracker({
      now: () => 0,
    })

    tracker.handlePressStart('feel', {
      pointerType: 'mouse',
      clientX: 10,
      clientY: 10,
      timeStamp: 100.2,
    })
    tracker.handlePressEnd('feel', {
      pointerType: 'mouse',
      clientX: 10,
      clientY: 10,
      timeStamp: 112.8,
    })

    expect(tracker.consume('feel')).toMatchObject({
      pointerType: 'mouse',
      pressDurationMs: 13,
    })
  })

  it('不会把事件时间戳和当前时钟混用成过期记录', () => {
    const tracker = createClickBehaviorTracker({
      now: () => 100000,
    })

    tracker.handlePressStart('feel', {
      pointerType: 'mouse',
      clientX: 10,
      clientY: 10,
      timeStamp: 100.2,
    })
    tracker.handlePressEnd('feel', {
      pointerType: 'mouse',
      clientX: 10,
      clientY: 10,
      timeStamp: 112.8,
    })

    expect(tracker.consume('feel')).not.toBeNull()
  })
})
