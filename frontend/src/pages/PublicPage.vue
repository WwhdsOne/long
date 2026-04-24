<script setup>
import BattlePage from './BattlePage.vue'
import MessagesPage from './MessagesPage.vue'
import ProfilePage from './ProfilePage.vue'
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
} = usePublicPageState()

registerPublicPageLifecycle()
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
    </nav>

    <section class="hero">
      <div class="hero__copy">
        <p class="hero__eyebrow">Long Vote Wall</p>
        <h1>登录账号，再狠狠干一票。</h1>
        <p class="hero__lede">
          平时点按钮照样冲榜；有活动 Boss 时，同一点击会把装备增量一起结算成伤害。
        </p>
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


    <BattlePage v-if="currentPublicPage === 'battle'" />
    <TalentsPage v-else-if="currentPublicPage === 'talents'" />
    <ProfilePage v-else-if="currentPublicPage === 'profile'" />
    <MessagesPage v-else />
  </main>
</template>
