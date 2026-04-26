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

describe('PublicPage 三合一边界', () => {
  it('提供战斗、天赋、三合一资料分区和消息导航，并默认进入战斗页', () => {
    expect(pageSource).toContain("import BattlePage from './BattlePage.vue'")
    expect(pageSource).toContain("import ArmoryPage from './ArmoryPage.vue'")
    expect(pageSource).toContain("import MessagesPage from './MessagesPage.vue'")
    expect(pageSource).toContain("id: 'battle'")
    expect(pageSource).toContain("id: 'talents'")
    expect(pageSource).toContain("id: 'resources'")
    expect(pageSource).toContain("id: 'inventory'")
    expect(pageSource).toContain("id: 'stats'")
    expect(pageSource).toContain("id: 'loadout'")
    expect(pageSource).toContain("id: 'messages'")
    expect(pageSource).toContain("currentPublicPage.value = 'battle'")
  })

  it('旧资料页接口已删除，不再请求 /api/player/profile', () => {
    expect(pageSource).not.toContain("fetch('/api/player/profile'")
    expect(pageSource).not.toContain('loadPlayerProfile')
    expect(pageSource).not.toContain('refreshProfileAfterMutation')
  })

  it('实时个人增量直接消费完整 user state', () => {
    const realtimeSegment = pageSource.slice(
      pageSource.indexOf('onUserDelta(payload)'),
      pageSource.indexOf('onClickAck(payload)'),
    )
    expect(realtimeSegment).toContain('applyUserState(payload)')
    expect(realtimeSegment).not.toContain('applyBattleUserState(payload)')
  })

  it('三合一页面不再使用资料页内部 tab', () => {
    const armorySource = readFileSync(path.resolve(currentDir, './ArmoryPage.vue'), 'utf8')
    expect(armorySource).toContain('armory-layout')
    expect(armorySource).toContain('armory-backpack-grid')
    expect(armorySource).not.toContain('player-hud__tabs')
    expect(armorySource).not.toContain('activeHudTab')
  })

  it('击杀与挂机共用战利品弹窗，并提供装备网格', () => {
    const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
    const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
    expect(stateSource).toContain('const rewardModal = ref(null)')
    expect(stateSource).toContain('openOnlineRewardModal')
    expect(stateSource).toContain('openAfkRewardModal')
    expect(battleSource).toContain('v-if="rewardModal"')
    expect(battleSource).toContain('class="reward-grid"')
  })
})
