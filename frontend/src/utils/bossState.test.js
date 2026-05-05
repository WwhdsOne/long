import { describe, expect, it } from 'vitest'

import { applyBossPartStateDeltas, mergeBossState } from './bossState'

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

  it('同一只活动 Boss 总血量回退保护生效时，仍会接收部位血量向下变化', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 80,
      maxHp: 100,
      parts: [
        {x: 0, y: 0, currentHp: 40, maxHp: 50, armor: 3, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 20, maxHp: 50, armor: 8, alive: true, type: 'heavy'},
      ],
    }
    const incoming = {
      id: 'boss-1',
      status: 'active',
      currentHp: 90,
      maxHp: 100,
      parts: [
        {x: 0, y: 0, currentHp: 40, maxHp: 50, armor: 3, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 10, maxHp: 50, armor: 8, alive: true, type: 'heavy'},
      ],
    }

    expect(mergeBossState(current, incoming)).toEqual({
      id: 'boss-1',
      status: 'active',
      currentHp: '80',
      maxHp: '100',
      parts: [
        {x: 0, y: 0, currentHp: '40', maxHp: '50', armor: '3', alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: '10', maxHp: '50', armor: '8', alive: true, type: 'heavy'},
      ],
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

  it('同一只活动 Boss 的后续增量没带 parts 时，保留现有部位数据', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 80,
      maxHp: 100,
      parts: [
        {x: 0, y: 0, currentHp: 40, maxHp: 50, armor: 3, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 20, maxHp: 50, armor: 8, alive: true, type: 'heavy'},
      ],
    }
    const incoming = {
      id: 'boss-1',
      status: 'active',
      currentHp: 60,
      maxHp: 100,
    }

    expect(mergeBossState(current, incoming)).toEqual({
      ...incoming,
      currentHp: '60',
      maxHp: '100',
      parts: [
        {x: 0, y: 0, currentHp: '40', maxHp: '50', armor: '3', alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: '20', maxHp: '50', armor: '8', alive: true, type: 'heavy'},
      ],
    })
  })

  it('同一只活动 Boss 的后续增量只带部分 parts 时，按坐标合并并保留未更新部位', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 80,
      maxHp: 100,
      parts: [
        {x: 0, y: 0, currentHp: 40, maxHp: 50, armor: 3, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 20, maxHp: 50, armor: 8, alive: true, type: 'heavy'},
      ],
    }
    const incoming = {
      id: 'boss-1',
      status: 'active',
      currentHp: 60,
      maxHp: 100,
      parts: [
        {x: 1, y: 0, currentHp: 10, maxHp: 50, armor: 6, alive: true, type: 'heavy'},
      ],
    }

    expect(mergeBossState(current, incoming)).toEqual({
      ...incoming,
      currentHp: '60',
      maxHp: '100',
      parts: [
        {x: 0, y: 0, currentHp: '40', maxHp: '50', armor: '3', alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: '10', maxHp: '50', armor: '6', alive: true, type: 'heavy'},
      ],
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

describe('applyBossPartStateDeltas', () => {
  it('click_ack 的部位增量会同步部位血量和 Boss 总血量', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 80,
      maxHp: 100,
      parts: [
        {x: 0, y: 0, currentHp: 40, maxHp: 50, armor: 3, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 40, maxHp: 50, armor: 8, alive: true, type: 'heavy'},
      ],
    }

    expect(applyBossPartStateDeltas(current, [
      {x: 1, y: 0, beforeHp: 40, afterHp: 31, damage: 9, partType: 'heavy'},
    ])).toEqual({
      id: 'boss-1',
      status: 'active',
      currentHp: '71',
      maxHp: '100',
      parts: [
        {x: 0, y: 0, currentHp: '40', maxHp: '50', armor: '3', alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: '31', maxHp: '50', armor: '8', alive: true, type: 'heavy'},
      ],
    })
  })

  it('过时的部位增量不会把血量回抬', () => {
    const current = {
      id: 'boss-1',
      status: 'active',
      currentHp: 71,
      maxHp: 100,
      parts: [
        {x: 0, y: 0, currentHp: 40, maxHp: 50, armor: 3, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 31, maxHp: 50, armor: 8, alive: true, type: 'heavy'},
      ],
    }

    expect(applyBossPartStateDeltas(current, [
      {x: 1, y: 0, beforeHp: 40, afterHp: 35, damage: 5, partType: 'heavy'},
    ])).toEqual({
      id: 'boss-1',
      status: 'active',
      currentHp: '71',
      maxHp: '100',
      parts: [
        {x: 0, y: 0, currentHp: '40', maxHp: '50', armor: '3', alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: '31', maxHp: '50', armor: '8', alive: true, type: 'heavy'},
      ],
    })
  })

  it('旧 Boss 的击杀增量晚到时，不会把新 Boss 同坐标部位错误打成残血或死亡', () => {
    const current = {
      id: 'boss-2',
      status: 'active',
      currentHp: 200,
      maxHp: 200,
      parts: [
        {x: 0, y: 0, currentHp: 100, maxHp: 100, armor: 0, alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: 100, maxHp: 100, armor: 0, alive: true, type: 'heavy'},
      ],
    }

    expect(applyBossPartStateDeltas(current, [
      {x: 1, y: 0, beforeHp: 8, afterHp: 0, damage: 8, partType: 'heavy'},
    ])).toEqual({
      id: 'boss-2',
      status: 'active',
      currentHp: '200',
      maxHp: '200',
      parts: [
        {x: 0, y: 0, currentHp: '100', maxHp: '100', armor: '0', alive: true, type: 'soft'},
        {x: 1, y: 0, currentHp: '100', maxHp: '100', armor: '0', alive: true, type: 'heavy'},
      ],
    })
  })
})
