import { assetUrl } from './assets'
import { effectAssetMap } from './effectAssetMap'

export function effectAssetUrl(filename) {
  const key = String(filename || '').trim()
  if (!key) return ''
  const remote = effectAssetMap[key]
  if (typeof remote === 'string' && remote.trim()) {
    return remote.trim()
  }
  return assetUrl(`effects/${key}`)
}
