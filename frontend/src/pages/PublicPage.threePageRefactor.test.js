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

describe('PublicPage 三页前台边界', () => {
  it('提供战斗、资料、消息三页导航，并默认进入战斗页', () => {
    expect(pageSource).toContain("import BattlePage from './BattlePage.vue'")
    expect(pageSource).toContain("import ProfilePage from './ProfilePage.vue'")
    expect(pageSource).toContain("import MessagesPage from './MessagesPage.vue'")
    expect(pageSource).toContain("const publicPages = [")
    expect(pageSource).toContain("id: 'battle'")
    expect(pageSource).toContain("id: 'profile'")
    expect(pageSource).toContain("id: 'messages'")
    expect(pageSource).toContain("currentPublicPage.value = 'battle'")
  })

  it('资料页进入时才请求完整资料接口', () => {
    expect(pageSource).toContain('async function loadPlayerProfile')
    expect(pageSource).toContain("fetch('/api/player/profile'")
    expect(pageSource).toContain("if (page === 'profile')")
    expect(pageSource).toContain('await loadPlayerProfile(true)')
  })

  it('实时个人增量只消费战斗字段，不刷新背包和商店资料', () => {
    const realtimeSegment = pageSource.slice(
      pageSource.indexOf('onUserDelta(payload)'),
      pageSource.indexOf('onClickAck(payload)'),
    )

    expect(realtimeSegment).toContain('applyBattleUserState(payload)')
    expect(realtimeSegment).not.toContain('applyUserState(payload)')
  })

  it('资料页操作后刷新资料接口，而不是依赖实时 payload 填充资料', () => {
    expect(pageSource).toContain('await refreshProfileAfterMutation(data)')
    expect(pageSource).toContain('async function refreshProfileAfterMutation')
  })
})
