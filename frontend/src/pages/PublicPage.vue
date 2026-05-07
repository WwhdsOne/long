<script setup>
import {computed, ref} from 'vue'
import BattlePage from './BattlePage.vue'
import ArmoryPage from './ArmoryPage.vue'
import MessagesPage from './MessagesPage.vue'
import ShopPage from './ShopPage.vue'
import TaskPage from './TaskPage.vue'
import TalentsPage from './TalentsPage.vue'
import {usePublicPageState} from './publicPageState'
import {formatCompact} from '../utils/formatNumber.js'
import wechatGroupImage from '../assets/community/wechat-group.png'

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

async function handleAuthClick() {
  if (isLoggedIn.value) {
    await resetNickname()
  } else {
    loginModalOpen.value = true
    errorMessage.value = ''
  }
}

async function handleLoginSubmit() {
  await submitNickname()
  if (nickname.value) {
    loginModalOpen.value = false
    errorMessage.value = ''
  }
}
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
      <span v-if="page.id === 'tasks' && hasClaimableTasks" class="public-nav__task-dot" aria-label="有可领取任务"></span>
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
                <img :src="wechatGroupImage" alt="Hai-World 微信群二维码" />
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
              <strong class="hero-info-card__value">github.com/WwhdsOne/long</strong>
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
      <div class="login-modal__backdrop" @click="loginModalOpen = false"></div>
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
          <button class="nickname-form__submit" type="submit">
            登录 / 首次认领
          </button>
        </form>
        <p v-if="errorMessage" class="feedback">{{ errorMessage }}</p>
      </div>
    </section>

    <BattlePage v-if="currentPublicPage === 'battle'" />
    <ShopPage v-else-if="currentPublicPage === 'shop'" />
    <TalentsPage v-else-if="currentPublicPage === 'talents'" />
    <TaskPage v-else-if="currentPublicPage === 'tasks'" />
    <ArmoryPage v-else-if="isArmoryPage(currentPublicPage)" :focus-section="currentPublicPage" />
    <MessagesPage v-else />

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
