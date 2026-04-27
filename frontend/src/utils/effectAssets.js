import ossUrlMap from '../../../pixel-assets/oss-url-map.json'
import { assetUrl } from './assets'

export function effectAssetUrl(filename) {
  const key = String(filename || '').trim()
  if (!key) return ''
  const remote = ossUrlMap[key]
  if (typeof remote === 'string' && remote.trim()) {
    return remote.trim()
  }
  return assetUrl(`effects/${key}`)
}

