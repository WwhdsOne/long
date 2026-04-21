const DEFAULT_BUFFER_MS = 150
const DEFAULT_EXPIRED_RETRY_DELAY_MS = 1000

export function resolveStarlightRefreshPlan({
  endsAtSeconds,
  nowMs = Date.now(),
  bufferMs = DEFAULT_BUFFER_MS,
  expiredRetryDelayMs = DEFAULT_EXPIRED_RETRY_DELAY_MS,
  lastExpiredEndsAtSeconds = 0,
} = {}) {
  const endsAt = Number(endsAtSeconds ?? 0)
  if (!Number.isFinite(endsAt) || endsAt <= 0) {
    return {
      delayMs: null,
      expiredRetryEndsAtSeconds: 0,
    }
  }

  const delayMs = endsAt * 1000 - nowMs + bufferMs
  if (delayMs > 0) {
    return {
      delayMs,
      expiredRetryEndsAtSeconds: 0,
    }
  }

  if (endsAt === Number(lastExpiredEndsAtSeconds || 0)) {
    return {
      delayMs: null,
      expiredRetryEndsAtSeconds: endsAt,
    }
  }

  return {
    delayMs: expiredRetryDelayMs,
    expiredRetryEndsAtSeconds: endsAt,
  }
}
