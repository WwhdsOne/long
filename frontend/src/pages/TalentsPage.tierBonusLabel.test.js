import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const backendTalentSource = readFileSync(path.resolve(currentDir, '../../../backend/internal/core/talent.go'), 'utf8')

describe('护甲树层满奖励文案', () => {
    it('第 2 层满文案不再显示崩塌触发 -30，只保留全伤害 +10%', () => {
        expect(backendTalentSource).toContain('0: "全伤害 +10%",')
        expect(backendTalentSource).toContain('1: "全伤害 +10%",')
        expect(backendTalentSource).not.toContain('1: "崩塌触发 -30 + 全伤害 +10%",')
    })
})
