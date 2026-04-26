import { mergeBossState } from './bossState'

export function mergeClickFallbackState(currentState, payload) {
  const nextState = {
    userStats: currentState.userStats ?? null,
    boss: currentState.boss ?? null,
    bossLeaderboard: Array.isArray(currentState.bossLeaderboard) ? currentState.bossLeaderboard : [],
    myBossStats: currentState.myBossStats ?? null,
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
  }
  if ('myBossStats' in payload) {
    nextState.myBossStats = payload.myBossStats ?? null
  }
  if ('recentRewards' in payload) {
    nextState.recentRewards = Array.isArray(payload.recentRewards) ? payload.recentRewards : []
  }

  return nextState
}
