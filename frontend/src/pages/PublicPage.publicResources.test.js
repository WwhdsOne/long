import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')

describe('PublicPage 公共资源拆分', () => {
  it('按钮列表改为维护当前页状态并通过分页接口加载', () => {
    expect(pageSource).toContain('const buttonPage = ref(1)')
    expect(pageSource).toContain('const buttonTotalPages = ref(1)')
    expect(pageSource).toContain('async function loadButtonPage(page)')
    expect(pageSource).toContain("/api/buttons/pages?page=")
  })

  it('Boss 掉落池改为通过独立资源接口按需拉取', () => {
    expect(pageSource).toContain('async function loadBossResources')
    expect(pageSource).toContain("'/api/boss/resources'")
  })

  it('最新公告改为跟随 announcementVersion 变化单独回源', () => {
    expect(pageSource).toContain('const announcementVersion = ref')
    expect(pageSource).toContain('async function loadLatestAnnouncement')
    expect(pageSource).toContain("'/api/announcements/latest'")
  })
})
