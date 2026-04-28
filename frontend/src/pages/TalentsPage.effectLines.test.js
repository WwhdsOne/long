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
    expect(pageSource).toContain('talentEffectDescriptions.value = talentState.value?.effectDescriptions || {}')
  })

  it('升级成功后用响应里的 effectLines 刷新浮层描述', () => {
    const upgradeSegment = pageSource.slice(
      pageSource.indexOf('async function handleNodeClick(item)'),
      pageSource.indexOf('function clearNode()'),
    )

    expect(upgradeSegment).toContain("fetch('/api/talents/upgrade'")
    expect(upgradeSegment).toContain('talentEffectLines.value = data.effectLines || talentEffectLines.value')
    expect(upgradeSegment).toContain('talentEffectDescriptions.value = data.effectDescriptions || talentEffectDescriptions.value')
    expect(upgradeSegment).not.toContain("showConfirm('确认升级'")
    expect(pageSource).toContain("showConfirm('确认洗点'")
  })

  it('浮层优先展示后端返回的动态效果描述，并保留静态文案兜底', () => {
    expect(pageSource).toContain('function effectDescription(def) {')
    expect(pageSource).toContain('return talentEffectDescriptions.value?.[def?.id]')
    expect(pageSource).toContain('|| def?.effectDescription')
    expect(pageSource).toContain('|| def?.description')
    expect(pageSource).toContain("|| '暂无效果说明'")
    expect(pageSource).toContain('{{ effectDescription(selectedNode) }}')
  })

  it('首次可学习节点也会显示消耗的天赋点', () => {
    expect(pageSource).toContain("nodeState(selectedNode) === 'upgradable' || nodeState(selectedNode) === 'available'")
    expect(pageSource).toContain('下一级消耗：{{ upgradeCost(selectedNode) }} 天赋点')
  })

  it('页面展示层锁规则、层满奖励和注意事项说明', () => {
    expect(pageSource).toContain('上一层所有节点到 Lv1，才能学习下一层节点')
    expect(pageSource).toContain('当前主系层满额外加成')
    expect(pageSource).toContain('其他注意事项')
    expect(pageSource).toContain('treeDefs.value?.trees?.[selectedTree.value]?.tierCompletionBonuses || {}')
  })
})
