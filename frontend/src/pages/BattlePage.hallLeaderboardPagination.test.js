import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('BattlePage 大厅底部分页总榜', () => {
    it('大厅底部新增 11-50 起步的分页总榜区', () => {
        expect(battleSource).toContain('hallLeaderboardSnapshot')
        expect(battleSource).toContain('hallLeaderboardPageEntries')
        expect(battleSource).toContain('11-50')
        expect(battleSource).toContain('大厅点击总榜')
        expect(battleSource).toContain('hall-leaderboard-panel')
    })

    it('分页区提供上一页下一页和四列布局', () => {
        expect(battleSource).toContain('hallLeaderboardColumns')
        expect(battleSource).toContain('上一页')
        expect(battleSource).toContain('下一页')
        expect(styleSource).toContain('.hall-leaderboard-panel__grid {')
        expect(styleSource).toContain('grid-template-columns: repeat(4, minmax(0, 1fr));')
    })
})
