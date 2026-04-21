export function buildPityProgress(counter = 0, threshold = 30) {
  const normalizedThreshold = Math.max(0, Number(threshold) || 0)
  const normalizedCurrent = Math.max(0, Math.min(Number(counter) || 0, normalizedThreshold))
  const total = normalizedThreshold + 1

  return {
    current: normalizedCurrent,
    threshold: normalizedThreshold,
    remaining: Math.max(1, total - normalizedCurrent),
    percent: total <= 1 ? 100 : Math.round((normalizedCurrent / normalizedThreshold) * 100),
    label: `${normalizedCurrent} / ${total}`,
  }
}
