<script setup>
import {computed, nextTick, onBeforeUnmount, ref} from 'vue'

import {usePublicPageState} from './publicPageState'

let turnstileScriptPromise = null

const {
  isLoggedIn,
  gold,
  stamina,
  shopItems,
  loadingShopItems,
  equippedBattleClickCursorImagePath,
  purchaseShopItem,
  purchaseStaminaFull,
  upgradeStaminaCap,
  equipShopItem,
  unequipShopItem,
} = usePublicPageState()

const DEFAULT_BATTLE_CLICK_CURSOR_IMAGE = 'https://hai-world2.oss-cn-beijing.aliyuncs.com/effects/click-sword_basic.png'
const currentCursorImage = computed(() => equippedBattleClickCursorImagePath.value || DEFAULT_BATTLE_CLICK_CURSOR_IMAGE)
const isRiskBanned = computed(() => Number(stamina.value?.riskBanUntil || 0) > Date.now() / 1000)
const staminaConfirmOpen = ref(false)
const staminaCapConfirmOpen = ref(false)
const staminaPurchaseSubmitting = ref(false)
const staminaCaptchaRequired = ref(false)
const staminaTurnstileSiteKey = ref('')
const staminaTurnstileToken = ref('')
const staminaTurnstileError = ref('')
const staminaTurnstileWidgetId = ref(null)
const staminaTurnstileContainer = ref(null)
const staminaBanHint = computed(() => (
    isRiskBanned.value ? '账号异常，当前不可手点/挂机/购买体力' : ''
))

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

function ensureTurnstileScript() {
  if (window.turnstile) {
    return Promise.resolve(window.turnstile)
  }
  if (turnstileScriptPromise) {
    return turnstileScriptPromise
  }
  turnstileScriptPromise = new Promise((resolve, reject) => {
    const existing = document.querySelector('script[data-turnstile-script="true"]')
    if (existing) {
      existing.addEventListener('load', () => resolve(window.turnstile), {once: true})
      existing.addEventListener('error', () => reject(new Error('load failed')), {once: true})
      return
    }
    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.dataset.turnstileScript = 'true'
    script.onload = () => resolve(window.turnstile)
    script.onerror = () => reject(new Error('load failed'))
    document.head.appendChild(script)
  })
  return turnstileScriptPromise
}

function clearStaminaTurnstileState() {
  staminaCaptchaRequired.value = false
  staminaTurnstileSiteKey.value = ''
  staminaTurnstileToken.value = ''
  staminaTurnstileError.value = ''
  if (window.turnstile && staminaTurnstileWidgetId.value !== null) {
    window.turnstile.remove?.(staminaTurnstileWidgetId.value)
  }
  staminaTurnstileWidgetId.value = null
  if (staminaTurnstileContainer.value) {
    staminaTurnstileContainer.value.innerHTML = ''
  }
}

function resetStaminaTurnstileWidget(message = '') {
  staminaTurnstileToken.value = ''
  staminaTurnstileError.value = message
  if (window.turnstile && staminaTurnstileWidgetId.value !== null) {
    window.turnstile.reset(staminaTurnstileWidgetId.value)
  }
}

async function renderStaminaTurnstile() {
  if (!staminaCaptchaRequired.value || !staminaTurnstileSiteKey.value) {
    return
  }
  await nextTick()
  if (!staminaTurnstileContainer.value) {
    return
  }
  try {
    const turnstile = await ensureTurnstileScript()
    if (!turnstile || !staminaCaptchaRequired.value) {
      return
    }
    if (staminaTurnstileWidgetId.value !== null) {
      turnstile.reset(staminaTurnstileWidgetId.value)
      return
    }
    staminaTurnstileContainer.value.innerHTML = ''
    staminaTurnstileWidgetId.value = turnstile.render(staminaTurnstileContainer.value, {
      sitekey: staminaTurnstileSiteKey.value,
      callback: handlePurchaseStaminaCaptchaSuccess,
      'expired-callback': handlePurchaseStaminaCaptchaExpired,
      'error-callback': handlePurchaseStaminaCaptchaError,
    })
  } catch {
    staminaTurnstileError.value = '验证服务暂时不可用，请稍后再试'
  }
}

async function submitPurchaseStaminaFull(turnstileToken = '') {
  if (staminaPurchaseSubmitting.value) {
    return
  }
  if (staminaCaptchaRequired.value && !turnstileToken) {
    staminaTurnstileError.value = '本次购买需要完成人机验证'
    return
  }

  staminaPurchaseSubmitting.value = true
  const result = await purchaseStaminaFull(turnstileToken)
  staminaPurchaseSubmitting.value = false

  if (result.ok) {
    clearStaminaTurnstileState()
    staminaConfirmOpen.value = false
    return
  }

  if (result.errorCode === 'CAPTCHA_REQUIRED') {
    if (!result.siteKey) {
      staminaTurnstileError.value = '验证服务暂时不可用，请稍后再试'
      return
    }
    staminaCaptchaRequired.value = true
    staminaTurnstileSiteKey.value = result.siteKey
    staminaTurnstileToken.value = ''
    staminaTurnstileError.value = '本次购买需要完成人机验证'
    await renderStaminaTurnstile()
    return
  }

  if (result.errorCode === 'CAPTCHA_INVALID') {
    resetStaminaTurnstileWidget('验证失败，请重试')
    return
  }

  if (result.errorCode === 'CAPTCHA_VERIFY_UNAVAILABLE') {
    resetStaminaTurnstileWidget('验证服务暂时不可用，请稍后再试')
  }
}

async function handlePurchaseStaminaFull() {
  await submitPurchaseStaminaFull(staminaTurnstileToken.value)
}

async function handlePurchaseStaminaCaptchaSuccess(token) {
  staminaTurnstileToken.value = token
  staminaTurnstileError.value = ''
  await submitPurchaseStaminaFull(token)
}

function handlePurchaseStaminaCaptchaExpired() {
  resetStaminaTurnstileWidget('验证已过期，请重新验证')
}

function handlePurchaseStaminaCaptchaError() {
  resetStaminaTurnstileWidget('验证失败，请重试')
}

async function handleUpgradeStaminaCap() {
  await upgradeStaminaCap()
  staminaCapConfirmOpen.value = false
}

function openStaminaConfirm() {
  if (!isLoggedIn.value || isRiskBanned.value) {
    return
  }
  clearStaminaTurnstileState()
  staminaConfirmOpen.value = true
}

function closeStaminaConfirm() {
  clearStaminaTurnstileState()
  staminaConfirmOpen.value = false
}

function openStaminaCapConfirm() {
  if (!isLoggedIn.value || isRiskBanned.value || !stamina.value?.nextCapUpgradeCost) {
    return
  }
  staminaCapConfirmOpen.value = true
}

function closeStaminaCapConfirm() {
  staminaCapConfirmOpen.value = false
}

onBeforeUnmount(() => {
  clearStaminaTurnstileState()
})
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
          <div class="shop-stamina-card">
            <span>体力</span>
            <strong>{{ stamina.current }} / {{ stamina.max }}</strong>
            <div class="shop-stamina-card__notice">
              <span>1 点体力 = 50 次点击</span>
              <span>体力归零后点击伤害锁定为 1</span>
              <span>挂机时伤害不受体力系统限制</span>
            </div>
            <div class="shop-stamina-card__actions">
              <button
                  class="nickname-form__submit"
                  type="button"
                  :disabled="!isLoggedIn || isRiskBanned"
                  @click="openStaminaConfirm"
              >
                购买体力
              </button>
              <button
                  class="nickname-form__ghost"
                  type="button"
                  :disabled="!isLoggedIn || isRiskBanned || !stamina.nextCapUpgradeCost"
                  @click="openStaminaCapConfirm"
              >
                升级体力上限
              </button>
            </div>
          </div>
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
      <p v-if="staminaBanHint" class="feedback-panel feedback-panel--compact">{{ staminaBanHint }}</p>
      <div class="shop-cursor-grid">
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
    <section v-if="staminaConfirmOpen" class="shop-stamina-modal boss-drop-modal" aria-label="确认购买体力">
      <div class="boss-drop-modal__backdrop" @click="closeStaminaConfirm"></div>
      <article class="boss-drop-modal__card shop-stamina-modal__card">
        <div class="boss-drop-modal__head">
          <strong>确认购买体力</strong>
        </div>
        <section class="boss-drop-modal__section">
          <div class="boss-drop-modal__section-head">
            <span>本次消耗</span>
            <strong>{{ stamina.nextFullBuyPrice }} 金币</strong>
          </div>
          <p class="shop-stamina-modal__desc">购买后将体力直接补满，并按当日次数刷新下一次价格。</p>
          <p class="shop-stamina-modal__desc">当前体力：{{ stamina.current }} / {{ stamina.max }}，体力上限等级：{{ stamina.maxLevel }} / 50</p>
          <div v-if="staminaCaptchaRequired" class="shop-turnstile-panel">
            <p class="shop-stamina-modal__desc shop-stamina-modal__desc--captcha">本次购买需要完成人机验证</p>
            <div ref="staminaTurnstileContainer" class="shop-turnstile-panel__widget"></div>
          </div>
          <p v-if="staminaTurnstileError" class="feedback-panel feedback-panel--compact">{{ staminaTurnstileError }}</p>
        </section>
        <div class="shop-stamina-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="closeStaminaConfirm">
            取消购买
          </button>
          <button
              class="nickname-form__submit"
              type="button"
              :disabled="!isLoggedIn || isRiskBanned || staminaPurchaseSubmitting || (staminaCaptchaRequired && !staminaTurnstileToken)"
              @click="handlePurchaseStaminaFull"
          >
            确认购买体力
          </button>
        </div>
      </article>
    </section>
    <section v-if="staminaCapConfirmOpen" class="shop-stamina-modal boss-drop-modal" aria-label="确认购买体力上限">
      <div class="boss-drop-modal__backdrop" @click="closeStaminaCapConfirm"></div>
      <article class="boss-drop-modal__card shop-stamina-modal__card">
        <div class="boss-drop-modal__head">
          <strong>确认购买体力上限</strong>
        </div>
        <section class="boss-drop-modal__section">
          <div class="boss-drop-modal__section-head">
            <span>本次消耗</span>
            <strong>{{ stamina.nextCapUpgradeCost }} 金币</strong>
          </div>
          <p class="shop-stamina-modal__desc">升级后永久提升 1 点体力上限，并同步提高可承载的手点额度。</p>
          <p class="shop-stamina-modal__desc">当前体力上限：{{ stamina.max }}，上限等级：{{ stamina.maxLevel }} / 50</p>
        </section>
        <div class="shop-stamina-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="closeStaminaCapConfirm">
            取消购买
          </button>
          <button
              class="nickname-form__submit"
              type="button"
              :disabled="!isLoggedIn || isRiskBanned || !stamina.nextCapUpgradeCost"
              @click="handleUpgradeStaminaCap"
          >
            确认购买体力上限
          </button>
        </div>
      </article>
    </section>
  </section>
</template>
