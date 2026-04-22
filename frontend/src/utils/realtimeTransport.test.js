import { describe, expect, it, vi } from 'vitest'

import { createRealtimeTransport } from './realtimeTransport'

class FakeWebSocket {
  constructor(url) {
    this.url = url
    this.sent = []
    this.closed = false
    this.onopen = null
    this.onmessage = null
    this.onerror = null
    this.onclose = null
  }

  send(payload) {
    this.sent.push(payload)
  }

  close() {
    this.closed = true
  }

  emitOpen() {
    this.onopen?.()
  }

  emitMessage(payload) {
    this.onmessage?.({ data: JSON.stringify(payload) })
  }

  emitError() {
    this.onerror?.(new Error('socket failed'))
  }

  emitClose() {
    this.onclose?.()
  }
}

class FakeEventSource {
  constructor(url) {
    this.url = url
    this.closed = false
    this.onopen = null
    this.onerror = null
    this.listeners = new Map()
  }

  addEventListener(name, listener) {
    this.listeners.set(name, listener)
  }

  close() {
    this.closed = true
  }

  emitOpen() {
    this.onopen?.()
  }

  emitError() {
    this.onerror?.(new Error('event source failed'))
  }

  emitEvent(name, payload) {
    this.listeners.get(name)?.({ data: JSON.stringify(payload) })
  }
}

describe('realtimeTransport', () => {
  it('WebSocket snapshot 会初始化页面状态并更新连接状态', () => {
    const snapshots = []
    const states = []
    const sockets = []
    const transport = createRealtimeTransport({
      createWebSocket(url) {
        const socket = new FakeWebSocket(url)
        sockets.push(socket)
        return socket
      },
      createEventSource() {
        throw new Error('should not create event source')
      },
      onSnapshot(publicState, userState) {
        snapshots.push({ publicState, userState })
      },
      onTransportState(nextState) {
        states.push(nextState)
      },
    })

    transport.connect({ nickname: '阿明' })
    sockets[0].emitOpen()
    sockets[0].emitMessage({
      type: 'snapshot',
      public: {
        buttons: [{ key: 'feel', count: 3 }],
      },
      user: {
        userStats: { nickname: '阿明', clickCount: 3 },
      },
    })

    expect(sockets[0].sent).toEqual([
      JSON.stringify({ type: 'hello', nickname: '阿明' }),
    ])
    expect(snapshots).toEqual([
      {
        publicState: {
          buttons: [{ key: 'feel', count: 3 }],
        },
        userState: {
          userStats: { nickname: '阿明', clickCount: 3 },
        },
      },
    ])
    expect(states.at(-1)).toEqual({
      connected: true,
      degraded: false,
      mode: 'ws',
    })
  })

  it('click_ack 走 WebSocket 发送点击并回调最小反馈', () => {
    const sockets = []
    const clickAcks = []
    const transport = createRealtimeTransport({
      createWebSocket(url) {
        const socket = new FakeWebSocket(url)
        sockets.push(socket)
        return socket
      },
      createEventSource() {
        throw new Error('should not create event source')
      },
      onClickAck(payload) {
        clickAcks.push(payload)
      },
    })

    transport.connect({ nickname: '阿明' })
    sockets[0].emitOpen()

    expect(transport.sendClick('feel')).toBe(true)
    expect(sockets[0].sent.at(-1)).toBe(JSON.stringify({ type: 'click', slug: 'feel' }))

    sockets[0].emitMessage({
      type: 'click_ack',
      payload: {
        button: { key: 'feel', count: 4 },
        delta: 1,
        critical: false,
      },
    })

    expect(clickAcks).toEqual([
      {
        button: { key: 'feel', count: 4 },
        delta: 1,
        critical: false,
      },
    ])
  })

  it('WebSocket 断开后会自动退回 SSE 并继续消费增量事件', () => {
    const sockets = []
    const sources = []
    const publicDeltas = []
    const states = []
    const errors = []
    const transport = createRealtimeTransport({
      createWebSocket(url) {
        const socket = new FakeWebSocket(url)
        sockets.push(socket)
        return socket
      },
      createEventSource(url) {
        const source = new FakeEventSource(url)
        sources.push(source)
        return source
      },
      onPublicDelta(payload) {
        publicDeltas.push(payload)
      },
      onTransportState(nextState) {
        states.push(nextState)
      },
      onTransportError(message) {
        errors.push(message)
      },
    })

    transport.connect({ nickname: '阿明' })
    sockets[0].emitOpen()
    sockets[0].emitClose()

    expect(sources).toHaveLength(1)
    expect(sources[0].url).toBe('/api/events?nickname=%E9%98%BF%E6%98%8E')
    sources[0].emitOpen()
    sources[0].emitEvent('public_state', {
      buttons: [{ key: 'feel', count: 5 }],
    })

    expect(publicDeltas).toEqual([
      {
        buttons: [{ key: 'feel', count: 5 }],
      },
    ])
    expect(states.at(-1)).toEqual({
      connected: true,
      degraded: true,
      mode: 'sse',
    })
    expect(errors.at(-1)).toContain('兼容模式')
  })
})
