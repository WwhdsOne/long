import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const talentSource = readFileSync(path.resolve(currentDir, '../../../backend/internal/vote/talent.go'), 'utf8')
const ossMapSource = readFileSync(path.resolve(currentDir, '../../../pixel-assets/oss-url-map.json'), 'utf8')

describe('暴击树死兆重构', () => {
  it('删除死兆共鸣节点并移除单节点前置', () => {
    expect(talentSource).not.toContain('"crit_omen_resonate"')
    expect(talentSource).toContain('"crit_bleed"')
    expect(talentSource).not.toContain('Prerequisite:')
  })

  it('末日审判图标复用死兆共鸣资源地址', () => {
    expect(ossMapSource).toContain('"talent-crit_doom_judgment.png"')
    expect(ossMapSource).toContain('talent-crit_omen_resonate.png')
  })
})
