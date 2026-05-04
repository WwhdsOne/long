import { describe, expect, it, vi } from 'vitest'

import { encodeRealtimeClickRequest, realtimeBinaryType } from './realtimeProto'
import { createRealtimeTransport } from './realtimeTransport'
import { realtime } from '../proto/realtime.js'

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
    this.onmessage?.({
      data: typeof payload === 'string' ? payload : JSON.stringify(payload),
    })
  }

  emitBinary(payload) {
    this.onmessage?.({ data: payload })
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

  it('public_delta 没有 leaderboard 字段时不会把已有榜单覆盖成空数组', () => {
    const publicDeltas = []
    const sockets = []
    const transport = createRealtimeTransport({
      createWebSocket(url) {
        const socket = new FakeWebSocket(url)
        sockets.push(socket)
        socket.emitOpen()
        return socket
      },
      createEventSource() {
        throw new Error('should not create event source')
      },
      onPublicDelta(payload) {
        publicDeltas.push(payload)
      },
    })

    transport.connect({ nickname: '阿明' })

    const encoded = realtime.PublicDelta.encode(realtime.PublicDelta.create({
      totalVotes: 10,
      roomId: 'hall',
    })).finish()
    const frame = new Uint8Array(1 + encoded.length)
    frame[0] = realtimeBinaryType.publicDelta
    frame.set(encoded, 1)

    sockets[0].emitBinary(frame.buffer)

    expect(publicDeltas).toEqual([
      {
        totalVotes: 10,
        roomId: 'hall',
        leaderboard: [],
        bossLeaderboard: [],
      },
    ])
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
    expect(sockets[0].sent.at(-1)).toEqual(encodeRealtimeClickRequest({
      slug: 'feel',
      comboCount: 0,
    }))

    const encodedAck = realtime.ClickAck.encode(realtime.ClickAck.create({
      button: { key: 'feel' },
      delta: 1,
      critical: false,
      myBossDamage: 61,
      bossLeaderboardCount: 2,
    })).finish()
    const ackFrame = new Uint8Array(1 + encodedAck.length)
    ackFrame[0] = realtimeBinaryType.clickAck
    ackFrame.set(encodedAck, 1)
    sockets[0].emitBinary(ackFrame.buffer)

    expect(clickAcks).toEqual([
      {
        button: { key: 'feel' },
        delta: 1,
        critical: false,
        myBossDamage: 61,
        bossLeaderboardCount: 2,
        partStateDeltas: [],
        talentEvents: [],
      },
    ])
  })

  it('online_count 可通过 WebSocket 与 SSE 回调在线人数', () => {
    const sockets = []
    const sources = []
    const onlineCounts = []
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
      onOnlineCount(payload) {
        onlineCounts.push(payload?.count)
      },
    })

    transport.connect({ nickname: '阿明' })
    sockets[0].emitOpen()
    sockets[0].emitMessage({
      type: 'online_count',
      payload: { count: 3 },
    })

    sockets[0].emitClose()
    sources[0].emitEvent('online_count', { count: 4 })

    expect(onlineCounts).toEqual([3, 4])
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
