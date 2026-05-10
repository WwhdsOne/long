function normalizeIntegerString(value, fallback = '0') {
    const raw = String(value ?? '').trim()
    return /^\d+$/.test(raw) ? raw.replace(/^0+(?=\d)/, '') || '0' : fallback
}

function normalizeBossPart(part) {
    if (!part || typeof part !== 'object') return part
    const damageAffinity = String(part.damageAffinity || '').trim()
    return {
        ...part,
        ...(damageAffinity ? {damageAffinity} : {}),
        maxHp: normalizeIntegerString(part.maxHp),
        currentHp: normalizeIntegerString(part.currentHp),
        armor: normalizeIntegerString(part.armor),
    }
}

function normalizeBossStaticPart(part) {
    if (!part || typeof part !== 'object') return part
    const damageAffinity = String(part.damageAffinity || '').trim()
    return {
        ...part,
        ...(damageAffinity ? {damageAffinity} : {}),
        maxHp: normalizeIntegerString(part.maxHp),
        armor: normalizeIntegerString(part.armor),
    }
}

function normalizeBossStaticPayload(bossStatic) {
    if (!bossStatic || typeof bossStatic !== 'object') {
        return null
    }
    return {
        ...bossStatic,
        maxHp: normalizeIntegerString(bossStatic.maxHp),
        goldOnKill: normalizeIntegerString(bossStatic.goldOnKill),
        stoneOnKill: normalizeIntegerString(bossStatic.stoneOnKill),
        talentPointsOnKill: normalizeIntegerString(bossStatic.talentPointsOnKill),
        parts: Array.isArray(bossStatic.parts) ? bossStatic.parts.map(normalizeBossStaticPart) : [],
    }
}

function normalizeBossRuntimePart(part) {
    if (!part || typeof part !== 'object') return part
    return {
        ...part,
        currentHp: normalizeIntegerString(part.currentHp),
        alive: Boolean(part.alive),
    }
}

function normalizeBossRuntimePayload(bossRuntime) {
    if (!bossRuntime || typeof bossRuntime !== 'object') {
        return null
    }
    return {
        ...bossRuntime,
        currentHp: normalizeIntegerString(bossRuntime.currentHp),
        parts: Array.isArray(bossRuntime.parts) ? bossRuntime.parts.map(normalizeBossRuntimePart) : [],
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
        if (!current) {
            mergedByKey.set(key, part)
            continue
        }

        const currentHp = hpBigInt(current.currentHp)
        const incomingHp = hpBigInt(part.currentHp)
        const useIncomingHp = incomingHp <= currentHp
        mergedByKey.set(key, {
            ...current,
            ...part,
            currentHp: useIncomingHp ? part.currentHp : current.currentHp,
            alive: useIncomingHp ? incomingHp > 0n : currentHp > 0n,
        })
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

export function buildBossStateFromSnapshot(payload) {
    const bossId = String(payload?.bossId || '').trim()
    const bossStatic = normalizeBossStaticPayload(payload?.bossStatic)
    const bossRuntime = normalizeBossRuntimePayload(payload?.bossRuntime)
    if (!bossId || !bossStatic || !bossRuntime) {
        return null
    }

    const runtimeByKey = new Map()
    for (const part of bossRuntime.parts) {
        runtimeByKey.set(bossPartKey(part), part)
    }

    return normalizeBossState({
        id: bossId,
        templateId: bossStatic.templateId,
        roomId: bossStatic.roomId,
        queueId: bossStatic.queueId,
        name: bossStatic.name,
        status: bossRuntime.status,
        maxHp: bossStatic.maxHp,
        currentHp: bossRuntime.currentHp,
        goldOnKill: bossStatic.goldOnKill,
        stoneOnKill: bossStatic.stoneOnKill,
        talentPointsOnKill: bossStatic.talentPointsOnKill,
        startedAt: bossStatic.startedAt,
        defeatedAt: bossRuntime.defeatedAt,
        parts: bossStatic.parts.map((part) => {
            const runtimePart = runtimeByKey.get(bossPartKey(part))
            const currentHp = runtimePart?.currentHp ?? part.maxHp
            return {
                ...part,
                currentHp,
                alive: runtimePart?.alive ?? hpBigInt(currentHp) > 0n,
            }
        }),
    })
}

export function applyBossDeltaMessage(currentState, payload) {
    const state = currentState && typeof currentState === 'object' ? currentState : {}
    const bossId = String(payload?.bossId || '').trim()
    const nextVersion = Number(payload?.bossVersion || 0)
    const bossStaticById = state.bossStaticById && typeof state.bossStaticById === 'object' ? state.bossStaticById : {}
    const currentBoss = normalizeBossState(state.boss)
    const currentVersion = Number(state.bossVersion || 0)

    if (!bossId || !Number.isFinite(nextVersion) || nextVersion <= 0) {
        return {
            bossStaticById,
            bossVersion: currentVersion,
            boss: currentBoss ?? null,
            shouldSync: false,
        }
    }
    if (!currentBoss || currentBoss.id !== bossId) {
        return {
            bossStaticById,
            bossVersion: currentVersion,
            boss: currentBoss ?? null,
            shouldSync: true,
        }
    }
    const bossStatic = normalizeBossStaticPayload(bossStaticById[bossId])
    if (!bossStatic) {
        return {
            bossStaticById,
            bossVersion: currentVersion,
            boss: currentBoss,
            shouldSync: true,
        }
    }
    if (nextVersion <= currentVersion) {
        return {
            bossStaticById,
            bossVersion: currentVersion,
            boss: currentBoss,
            shouldSync: false,
        }
    }
    if (currentVersion > 0 && nextVersion !== currentVersion + 1) {
        return {
            bossStaticById,
            bossVersion: currentVersion,
            boss: currentBoss,
            shouldSync: true,
        }
    }

    const runtime = normalizeBossRuntimePayload(payload?.bossRuntime)
    const staticPartsByKey = new Map()
    for (const part of bossStatic.parts) {
        staticPartsByKey.set(bossPartKey(part), part)
    }
    const incomingParts = Array.isArray(runtime?.parts)
        ? runtime.parts.map((part) => ({
            ...(staticPartsByKey.get(bossPartKey(part)) || {}),
            ...part,
        }))
        : []
    const nextBoss = mergeBossState(currentBoss, {
        id: bossId,
        templateId: bossStatic.templateId,
        roomId: bossStatic.roomId,
        queueId: bossStatic.queueId,
        name: bossStatic.name,
        status: runtime?.status ?? currentBoss.status,
        maxHp: bossStatic.maxHp,
        currentHp: runtime?.currentHp ?? currentBoss.currentHp,
        goldOnKill: bossStatic.goldOnKill,
        stoneOnKill: bossStatic.stoneOnKill,
        talentPointsOnKill: bossStatic.talentPointsOnKill,
        startedAt: bossStatic.startedAt,
        defeatedAt: runtime?.defeatedAt ?? currentBoss.defeatedAt,
        parts: incomingParts,
    })

    return {
        bossStaticById,
        bossVersion: nextVersion,
        boss: nextBoss,
        shouldSync: false,
    }
}

function hpBigInt(value) {
    try {
        return BigInt(normalizeIntegerString(value))
    } catch {
        return 0n
    }
}

function bigIntToString(value) {
    return value < 0n ? '0' : value.toString()
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
    return {
        ...normalizedCurrent,
        ...normalizedIncoming,
        currentHp: incomingHp > currentHp ? normalizedCurrent.currentHp : normalizedIncoming.currentHp,
        parts: mergeBossParts(normalizedCurrent.parts, normalizedIncoming.parts),
    }
}

export function applyBossPartStateDeltas(currentBoss, deltas) {
    if (!currentBoss || typeof currentBoss !== 'object') {
        return currentBoss ?? null
    }
    if (!Array.isArray(deltas) || deltas.length === 0) {
        return normalizeBossState(currentBoss)
    }

    const normalizedBoss = normalizeBossState(currentBoss)
    if (!Array.isArray(normalizedBoss.parts) || normalizedBoss.parts.length === 0) {
        return normalizedBoss
    }

    let nextBossHp = hpBigInt(normalizedBoss.currentHp)
    const nextParts = normalizedBoss.parts.map((part) => {
        const matched = deltas.find((delta) => Number(delta?.x) === Number(part.x) && Number(delta?.y) === Number(part.y))
        if (!matched) {
            return part
        }

        const currentPartHp = hpBigInt(part.currentHp)
        const beforePartHp = hpBigInt(matched.beforeHp)
        const nextPartHp = hpBigInt(matched.afterHp)
        if (currentPartHp > beforePartHp || nextPartHp >= currentPartHp) {
            return part
        }

        nextBossHp -= currentPartHp - nextPartHp
        return {
            ...part,
            currentHp: bigIntToString(nextPartHp),
            alive: nextPartHp > 0n,
        }
    })

    return {
        ...normalizedBoss,
        currentHp: bigIntToString(nextBossHp),
        parts: nextParts,
    }
}
