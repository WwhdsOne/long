import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const source = readFileSync(path.resolve(currentDir, './RoomSelector.vue'), 'utf8')
const compactSource = source.replace(/\s+/g, ' ')

describe('RoomSelector 紧凑卡片', () => {
    it('房间卡仅展示 Boss 均血，不再展示战斗力准入', () => {
        expect(source).toContain("import RoomSwitchCooldownTag from './RoomSwitchCooldownTag.vue'")
        expect(compactSource).toContain('<RoomSwitchCooldownTag :cooldown-remaining-seconds="cooldownRemainingSeconds"/>')
        expect(source).toContain("const displayName = String(room?.displayName || '').trim()")
        expect(source).toContain("return defaultRoomLabel(room?.id)")
        expect(source).toContain("if (props.cooldownRemainingSeconds > 0) return '冷却未结束'")
        expect(source).toContain('function roomAvgHpText(room) {')
        expect(source).toContain('<span class="room-selector__bossline">')
        expect(source).toContain('<small>Boss平均血量</small>')
        expect(source).toContain('<span class="room-selector__action">{{ roomActionLabel(room) }}</span>')
        expect(source).not.toContain('roomBattlePowerRangeText')
        expect(source).not.toContain('准入 ')
    })

    it('大厅房间卡带有卡片本体呼吸光和沿边缘巡游的光点', () => {
        expect(source).toContain('<span class="room-selector__surface" aria-hidden="true"></span>')
        expect(source).toContain("'--room-orbit-delay':")
        expect(source).toContain("'--room-orbit-duration': `${5.8 + (offset % 3) * 0.45}s`")
        expect(source).toContain('.room-selector__surface {')
        expect(source).toContain('@keyframes room-card-surface-breathe')
        expect(source).toContain('@keyframes room-card-breathe')
        expect(source).toContain('@keyframes room-card-orbit')
        expect(source).toContain('.room-selector__surface {')
        expect(source).toContain('animation: room-card-surface-breathe')
        expect(source).toContain('linear-gradient(180deg, rgba(var(--room-accent-rgb), 0.14), rgba(var(--room-accent-rgb), 0.05) 34%, rgba(255, 255, 255, 0.03) 100%)')
        expect(source).toContain('opacity: 0.5;')
        expect(source).toContain('filter: brightness(0.98) saturate(1.12);')
        expect(source).toContain('opacity: 0.86;')
        expect(source).toContain('filter: brightness(1.08) saturate(1.32);')
        expect(source).toContain('transform: scale(1.018);')
        expect(source).toContain('animation: room-card-breathe')
        expect(source).toContain('animation: room-card-orbit')
        expect(source).toContain('-webkit-mask-composite: xor')
        expect(source).toContain('mask-composite: exclude')
        expect(source).toContain('padding: 1px')
        expect(source).toContain('width: 6px')
        expect(source).toContain('height: 6px')
        expect(source).toContain('background: rgba(var(--room-accent-rgb), 0.98);')
        expect(source).toContain('-100px 0 32px rgba(var(--room-accent-rgb), 0.08)')
        expect(source).toContain('rotate(90deg)')
        expect(source).toContain('.room-selector__button--locked:not(.room-selector__button--active)::before')
        expect(source).toContain('.room-selector__button--locked:not(.room-selector__button--active)::after')
        expect(source).not.toContain('background: rgb(118 227 214 / 0.98);')
        expect(source).not.toContain('background: radial-gradient(circle at 50% 50%')
    })
})
