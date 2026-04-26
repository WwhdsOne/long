import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

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

describe('PublicPage 公共资源拆分', () => {
  it('页面挂载时会先走一次 HTTP 首屏加载，再连接实时链路', () => {
    expect(pageSource).toContain('await loadState()')
    expect(pageSource).toContain('connectRealtime(nickname.value)')
  })

  it('按钮列表直接使用首屏完整列表，不再请求公共按钮分页接口', () => {
    expect(pageSource).not.toContain('async function loadButtonPage(page)')
    expect(pageSource).not.toContain('/api/buttons/pages')
    expect(pageSource).not.toContain('buttonTotalPages')
  })

  it('Boss 掉落池改为通过独立资源接口按需拉取', () => {
    expect(pageSource).toContain('async function loadBossResources')
    expect(pageSource).toContain("'/api/boss/resources'")
  })

  it('最新公告改为跟随 announcementVersion 变化单独回源', () => {
    expect(pageSource).toContain('const announcementVersion = ref')
    expect(pageSource).toContain('async function loadLatestAnnouncement')
    expect(pageSource).toContain("'/api/announcements/latest'")
  })
})
