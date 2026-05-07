import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminPage.vue'), 'utf8')
const adminStateSource = readFileSync(path.resolve(currentDir, './admin/useAdminPage.js'), 'utf8')
const adminActionSource = readFileSync(path.resolve(currentDir, './admin/useAdminPageActions.js'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './admin/state.js'), 'utf8')
const tabSource = readFileSync(path.resolve(currentDir, '../components/admin/AdminBlacklistTab.vue'), 'utf8')
const compactPageSource = pageSource.replace(/\s+/g, ' ')

describe('AdminPage 黑名单管理接线', () => {
    it('管理页挂载黑名单 Tab 并提供入口按钮', () => {
        expect(pageSource).toContain("import AdminBlacklistTab from '../components/admin/AdminBlacklistTab.vue'")
        expect(pageSource).toContain("admin.activeTab === 'blacklist'")
        expect(compactPageSource).toContain('>黑名单 </button>')
    })

    it('后台状态层提供黑名单列表和时间格式化能力', () => {
        expect(adminStateSource).toContain('const blacklistPage = ref(emptyBlacklistPage())')
        expect(adminStateSource).toContain('const loadingBlacklist = ref(false)')
        expect(adminStateSource).toContain('async function fetchBlacklist()')
        expect(stateSource).toContain('export function formatDuration(seconds)')
        expect(stateSource).toContain("year: 'numeric'")
        expect(stateSource).toContain("second: '2-digit'")
    })

    it('后台动作层支持手动解封', () => {
        expect(adminActionSource).toContain('async function unblockBlacklistEntry(clientId, nickname)')
        expect(adminActionSource).toContain("/api/admin/blacklist/${encodeURIComponent(clientId)}/unblock")
    })

    it('黑名单面板展示昵称、起止时间和手动解封按钮', () => {
        expect(tabSource).toContain('封禁开始：{{ formatTime(entry.blockedAt) }}')
        expect(tabSource).toContain('封禁结束：{{ formatTime(entry.blockedUntil) }}')
        expect(tabSource).toContain('剩余时间：{{ formatDuration(entry.remainingSeconds) }}')
        expect(tabSource).toContain('手动解封')
    })

    it('IP 封禁条目展示具体 IP', () => {
        expect(tabSource).toContain("entry.clientId.startsWith('ip:')")
        expect(tabSource).toContain("IP：{{ entry.clientId.slice(3) }}")
    })
})
