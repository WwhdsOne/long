import { describe, expect, it } from 'vitest'

import { mergeBossState } from './bossState'

describe('mergeBossState', () => {
  it('同一只活动 Boss 收到更高血量的旧消息时，不会把血量回退', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 80,
      maxHp: 100,
    }
    const incoming = {
      id: 'boss-1',
      status: 'active',
      currentHp: 90,
      maxHp: 100,
    }

    expect(mergeBossState(current, incoming)).toEqual({
      ...current,
      currentHp: '80',
      maxHp: '100',
      parts: [],
    })
  })

  it('同一只活动 Boss 收到更低血量的新消息时，会继续向下更新', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 80,
      maxHp: 100,
    }
    const incoming = {
      id: 'boss-1',
      status: 'active',
      currentHp: 70,
      maxHp: 100,
    }

    expect(mergeBossState(current, incoming)).toEqual({
      ...incoming,
      currentHp: '70',
      maxHp: '100',
      parts: [],
    })
  })

  it('切换到下一只 Boss 时，即使血量更高也要接受', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 1,
      maxHp: 100,
    }
    const incoming = {
      id: 'boss-2',
      status: 'active',
      currentHp: 500,
      maxHp: 500,
    }

    expect(mergeBossState(current, incoming)).toEqual({
      ...incoming,
      currentHp: '500',
      maxHp: '500',
      parts: [],
    })
  })

  it('大整数血量用字符串比较时，也不会把旧消息回退覆盖成更高血量', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: '9223372036854775799',
      maxHp: '9223372036854775800',
    }
    const incoming = {
      id: 'boss-1',
      status: 'active',
      currentHp: '9223372036854775800',
      maxHp: '9223372036854775800',
    }

    expect(mergeBossState(current, incoming)).toEqual({
      ...current,
      parts: [],
    })
  })
})
