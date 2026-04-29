import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')
const docsPath = path.resolve(currentDir, '../../../docs/developer-reference/2026-04-29-BattlePage部位状态面板复用说明.md')

describe('BattlePage 部位状态面板', () => {
  it('左侧新增可复用的部位状态面板，并接入剥皮与护甲崩塌状态', () => {
    expect(stateSource).toContain('const partStatusList = computed(() => {')
    expect(stateSource).toContain('const skinnerMap = cs?.skinnerParts || {}')
    expect(stateSource).toContain("statusKey: 'collapse'")
    expect(stateSource).toContain("statusLabel: '护甲崩塌'")
    expect(stateSource).toContain("statusLabel: '剥皮'")
    expect(stateSource).toContain('remainingSec:')
    expect(stateSource).toContain('progress:')
    expect(battleSource).toContain('partStatusList')
    expect(battleSource).toContain('class="part-status-panel"')
    expect(battleSource).toContain('class="part-status-panel__title">部位状态</div>')
    expect(battleSource).toContain('{{ s.statusLabel }}')
    expect(battleSource).toContain('{{ s.remainingSec }}s')
    expect(battleSource).toContain('class="part-status-panel__bar"')
    expect(battleSource).toContain('class="part-status-panel__bar-fill"')
    expect(battleSource).not.toContain('class="collapse-panel"')
  })

  it('部位状态面板复用累计进度面板的容器视觉，而不是新造一整套卡片风格', () => {
    expect(styleSource).toContain('.part-progress-panel,\n.part-status-panel {')
    expect(styleSource).toContain('.part-progress-panel__title,\n.part-status-panel__title {')
    expect(styleSource).toContain('.part-progress-panel__item,\n.part-status-panel__item {')
    expect(styleSource).toContain('.part-progress-panel__name,\n.part-status-panel__name {')
    expect(styleSource).toContain('.part-status-panel__bar {')
    expect(styleSource).toContain('.part-status-panel__bar-fill {')
  })
})

describe('BattlePage 全局状态面板', () => {
  it('左侧全局状态统一走可复用的 status-panel 列表，而不是保留专属卡片模板', () => {
    expect(stateSource).toContain('const globalStatusList = computed(() => {')
    expect(stateSource).toContain("kind: 'combo'")
    expect(stateSource).toContain("kind: 'omen'")
    expect(stateSource).toContain("kind: 'final_cut'")
    expect(battleSource).toContain('globalStatusList')
    expect(battleSource).toContain('class="status-panel"')
    expect(battleSource).toContain('class="status-panel__title">{{ status.title }}</div>')
    expect(battleSource).not.toContain('class="combo-box"')
    expect(battleSource).not.toContain('class="omen-panel"')
    expect(battleSource).not.toContain('class="final-cut-cooldown-panel"')
  })
})

describe('BattlePage 部位状态面板开发参考', () => {
  it('补充了部位状态面板复用说明文档', () => {
    expect(existsSync(docsPath)).toBe(true)
    const docsSource = readFileSync(docsPath, 'utf8')
    expect(docsSource).toContain('部位累计型')
    expect(docsSource).toContain('部位状态型')
    expect(docsSource).toContain('skinnerParts')
    expect(docsSource).toContain('collapsePartKeys')
    expect(docsSource).toContain('partStatusList')
    expect(docsSource).toContain('BattlePage.vue')
    expect(docsSource).toContain('publicPageState.js')
  })
})
