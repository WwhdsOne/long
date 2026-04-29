import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const appSource = readFileSync(path.resolve(currentDir, '../App.vue'), 'utf8')
const publicPageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')

describe('BattleFxGalleryPage 隐藏入口', () => {
  it('通过 /internal/battle-fx-gallery 路径挂载', () => {
    expect(appSource).toContain("import BattleFxGalleryPage from './pages/BattleFxGalleryPage.vue'")
    expect(appSource).toContain("currentPath.startsWith('/internal/battle-fx-gallery')")
  })

  it('不出现在公开前台容器中', () => {
    expect(publicPageSource).not.toContain('BattleFxGalleryPage')
    expect(publicPageSource).not.toContain('battle-fx-gallery')
  })
})
