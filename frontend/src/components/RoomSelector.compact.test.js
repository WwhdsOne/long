import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const source = readFileSync(path.resolve(currentDir, './RoomSelector.vue'), 'utf8')

describe('RoomSelector 紧凑卡片', () => {
  it('房间卡仅展示 Boss 均血，不再展示战斗力准入', () => {
    expect(source).toContain("import RoomSwitchCooldownTag from './RoomSwitchCooldownTag.vue'")
    expect(source).toContain('<RoomSwitchCooldownTag :cooldown-remaining-seconds="cooldownRemainingSeconds" />')
    expect(source).toContain("const displayName = String(room?.displayName || '').trim()")
    expect(source).toContain("return defaultRoomLabel(room?.id)")
    expect(source).toContain("if (props.cooldownRemainingSeconds > 0) return '冷却未结束'")
    expect(source).toContain('function roomAvgHpText(room) {')
    expect(source).toContain('<span class="room-selector__bossline">')
    expect(source).toContain('<small>Boss平均血量</small>')
    expect(source).toContain('<span class="room-selector__action">{{ roomActionLabel(room) }}</span>')
    expect(source).not.toContain('roomBattlePowerRangeText')
    expect(source).not.toContain('准入 ')
  })
})
