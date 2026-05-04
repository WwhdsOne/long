const SUFFIXES = [
  { threshold: 1e9, suffix: 'B' },
  { threshold: 1e6, suffix: 'M' },
  { threshold: 1e3, suffix: 'K' },
  { threshold: 1, suffix: '' },
]

const BIGINT_SUFFIXES = [
  { threshold: 1000000000n, suffix: 'B' },
  { threshold: 1000000n, suffix: 'M' },
  { threshold: 1000n, suffix: 'K' },
  { threshold: 1n, suffix: '' },
]

function integerString(value) {
  if (typeof value === 'bigint') return value.toString()
  if (typeof value === 'number') {
    if (!Number.isFinite(value)) return null
    return Number.isInteger(value) ? Math.trunc(value).toString() : null
  }
  const raw = String(value ?? '').trim()
  return /^-?\d+$/.test(raw) ? raw : null
}

function bigintValue(value) {
  const raw = integerString(value)
  if (raw == null) return null
  try {
    return BigInt(raw)
  } catch {
    return null
  }
}

function groupDigits(raw) {
  const negative = raw.startsWith('-')
  const digits = negative ? raw.slice(1) : raw
  const grouped = digits.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return negative ? `-${grouped}` : grouped
}

export function formatIntegerExact(value) {
  const raw = integerString(value)
  if (raw == null) return '0'
  return groupDigits(raw)
}

export function ratioPercent(current, max) {
  const currentValue = bigintValue(current)
  const maxValue = bigintValue(max)
  if (currentValue == null || maxValue == null || maxValue <= 0n) return 0
  const clamped = currentValue < 0n ? 0n : (currentValue > maxValue ? maxValue : currentValue)
  const basisPoints = (clamped * 10000n) / maxValue
  return Number(basisPoints) / 100
}

export function formatCompact(value) {
  const big = bigintValue(value)
  if (big != null) {
    const negative = big < 0n
    const abs = negative ? -big : big
    for (const { threshold, suffix } of BIGINT_SUFFIXES) {
      if (abs >= threshold) {
        const whole = abs / threshold
        const decimal = (abs % threshold) * 10n / threshold
        const formatted = decimal > 0n ? `${whole}.${decimal}` : whole.toString()
        return (negative ? '-' : '') + formatted + suffix
      }
    }
    return '0'
  }

  const num = Number(value ?? 0)
  if (!Number.isFinite(num)) return '0'

  for (const { threshold, suffix } of SUFFIXES) {
    if (Math.abs(num) >= threshold) {
      const scaled = num / threshold
      const formatted = scaled % 1 === 0 ? scaled.toString() : Number(scaled.toFixed(1)).toString()
      return formatted + suffix
    }
  }
  return '0'
}
