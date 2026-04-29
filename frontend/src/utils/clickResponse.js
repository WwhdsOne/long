import { mergeBossState } from './bossState'

export function mergeClickFallbackState(currentState, payload) {
  const nextState = {
    userStats: currentState.userStats ?? null,
    boss: currentState.boss ?? null,
    bossLeaderboard: Array.isArray(currentState.bossLeaderboard) ? currentState.bossLeaderboard : [],
    bossLeaderboardCount: Number.isFinite(currentState.bossLeaderboardCount) ? currentState.bossLeaderboardCount : (Array.isArray(currentState.bossLeaderboard) ? currentState.bossLeaderboard.length : 0),
    myBossStats: currentState.myBossStats ?? null,
    myBossDamage: Number.isFinite(currentState.myBossDamage) ? currentState.myBossDamage : (currentState.myBossStats?.damage ?? 0),
    recentRewards: Array.isArray(currentState.recentRewards) ? currentState.recentRewards : [],
  }

  if (!payload || typeof payload !== 'object') {
    return nextState
  }
  if ('userStats' in payload) {
    nextState.userStats = payload.userStats ?? null
  }
  if ('boss' in payload) {
    nextState.boss = mergeBossState(nextState.boss, payload.boss)
  }
  if ('bossLeaderboard' in payload) {
    nextState.bossLeaderboard = Array.isArray(payload.bossLeaderboard) ? payload.bossLeaderboard : nextState.bossLeaderboard
    nextState.bossLeaderboardCount = nextState.bossLeaderboard.length
  }
  if ('bossLeaderboardCount' in payload) {
    nextState.bossLeaderboardCount = Math.max(0, Number(payload.bossLeaderboardCount ?? 0))
  }
  if ('myBossStats' in payload) {
    nextState.myBossStats = payload.myBossStats ?? null
    nextState.myBossDamage = nextState.myBossStats?.damage ?? nextState.myBossDamage
  }
  if ('myBossDamage' in payload) {
    nextState.myBossDamage = Math.max(0, Number(payload.myBossDamage ?? 0))
  }
  if ('recentRewards' in payload) {
    nextState.recentRewards = Array.isArray(payload.recentRewards) ? payload.recentRewards : []
  }

  return nextState
}
