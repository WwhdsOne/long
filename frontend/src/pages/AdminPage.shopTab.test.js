import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './AdminPage.vue'), 'utf8')
const adminStateSource = readFileSync(path.resolve(currentDir, './admin/useAdminPage.js'), 'utf8')
const adminActionSource = readFileSync(path.resolve(currentDir, './admin/useAdminPageActions.js'), 'utf8')

describe('AdminPage 商店管理接线', () => {
  it('后台容器挂载商店 Tab 组件并提供入口按钮', () => {
    expect(pageSource).toContain("import AdminShopTab from '../components/admin/AdminShopTab.vue'")
    expect(pageSource).toContain("admin.activeTab === 'shop'")
    expect(pageSource).toContain('>商店<')
  })

  it('后台状态层提供商店列表和表单状态', () => {
    expect(adminStateSource).toContain('const shopItems = ref([])')
    expect(adminStateSource).toContain('const shopItemForm = ref(emptyShopItemForm())')
    expect(adminStateSource).toContain('fetchShopItems')
  })

  it('后台动作层支持保存、删除和图片上传', () => {
    expect(adminActionSource).toContain('async function saveShopItem()')
    expect(adminActionSource).toContain('async function deleteShopItem(itemId)')
    expect(adminActionSource).toContain('async function uploadShopImage(event)')
    expect(adminActionSource).toContain('async function uploadShopPreviewImage(event)')
    expect(adminActionSource).toContain('async function uploadShopCursorImage(event)')
  })
})
