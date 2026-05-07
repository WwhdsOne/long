import {describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './BattleFxGalleryPage.vue'), 'utf8')
const canvasSource = readFileSync(path.resolve(currentDir, '../components/PixelEffectCanvas.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')
const compactPageSource = pageSource.replace(/\s+/g, ' ')

describe('BattleFxGalleryPage 页面结构', () => {
    it('包含上方小技能区与下方 5x5 终极技能演示区', () => {
        expect(pageSource).toContain('小技能特效')
        expect(pageSource).toContain('5x5 终极技能演示')
        expect(pageSource).toContain('bfxg__gallery')
        expect(pageSource).toContain('bfxg__ultimate-card')
    })

    it('覆盖后端实际存活的 8 个 EffectType', () => {
        const keys = [
            'storm_combo', 'auto_strike', 'bleed', 'final_cut',
            'collapse_trigger', 'judgment_day', 'doom_mark', 'silver_storm',
        ]
        for (const key of keys) {
            expect(pageSource).toContain(key)
        }
    })

    it('不包含后端已移除的特效 key', () => {
        expect(pageSource).not.toContain('omen_harvest')
        expect(pageSource).not.toContain('crosshair-mark')
        expect(pageSource).not.toContain('doom_judgment')
    })

    it('使用 PixelEffectCanvas 组件渲染 Canvas 像素动画', () => {
        expect(pageSource).toContain("import PixelEffectCanvas from '../components/PixelEffectCanvas.vue'")
        expect(pageSource).toContain('<PixelEffectCanvas')
        expect(pageSource).toContain(':effect="fx.key"')
        expect(compactPageSource).toContain('<PixelEffectCanvas :effect="ultimate.key" :size="ultimate.size || 90" :loop="true"/>')
    })

    it('不依赖 OSS 图片或外链资源', () => {
        expect(pageSource).not.toContain('oss-cn-beijing')
        expect(pageSource).not.toContain('https://hai-world2')
        expect(canvasSource).not.toContain('img src')
        expect(canvasSource).not.toContain('new Image')
    })

    it('Canvas 组件包含全部 8 种渲染器', () => {
        const rendererKeys = [
            'storm_combo:', 'auto_strike:', 'bleed:', 'final_cut:',
            'collapse_trigger:', 'judgment_day:', 'doom_mark:', 'silver_storm:',
        ]
        for (const key of rendererKeys) {
            expect(canvasSource).toContain(key)
        }
    })

    it('白银风暴改为 10 道自上而下依次就位，整条路径保留后统一碎裂', () => {
        const silverStormSection = canvasSource.slice(
            canvasSource.indexOf('// ---- 9. silver_storm'),
            canvasSource.indexOf('// ---- 10. crosshair-mark'),
        )
        expect(silverStormSection).toContain('screenFlash')
        expect(silverStormSection).toContain('for (let i = 0; i < 10; i++)')
        expect(silverStormSection).toContain('width: 1 + Math.floor(Math.random() * 2)')
        expect(silverStormSection).toContain('topX')
        expect(silverStormSection).toContain('botX')
        expect(silverStormSection).toContain('headDist')
        expect(silverStormSection).toContain("s.phase = 'hold'")
        expect(silverStormSection).toContain('s.timer >= 500')
        expect(silverStormSection).toContain("s.phase = 'shatter'")
        expect(silverStormSection).toContain('s.shardPixels.push(makeParticle(')
        expect(silverStormSection).toContain('x: px + w * 1.1 + rnd(0.6)')
    })

    it('CSS 不包含 animation 关键帧（针对 demo 区域）', () => {
        const demoSection = styleSource.substring(styleSource.indexOf('battle-fx-gallery-page'))
        expect(demoSection).not.toContain('@keyframes')
        expect(demoSection).not.toContain('animation:')
    })

    it('终极技能区明确写出 5x5 画布，并复刻战斗页同款 Boss 网格尺寸与覆盖层', () => {
        expect(pageSource).toContain('终极技能画布为 5x5')
        expect(pageSource).toContain('const ultimateSkills = [')
        expect(pageSource).toContain("key: 'final_cut'")
        expect(pageSource).toContain("trigger: 'final_cut'")
        expect(pageSource).toContain("key: 'silver_storm'")
        expect(pageSource).toContain('class="boss-part-grid bfxg__ultimate-grid"')
        expect(pageSource).toContain('class="bfxg__ultimate-overlay"')
        expect(styleSource).toContain('.bfxg__ultimate-grid {')
        expect(styleSource).toContain('max-width: 560px;')
        expect(styleSource).toContain('.bfxg__ultimate-overlay canvas {')
        expect(styleSource).toContain('width: 100% !important;')
    })
})
