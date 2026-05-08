<script setup>
import {computed, nextTick, onBeforeUnmount, ref} from 'vue'
import BattlePage from './BattlePage.vue'
import ArmoryPage from './ArmoryPage.vue'
import MessagesPage from './MessagesPage.vue'
import ShopPage from './ShopPage.vue'
import TaskPage from './TaskPage.vue'
import TalentsPage from './TalentsPage.vue'
import {usePublicPageState} from './publicPageState'
import {formatCompact} from '../utils/formatNumber.js'
import wechatGroupImage from '../assets/community/wechat-group.png'

let loginTurnstileScriptPromise = null

const {
  publicPages,
  currentPublicPage,
  navigatePublicPage,
  hasClaimableTasks,
  syncing,
  syncLabel,
  combatStats,
  announcementModalOpen,
  latestAnnouncement,
  closeAnnouncementModal,
  formatTime,
  registerPublicPageLifecycle,
  isLoggedIn,
  nickname,
  nicknameDraft,
  passwordDraft,
  errorMessage,
  submitNickname,
  resetNickname,
} = usePublicPageState()

registerPublicPageLifecycle()

const loginModalOpen = ref(false)
const loginSubmitting = ref(false)
const loginCaptchaRequired = ref(false)
const loginTurnstileSiteKey = ref('')
const loginTurnstileToken = ref('')
const loginTurnstileError = ref('')
const loginTurnstileWidgetId = ref(null)
const loginTurnstileContainer = ref(null)
const armoryPageIDs = new Set(['resources', 'inventory', 'stats', 'loadout'])
const isBattlePage = computed(() => currentPublicPage.value === 'battle')
const myBattlePower = computed(() => (
    Math.max(0, Number(combatStats.value?.attackPower || 0)) +
    Math.max(0, Number(combatStats.value?.normalDamage || 0)) +
    Math.max(0, Number(combatStats.value?.criticalDamage || 0))
))
const heroBattlePowerLabel = computed(() => (
    isLoggedIn.value ? formatCompact(myBattlePower.value) : '登录后激活'
))

function isArmoryPage(pageID) {
  return armoryPageIDs.has(pageID)
}

function ensureLoginTurnstileScript() {
  if (window.turnstile) {
    return Promise.resolve(window.turnstile)
  }
  if (loginTurnstileScriptPromise) {
    return loginTurnstileScriptPromise
  }
  loginTurnstileScriptPromise = new Promise((resolve, reject) => {
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
  return loginTurnstileScriptPromise
}

function clearLoginTurnstileState() {
  loginCaptchaRequired.value = false
  loginTurnstileSiteKey.value = ''
  loginTurnstileToken.value = ''
  loginTurnstileError.value = ''
  if (window.turnstile && loginTurnstileWidgetId.value !== null) {
    window.turnstile.remove?.(loginTurnstileWidgetId.value)
  }
  loginTurnstileWidgetId.value = null
  if (loginTurnstileContainer.value) {
    loginTurnstileContainer.value.innerHTML = ''
  }
}

function resetLoginTurnstileWidget(message = '') {
  loginTurnstileToken.value = ''
  loginTurnstileError.value = message
  if (window.turnstile && loginTurnstileWidgetId.value !== null) {
    window.turnstile.reset(loginTurnstileWidgetId.value)
  }
}

async function renderLoginTurnstile() {
  if (!loginCaptchaRequired.value || !loginTurnstileSiteKey.value) {
    return
  }
  await nextTick()
  if (!loginTurnstileContainer.value) {
    return
  }
  try {
    const turnstile = await ensureLoginTurnstileScript()
    if (!turnstile || !loginCaptchaRequired.value) {
      return
    }
    if (loginTurnstileWidgetId.value !== null) {
      turnstile.reset(loginTurnstileWidgetId.value)
      return
    }
    loginTurnstileContainer.value.innerHTML = ''
    loginTurnstileWidgetId.value = turnstile.render(loginTurnstileContainer.value, {
      sitekey: loginTurnstileSiteKey.value,
      callback: handleLoginCaptchaSuccess,
      'expired-callback': handleLoginCaptchaExpired,
      'error-callback': handleLoginCaptchaError,
    })
  } catch {
    loginTurnstileError.value = '验证服务暂时不可用，请稍后再试'
  }
}

async function submitNicknameWithCaptcha(turnstileToken = '') {
  if (loginSubmitting.value) {
    return
  }
  if (loginCaptchaRequired.value && !turnstileToken) {
    loginTurnstileError.value = '登录前需要先完成人机验证'
    return
  }

  loginSubmitting.value = true
  const result = await submitNickname(turnstileToken)
  loginSubmitting.value = false

  if (result?.ok) {
    clearLoginTurnstileState()
    loginModalOpen.value = false
    errorMessage.value = ''
    return
  }

  if (result?.errorCode === 'CAPTCHA_REQUIRED') {
    if (!result.siteKey) {
      loginTurnstileError.value = '验证服务暂时不可用，请稍后再试'
      return
    }
    loginCaptchaRequired.value = true
    loginTurnstileSiteKey.value = result.siteKey
    loginTurnstileToken.value = ''
    loginTurnstileError.value = '登录前需要先完成人机验证'
    await renderLoginTurnstile()
    return
  }

  if (result?.errorCode === 'CAPTCHA_INVALID') {
    resetLoginTurnstileWidget('验证失败，请重试')
    return
  }

  if (result?.errorCode === 'CAPTCHA_VERIFY_UNAVAILABLE') {
    resetLoginTurnstileWidget('验证服务暂时不可用，请稍后再试')
    return
  }

  if (loginCaptchaRequired.value) {
    resetLoginTurnstileWidget('')
  }
}

async function handleAuthClick() {
  if (isLoggedIn.value) {
    await resetNickname()
  } else {
    loginModalOpen.value = true
    errorMessage.value = ''
    clearLoginTurnstileState()
  }
}

async function handleLoginSubmit() {
  await submitNicknameWithCaptcha(loginTurnstileToken.value)
}

function closeLoginModal() {
  clearLoginTurnstileState()
  errorMessage.value = ''
  loginModalOpen.value = false
}

async function handleLoginCaptchaSuccess(token) {
  loginTurnstileToken.value = token
  loginTurnstileError.value = ''
  await submitNicknameWithCaptcha(token)
}

function handleLoginCaptchaExpired() {
  resetLoginTurnstileWidget('验证已过期，请重新验证')
}

function handleLoginCaptchaError() {
  resetLoginTurnstileWidget('验证失败，请重试')
}

onBeforeUnmount(() => {
  clearLoginTurnstileState()
})
</script>

<template>
  <nav class="public-nav" :class="{ 'public-nav--battle': isBattlePage }" aria-label="前台导航">
    <button
        v-for="page in publicPages"
        :key="page.id"
        class="public-nav__item"
        :class="{ 'public-nav__item--active': currentPublicPage === page.id }"
        type="button"
        @click="navigatePublicPage(page.id)"
    >
      <span>{{ page.label }}</span>
      <span v-if="page.id === 'tasks' && hasClaimableTasks" class="public-nav__task-dot"
            aria-label="有可领取任务"></span>
    </button>
    <button
        class="public-nav__item public-nav__auth"
        type="button"
        @click="handleAuthClick"
    >
      {{ isLoggedIn ? '退出登录' : '登录/注册' }}
    </button>
  </nav>
  <main class="page-shell">

    <template v-if="!isBattlePage">
      <div class="page-shell__glow page-shell__glow--pink"></div>
      <div class="page-shell__glow page-shell__glow--blue"></div>
      <div class="page-shell__glow page-shell__glow--yellow"></div>
      <section class="hero">
        <div class="hero__copy">
          <p class="hero__eyebrow">Hai-World</p>
          <h1>狠狠干一票。</h1>
        </div>

        <div class="hero__status">
          <span class="live-pill" :class="{ 'live-pill--syncing': syncing }">
            <span class="live-pill__dot"></span>
            {{ syncLabel }}
          </span>
          <article class="power-glory-card" aria-label="当前用户战斗力">
            <span class="power-glory-card__eyebrow">当前用户战斗力</span>
            <strong class="power-glory-card__value">{{ heroBattlePowerLabel }}</strong>
          </article>
          <div class="hero__link-grid" aria-label="项目相关入口">
            <article class="hero-info-card hero-info-card--community" aria-label="游戏交流群">
              <span class="hero-info-card__eyebrow">游戏交流群</span>
              <strong class="hero-info-card__value">鼠标悬停查看微信群</strong>
              <div class="hero-info-card__preview" role="img" aria-label="微信群二维码预览">
                <img :src="wechatGroupImage" alt="Hai-World 微信群二维码"/>
              </div>
            </article>
            <a
                class="hero-info-card hero-info-card--github"
                href="https://github.com/WwhdsOne/long"
                target="_blank"
                rel="noreferrer"
                aria-label="项目地址 GitHub 仓库"
            >
              <span class="hero-info-card__eyebrow">项目地址</span>
              <strong class="hero-info-card__value">点击进入项目仓库</strong>
            </a>
          </div>
        </div>
      </section>
    </template>


    <section v-if="announcementModalOpen && latestAnnouncement" class="announcement-modal" aria-label="更新公告">
      <div class="announcement-modal__backdrop" @click="closeAnnouncementModal"></div>
      <article class="announcement-modal__card">
        <p class="vote-stage__eyebrow">更新内容公告</p>
        <strong>{{ latestAnnouncement.title }}</strong>
        <p class="announcement-modal__time">{{ formatTime(latestAnnouncement.publishedAt) }}</p>
        <p class="social-card__copy social-card__copy--multiline">{{ latestAnnouncement.content }}</p>
        <div class="announcement-modal__actions">
          <button class="nickname-form__submit" type="button" @click="closeAnnouncementModal">我知道了</button>
        </div>
      </article>
    </section>


    <section v-if="loginModalOpen" class="login-modal" aria-label="登录">
      <div class="login-modal__backdrop" @click="closeLoginModal"></div>
      <div class="login-modal__card">
        <div class="login-modal__header">
          <p class="vote-stage__eyebrow">账号</p>
          <strong>登录 / 注册</strong>
        </div>
        <p class="login-modal__note">输入昵称和密码，首次输入自动注册</p>
        <p class="login-modal__hint">昵称规则：最多 20 字，不得包含敏感词（不区分大小写）。</p>
        <form class="nickname-form login-modal__form" @submit.prevent="handleLoginSubmit">
          <input
              v-model="nicknameDraft"
              class="nickname-form__input"
              type="text"
              maxlength="20"
              placeholder="比如：阿明"
          />
          <input
              v-model="passwordDraft"
              class="nickname-form__input"
              type="password"
              placeholder="输入密码"
          />
          <div v-if="loginCaptchaRequired" class="login-turnstile-panel">
            <p class="login-modal__hint login-modal__hint--captcha">登录前需要先完成人机验证</p>
            <div ref="loginTurnstileContainer" class="login-turnstile-panel__widget"></div>
          </div>
          <button class="nickname-form__submit" type="submit" :disabled="loginSubmitting || (loginCaptchaRequired && !loginTurnstileToken)">
            登录 / 首次认领
          </button>
        </form>
        <p v-if="loginTurnstileError" class="feedback">{{ loginTurnstileError }}</p>
        <p v-if="errorMessage" class="feedback">{{ errorMessage }}</p>
      </div>
    </section>

    <BattlePage v-if="currentPublicPage === 'battle'"/>
    <ShopPage v-else-if="currentPublicPage === 'shop'"/>
    <TalentsPage v-else-if="currentPublicPage === 'talents'"/>
    <TaskPage v-else-if="currentPublicPage === 'tasks'"/>
    <ArmoryPage v-else-if="isArmoryPage(currentPublicPage)" :focus-section="currentPublicPage"/>
    <MessagesPage v-else/>

    <footer class="site-footer" aria-label="网站备案信息">
      <a
          class="site-footer__link"
          href="https://beian.miit.gov.cn/"
          target="_blank"
          rel="noreferrer"
      >
        京ICP备2025120689号-2
      </a>
    </footer>
  </main>
</template>
