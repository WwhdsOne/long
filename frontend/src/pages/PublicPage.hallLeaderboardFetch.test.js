import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')

describe('PublicPage 大厅分页总榜拉取', () => {
  it('大厅分页总榜改为进入 hall 时通过 HTTP 单独拉取', () => {
    expect(stateSource).toContain('async function loadHallLeaderboardSnapshot()')
    expect(stateSource).toContain("'/api/leaderboard?offset=10&limit=200'")
    expect(stateSource).toContain('hallLeaderboardSnapshot.value = Array.isArray(payload?.leaderboard)')
  })

  it('停留 hall 期间不刷新，离开再进入才重新拉取', () => {
    expect(battleSource).toContain("watch(() => currentRoomId.value, async (next, prev) => {")
    expect(battleSource).toContain("if (String(next || '') === HALL_ROOM_ID && String(prev || '') !== HALL_ROOM_ID)")
    expect(battleSource).toContain('await loadHallLeaderboardSnapshot()')
    expect(battleSource).toContain("if (String(next || '') !== HALL_ROOM_ID)")
    expect(battleSource).toContain('resetHallLeaderboardSnapshot()')
  })
})
