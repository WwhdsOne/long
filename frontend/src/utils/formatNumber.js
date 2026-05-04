const SUFFIXES = [
  { threshold: 1000000000000000000n, suffix: 'Qi' },
  { threshold: 1000000000000000n, suffix: 'Qa' },
  { threshold: 1000000000000n, suffix: 'T' },
  { threshold: 1000000000n, suffix: 'B' },
  { threshold: 1000000n, suffix: 'M' },
  { threshold: 1000n, suffix: 'K' },
]

const COMPACT_SUFFIXES = [
  { threshold: 1e9, suffix: 'B' },
  { threshold: 1e6, suffix: 'M' },
  { threshold: 1e3, suffix: 'K' },
  { threshold: 1, suffix: '' },
]

const BIGINT_COMPACT_SUFFIXES = [
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

function scientificIntegerString(raw) {
  const negative = raw.startsWith('-')
  const digits = (negative ? raw.slice(1) : raw).replace(/^0+/, '') || '0'
  if (digits === '0') return '0'
  const exponent = digits.length - 1
  const fractional = digits.slice(1).replace(/0+$/, '')
  const mantissa = fractional ? `${digits[0]}.${fractional}` : digits[0]
  const formatted = exponent > 0 ? `${mantissa}e+${exponent}` : mantissa
  return negative ? `-${formatted}` : formatted
}

function suffixIntegerString(raw) {
  const big = BigInt(raw)
  const negative = big < 0n
  const abs = negative ? -big : big
  if (abs < 1000n) return raw

  const scientificThreshold = SUFFIXES[0].threshold * 1000n
  if (abs >= scientificThreshold) return scientificIntegerString(raw)

  for (const { threshold, suffix } of SUFFIXES) {
    if (abs < threshold) continue
    const whole = abs / threshold
    const decimal = (abs % threshold) * 10n / threshold
    const formatted = decimal > 0n ? `${whole}.${decimal}` : whole.toString()
    return `${negative ? '-' : ''}${formatted}${suffix}`
  }

  return raw
}

export function formatIntegerExact(value) {
  const raw = integerString(value)
  if (raw == null) return '0'
  return suffixIntegerString(raw)
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
    for (const { threshold, suffix } of BIGINT_COMPACT_SUFFIXES) {
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

  for (const { threshold, suffix } of COMPACT_SUFFIXES) {
    if (Math.abs(num) >= threshold) {
      const scaled = num / threshold
      const formatted = scaled % 1 === 0 ? scaled.toString() : Number(scaled.toFixed(1)).toString()
      return formatted + suffix
    }
  }
  return '0'
}
