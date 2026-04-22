import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')

describe('PublicPage 点击响应链路', () => {
  it('点击请求会显式上报当前 SSE 连接状态', () => {
    expect(pageSource).toContain('buildClickRequestBody(nickname.value, liveConnected.value)')
  })

  it('点击成功后不会把断连状态直接改成已连接', () => {
    const clickSegment = pageSource.slice(
      pageSource.indexOf('async function clickButton'),
      pageSource.indexOf('async function postEquipmentAction'),
    )

    expect(clickSegment).not.toContain('liveConnected.value = true')
  })
})
