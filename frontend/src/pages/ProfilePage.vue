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
<section class="stage-layout stage-layout--single">
      <aside class="player-hud player-hud--page">
        <section class="player-hud__shell">
          <div class="player-hud__head">
            <div>
              <p class="vote-stage__eyebrow">角色资料</p>
              <strong>{{ isLoggedIn ? nickname : '未登录角色' }}</strong>
            </div>
            <span class="player-hud__pill">{{ isLoggedIn ? '已上墙' : '访客' }}</span>
          </div>

          <p class="player-hud__copy">{{ profileNotice || (isLoggedIn ? `你现在登录的是 ${nickname}。进入本页会刷新背包、属性和装备。` : '先输入昵称和密码登录；第一次使用该昵称时会直接为它设置密码。') }}
          </p>

          <form class="nickname-form player-hud__form" @submit.prevent="submitNickname">
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
              {{ isLoggedIn ? '切换账号' : '登录 / 首次认领' }}
            </button>
          </form>

          <button
              v-if="isLoggedIn"
              class="nickname-form__ghost player-hud__reset"
              type="button"
              @click="resetNickname"
          >
            退出登录
          </button>

          <div class="player-hud__tabs">
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'inventory' }"
                type="button"
                @click="selectHudTab('inventory')"
            >
              背包
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'stats' }"
                type="button"
                @click="selectHudTab('stats')"
            >
              属性
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'loadout' }"
                type="button"
                @click="selectHudTab('loadout')"
            >
              装备栏
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'heroes' }"
                type="button"
                @click="selectHudTab('heroes')"
            >
              小小英雄
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'forge' }"
                type="button"
                @click="selectHudTab('forge')"
            >
              强化
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'shop' }"
                type="button"
                @click="selectHudTab('shop')"
            >
              商店
            </button>
          </div>

          <p v-if="profileLoading" class="feedback">资料刷新中...</p>
          <div class="player-hud__content">
            <section v-if="activeHudTab === 'inventory'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">背包</p>
                <strong>{{ inventory.length }} 件</strong>
              </div>

              <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
                <p>先去打 Boss 或等后台发装备，背包就会慢慢满起来。</p>
              </div>

              <ul v-else class="inventory-list">
                <li v-for="item in inventory" :key="item.itemId" class="inventory-item inventory-item--panel">
                  <div class="inventory-item__top">
                    <div class="inventory-item__main">
                      <strong>
                        <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                        <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                      </strong>
                      <div class="inventory-item__meta">
                        <span class="inventory-item__chip">{{ formatRarityLabel(item.rarity) }}</span>
                        <span class="inventory-item__chip">类型:{{ item.slot || '未分类' }}</span>
                        <span class="inventory-item__chip">库存:{{ item.quantity }}</span>
                        <span class="inventory-item__chip">强化:{{
                            item.enhanceLevel ? `+${item.enhanceLevel}` : '未强化'
                          }}</span>
                        <span class="inventory-item__chip">可分解:{{ salvageableEquipmentCount(item) }}</span>
                      </div>
                    </div>
                  </div>

                  <ul class="inventory-item__stats inventory-item__stats--stacked">
                    <li v-for="line in formatItemStatLines(item)" :key="line">
                      {{ line }}
                    </li>
                  </ul>

                  <div class="inventory-item__footer">
                    <span
                        class="inventory-item__state"
                        :class="{ 'inventory-item__state--active': item.equipped }"
                    >
                      {{ item.equipped ? '已穿戴' : '待命中' }}
                    </span>

                    <div class="inventory-item__actions">
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || actioningItemId === item.itemId"
                          @click="item.equipped ? postEquipmentAction(item.itemId, 'unequip') : postEquipmentAction(item.itemId, 'equip')"
                      >
                        {{ item.equipped ? '卸下' : '穿戴' }}
                      </button>
                    </div>
                  </div>
                </li>
              </ul>
            </section>

            <section v-else-if="activeHudTab === 'stats'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">战斗属性</p>
                <strong>{{ isLoggedIn ? nickname : '未登录' }}</strong>
              </div>

              <div class="me-card__stats">
                <article>
                  <span>普通伤害</span>
                  <strong>{{ normalDamage }}</strong>
                </article>
                <article>
                  <span>暴击伤害</span>
                  <strong>{{ criticalDamage }}</strong>
                </article>
                <article>
                  <span>暴击率</span>
                  <strong>{{ formatNumber(combatStats.criticalChancePercent, 2) }}%</strong>
                </article>
                <article>
                  <span>我的 Boss 伤害</span>
                  <strong>{{ myBossDamage }}</strong>
                </article>
                <article>
                  <span>我的点击</span>
                  <strong>{{ isLoggedIn ? myClicks : '--' }}</strong>
                </article>
                <article>
                  <span>我的排名</span>
                  <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
                </article>
              </div>
            </section>

            <section v-else-if="activeHudTab === 'loadout'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">装备栏</p>
                <strong>{{ equippedItems.length }} / 3</strong>
              </div>

              <div class="loadout-grid">
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>武器</span>
                    <strong v-if="loadout.weapon">
                      <span v-if="equipmentNameParts(loadout.weapon).prefix">{{ equipmentNameParts(loadout.weapon).prefix }}</span>
                      <span :class="equipmentNameClass(loadout.weapon)">{{ equipmentNameParts(loadout.weapon).text }}</span>
                    </strong>
                    <strong v-else>未穿戴</strong>
                  </div>
                  <ul v-if="loadout.weapon" class="loadout-slot__attrs">
                    <li>{{ formatRarityLabel(loadout.weapon.rarity) }}</li>
                    <li v-for="line in formatItemStatLines(loadout.weapon)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>护甲</span>
                    <strong v-if="loadout.armor">
                      <span v-if="equipmentNameParts(loadout.armor).prefix">{{ equipmentNameParts(loadout.armor).prefix }}</span>
                      <span :class="equipmentNameClass(loadout.armor)">{{ equipmentNameParts(loadout.armor).text }}</span>
                    </strong>
                    <strong v-else>未穿戴</strong>
                  </div>
                  <ul v-if="loadout.armor" class="loadout-slot__attrs">
                    <li>{{ formatRarityLabel(loadout.armor.rarity) }}</li>
                    <li v-for="line in formatItemStatLines(loadout.armor)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
                <article class="loadout-slot">
                  <div class="loadout-slot__main">
                    <span>饰品</span>
                    <strong v-if="loadout.accessory">
                      <span v-if="equipmentNameParts(loadout.accessory).prefix">{{ equipmentNameParts(loadout.accessory).prefix }}</span>
                      <span :class="equipmentNameClass(loadout.accessory)">{{ equipmentNameParts(loadout.accessory).text }}</span>
                    </strong>
                    <strong v-else>未穿戴</strong>
                  </div>
                  <ul v-if="loadout.accessory" class="loadout-slot__attrs">
                    <li>{{ formatRarityLabel(loadout.accessory.rarity) }}</li>
                    <li v-for="line in formatItemStatLines(loadout.accessory)" :key="line">
                      {{ line }}
                    </li>
                  </ul>
                  <p v-else class="loadout-slot__empty">暂无属性</p>
                </article>
              </div>
            </section>

            <section v-else-if="activeHudTab === 'heroes'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">小小英雄</p>
                <strong>{{ heroCount }} 位</strong>
              </div>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>当前出战</span>
                  <strong>{{ activeHero?.name || '未派出' }}</strong>
                </div>
                <div v-if="activeHero?.imagePath" class="player-hud__hero-active">
                  <img
                      class="player-hud__hero-portrait"
                      :src="activeHero.imagePath"
                      :alt="heroImageAlt(activeHero)"
                  />
                </div>
                <p class="player-hud__note">
                  {{
                    activeHero
                        ? `${formatItemStats(activeHero)}，${formatHeroTrait(activeHero)}`
                        : '先去打 Boss 拿到一位英雄，再派出去陪你冲榜。'
                  }}
                </p>
              </section>

              <div v-if="heroes.length === 0" class="leaderboard-list leaderboard-list--empty">
                <p>你还没有招募到任何小小英雄。</p>
              </div>

              <ul v-else class="inventory-list">
                <li v-for="hero in heroes" :key="hero.heroId" class="inventory-item inventory-item--panel">
                  <div class="inventory-item__top">
                    <img
                        v-if="hero.imagePath"
                        class="inventory-item__avatar inventory-item__avatar--hero"
                        :src="hero.imagePath"
                        :alt="heroImageAlt(hero)"
                    />
                    <div class="inventory-item__main">
                      <strong>{{ hero.name }}</strong>
                      <div class="inventory-item__meta">
                        <span class="inventory-item__chip">库存:{{ hero.quantity }}</span>
                        <span class="inventory-item__chip">{{ hero.active ? '出战中' : '待命中' }}</span>
                        <span class="inventory-item__chip">觉醒:{{ hero.awakenLevel || 0 }} / {{ hero.awakenCap || '∞' }}</span>
                        <span class="inventory-item__chip">可分解:{{ salvageableHeroCount(hero) }}</span>
                      </div>
                    </div>
                  </div>

                  <ul class="inventory-item__stats inventory-item__stats--stacked">
                    <li>{{ formatItemStats(hero) }}</li>
                    <li>{{ formatHeroTrait(hero) }}</li>
                  </ul>

                  <div class="inventory-item__footer">
                    <span
                        class="inventory-item__state"
                        :class="{ 'inventory-item__state--active': hero.active }"
                    >
                      {{ hero.active ? '已出战' : '未出战' }}
                    </span>

                    <div class="inventory-item__actions">
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || actioningItemId === hero.heroId"
                          @click="hero.active ? postHeroAction(hero.heroId, 'unequip') : postHeroAction(hero.heroId, 'equip')"
                      >
                        {{ hero.active ? '收回' : '出战' }}
                      </button>
                    </div>
                  </div>
                </li>
              </ul>
            </section>

            <section v-else-if="activeHudTab === 'forge'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">原石强化</p>
                <strong>{{ gems }} 原石</strong>
              </div>

              <div class="forge-grid">
                <article class="forge-summary">
                  <span>当前余额</span>
                  <strong>{{ gems }} 原石</strong>
                  <p>重复装备和重复英雄都可以分解成原石，再投入强化和觉醒。</p>
                  <p>每件装备或者小小英雄可以分解为1原石</p>
                </article>
                <article class="forge-summary">
                  <span>本期价格</span>
                  <strong>强化 {{ EQUIPMENT_ENHANCE_COST }} · 觉醒 {{ HERO_AWAKEN_COST }}</strong>
                  <p>装备与英雄都按模板上限成长，每次只提升一个基础属性，并直接返回本次成长结果。</p>
                </article>
                <article class="forge-summary">
                  <span>强化规则</span>
                  <strong>三项基础属性等概率命中</strong>
                  <p class="forge-summary__tips">
                    <span>仅提升点击 / 暴击 / 暴击率中的一项。</span>
                    <span>{{ GROWTH_FORMULA_TEXT }}。</span>
                    <span>暴击率每次固定 +0.20%。</span>
                  </p>
                </article>
              </div>

              <article
                  v-if="lastForgeResult"
                  class="forge-result"
                  :class="{ 'forge-result--jackpot': lastForgeResult.jackpot }"
              >
                <span>{{ lastForgeResult.kind }}</span>
                <strong>{{ lastForgeResult.targetName || lastForgeResult.targetId }}</strong>
                <p class="player-hud__note">
                  {{ lastForgeResult.rewardSummary }} · 原石 {{ lastForgeResult.gemsDelta > 0 ? '+' : '' }}{{ lastForgeResult.gemsDelta }} · 余额 {{ lastForgeResult.remainingGems }}
                </p>
              </article>

              <section class="player-hud__info-block">
                <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>背包里还没有装备，先去打 Boss 再回来强化。</p>
                </div>
                <ul v-else class="forge-action-list">
                  <li v-for="item in inventory" :key="`forge-${item.itemId}`">
                    <div class="forge-action-list__copy">
                      <strong>
                        <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                        <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                      </strong>
                      <div class="forge-action-list__meta">
                        <span>{{ formatRarityLabel(item.rarity) }}</span>
                        <span>可分解 {{ salvageableEquipmentCount(item) }} 件</span>
                        <span>强化 {{ item.enhanceLevel || 0 }} / {{ item.enhanceCap || '∞' }}</span>
                        <span>每次 {{ EQUIPMENT_ENHANCE_COST }} 原石</span>
                      </div>
                      <p>{{ equipmentEnhanceHint(item) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn || !salvageableEquipmentCount(item) || actioningItemId === item.itemId"
                          @click="salvageEquipment(item)"
                      >
                        分解 x{{ salvageableEquipmentCount(item) }}
                      </button>
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < EQUIPMENT_ENHANCE_COST || actioningItemId === item.itemId || (item.enhanceCap > 0 && item.enhanceLevel >= item.enhanceCap)"
                          @click="enhanceEquipment(item)"
                      >
                        强化
                      </button>
                    </div>
                  </li>
                </ul>
              </section>

              <section class="player-hud__info-block">
                <div v-if="heroes.length === 0" class="leaderboard-list leaderboard-list--empty">
                  <p>你还没有招募到英雄，先去 Boss 池碰碰运气。</p>
                </div>
                <ul v-else class="forge-action-list">
                  <li v-for="hero in heroes" :key="`awaken-${hero.heroId}`">
                    <div class="forge-action-list__copy">
                      <strong>{{ hero.name }}</strong>
                      <div class="forge-action-list__meta">
                        <span>可分解 {{ salvageableHeroCount(hero) }} 个</span>
                        <span>觉醒 {{ hero.awakenLevel || 0 }} / {{ hero.awakenCap || '∞' }}</span>
                        <span>每次 {{ HERO_AWAKEN_COST }} 原石</span>
                      </div>
                      <p>{{ heroAwakenHint(hero) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn || !salvageableHeroCount(hero) || actioningItemId === hero.heroId"
                          @click="salvageHero(hero)"
                      >
                        分解 x{{ salvageableHeroCount(hero) }}
                      </button>
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < HERO_AWAKEN_COST || actioningItemId === hero.heroId || (hero.awakenCap > 0 && hero.awakenLevel >= hero.awakenCap)"
                          @click="awakenHero(hero)"
                      >
                        觉醒
                      </button>
                    </div>
                  </li>
                </ul>
              </section>
            </section>

            <section v-else-if="activeHudTab === 'shop'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">外观商店</p>
                <strong>{{ gems }} 原石</strong>
              </div>

              <div class="forge-grid">
                <article class="forge-summary">
                  <span>已拥有外观</span>
                  <strong>{{ ownedCosmetics.length }} 件</strong>
                  <p>一期只卖轨迹和点击特效，全部拆件售卖，不碰任何数值。</p>
                </article>
                <article class="forge-summary">
                  <span>当前装备</span>
                  <strong>{{ equippedCosmeticSummary.trailName }} / {{ equippedCosmeticSummary.impactName }}</strong>
                  <p>轨迹和点击特效可以自由混搭，星光按钮上会自动降透明度。</p>
                </article>
              </div>

              <section class="cosmetic-preview">
                <div class="player-hud__mini-head">
                  <span>试衣预览</span>
                  <strong>{{ selectedCosmeticSummary.trailName }} / {{ selectedCosmeticSummary.impactName }}</strong>
                </div>
                <div class="cosmetic-preview__stage">
                  <div class="cosmetic-preview__copy">
                    <span>仅自己可见</span>
                    <strong>普通点击、挂机点击和星光按钮都会自动切换到对应表现。</strong>
                    <p>星光态会压制外观亮度，避免和系统提示抢焦点。</p>
                  </div>
                  <span
                      v-if="previewEffectConfig.trailTheme"
                      class="cosmetic-preview__trail"
                      :class="[previewEffectConfig.trailClass, cosmeticModeClasses(previewEffectConfig)]"
                  ></span>
                  <span
                      v-if="previewEffectConfig.impactTheme"
                      class="cosmetic-preview__impact"
                      :class="[previewEffectConfig.impactClass, cosmeticModeClasses(previewEffectConfig)]"
                  >
                    <span
                        v-for="dot in previewDots"
                        :key="`preview-${dot}`"
                        class="cosmetic-preview__dot"
                    ></span>
                  </span>
                </div>
                <div class="cosmetic-preview__actions">
                  <button
                      class="inventory-item__action"
                      type="button"
                      :disabled="!canApplyCosmeticSelection || actioningItemId === 'cosmetic-loadout'"
                      @click="equipSelectedCosmetics"
                  >
                    应用当前搭配
                  </button>
                  <button
                      class="nickname-form__ghost"
                      type="button"
                      :disabled="actioningItemId === 'cosmetic-loadout'"
                      @click="syncCosmeticDraft(equippedCosmetics)"
                  >
                    恢复已装备
                  </button>
                  <button
                      class="nickname-form__ghost"
                      type="button"
                      :disabled="actioningItemId === 'cosmetic-loadout'"
                      @click="syncCosmeticDraft(createEmptyCosmeticLoadout())"
                  >
                    清空搭配
                  </button>
                </div>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>轨迹</span>
                  <strong>{{ cosmeticCollections.trails.length }} 件</strong>
                </div>
                <ul class="shop-grid">
                  <li
                      v-for="item in cosmeticCollections.trails"
                      :key="item.cosmeticId"
                      class="shop-card"
                      :class="{
                        'shop-card--owned': item.owned,
                        'shop-card--equipped': item.equipped,
                        'shop-card--selected': selectedCosmeticLoadout.trailId === item.cosmeticId,
                      }"
                  >
                    <div class="shop-card__preview" :class="`cosmetic-theme--${item.preview?.theme || 'ribbon'}`">
                      <span class="shop-card__preview-mark"></span>
                    </div>
                    <div>
                      <strong>{{ item.name }}</strong>
                      <p>{{ formatRarityLabel(item.rarity) }} · 轨迹 · {{ cosmeticStatusText(item) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          v-if="!item.owned"
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < item.price || actioningItemId === item.cosmeticId"
                          @click="purchaseCosmetic(item)"
                      >
                        购买
                      </button>
                      <button
                          v-else
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn"
                          @click="selectCosmeticItem(item)"
                      >
                        {{ selectedCosmeticLoadout.trailId === item.cosmeticId ? '已选中' : '选这条轨迹' }}
                      </button>
                    </div>
                  </li>
                </ul>
              </section>

              <section class="player-hud__info-block">
                <div class="player-hud__mini-head">
                  <span>点击特效</span>
                  <strong>{{ cosmeticCollections.impacts.length }} 件</strong>
                </div>
                <ul class="shop-grid">
                  <li
                      v-for="item in cosmeticCollections.impacts"
                      :key="item.cosmeticId"
                      class="shop-card"
                      :class="{
                        'shop-card--owned': item.owned,
                        'shop-card--equipped': item.equipped,
                        'shop-card--selected': selectedCosmeticLoadout.impactId === item.cosmeticId,
                      }"
                  >
                    <div class="shop-card__preview" :class="`cosmetic-theme--${item.preview?.theme || 'ribbon'}`">
                      <span class="shop-card__preview-mark"></span>
                    </div>
                    <div>
                      <strong>{{ item.name }}</strong>
                      <p>{{ formatRarityLabel(item.rarity) }} · 点击特效 · {{ cosmeticStatusText(item) }}</p>
                    </div>
                    <div class="inventory-item__actions">
                      <button
                          v-if="!item.owned"
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || gems < item.price || actioningItemId === item.cosmeticId"
                          @click="purchaseCosmetic(item)"
                      >
                        购买
                      </button>
                      <button
                          v-else
                          class="nickname-form__ghost"
                          type="button"
                          :disabled="!isLoggedIn"
                          @click="selectCosmeticItem(item)"
                      >
                        {{ selectedCosmeticLoadout.impactId === item.cosmeticId ? '已选中' : '选这个特效' }}
                      </button>
                    </div>
                  </li>
                </ul>
              </section>
            </section>

          </div>
        </section>
      </aside>


    </section>
</template>
