import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const componentSource = readFileSync(path.resolve(currentDir, './AdminBossTab.vue'), 'utf8')
const actionSource = readFileSync(path.resolve(currentDir, '../../pages/admin/useAdminPageActions.js'), 'utf8')

describe('AdminBossTab 部位血量口径', () => {
  it('Boss 总血量只读展示并由部位最大血量合计决定', () => {
    expect(componentSource).toContain('bossPartTotalHp')
    expect(componentSource).toContain(':value="bossPartTotalHp"')
    expect(componentSource).toContain('readonly')
    expect(componentSource).toContain('由部位总血量决定')
    expect(actionSource).toContain('sumBossPartMaxHp')
    expect(actionSource).toContain('maxHp: sumBossPartMaxHp(bossForm.value.layout)')
  })
})
