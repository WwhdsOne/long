import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { AUTO_CLICK_INTERVAL_MS, createAutoClickLoop } from './autoClicker'

describe('createAutoClickLoop', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('会按固定间隔持续触发', () => {
    const onTick = vi.fn()
    const loop = createAutoClickLoop({ onTick })

    loop.start()

    vi.advanceTimersByTime(AUTO_CLICK_INTERVAL_MS - 1)
    expect(onTick).not.toHaveBeenCalled()

    vi.advanceTimersByTime(1)
    expect(onTick).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(AUTO_CLICK_INTERVAL_MS)
    expect(onTick).toHaveBeenCalledTimes(2)
  })

  it('重复开启不会创建并行定时器', () => {
    const onTick = vi.fn()
    const loop = createAutoClickLoop({ onTick })

    loop.start()
    loop.start()

    vi.advanceTimersByTime(AUTO_CLICK_INTERVAL_MS * 2)
    expect(onTick).toHaveBeenCalledTimes(2)
  })

  it('停止后不会继续触发', () => {
    const onTick = vi.fn()
    const loop = createAutoClickLoop({ onTick })

    loop.start()
    vi.advanceTimersByTime(AUTO_CLICK_INTERVAL_MS)
    loop.stop()
    vi.advanceTimersByTime(AUTO_CLICK_INTERVAL_MS * 3)

    expect(onTick).toHaveBeenCalledTimes(1)
    expect(loop.isRunning()).toBe(false)
  })
})
