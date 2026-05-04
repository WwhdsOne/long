function normalizeIntegerString(value, fallback = '0') {
  const raw = String(value ?? '').trim()
  return /^\d+$/.test(raw) ? raw.replace(/^0+(?=\d)/, '') || '0' : fallback
}

function normalizeBossPart(part) {
  if (!part || typeof part !== 'object') return part
  return {
    ...part,
    maxHp: normalizeIntegerString(part.maxHp),
    currentHp: normalizeIntegerString(part.currentHp),
    armor: normalizeIntegerString(part.armor),
  }
}

export function normalizeBossState(boss) {
  if (!boss || typeof boss !== 'object') return boss ?? null
  return {
    ...boss,
    maxHp: normalizeIntegerString(boss.maxHp),
    currentHp: normalizeIntegerString(boss.currentHp),
    parts: Array.isArray(boss.parts) ? boss.parts.map(normalizeBossPart) : [],
  }
}

function hpBigInt(value) {
  try {
    return BigInt(normalizeIntegerString(value))
  } catch {
    return 0n
  }
}

export function mergeBossState(currentBoss, incomingBoss) {
  if (!incomingBoss || typeof incomingBoss !== 'object') {
    return incomingBoss ?? null
  }
  const normalizedIncoming = normalizeBossState(incomingBoss)
  if (!currentBoss || typeof currentBoss !== 'object') {
    return normalizedIncoming
  }
  const normalizedCurrent = normalizeBossState(currentBoss)

  if (normalizedCurrent.id !== normalizedIncoming.id) {
    return normalizedIncoming
  }
  if (normalizedCurrent.status !== 'active' || normalizedIncoming.status !== 'active') {
    return normalizedIncoming
  }

  const currentHp = hpBigInt(normalizedCurrent.currentHp)
  const incomingHp = hpBigInt(normalizedIncoming.currentHp)
  if (incomingHp > currentHp) {
    return normalizedCurrent
  }

  return normalizedIncoming
}
