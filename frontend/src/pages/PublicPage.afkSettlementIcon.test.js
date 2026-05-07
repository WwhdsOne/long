import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

function sliceBetween(startMarker, endMarker) {
    const start = stateSource.indexOf(startMarker)
    const end = stateSource.indexOf(endMarker, start)
    return stateSource.slice(start, end)
}

describe('挂机结算图标时序', () => {
    it('先刷新资料再拉取挂机结算，避免奖励弹窗拿不到最新背包图标', () => {
        expect(stateSource).toContain('async function refreshProfileAndLoadAfkSettlement()')
        expect(stateSource).toContain('await loadPlayerProfile()')
        expect(stateSource).toContain('await loadAfkSettlement()')

        const helperSegment = sliceBetween(
            'async function refreshProfileAndLoadAfkSettlement()',
            'async function submitNickname()',
        )
        expect(helperSegment.indexOf('await loadPlayerProfile()')).toBeLessThan(helperSegment.indexOf('await loadAfkSettlement()'))
    })

    it('登录、恢复会话和切回页面都走统一刷新流程', () => {
        const submitSegment = sliceBetween('async function submitNickname()', 'async function resetNickname()')
        const sessionSegment = sliceBetween('async function loadPlayerSession()', 'function registerPublicPageLifecycle()')
        const lifecycleSegment = sliceBetween('function registerPublicPageLifecycle()', 'onBeforeUnmount(() => {')

        expect(submitSegment).toContain('await refreshProfileAndLoadAfkSettlement()')
        expect(sessionSegment).toContain('await refreshProfileAndLoadAfkSettlement()')
        expect(lifecycleSegment).toContain('void refreshProfileAndLoadAfkSettlement()')
        expect(lifecycleSegment).not.toContain('void loadAfkSettlement()')
    })
})
