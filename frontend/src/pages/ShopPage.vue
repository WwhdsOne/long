<script setup>
import {computed} from 'vue'

import {usePublicPageState} from './publicPageState'

const {
  isLoggedIn,
  gold,
  shopItems,
  loadingShopItems,
  equippedBattleClickCursorImagePath,
  purchaseShopItem,
  equipShopItem,
  unequipShopItem,
} = usePublicPageState()

const DEFAULT_BATTLE_CLICK_CURSOR_IMAGE = 'https://hai-world2.oss-cn-beijing.aliyuncs.com/effects/click-sword_basic.png'
const currentCursorImage = computed(() => equippedBattleClickCursorImagePath.value || DEFAULT_BATTLE_CLICK_CURSOR_IMAGE)

function buttonLabel(item) {
  if (!isLoggedIn.value) return '先登录'
  if (item.equipped) return '使用中'
  if (item.owned) return '使用'
  return '购买'
}

function buttonDisabled(item) {
  if (!isLoggedIn.value) return true
  if (item.equipped) return true
  if (!item.owned && Number(item.priceGold || 0) > Number(gold.value || 0)) return true
  return false
}

async function handleShopAction(item) {
  if (item.equipped || !isLoggedIn.value) {
    return
  }
  if (item.owned) {
    await equipShopItem(item.itemId)
    return
  }
  await purchaseShopItem(item.itemId)
}
</script>

<template>
  <section class="armory-layout shop-layout">
    <article class="armory-panel shop-panel shop-panel--cursor">
      <div class="shop-panel__header">
        <div>
          <p class="vote-stage__eyebrow">商店</p>
          <strong>战斗点击图标</strong>
        </div>
        <div class="shop-panel__summary">
          <div class="shop-panel__gold">
            <span>金币</span>
            <strong>{{ gold }}</strong>
          </div>
          <div class="shop-current-cursor">
            <img class="shop-current-cursor__image" :src="currentCursorImage" alt="当前点击图标预览"/>
            <div class="shop-current-cursor__meta">
              <span>{{ equippedBattleClickCursorImagePath ? '当前已装备点击图标' : '当前使用默认点击图标' }}</span>
              <button
                  v-if="equippedBattleClickCursorImagePath"
                  class="nickname-form__submit shop-current-cursor__reset-btn"
                  type="button"
                  :disabled="!isLoggedIn"
                  @click="unequipShopItem"
              >
                恢复默认
              </button>
            </div>
          </div>
        </div>
      </div>

      <p v-if="loadingShopItems" class="feedback-panel">商店加载中...</p>
      <div v-else class="shop-cursor-grid">
        <article v-for="item in shopItems" :key="item.itemId" class="shop-cursor-card">
          <div class="shop-cursor-card__visual">
            <img
                v-if="item.previewImagePath || item.imagePath"
                class="shop-cursor-card__image"
                :src="item.previewImagePath || item.imagePath"
                :alt="item.imageAlt || item.title"
            />
            <span v-else class="shop-cursor-card__fallback">?</span>
          </div>
          <div class="shop-cursor-card__main">
            <strong class="shop-cursor-card__title">{{ item.title }}</strong>
            <span class="shop-cursor-card__price">{{ item.priceGold }} 金币</span>
          </div>
          <p class="shop-cursor-card__desc">{{ item.description || '永久点击图标外观。' }}</p>
          <div class="shop-cursor-card__action">
            <button
                class="nickname-form__submit"
                type="button"
                :disabled="buttonDisabled(item)"
                @click="handleShopAction(item)"
            >
              {{ buttonLabel(item) }}
            </button>
          </div>
        </article>
      </div>
    </article>
  </section>
</template>
