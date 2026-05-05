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

function bossPartKey(part) {
  if (!part || typeof part !== 'object') return ''
  return `${Number(part.x)}:${Number(part.y)}`
}

function mergeBossParts(currentParts, incomingParts) {
  const normalizedCurrent = Array.isArray(currentParts) ? currentParts.map(normalizeBossPart) : []
  const normalizedIncoming = Array.isArray(incomingParts) ? incomingParts.map(normalizeBossPart) : []
  if (normalizedCurrent.length === 0) {
    return normalizedIncoming
  }
  if (normalizedIncoming.length === 0) {
    return normalizedCurrent
  }

  const mergedByKey = new Map()
  for (const part of normalizedCurrent) {
    mergedByKey.set(bossPartKey(part), part)
  }
  for (const part of normalizedIncoming) {
    const key = bossPartKey(part)
    const current = mergedByKey.get(key)
    mergedByKey.set(key, current ? {...current, ...part} : part)
  }

  return normalizedCurrent.map((part) => mergedByKey.get(bossPartKey(part)) || part)
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

  return {
    ...normalizedCurrent,
    ...normalizedIncoming,
    parts: mergeBossParts(normalizedCurrent.parts, normalizedIncoming.parts),
  }
}
