import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')
const shopPageSource = readFileSync(path.resolve(currentDir, './ShopPage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')
const battleSource = readFileSync(path.resolve(currentDir, './BattlePage.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')

describe('PublicPage 商店接线', () => {
    it('前台导航新增独立商店页并挂载 ShopPage 组件', () => {
        expect(pageSource).toContain("import ShopPage from './ShopPage.vue'")
        expect(pageSource).toContain("currentPublicPage === 'shop'")
        expect(stateSource).toContain("id: 'shop'")
        expect(stateSource).toContain("path: '/shop'")
        expect(stateSource).toContain("return 'shop'")
    })

    it('公共状态层提供商店列表加载、购买和使用动作', () => {
        expect(stateSource).toContain('const shopItems = ref([])')
        expect(stateSource).toContain('const loadingShopItems = ref(false)')
        expect(stateSource).toContain('async function loadShopItems()')
        expect(stateSource).toContain("fetch('/api/shop/items'")
        expect(stateSource).toContain('async function purchaseShopItem(itemId)')
        expect(stateSource).toContain('/api/shop/items/${encodeURIComponent(itemId)}/purchase')
        expect(stateSource).toContain('async function equipShopItem(itemId)')
        expect(stateSource).toContain('/api/shop/items/${encodeURIComponent(itemId)}/equip')
    })

    it('战斗页点击光标优先使用用户态中的商店皮肤并保留默认回退', () => {
        expect(stateSource).toContain('const equippedBattleClickCursorImagePath = ref(\'\')')
        expect(stateSource).toContain("'equippedBattleClickCursorImagePath' in payload")
        expect(shopPageSource).toContain('DEFAULT_BATTLE_CLICK_CURSOR_IMAGE')
        expect(shopPageSource).toContain("computed(() => equippedBattleClickCursorImagePath.value || DEFAULT_BATTLE_CLICK_CURSOR_IMAGE)")
        expect(shopPageSource).toContain('class="shop-current-cursor__image"')
        expect(shopPageSource).toContain('class="shop-cursor-card__image"')
        expect(shopPageSource).toContain('shop-panel shop-panel--cursor')
        expect(shopPageSource).toContain('class="shop-panel__header"')
        expect(shopPageSource).toContain('class="shop-panel__summary"')
        expect(shopPageSource).toContain('class="shop-cursor-grid"')
        expect(shopPageSource).toContain('class="shop-cursor-card"')
        expect(shopPageSource).toContain('class="shop-cursor-card__main"')
        expect(shopPageSource).toContain('class="shop-cursor-card__title"')
        expect(shopPageSource).toContain('class="shop-cursor-card__price"')
        expect(styleSource).toContain('.shop-current-cursor__image {')
        expect(styleSource).toContain('width: 72px;')
        expect(styleSource).toContain('height: 72px;')
        expect(styleSource).toContain('.shop-cursor-card__image {')
        expect(styleSource).toContain('object-fit: contain;')
        expect(styleSource).toContain('.shop-cursor-grid {')
        expect(styleSource).toContain('grid-template-columns: repeat(4, minmax(0, 1fr));')
        expect(styleSource).toContain('.shop-cursor-card {')
        expect(styleSource).toContain('grid-template-columns: 72px minmax(0, 1fr);')
        expect(battleSource).toContain('DEFAULT_BOSS_SWORD_CURSOR_URL')
        expect(battleSource).toContain('equippedBattleClickCursorImagePath')
        expect(battleSource).toContain("computed(() => equippedBattleClickCursorImagePath.value || DEFAULT_BOSS_SWORD_CURSOR_URL)")
    })
})
