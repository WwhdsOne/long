export function formatDropRate(value) {
  const numeric = Number(value ?? 0)
  if (!Number.isFinite(numeric) || numeric <= 0) {
    return '0%'
  }

  const rounded = Math.round(numeric * 100) / 100
  if (Number.isInteger(rounded)) {
    return `${rounded}%`
  }
  return `${rounded.toFixed(2)}%`
}
