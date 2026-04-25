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

  it('实时个人增量只消费战斗字段，不刷新背包资料', () => {
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

  it('战斗页使用按钮主区加右侧排行的两列布局', () => {
    const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

    expect(pageSource).toContain('class="stage-layout stage-layout--battle"')
    expect(styleSource.replace(/\s+/g, ' ')).toContain(
      '.stage-layout--battle { grid-template-columns: minmax(0, 1fr) minmax(280px, 320px);',
    )
  })

  it('战斗页将世界 Boss 融入投票墙，并只用小字入口打开掉落池弹窗', () => {
    const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

    expect(battleSource).not.toContain('class="boss-stage social-card"')
    expect(battleSource).toContain('世界 Boss 战场')
    expect(battleSource).toContain('const bossDropPool = computed')
    expect(battleSource).toContain('class="boss-drop-link"')
    expect(battleSource).toContain('@click="openBossDropPool"')
    expect(battleSource).not.toContain('class="boss-drop-pool"')
    expect(battleSource).toContain('class="boss-drop-modal"')
    expect(battleSource).toContain('装备掉落')
    expect(battleSource).toContain('class="boss-drop-card boss-drop-card--detail"')
    expect(battleSource).toContain('class="boss-drop-card__details"')
    expect(battleSource).not.toContain('@click="openBossDropDetail')
    expect(battleSource).not.toContain('selectedBossDrop')
    expect(battleSource).not.toContain('Boss 英雄池')
  })

  it('战斗页只呈现 Boss 分区，不再显示常规按钮墙', () => {
    const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

    expect(battleSource).toContain('const bossZones = computed')
    expect(battleSource).toContain('pickButtonForBossPart')
    expect(battleSource).toContain('class="boss-part-cell boss-zone-button"')
    expect(battleSource).toContain('boss-zone-button__label')
    expect(battleSource).toContain('zone.assignedButton.label')
    expect(battleSource).toContain('zone.assignedButton.count')
    expect(battleSource).toContain('@click="clickBossZone(zone)"')
    expect(battleSource).not.toContain('class="button-grid"')
    expect(battleSource).not.toContain('class="vote-card"')
    expect(battleSource).not.toContain('v-for="button in displayedButtons"')
  })

  it('战斗页 Boss 面板不吸顶浮动，并收紧部件网格尺寸', () => {
    const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')
    const compactStyleSource = styleSource.replace(/\s+/g, ' ')
    const bossHudStyle = compactStyleSource.slice(
      compactStyleSource.indexOf('.vote-stage__boss-hud {'),
      compactStyleSource.indexOf('.vote-stage__boss-hud-head {'),
    )

    expect(bossHudStyle).not.toContain('position: sticky')
    expect(bossHudStyle).not.toContain('top: 14px')
    expect(compactStyleSource).toContain('max-width: min(100%, 560px);')
    expect(compactStyleSource).toContain('padding: 10px;')
  })
})
