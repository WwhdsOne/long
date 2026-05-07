import {assetUrl} from './assets'

const soundRegistry = new Map([
    ['battle.click.soft', 'sfx/battle/click/soft.wav'],
    ['battle.click.heavy', 'sfx/battle/click/heavy.wav'],
    ['battle.click.weak', 'sfx/battle/click/weak.wav'],
    ['battle.trigger.pursuit', 'sfx/battle/trigger/pursuit.wav'],
    ['battle.trigger.armor-collapse', 'sfx/battle/trigger/armor-collapse.wav'],
    ['battle.trigger.silver-storm', 'sfx/battle/trigger/silver-storm.wav'],
    ['battle.trigger.judgment-day', 'sfx/battle/trigger/judgment-day.wav'],
    ['battle.trigger.final-cut', 'sfx/battle/trigger/final-cut.wav'],
    ['battle.trigger.auto-strike', 'sfx/battle/trigger/auto-strike.wav'],
])

const soundAliases = {
    soft: 'battle.click.soft',
    heavy: 'battle.click.heavy',
    weak: 'battle.click.weak',
    'click-soft': 'battle.click.soft',
    'click-heavy': 'battle.click.heavy',
    'click-weak': 'battle.click.weak',
    '软组织': 'battle.click.soft',
    '重甲': 'battle.click.heavy',
    '弱点': 'battle.click.weak',
    pursuit: 'battle.trigger.pursuit',
    storm_combo: 'battle.trigger.pursuit',
    silver_storm: 'battle.trigger.silver-storm',
    '白银风暴': 'battle.trigger.silver-storm',
    collapse_trigger: 'battle.trigger.armor-collapse',
    '护甲崩塌': 'battle.trigger.armor-collapse',
    judgment_day: 'battle.trigger.judgment-day',
    judgment: 'battle.trigger.judgment-day',
    judgement: 'battle.trigger.judgment-day',
    '审判日': 'battle.trigger.judgment-day',
    final_cut: 'battle.trigger.final-cut',
    '终末血斩': 'battle.trigger.final-cut',
    auto_strike: 'battle.trigger.auto-strike',
    '自动锤击音效': 'battle.trigger.auto-strike',
}

const triggerSoundPolicies = {
    'battle.trigger.pursuit': {
        delayMs: 500,
    },
    'battle.trigger.silver-storm': {
        cooldownMs: 3200,
    },
    'battle.trigger.judgment-day': {
        delayMs: 1000,
    },
    'battle.trigger.final-cut': {
        delayMs: 700,
    },
    'battle.trigger.auto-strike': {
        delayMs: 1000,
        layers: [
            {delayMs: 35, volume: 0.72},
        ],
    },
}

const triggerSoundLastPlayedAt = new Map()

function normalizeSoundKey(value) {
    return String(value || '').trim().toLowerCase()
}

function clampVolume(value) {
    const normalized = Number(value)
    if (!Number.isFinite(normalized)) {
        return 1
    }
    return Math.min(1, Math.max(0, normalized))
}

export function registerSoundEffect(id, path) {
    const key = String(id || '').trim()
    const nextPath = String(path || '').trim()
    if (!key || !nextPath) {
        return false
    }
    soundRegistry.set(key, nextPath)
    return true
}

export function resolveSoundEffectId(key) {
    const normalized = normalizeSoundKey(key)
    if (!normalized) {
        return ''
    }
    if (soundRegistry.has(normalized)) {
        return normalized
    }
    return soundAliases[normalized] || ''
}

export function resolveSoundEffectUrl(key) {
    const id = resolveSoundEffectId(key)
    if (!id) {
        return ''
    }
    const path = soundRegistry.get(id)
    return path ? assetUrl(path) : ''
}

export function playSoundEffect(key, options = {}) {
    const url = resolveSoundEffectUrl(key)
    if (!url || typeof Audio !== 'function') {
        return false
    }

    const audio = new Audio(url)
    audio.preload = 'auto'
    audio.volume = clampVolume(options.volume ?? 1)
    audio.play().catch(() => {
    })
    return true
}

function scheduleSoundEffect(key, options = {}) {
    const delayMs = Math.max(0, Number(options.delayMs) || 0)
    const invoke = () => playSoundEffect(key, options)
    if (delayMs <= 0 || typeof setTimeout !== 'function') {
        return invoke()
    }
    setTimeout(invoke, delayMs)
    return true
}

export function playBattlePartSound(partType, options = {}) {
    const effectKey = resolveSoundEffectId(partType)
    const clickableKey = {
        'battle.click.soft': 'battle.click.soft',
        'battle.click.heavy': 'battle.click.heavy',
        'battle.click.weak': 'battle.click.weak',
    }[effectKey]
    if (!clickableKey) {
        return false
    }
    return playSoundEffect(clickableKey, options)
}

export function playBattleTriggerSound(effectType, options = {}) {
    const effectKey = resolveSoundEffectId(effectType)
    if (!effectKey) {
        return false
    }

    const policy = triggerSoundPolicies[effectKey] || {}
    const cooldownMs = Math.max(0, Number(options.cooldownMs ?? policy.cooldownMs) || 0)
    if (cooldownMs > 0) {
        const now = Date.now()
        const lastPlayedAt = Number(triggerSoundLastPlayedAt.get(effectKey) || 0)
        if (lastPlayedAt > 0 && now - lastPlayedAt < cooldownMs) {
            return false
        }
        triggerSoundLastPlayedAt.set(effectKey, now)
    }

    const baseDelayMs = Math.max(0, Number(options.delayMs ?? policy.delayMs) || 0)
    const baseVolume = options.volume ?? policy.volume ?? 1
    const layers = Array.isArray(policy.layers) ? policy.layers : []

    let scheduled = scheduleSoundEffect(effectKey, {
        ...options,
        delayMs: baseDelayMs,
        volume: baseVolume,
    })
    for (const layer of layers) {
        scheduled = scheduleSoundEffect(effectKey, {
            ...options,
            delayMs: baseDelayMs + Math.max(0, Number(layer?.delayMs) || 0),
            volume: layer?.volume ?? baseVolume,
        }) || scheduled
    }
    return scheduled
}
