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

export function summarizePointerTrajectory(points = []) {
  const normalized = points
    .filter((point) => Number.isFinite(point?.x) && Number.isFinite(point?.y) && Number.isFinite(point?.t))
    .map((point) => ({
      x: Number(point.x),
      y: Number(point.y),
      t: Number(point.t),
    }))

  if (normalized.length < 2) {
    return {
      pointCount: normalized.length,
      pathDistance: 0,
      displacement: 0,
      curvature: 0,
      speedVariance: 0,
    }
  }

  let pathDistance = 0
  let curvature = 0
  const speeds = []

  for (let index = 1; index < normalized.length; index += 1) {
    const dx = normalized[index].x - normalized[index - 1].x
    const dy = normalized[index].y - normalized[index - 1].y
    const dt = normalized[index].t - normalized[index - 1].t
    if (dt <= 0) {
      continue
    }

    const distance = Math.hypot(dx, dy)
    pathDistance += distance
    speeds.push(distance / dt)

    if (index >= 2) {
      const prevDX = normalized[index - 1].x - normalized[index - 2].x
      const prevDY = normalized[index - 1].y - normalized[index - 2].y
      const prevAngle = Math.atan2(prevDY, prevDX)
      const nextAngle = Math.atan2(dy, dx)
      let turn = Math.abs(nextAngle - prevAngle)
      if (turn > Math.PI) {
        turn = (Math.PI * 2) - turn
      }
      curvature += turn
    }
  }

  const displacement = Math.hypot(
    normalized.at(-1).x - normalized[0].x,
    normalized.at(-1).y - normalized[0].y,
  )

  const mean = speeds.reduce((total, value) => total + value, 0) / Math.max(speeds.length, 1)
  const variance = speeds.reduce((total, value) => {
    const diff = value - mean
    return total + (diff * diff)
  }, 0) / Math.max(speeds.length, 1)

  return {
    pointCount: normalized.length,
    pathDistance,
    displacement,
    curvature,
    speedVariance: mean > 0 ? Math.sqrt(variance) / mean : 0,
  }
}

function pointFromEvent(event, timestamp) {
  return {
    x: Number(event?.clientX || 0),
    y: Number(event?.clientY || 0),
    t: Number(timestamp || 0),
  }
}

function trimTrajectory(points, maxPoints) {
  if (!Array.isArray(points)) {
    return []
  }
  const filtered = points.filter((point) => Number.isFinite(point?.x) && Number.isFinite(point?.y) && Number.isFinite(point?.t))
  if (filtered.length <= maxPoints) {
    return filtered
  }
  return filtered.slice(filtered.length - maxPoints)
}

export function createClickBehaviorTracker(options = {}) {
  const now = options.now || (() => Date.now())
  const trailWindowMs = Number(options.trailWindowMs || 600)
  const freshnessWindowMs = Number(options.freshnessWindowMs || 1500)
  const maxPoints = Number(options.maxPoints || 12)
  const recentTrail = []
  const activePresses = new Map()
  const completedPresses = new Map()

  function appendRecentPoint(event) {
    const timestamp = eventTimestamp(event, now)
    recentTrail.push({
      ...pointFromEvent(event, timestamp),
      pointerType: String(event?.pointerType || 'mouse'),
    })
    while (recentTrail.length > maxPoints * 4) {
      recentTrail.shift()
    }
    while (recentTrail.length > 0 && timestamp - recentTrail[0].t > trailWindowMs) {
      recentTrail.shift()
    }
  }

  function recentPrelude(pointerType, timestamp) {
    return trimTrajectory(
      recentTrail.filter((point) => {
        if (timestamp - point.t > trailWindowMs) {
          return false
        }
        return !pointerType || point.pointerType === pointerType
      }),
      maxPoints,
    )
  }

  return {
    handleGlobalPointerMove(event) {
      appendRecentPoint(event)
    },
    handlePressStart(key, event) {
      const timestamp = eventTimestamp(event, now)
      const pointerType = String(event?.pointerType || 'mouse')
      const trajectory = [...recentPrelude(pointerType, timestamp), pointFromEvent(event, timestamp)]
      activePresses.set(String(key), {
        pointerType,
        startedAt: timestamp,
        trajectory: trimTrajectory(trajectory, maxPoints),
      })
    },
    handlePressEnd(key, event) {
      const normalizedKey = String(key)
      const active = activePresses.get(normalizedKey)
      if (!active) {
        return
      }

      const timestamp = eventTimestamp(event, now)
      const trajectory = trimTrajectory([...active.trajectory, pointFromEvent(event, timestamp)], maxPoints)
      completedPresses.set(normalizedKey, {
        pointerType: active.pointerType,
        pressDurationMs: Math.max(0, Math.round(timestamp - active.startedAt)),
        trajectory,
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

      const origin = record.trajectory[0]?.t || 0
      return {
        pointerType: record.pointerType,
        pressDurationMs: record.pressDurationMs,
        trajectory: record.trajectory.map((point) => ({
          x: point.x,
          y: point.y,
          t: Math.max(0, Math.round(point.t - origin)),
        })),
      }
    },
    clear() {
      recentTrail.length = 0
      activePresses.clear()
      completedPresses.clear()
    },
  }
}
