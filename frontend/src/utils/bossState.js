export function mergeBossState(currentBoss, incomingBoss) {
  if (!incomingBoss || typeof incomingBoss !== 'object') {
    return incomingBoss ?? null
  }
  if (!currentBoss || typeof currentBoss !== 'object') {
    return incomingBoss
  }

  if (currentBoss.id !== incomingBoss.id) {
    return incomingBoss
  }
  if (currentBoss.status !== 'active' || incomingBoss.status !== 'active') {
    return incomingBoss
  }

  const currentHp = Number(currentBoss.currentHp ?? 0)
  const incomingHp = Number(incomingBoss.currentHp ?? 0)
  if (Number.isFinite(currentHp) && Number.isFinite(incomingHp) && incomingHp > currentHp) {
    return currentBoss
  }

  return incomingBoss
}
