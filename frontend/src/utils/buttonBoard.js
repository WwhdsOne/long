function normalizeText(value) {
  return String(value || '').trim().toLowerCase()
}

export function collectButtonTags(buttons) {
  const tags = new Set()
  for (const button of Array.isArray(buttons) ? buttons : []) {
    for (const tag of Array.isArray(button?.tags) ? button.tags : []) {
      const normalized = String(tag || '').trim()
      if (normalized) {
        tags.add(normalized)
      }
    }
  }

  return Array.from(tags).sort((left, right) => left.localeCompare(right, 'zh-CN'))
}

export function filterAndSortButtons(buttons, options = {}) {
  const selectedTag = String(options.selectedTag || '全部').trim()
  const query = normalizeText(options.query)
  const activeKeys = new Set(Array.isArray(options.activeStarlightKeys) ? options.activeStarlightKeys : [])

  const filtered = (Array.isArray(buttons) ? buttons : []).filter((button) => {
    const tags = Array.isArray(button?.tags) ? button.tags.map((tag) => String(tag || '').trim()).filter(Boolean) : []
    if (selectedTag && selectedTag !== '全部' && !tags.includes(selectedTag)) {
      return false
    }

    if (!query) {
      return true
    }

    const label = normalizeText(button?.label)
    if (label.includes(query)) {
      return true
    }

    return tags.some((tag) => normalizeText(tag).includes(query))
  })

  return filtered.slice().sort((left, right) => {
    const leftActive = activeKeys.has(left.key)
    const rightActive = activeKeys.has(right.key)
    if (leftActive !== rightActive) {
      return leftActive ? -1 : 1
    }
    if ((left.sort ?? 0) !== (right.sort ?? 0)) {
      return (left.sort ?? 0) - (right.sort ?? 0)
    }
    return String(left.key || '').localeCompare(String(right.key || ''), 'zh-CN')
  })
}

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
