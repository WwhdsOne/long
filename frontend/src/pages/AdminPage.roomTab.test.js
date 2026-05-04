import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminPage.vue'), 'utf8')
const adminStateSource = readFileSync(path.resolve(currentDir, './admin/useAdminPage.js'), 'utf8')
const adminActionSource = readFileSync(path.resolve(currentDir, './admin/useAdminPageActions.js'), 'utf8')

describe('AdminPage 房间管理接线', () => {
  it('后台容器挂载房间管理 Tab 组件并提供入口按钮', () => {
    expect(pageSource).toContain("import AdminRoomTab from '../components/admin/AdminRoomTab.vue'")
    expect(pageSource).toContain("admin.activeTab === 'rooms'")
    expect(pageSource).toContain('>房间<')
    expect(pageSource).toContain("room.displayName || `房间 ${room.id}`")
  })

  it('后台状态层提供房间管理列表和加载动作', () => {
    expect(adminStateSource).toContain('const adminRoomSettings = ref([])')
    expect(adminStateSource).toContain('fetchAdminRoomSettings')
    expect(adminStateSource).toContain('fetchAdminRooms')
  })

  it('后台动作层支持保存房间显示名', () => {
    expect(adminActionSource).toContain('async function saveRoomDisplayName(roomId, displayName)')
    expect(adminActionSource).toContain("/api/admin/rooms/${encodeURIComponent(roomId)}")
    expect(adminActionSource).toContain('fetchAdminRooms')
  })
})
