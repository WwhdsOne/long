import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './TalentsPage.vue'), 'utf8')

describe('TalentsPage effectLines 响应链路', () => {
  it('升级成本展示与后端幂次公式保持一致', () => {
    expect(pageSource).toContain('const talentCostLevelExponent = 0.85')
    expect(pageSource).toContain('const talentCostMultiplier = 1.8')
    expect(pageSource).toContain('Math.round(def.cost * Math.pow(targetLevel, talentCostLevelExponent) * talentCostMultiplier)')
    expect(pageSource).toContain('for (let level = fromLevel + 1; level <= toLevel; level += 1)')
    expect(pageSource).not.toContain('targetLevel * 1.5')
  })

  it('初次加载从 /api/talents/state 读取后端 effectLines', () => {
    expect(pageSource).toContain("fetch('/api/talents/state'")
    expect(pageSource).toContain('talentEffectLines.value = talentState.value?.effectLines || {}')
  })

  it('升级成功后用响应里的 effectLines 刷新浮层描述', () => {
    const upgradeSegment = pageSource.slice(
      pageSource.indexOf('async function handleNodeClick(item)'),
      pageSource.indexOf('function clearNode()'),
    )

    expect(upgradeSegment).toContain("fetch('/api/talents/upgrade'")
    expect(upgradeSegment).toContain('talentEffectLines.value = data.effectLines || talentEffectLines.value')
  })
})
