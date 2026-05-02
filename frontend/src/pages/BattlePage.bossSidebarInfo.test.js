import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('BattlePage Boss 右侧整合信息', () => {
  it('把伤害、榜单、掉落资格和掉落池入口整合到5x5区域右下卡片', () => {
    expect(battleSource).toContain('class="boss-right-summary"')
    expect(battleSource).toContain('class="boss-right-summary__stats"')
    expect(battleSource).toContain('我的伤害 {{ myBossDamage }}')
    expect(battleSource).toContain('Boss 榜 {{ bossLeaderboardCount }} 人')
    expect(battleSource).toContain('class="boss-right-summary__rule"')
    expect(battleSource).toContain('对 Boss 造成至少 1% 生命值的伤害，才有资格掉落装备与资源。')
    expect(battleSource).toContain('class="boss-right-summary__drop"')
    expect(battleSource).toContain('点击查看 Boss 掉落池')
    expect(battleSource).toContain('{{ bossDropPool.length }} 件掉落物')
  })

  it('移除战斗区下方旧的Boss信息块，避免重复展示', () => {
    expect(battleSource).not.toContain('class="vote-stage__boss-hud-stats"')
    expect(battleSource).not.toContain('class="vote-stage__boss-note"')
  })

  it('为右侧整合信息卡定义独立样式', () => {
    expect(styleSource).toContain('.boss-right-summary {')
    expect(styleSource).toContain('.boss-right-summary__stats {')
    expect(styleSource).toContain('.boss-right-summary__drop {')
  })

  it('Boss 掉落池复用背包多行属性格式，并补齐三种部位增伤', () => {
    expect(battleSource).toContain('v-for="line in formatItemStatLines(item)"')
    expect(battleSource).not.toContain('{{ formatItemStats(item) }}')
    expect(stateSource).toContain('function normalizeDisplayPercent(value) {')
    expect(stateSource).toContain('if (item?.partTypeDamageSoft) lines.push(`软组织伤害 ${formatDisplayPercent(item.partTypeDamageSoft)}%`)')
    expect(stateSource).toContain('if (item?.partTypeDamageHeavy) lines.push(`重甲伤害 ${formatDisplayPercent(item.partTypeDamageHeavy)}%`)')
    expect(stateSource).toContain('if (item?.partTypeDamageWeak) lines.push(`弱点伤害 ${formatDisplayPercent(item.partTypeDamageWeak)}%`)')
    expect(stateSource).toContain('lines.push(`护甲穿透 ${formatDisplayPercent(item.armorPenPercent)}%`)')
  })
})
