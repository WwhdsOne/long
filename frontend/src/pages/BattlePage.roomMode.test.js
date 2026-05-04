import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('BattlePage 房间战斗页', () => {
  it('顶部统计只保留点击与 Boss 击杀四项', () => {
    expect(battleSource).toContain('我的点击')
    expect(battleSource).toContain('总点击')
    expect(battleSource).toContain('我的Boss击杀数')
    expect(battleSource).toContain('总Boss击杀数')
  })

  it('大厅和战斗房间按 hall 态切换显示内容', () => {
    expect(battleSource).toContain("const HALL_ROOM_ID = 'hall'")
    expect(battleSource).toContain('const isHallRoom = computed(() => String(currentRoomId.value || \'\') === HALL_ROOM_ID)')
    expect(battleSource).toContain("const displayName = String(currentRoom.value?.displayName || '').trim()")
    expect(battleSource).toContain('v-if="isHallRoom"')
    expect(battleSource).toContain('v-if="!isHallRoom"')
    expect(battleSource).toContain('当前处于大厅。这里只显示战线分流和点击总榜，不显示 Boss 战斗区与 Boss 伤害榜。')
  })

  it('战斗房间允许随时退出，并复用公共切房冷却标签', () => {
    expect(battleSource).toContain('const roomJoinCooldownRemainingSeconds = computed(() => {')
    expect(battleSource).toContain("import RoomSwitchCooldownTag from '../components/RoomSwitchCooldownTag.vue'")
    expect(battleSource).toContain('<RoomSwitchCooldownTag :cooldown-remaining-seconds="roomJoinCooldownRemainingSeconds" />')
    expect(battleSource).toContain(':disabled="roomSwitching"')
    expect(battleSource).not.toContain('const roomExitLocked = computed(() => roomExitCooldownRemainingSeconds.value > 0)')
    expect(battleSource).not.toContain('function formatRoomExitCooldown(seconds) {')
  })

  it('房间冷却结束时间由 /api/rooms 和切房返回同步', () => {
    expect(stateSource).toContain('const roomSwitchCooldownEndsAt = ref(0)')
    expect(stateSource).toContain('function setRoomSwitchCooldown(remainingSeconds) {')
    expect(stateSource).toContain('setRoomSwitchCooldown(payload?.switchCooldownRemainingSeconds ?? 0)')
    expect(stateSource).toContain('setRoomSwitchCooldown(payload?.cooldownRemainingSeconds ?? payload?.switchCooldownRemainingSeconds ?? 0)')
    expect(stateSource).toContain('onPublicDelta(payload) {')
    expect(stateSource).toContain('applyPublicState(payload)')
    expect(stateSource).toContain('void loadRooms()')
  })

  it('退出当前房间直接切到 hall', () => {
    expect(battleSource).toContain('joinRoom(HALL_ROOM_ID)')
  })
})
