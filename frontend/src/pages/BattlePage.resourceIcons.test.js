import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('BattlePage 资源掉落图标', () => {
    it('Boss 掉落池资源卡显示金币强化石刻印石和天赋点图标', () => {
        expect(pageSource).toContain('const resourceIcons = {')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E9%87%91%E5%B8%81.png')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E5%BC%BA%E5%8C%96%E7%9F%B3.png')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E5%88%BB%E5%8D%B0%E7%9F%B3.png')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E5%A4%A9%E8%B5%8B%E7%82%B9.png')
        expect(pageSource).toContain('boss-drop-card__resource-head')
        expect(pageSource).toContain('boss-drop-card__resource-icon')
        expect(pageSource).toContain('可获取刻印石量 : {{ bossInscriptionStoneRange.min }} ~ {{ bossInscriptionStoneRange.max }}')
    })

    it('公共资源状态会读取 Boss 刻印石区间', () => {
        expect(stateSource).toContain('const bossInscriptionStoneRange = ref({min: 0, max: 0})')
        expect(stateSource).toContain("bossInscriptionStoneRange.value = payload?.inscriptionStoneRange ?? {min: 0, max: 0}")
        expect(stateSource).toContain("if ('bossInscriptionStoneRange' in payload && payload.bossInscriptionStoneRange) {")
    })

    it('图标样式在装备页和 Boss 掉落池里都放大显示', () => {
        expect(styleSource).toContain('.armory-backpack-resources__icon')
        expect(styleSource).toContain('width: 40px;')
        expect(styleSource).toContain('.boss-drop-card__resource-icon')
        expect(styleSource).toContain('width: 28px;')
    })
})
