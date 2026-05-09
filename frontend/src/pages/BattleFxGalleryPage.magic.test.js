import {describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './BattleFxGalleryPage.vue'), 'utf8')
const canvasSource = readFileSync(path.resolve(currentDir, '../components/PixelEffectCanvas.vue'), 'utf8')

describe('BattleFxGalleryPage 魔法特效入口', () => {
    it('图鉴墙包含魔法爆裂、裂解和星陨潮爆', () => {
        expect(pageSource).toContain("key: 'magic_burst'")
        expect(pageSource).toContain("key: 'magic_rupture'")
        expect(pageSource).toContain("key: 'magic_starfall'")
    })

    it('魔法三特效使用独立渲染器，而不是继续复用旧的暴击和碎盾动画', () => {
        expect(canvasSource).toContain('const magicBurstRenderer = {')
        expect(canvasSource).toContain('const magicRuptureRenderer = {')
        expect(canvasSource).toContain('const magicStarfallRenderer = {')
        expect(canvasSource).toContain('function drawPixelRing(ctx, cx, cy, radius, thickness, color) {')
        expect(canvasSource).toContain('magic_burst: magicBurstRenderer')
        expect(canvasSource).toContain('magic_rupture: magicRuptureRenderer')
        expect(canvasSource).toContain('magic_starfall: magicStarfallRenderer')
        expect(canvasSource).toContain("drawPixelRing(ctx, cx, cy, outerRingRadius, 2.4, '#93c5fd')")
        expect(canvasSource).toContain("drawPixelRing(ctx, cx, cy, innerRingRadius, 2.2, '#1d4ed8')")
    })

    it('图鉴页文案同步描述新的闪电、法球碎裂和中心扩散表现', () => {
        expect(pageSource).toContain('一道蓝白闪电从上方霹雳直落到目标格，命中后会回卷出电弧和余辉火花')
        expect(pageSource).toContain('几颗奥术蓝紫法球在格心短暂显现')
        expect(pageSource).toContain('一颗蓝紫流星从左上斜坠砸进 5x5 中心，落地先炸出陨坑感爆裂')
    })

    it('星陨潮爆先有左上流星命中陨坑爆裂，再进入高密度双层扩散环', () => {
        expect(canvasSource).toContain("phase: 'meteor'")
        expect(canvasSource).toContain('meteorTrailPixels: []')
        expect(canvasSource).toContain('meteorShards: []')
        expect(canvasSource).toContain('meteorStartX: size * 0.12')
        expect(canvasSource).toContain('meteorStartY: size * 0.14')
        expect(canvasSource).toContain("s.phase = 'impact'")
        expect(canvasSource).toContain("s.phase = 'expand'")
        expect(canvasSource).toContain("if (s.phase === 'impact') {")
        expect(canvasSource).toContain('const debrisKick = Math.max(0, 1 - s.timer / 120) * 4')
        expect(canvasSource).toContain("drawPixelRing(ctx, cx, cy, craterRadius, 2.8, '#7dd3fc')")
        expect(canvasSource).toContain("ctx.fillStyle = '#312e81'")
        expect(canvasSource).toContain("ctx.fillStyle = '#60a5fa'")
        expect(canvasSource).toContain("ctx.fillStyle = '#1e1b4b'")
        expect(canvasSource).toContain("if (s.phase === 'expand' || s.phase === 'fade' || s.phase === 'done') {")
        expect(canvasSource).toContain('const ringCount = 84')
        expect(canvasSource).toContain('const innerBandOffset = Math.max(3, size * 0.022)')
        expect(canvasSource).toContain('const wobble = Math.sin(progress * 4 + i * 0.72) * size * 0.006')
        expect(canvasSource).toContain('const outerRadius = s.waveRadius + wobble')
        expect(canvasSource).toContain('const innerRadius = Math.max(0, outerRadius - innerBandOffset)')
        expect(canvasSource).toContain("color: i % 18 === 0 ? '#fcd34d' : i % 6 === 0 ? '#c4b5fd' : '#7dd3fc'")
        expect(canvasSource).toContain("color: i % 18 === 0 ? '#f59e0b' : i % 6 === 0 ? '#60a5fa' : '#1e3a8a'")
        expect(canvasSource).toContain("Math.random() < 0.12 ? '#fcd34d' : Math.random() < 0.22 ? '#fde68a' : Math.random() < 0.62 ? '#60a5fa' : '#8b5cf6'")
    })

    it('奥术爆裂延长了命中后表现，并加入回卷电弧与余辉环', () => {
        expect(canvasSource).toContain("s.phase = 'afterglow'")
        expect(canvasSource).toContain('arcPixels: []')
        expect(canvasSource).toContain('ringPixels: []')
        expect(canvasSource).toContain('for (let i = 0; i < 18; i++) {')
        expect(canvasSource).toContain('for (let branch = 0; branch < 3; branch++) {')
        expect(canvasSource).toContain('const flashSize = Math.max(12, Math.round(size * 0.34))')
        expect(canvasSource).toContain("if (s.timer > 220) {")
        expect(canvasSource).toContain("return s.phase === 'done' && s.timer > 700")
    })

    it('星陨潮爆结束判定比原先更晚，避免刚出流星就被提前收掉', () => {
        expect(canvasSource).toContain("if (s.timer > 420) {")
        expect(canvasSource).toContain('const progress = Math.min(1, s.timer / 2600)')
        expect(canvasSource).toContain("return s.phase === 'done' && s.timer > 300")
    })
})
