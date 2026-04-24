function safeCrypto() {
  return globalThis.crypto?.subtle ?? null
}

function encodeText(value) {
  return new TextEncoder().encode(String(value ?? ''))
}

function eventTimestamp(event, fallbackNow) {
  const timestamp = Number(event?.timeStamp)
  if (Number.isFinite(timestamp) && timestamp >= 0) {
    return timestamp
  }
  return Number(fallbackNow())
}

export async function sha256Hex(value) {
  const subtle = safeCrypto()
  if (!subtle) {
    throw new Error('当前环境不支持指纹摘要')
  }

  const digest = await subtle.digest('SHA-256', encodeText(value))
  return Array.from(new Uint8Array(digest))
    .map((item) => item.toString(16).padStart(2, '0'))
    .join('')
}

export async function collectFingerprintHash(win = globalThis) {
  const navigator = win.navigator ?? {}
  const screen = win.screen ?? {}
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || ''

  const source = {
    ua: navigator.userAgent || '',
    language: navigator.language || '',
    languages: Array.isArray(navigator.languages) ? navigator.languages.join(',') : '',
    platform: navigator.platform || '',
    timezone,
    screenWidth: Number(screen.width || 0),
    screenHeight: Number(screen.height || 0),
    colorDepth: Number(screen.colorDepth || 0),
    hardwareConcurrency: Number(navigator.hardwareConcurrency || 0),
    deviceMemory: Number(navigator.deviceMemory || 0),
    maxTouchPoints: Number(navigator.maxTouchPoints || 0),
    webdriver: Boolean(navigator.webdriver),
  }

  return sha256Hex(JSON.stringify(source))
}

export async function buildFingerprintProof({ fingerprintHash, ticket, challengeNonce }) {
  return sha256Hex(`${String(fingerprintHash ?? '').trim()}:${String(ticket ?? '').trim()}:${String(challengeNonce ?? '').trim()}`)
}

export function createClickBehaviorTracker(options = {}) {
  const now = options.now || (() => Date.now())
  const freshnessWindowMs = Number(options.freshnessWindowMs || 1500)
  const activePresses = new Map()
  const completedPresses = new Map()

  return {
    handlePressStart(key, event) {
      activePresses.set(String(key), {
        pointerType: String(event?.pointerType || 'mouse'),
        startedAt: eventTimestamp(event, now),
      })
    },
    handlePressEnd(key, event) {
      const normalizedKey = String(key)
      const active = activePresses.get(normalizedKey)
      if (!active) {
        return
      }

      const timestamp = eventTimestamp(event, now)
      completedPresses.set(normalizedKey, {
        pointerType: active.pointerType,
        pressDurationMs: Math.max(0, Math.round(timestamp - active.startedAt)),
        completedAt: Number(now()),
      })
      activePresses.delete(normalizedKey)
    },
    handlePressCancel(key) {
      activePresses.delete(String(key))
      completedPresses.delete(String(key))
    },
    consume(key) {
      const normalizedKey = String(key)
      const record = completedPresses.get(normalizedKey)
      completedPresses.delete(normalizedKey)
      if (!record) {
        return null
      }
      if (now() - record.completedAt > freshnessWindowMs) {
        return null
      }

      return {
        pointerType: record.pointerType,
        pressDurationMs: record.pressDurationMs,
      }
    },
    clear() {
      activePresses.clear()
      completedPresses.clear()
    },
  }
}
