import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const appSource = readFileSync(path.resolve(currentDir, '../App.vue'), 'utf8')
const publicPageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')

describe('TalentBuffDemoPage 隐藏入口', () => {
  it('只通过隐藏路径挂载 demo 页面', () => {
    expect(appSource).toContain("import TalentBuffDemoPage from './pages/TalentBuffDemoPage.vue'")
    expect(appSource).toContain("currentPath.startsWith('/__talent-buff-demo')")
  })

  it('不出现在公开前台容器中', () => {
    expect(publicPageSource).not.toContain('TalentBuffDemoPage')
    expect(publicPageSource).not.toContain('__talent-buff-demo')
  })
})
