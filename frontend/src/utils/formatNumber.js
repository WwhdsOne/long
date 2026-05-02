const SUFFIXES = [
  { threshold: 1e9, suffix: 'B' },
  { threshold: 1e6, suffix: 'M' },
  { threshold: 1e3, suffix: 'K' },
  { threshold: 1, suffix: '' },
]

export function formatCompact(value) {
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
