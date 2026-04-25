import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = [
  './PublicPage.vue',
  './BattlePage.vue',
  './ProfilePage.vue',
  './MessagesPage.vue',
  './publicPageState.js',
]
  .map((file) => readFileSync(path.resolve(currentDir, file), 'utf8'))
  .join('\n')

describe('PublicPage 点击响应链路', () => {
  it('点击前会先申请一次性票据，并显式上报当前实时连接状态', () => {
    expect(pageSource).toContain("ensureRealtimeTransport().requestClickTicket(key, nextFingerprintHash)")
    expect(pageSource).toContain('buildClickRequestBody(ticketInfo.ticket, liveConnected.value, behavior)')
    expect(pageSource).toContain('consumeClickBehavior(key)')
    expect(pageSource).toContain('buildFingerprintProof({')
    expect(pageSource).toContain('fingerprintHash')
    expect(pageSource).toContain('fingerprintProof')
    expect(pageSource).toContain('@pointerdown="handleBossZonePressStart(zone, $event)"')
    expect(pageSource).toContain('@pointerup="handleBossZonePressEnd(zone, $event)"')
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
    expect(pageSource).toContain("fetch('/api/auto-click/start'")
    expect(pageSource).toContain("fetch('/api/auto-click/stop'")
  })
})
