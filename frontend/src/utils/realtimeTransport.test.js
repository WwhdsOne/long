import {describe, expect, it} from 'vitest'

import {decodeRealtimeBinaryMessage, encodeRealtimeClickRequest, realtimeBinaryType} from './realtimeProto'
import {createRealtimeTransport} from './realtimeTransport'
import {realtime} from '../proto/realtime.js'

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
        this.onmessage?.({data: payload})
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
        this.listeners.get(name)?.({data: JSON.stringify(payload)})
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
                snapshots.push({publicState, userState})
            },
            onTransportState(nextState) {
                states.push(nextState)
            },
        })

        transport.connect({nickname: '阿明'})
        sockets[0].emitOpen()
        sockets[0].emitMessage({
            type: 'snapshot',
            public: {
                buttons: [{key: 'feel', count: 3}],
            },
            user: {
                userStats: {nickname: '阿明', clickCount: 3},
            },
        })

        expect(sockets[0].sent).toEqual([
            JSON.stringify({type: 'hello', nickname: '阿明'}),
        ])
        expect(snapshots).toEqual([
            {
                publicState: {
                    buttons: [{key: 'feel', count: 3}],
                },
                userState: {
                    userStats: {nickname: '阿明', clickCount: 3},
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

        transport.connect({nickname: '阿明'})

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

        transport.connect({nickname: '阿明'})
        sockets[0].emitOpen()

        expect(transport.sendClick('feel')).toBe(true)
        expect(sockets[0].sent.at(-1)).toEqual(encodeRealtimeClickRequest({
            slug: 'feel',
            comboCount: 0,
        }))

        const encodedAck = realtime.ClickAck.encode(realtime.ClickAck.create({
            button: {key: 'feel'},
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
                button: {key: 'feel'},
                delta: 1,
                critical: false,
                myBossDamage: 61,
                bossLeaderboardCount: 2,
                partStateDeltas: [],
                talentEvents: [],
            },
        ])
    })

    it('click_ack 的 0 坐标部位增量不会在二进制解码时丢失 x/y', () => {
        const encodedAck = realtime.ClickAck.encode(realtime.ClickAck.create({
            button: {key: 'boss-part:0-2'},
            delta: 1,
            critical: false,
            partStateDeltas: [
                {x: 0, y: 2, damage: 9, beforeHp: 40, afterHp: 31, partType: 'soft'},
            ],
        })).finish()
        const ackFrame = new Uint8Array(1 + encodedAck.length)
        ackFrame[0] = realtimeBinaryType.clickAck
        ackFrame.set(encodedAck, 1)

        expect(decodeRealtimeBinaryMessage(ackFrame)).toEqual({
            type: 'click_ack',
            payload: {
                button: {key: 'boss-part:0-2'},
                delta: 1,
                critical: false,
                partStateDeltas: [
                    {x: 0, y: 2, damage: 9, beforeHp: 40, afterHp: 31, partType: 'soft'},
                ],
                talentEvents: [],
            },
        })
    })

    it('public_delta 的 0 坐标 Boss 部位不会在二进制解码时丢失 x/y', () => {
        const encoded = realtime.PublicDelta.encode(realtime.PublicDelta.create({
            totalVotes: 10,
            roomId: '1',
            boss: {
                id: 'boss-1',
                status: 'active',
                maxHp: 100,
                currentHp: 91,
                parts: [
                    {x: 0, y: 0, type: 'soft', maxHp: 50, currentHp: 41, armor: 3, alive: true},
                    {x: 1, y: 0, type: 'heavy', maxHp: 50, currentHp: 50, armor: 8, alive: true},
                ],
            },
        })).finish()
        const frame = new Uint8Array(1 + encoded.length)
        frame[0] = realtimeBinaryType.publicDelta
        frame.set(encoded, 1)

        expect(decodeRealtimeBinaryMessage(frame)).toMatchObject({
            type: 'public_delta',
            payload: {
                totalVotes: 10,
                roomId: '1',
                boss: {
                    id: 'boss-1',
                    status: 'active',
                    maxHp: 100,
                    currentHp: 91,
                    parts: [
                        {x: 0, y: 0, type: 'soft', maxHp: 50, currentHp: 41, armor: 3, alive: true},
                        {x: 1, y: 0, type: 'heavy', maxHp: 50, currentHp: 50, armor: 8, alive: true},
                    ],
                },
            },
        })
    })

    it('public_meta 未携带 leaderboard 时不会在二进制解码时补成空数组', () => {
        const encoded = realtime.PublicMeta.encode(realtime.PublicMeta.create({
            announcementVersion: 'ann-2',
        })).finish()
        const frame = new Uint8Array(1 + encoded.length)
        frame[0] = realtimeBinaryType.publicMeta
        frame.set(encoded, 1)

        expect(decodeRealtimeBinaryMessage(frame)).toEqual({
            type: 'public_meta',
            payload: {
                announcementVersion: 'ann-2',
            },
        })
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

        transport.connect({nickname: '阿明'})
        sockets[0].emitOpen()
        sockets[0].emitMessage({
            type: 'online_count',
            payload: {count: 3},
        })

        sockets[0].emitClose()
        sources[0].emitEvent('online_count', {count: 4})

        expect(onlineCounts).toEqual([3, 4])
    })

    it('public_meta 可通过 WebSocket 与 SSE 回调低频公共态', () => {
        const sockets = []
        const sources = []
        const publicMetas = []
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
            onPublicMeta(payload) {
                publicMetas.push(payload)
            },
        })

        transport.connect({nickname: '阿明'})
        sockets[0].emitOpen()
        const encoded = realtime.PublicMeta.encode(realtime.PublicMeta.create({
            announcementVersion: 'ann-1',
            leaderboard: [{rank: 2, nickname: '小红', clickCount: 0}],
            bossLeaderboard: [{rank: 1, nickname: '阿明', damage: 0}],
        })).finish()
        const frame = new Uint8Array(1 + encoded.length)
        frame[0] = realtimeBinaryType.publicMeta
        frame.set(encoded, 1)
        sockets[0].emitBinary(frame.buffer)

        sockets[0].emitClose()
        sources[0].emitEvent('public_meta', {
            bossLeaderboard: [{rank: 1, nickname: '阿明', damage: 66}],
        })

        expect(publicMetas).toEqual([
            {
                announcementVersion: 'ann-1',
                leaderboard: [{rank: 2, nickname: '小红', clickCount: 0}],
                bossLeaderboard: [{rank: 1, nickname: '阿明', damage: 0}],
            },
            {
                bossLeaderboard: [{rank: 1, nickname: '阿明', damage: 66}],
            },
        ])
    })

    it('room_state 可通过 WebSocket 与 SSE 回调房间列表', () => {
        const sockets = []
        const sources = []
        const roomStates = []
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
            onRoomState(payload) {
                roomStates.push(payload)
            },
        })

        transport.connect({nickname: '阿明'})
        sockets[0].emitOpen()

        const encoded = realtime.RoomState.encode(realtime.RoomState.create({
            currentRoomId: '2',
            switchCooldownRemainingSeconds: 6,
            rooms: [{id: '2', displayName: '二线', current: true, onlineCount: 4}],
        })).finish()
        const frame = new Uint8Array(1 + encoded.length)
        frame[0] = realtimeBinaryType.roomState
        frame.set(encoded, 1)
        sockets[0].emitBinary(frame.buffer)

        sockets[0].emitClose()
        sources[0].emitEvent('room_state', {
            currentRoomId: 'hall',
            switchCooldownRemainingSeconds: 0,
            rooms: [{id: '1', displayName: '一线', current: false, onlineCount: 2}],
        })

        expect(roomStates).toEqual([
            {
                currentRoomId: '2',
                switchCooldownRemainingSeconds: 6,
                rooms: [{id: '2', displayName: '二线', current: true, onlineCount: 4}],
            },
            {
                currentRoomId: 'hall',
                switchCooldownRemainingSeconds: 0,
                rooms: [{id: '1', displayName: '一线', current: false, onlineCount: 2}],
            },
        ])
    })

    it('user_delta 的 0 值不会在二进制解码时丢字段', () => {
        const encoded = realtime.UserDelta.encode(realtime.UserDelta.create({
            gold: 0,
            stones: 0,
            talentPoints: 0,
            talentEvents: [
                {talentId: 'doom-mark', partX: 0, partY: 2, extraDamage: 0},
            ],
        })).finish()
        const frame = new Uint8Array(1 + encoded.length)
        frame[0] = realtimeBinaryType.userDelta
        frame.set(encoded, 1)

        expect(decodeRealtimeBinaryMessage(frame)).toEqual({
            type: 'user_delta',
            payload: {
                gold: 0,
                recentRewards: [],
                stones: 0,
                talentPoints: 0,
                talentEvents: [
                    {
                        talentId: 'doom-mark',
                        extraDamage: 0,
                        partX: 0,
                        partY: 2,
                    },
                ],
            },
        })
    })

    it('room_state 的 0 值不会在二进制解码时丢字段', () => {
        const encoded = realtime.RoomState.encode(realtime.RoomState.create({
            currentRoomId: '2',
            switchCooldownRemainingSeconds: 0,
            rooms: [{
                id: '2',
                displayName: '二线',
                current: true,
                onlineCount: 0,
                currentBossHp: 0,
                currentBossMaxHp: 0,
                currentBossAvgHp: 0,
            }],
        })).finish()
        const frame = new Uint8Array(1 + encoded.length)
        frame[0] = realtimeBinaryType.roomState
        frame.set(encoded, 1)

        expect(decodeRealtimeBinaryMessage(frame)).toEqual({
            type: 'room_state',
            payload: {
                currentRoomId: '2',
                switchCooldownRemainingSeconds: 0,
                rooms: [{
                    id: '2',
                    displayName: '二线',
                    current: true,
                    onlineCount: 0,
                    currentBossHp: 0,
                    currentBossMaxHp: 0,
                    currentBossAvgHp: 0,
                }],
            },
        })
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

        transport.connect({nickname: '阿明'})
        sockets[0].emitOpen()
        sockets[0].emitClose()

        expect(sources).toHaveLength(1)
        expect(sources[0].url).toBe('/api/events?nickname=%E9%98%BF%E6%98%8E')
        sources[0].emitOpen()
        sources[0].emitEvent('public_state', {
            buttons: [{key: 'feel', count: 5}],
        })

        expect(publicDeltas).toEqual([
            {
                buttons: [{key: 'feel', count: 5}],
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
