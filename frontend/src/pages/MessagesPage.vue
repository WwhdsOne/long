<script setup>
import {usePublicPageState} from './publicPageState'

const {
  ANNOUNCEMENT_READ_KEY,
  ANNOUNCEMENT_CACHE_KEY,
  AUTO_CLICK_RATE_LABEL,
  EQUIPMENT_ENHANCE_COST,
  HERO_AWAKEN_COST,
  GROWTH_FORMULA_TEXT,
  HERO_GROWTH_FORMULA_TEXT,
  publicPages,
  buttons,
  firstPageButtons,
  buttonPage,
  buttonPageSize,
  buttonTotalPages,
  buttonTotalCount,
  buttonTotalVotes,
  leaderboard,
  boss,
  bossLeaderboard,
  bossLoot,
  bossHeroLoot,
  starlight,
  announcementVersion,
  latestAnnouncement,
  announcements,
  myBossStats,
  inventory,
  heroes,
  activeHero,
  loadout,
  combatStats,
  recentRewards,
  lastReward,
  userStats,
  nickname,
  nicknameDraft,
  passwordDraft,
  loading,
  syncing,
  errorMessage,
  pendingKeys,
  actioningItemId,
  activeHudTab,
  lastUpdatedAt,
  liveConnected,
  criticalBursts,
  bossHistory,
  bossHistoryQuery,
  loadingBossHistory,
  bossHistoryLoaded,
  bossHistoryError,
  selectedButtonTag,
  buttonSearch,
  loadingAnnouncements,
  announcementsLoaded,
  announcementError,
  loadingBossResources,
  latestAnnouncementLoaded,
  announcementModalOpen,
  messages,
  messageNextCursor,
  loadingMessages,
  postingMessage,
  messageDraft,
  messageError,
  autoClickEnabled,
  autoClickTargetKey,
  gems,
  ownedCosmetics,
  equippedCosmetics,
  cosmeticDraft,
  shopCatalog,
  lastForgeResult,
  cosmeticBursts,
  fingerprintHash,
  currentPublicPage,
  profileLoading,
  profileLoaded,
  profileNotice,
  starlightTimer,
  lastExpiredStarlightEndsAt,
  lastBossResourceVersion,
  burstTimers,
  cosmeticTimers,
  pendingClickSources,
  clickBehaviorTracker,
  buttonCount,
  totalVotes,
  buttonTags,
  activeStarlightKeys,
  displayedButtons,
  syncLabel,
  isLoggedIn,
  myClicks,
  myRank,
  myBossDamage,
  effectiveIncrement,
  normalDamage,
  criticalDamage,
  autoClickTargetButton,
  autoClickTargetLabel,
  canStartAutoClick,
  autoClickStatus,
  bossStatusLabel,
  bossProgress,
  equippedItems,
  heroCount,
  cosmeticCollections,
  selectedCosmeticLoadout,
  selectedCosmeticSummary,
  equippedCosmeticSummary,
  canApplyCosmeticSelection,
  previewEffectConfig,
  previewDots,
  displayedRecentRewards,
  recentRewardTitle,
  recentRewardNote,
  filteredBossHistory,
  emptyLoadout,
  defaultCombatStats,
  formatItemStats,
  formatItemStatLines,
  equipmentNameParts,
  equipmentNameClass,
  formatEnhanceCap,
  formatAwakenCap,
  formatHeroTrait,
  heroImageAlt,
  normalizeNickname,
  resolvePublicPage,
  navigatePublicPage,
  activatePublicPage,
  handlePublicRouteChange,
  isStarlightButton,
  clearStarlightTimer,
  scheduleStarlightRefresh,
  formatBossTime,
  topBossDamage,
  formatTime,
  formatNumber,
  formatStatWithDelta,
  formatPercentWithDelta,
  formatHeroEffect,
  salvageableEquipmentCount,
  salvageableHeroCount,
  equipmentEnhanceHint,
  heroAwakenHint,
  dotIndexes,
  cosmeticModeClasses,
  syncCosmeticDraft,
  readErrorMessage,
  normalizePageNumber,
  updateCurrentPageButtons,
  applyButtonPagePayload,
  bossResourceVersion,
  readCachedLatestAnnouncement,
  writeCachedLatestAnnouncement,
  restoreCachedLatestAnnouncement,
  maybePromptAnnouncement,
  closeAnnouncementModal,
  loadBossResources,
  loadLatestAnnouncement,
  loadAnnouncements,
  loadMessages,
  submitMessage,
  validateNicknameWithServer,
  loadBossHistory,
  markUpdated,
  selectHudTab,
  applyState,
  applyPublicState,
  applyUserState,
  applyBattleUserState,
  applyPlayerProfileState,
  applyClickResult,
  clearUserRealtimeState,
  clearPendingClicks,
  applyRealtimeSnapshot,
  ensureRealtimeTransport,
  connectRealtime,
  clearCriticalBurst,
  triggerCriticalBurst,
  clearCosmeticBurst,
  handlePressStart,
  handlePressEnd,
  handlePressCancel,
  ensureFingerprintHash,
  consumeClickBehavior,
  triggerCosmeticBurst,
  currentNicknameQuery,
  loadButtonPage,
  syncAutoClickTarget,
  applyAutoClickStatus,
  clearAutoClickLocalState,
  loadAutoClickStatus,
  syncAutoClickTargetOnServer,
  startAutoClick,
  stopAutoClick,
  toggleAutoClick,
  requestClickTicket,
  loadState,
  loadPlayerProfile,
  refreshProfileAfterMutation,
  clickButton,
  postEquipmentAction,
  postHeroAction,
  salvageEquipment,
  enhanceEquipment,
  salvageHero,
  awakenHero,
  purchaseCosmetic,
  selectCosmeticItem,
  equipSelectedCosmetics,
  submitNickname,
  resetNickname,
  clearPlayerSessionState,
  loadPlayerSession,
  registerPublicPageLifecycle,
} = usePublicPageState()
</script>

<template>
<section class="stage-layout stage-layout--messages stage-layout--single">
      <aside class="player-hud player-hud--page">
        <section class="player-hud__shell">
          <div class="player-hud__head">
            <div>
              <p class="vote-stage__eyebrow">公共消息</p>
              <strong>{{ isLoggedIn ? nickname : '未登录角色' }}</strong>
            </div>
            <span class="player-hud__pill">{{ isLoggedIn ? '已上墙' : '访客' }}</span>
          </div>

          <p class="player-hud__copy">消息页保留留言、公告和规则信息；战斗实时链路继续在后台保持连接。</p>

          <form class="nickname-form player-hud__form" @submit.prevent="submitNickname">
            <input v-model="nicknameDraft" class="nickname-form__input" type="text" maxlength="20" placeholder="比如：阿明" />
            <input v-model="passwordDraft" class="nickname-form__input" type="password" placeholder="输入密码" />
            <button class="nickname-form__submit" type="submit">登录 / 创建</button>
          </form>

          <button v-if="isLoggedIn" class="nickname-form__ghost player-hud__reset" type="button" @click="resetNickname">退出当前账号</button>

          <div class="player-hud__content messages-page__grid">
            <section class="player-hud__panel messages-page__feed">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">公共留言墙</p>
                <strong>{{ messages.length }} 条</strong>
              </div>

              <form class="admin-form player-hud__message-form" @submit.prevent="submitMessage">
                <textarea v-model="messageDraft" class="nickname-form__input admin-textarea" rows="4" maxlength="200" placeholder="说点什么，所有人都能看到。"></textarea>
                <button class="nickname-form__submit" type="submit" :disabled="postingMessage || !isLoggedIn">{{ postingMessage ? '发送中...' : '发送留言' }}</button>
              </form>

              <p v-if="messageError" class="feedback feedback--error">{{ messageError }}</p>
              <div v-if="loadingMessages" class="leaderboard-list leaderboard-list--empty"><p>留言加载中...</p></div>
              <div v-else-if="messages.length === 0" class="leaderboard-list leaderboard-list--empty"><p>还没有留言，先写第一条。</p></div>
              <ul v-else class="history-list">
                <li v-for="item in messages" :key="item.id" class="history-item">
                  <div class="history-item__head"><strong>{{ item.nickname }}</strong><span>{{ formatTime(item.createdAt) }}</span></div>
                  <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
                </li>
              </ul>
              <button v-if="messageNextCursor" class="nickname-form__ghost player-hud__retry" type="button" :disabled="loadingMessages" @click="loadMessages(messageNextCursor, true)">加载更多</button>
            </section>

            <aside class="messages-page__side">
              <section class="player-hud__info-block">
                <div class="player-hud__mini-head"><span>最新公告</span><strong>{{ latestAnnouncement?.title || '暂无' }}</strong></div>
                <p class="player-hud__note player-hud__note--multiline">{{ latestAnnouncement?.content || '当前还没有新的站内公告。' }}</p>
              </section>
              <section class="player-hud__info-block">
                <div class="player-hud__mini-head"><span>规则</span><strong>公开留言</strong></div>
                <p class="player-hud__note player-hud__note--multiline">留言按时间倒序展示；发送前需要登录当前账号。</p>
              </section>
              <section class="player-hud__info-block">
                <div class="player-hud__mini-head"><span>在线玩家</span><strong>{{ leaderboard.length }} 人</strong></div>
                <ol v-if="leaderboard.length > 0" class="leaderboard-list">
                  <li v-for="entry in leaderboard" :key="entry.nickname" class="leaderboard-list__item" :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }">
                    <span class="leaderboard-list__rank">#{{ entry.rank }}</span><span class="leaderboard-list__name">{{ entry.nickname }}</span><strong class="leaderboard-list__count">{{ entry.clickCount }}</strong>
                  </li>
                </ol>
                <div v-else class="leaderboard-list leaderboard-list--empty"><p>暂无在线玩家数据。</p></div>
              </section>
            </aside>
          </div>
        </section>
      </aside>
    </section>
</template>
