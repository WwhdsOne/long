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
        expect(compactPageSource).toContain('>风险名单 </button>')
    })

    it('后台状态层提供黑名单列表和时间格式化能力', () => {
        expect(adminStateSource).toContain('const blacklistPage = ref(emptyBlacklistPage())')
        expect(adminStateSource).toContain('const loadingBlacklist = ref(false)')
        expect(adminStateSource).toContain('async function fetchBlacklist()')
        expect(stateSource).toContain('export function formatDuration(seconds)')
        expect(stateSource).toContain("year: 'numeric'")
        expect(stateSource).toContain("second: '2-digit'")
    })

    it('后台动作层支持清除账号风险状态', () => {
        expect(adminActionSource).toContain('async function unblockBlacklistEntry(nickname)')
        expect(adminActionSource).toContain("/api/admin/blacklist/${encodeURIComponent(nickname)}/unblock")
    })

    it('风险面板展示昵称、积分和可选封禁时间', () => {
        expect(tabSource).toContain('当前积分：{{ entry.score }}')
        expect(tabSource).toContain('v-if="entry.banUntil"')
        expect(tabSource).toContain('封禁截止：{{ formatTime(entry.banUntil) }}')
        expect(tabSource).toContain('清除风险状态')
    })

    it('风险列表标准化昵称、积分和封禁截止时间', () => {
        expect(stateSource).toContain('score: Math.max(0, Number(entry?.score ?? 0))')
        expect(stateSource).toContain('banUntil: Math.max(0, Number(entry?.banUntil ?? 0))')
    })
})
