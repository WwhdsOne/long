import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = [
    './PublicPage.vue',
    './BattlePage.vue',
    './ArmoryPage.vue',
    './MessagesPage.vue',
    './publicPageState.js',
]
    .map((file) => readFileSync(path.resolve(currentDir, file), 'utf8'))
    .join('\n')

describe('PublicPage 点击响应链路', () => {
    it('Boss 攻击通过 WebSocket 发送，不再走攻击 POST 接口', () => {
        const clickSegment = pageSource.slice(
            pageSource.indexOf('async function clickButton'),
            pageSource.indexOf('async function submitNickname'),
        )

        expect(clickSegment).toContain('ensureRealtimeTransport().sendClick')
        expect(clickSegment).not.toContain('/api/boss/parts/')
    })

    it('点击成功后不会把断连状态直接改成已连接', () => {
        const clickSegment = pageSource.slice(
            pageSource.indexOf('async function clickButton'),
            pageSource.indexOf('async function submitNickname'),
        )

        expect(clickSegment).not.toContain('liveConnected.value = true')
    })

    it('页面不再使用本地 setTimeout 挂机循环', () => {
        expect(pageSource).not.toContain('createAutoClickLoop')
        expect(pageSource).toContain('async function startAutoClick')
        expect(pageSource).toContain('async function stopAutoClick')
    })
})
