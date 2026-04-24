export const DEFAULT_RARITY = '普通'

export const RARITY_OPTIONS = ['普通', '优秀', '稀有', '史诗', '传说', '至臻']

export function normalizeRarity(value) {
  return RARITY_OPTIONS.includes(value) ? value : DEFAULT_RARITY
}

export function getRarityClassName(value) {
  switch (normalizeRarity(value)) {
    case '优秀':
      return 'rarity-text rarity-text--fine'
    case '稀有':
      return 'rarity-text rarity-text--rare'
    case '史诗':
      return 'rarity-text rarity-text--epic'
    case '传说':
      return 'rarity-text rarity-text--legendary'
    case '至臻':
      return 'rarity-text rarity-text--supreme rarity-text--animated'
    default:
      return 'rarity-text rarity-text--common'
  }
}

export function formatRarityLabel(value) {
  return `${normalizeRarity(value)}`
}

export function splitEquipmentName(name) {
  const normalized = String(name || '')
  const match = normalized.match(/^(\p{Extended_Pictographic}\s*)(.+)$/u)
  if (!match) {
    return {
      prefix: '',
      text: normalized,
    }
  }

  return {
    prefix: match[1],
    text: match[2],
  }
}
