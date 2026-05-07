import {describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {effectAssetUrl} from './effectAssets'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const effectAssetsSource = readFileSync(path.resolve(currentDir, './effectAssets.js'), 'utf8')

describe('effectAssets', () => {
    it('只依赖前端本地映射，不再引用 pixel-assets 目录', () => {
        expect(effectAssetsSource).not.toContain('pixel-assets/oss-url-map.json')
        expect(effectAssetsSource).toContain("./effectAssetMap")
    })

    it('会优先返回前端集中映射里的远端地址', () => {
        expect(effectAssetUrl('talent-crit_doom_judgment.png')).toBe(
            'https://hai-world2.oss-cn-beijing.aliyuncs.com/effects/talent-crit_omen_resonate.png',
        )
    })

    it('未命中映射时会回退到本地 effects 目录', () => {
        expect(effectAssetUrl('future-effect.png')).toBe('/effects/future-effect.png')
    })
})
