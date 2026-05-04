import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('BattlePage 审判日 HUD', () => {
  it('审判日进度使用独立的 partJudgmentDayCount，而不是复用破甲计数', () => {
    expect(stateSource).toContain('const judgmentDayMap = cs?.partJudgmentDayCount || {}')
    expect(stateSource).toContain('const rawJudgmentDay = Number(judgmentDayMap[key]) || 0')
    expect(stateSource).toContain("const jdCount = (part.type === 'heavy' && !jdOnCooldown) ? rawJudgmentDay : 0")
    expect(stateSource).not.toContain("const jdCount = (part.type === 'heavy' && !jdOnCooldown) ? rawHeavy : 0")
  })
})
