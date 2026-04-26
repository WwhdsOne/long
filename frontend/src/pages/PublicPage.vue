<script setup>
import {ref} from 'vue'
import BattlePage from './BattlePage.vue'
import ArmoryPage from './ArmoryPage.vue'
import MessagesPage from './MessagesPage.vue'
import TalentsPage from './TalentsPage.vue'
import {usePublicPageState} from './publicPageState'

const {
  publicPages,
  currentPublicPage,
  navigatePublicPage,
  syncing,
  syncLabel,
  lastUpdatedAt,
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
  <main class="page-shell">
    <div class="page-shell__glow page-shell__glow--pink"></div>
    <div class="page-shell__glow page-shell__glow--blue"></div>
    <div class="page-shell__glow page-shell__glow--yellow"></div>

    <nav class="public-nav" aria-label="前台导航">
      <button
          v-for="page in publicPages"
          :key="page.id"
          class="public-nav__item"
          :class="{ 'public-nav__item--active': currentPublicPage === page.id }"
          type="button"
          @click="navigatePublicPage(page.id)"
      >
        {{ page.label }}
      </button>
      <button
          class="public-nav__item public-nav__auth"
          type="button"
          @click="handleAuthClick"
      >
        {{ isLoggedIn ? '退出登录' : '登录/注册' }}
      </button>
    </nav>

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
        <span class="hero__time">最近刷新 {{ lastUpdatedAt || '--:--:--' }}</span>
        <a class="hero__admin-link" href="/admin">管理后台</a>
      </div>
    </section>


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
    <TalentsPage v-else-if="currentPublicPage === 'talents'" />
    <ArmoryPage v-else-if="isArmoryPage(currentPublicPage)" :focus-section="currentPublicPage" />
    <MessagesPage v-else />
  </main>
</template>
