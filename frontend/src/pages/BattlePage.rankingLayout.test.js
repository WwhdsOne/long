import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('BattlePage 右侧榜单布局', () => {
    it('收窄战斗页右侧榜单列宽，把空间让给左侧战斗区', () => {
        expect(styleSource).toContain('.stage-layout--battle {')
        expect(styleSource).toContain('grid-template-columns: minmax(0, 1fr) minmax(220px, 260px);')
    })

    it('缩小右侧榜单内容字号和内边距', () => {
        expect(styleSource).toContain('.leaderboard-list__item {')
        expect(styleSource).toContain('padding: 11px 12px;')
        expect(styleSource).toContain('.leaderboard-list__name {')
        expect(styleSource).toContain('font-size: 0.88rem;')
        expect(styleSource).toContain('.leaderboard-list__rank,')
        expect(styleSource).toContain('font-size: 1.02rem;')
        expect(styleSource).toContain('.leaderboard-card__hint {')
        expect(styleSource).toContain('font-size: 0.74rem;')
    })
})
