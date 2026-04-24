<script setup>
import {computed, ref} from 'vue'

import {usePublicPageState} from './publicPageState'

const {
  AUTO_CLICK_RATE_LABEL,
  buttons,
  boss,
  bossLeaderboard,
  bossLoot,
  bossHeroLoot,
  nickname,
  loading,
  errorMessage,
  pendingKeys,
  criticalBursts,
  autoClickEnabled,
  buttonCount,
  totalVotes,
  displayedButtons,
  isLoggedIn,
  myClicks,
  myRank,
  myBossDamage,
  effectiveIncrement,
  autoClickTargetLabel,
  canStartAutoClick,
  autoClickStatus,
  bossStatusLabel,
  bossProgress,
  displayedRecentRewards,
  recentRewardTitle,
  formatDropRate,
  formatRarityLabel,
  formatItemStats,
  equipmentNameParts,
  equipmentNameClass,
  formatEnhanceCap,
  formatAwakenCap,
  formatHeroTrait,
  heroImageAlt,
  handlePressStart,
  handlePressEnd,
  handlePressCancel,
  toggleAutoClick,
  clickButton,
} = usePublicPageState()

const bossDropModalOpen = ref(false)

const bossDropPool = computed(() => [
  ...bossLoot.value.map((item) => ({
    id: `equipment:${item.itemId}`,
    type: 'equipment',
    label: '装备',
    item,
  })),
  ...bossHeroLoot.value.map((hero) => ({
    id: `hero:${hero.heroId}`,
    type: 'hero',
    label: '英雄',
    item: hero,
  })),
])

function openBossDropPool() {
  bossDropModalOpen.value = true
}

function closeBossDropPool() {
  bossDropModalOpen.value = false
}
</script>

<template>
<section class="stats-band stats-band--wide" aria-label="实时统计">
      <article class="stats-band__card">
        <span class="stats-band__label">当前按钮</span>
        <strong>{{ buttonCount }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">累计点击</span>
        <strong>{{ totalVotes }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">我的点击</span>
        <strong>{{ isLoggedIn ? myClicks : '先登录' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">我的排名</span>
        <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
      </article>
      <article class="stats-band__card">
        <span class="stats-band__label">单击增量</span>
        <strong>+{{ effectiveIncrement }}</strong>
      </article>
    </section>

    <section class="stage-layout stage-layout--battle">
      <section class="vote-stage">
        <div class="vote-stage__head">
          <div>
            <p class="vote-stage__eyebrow">现场投票墙 · 世界 Boss</p>
            <h2>{{ boss?.name || '看见哪个想按，就直接拍下去。' }}</h2>
            <p class="vote-stage__hint vote-stage__hint--wide">
              {{
                !boss
                    ? '当前休战中，按钮依然正常计票。'
                    : boss.status === 'active'
                        ? '全服正在集火当前 Boss，每次点击都会把装备加成一起折算成伤害。'
                        : '这只 Boss 已经倒下，等待后台开启下一只。'
              }}
            </p>
          </div>
          <p v-if="!errorMessage" class="vote-stage__hint">
            {{ isLoggedIn ? `现在上墙的是 ${nickname}` : '先登录账号，再开始冲榜。' }}
          </p>
        </div>

        <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

        <section class="vote-stage__boss-hud vote-stage__boss-hud--merged">
          <div class="vote-stage__boss-hud-head">
            <div>
              <p class="vote-stage__eyebrow">当前 Boss</p>
              <strong>{{ boss?.name || '休战中' }}</strong>
            </div>
            <div class="boss-stage__meta">
              <span class="boss-stage__pill">{{ bossStatusLabel }}</span>
              <strong v-if="boss">HP {{ boss.currentHp }} / {{ boss.maxHp }}</strong>
              <strong v-else>我的伤害 {{ myBossDamage }}</strong>
            </div>
          </div>
          <div v-if="boss" class="boss-stage__bar boss-stage__bar--compact">
            <span class="boss-stage__bar-fill" :style="{ width: `${bossProgress}%` }"></span>
          </div>
          <div class="vote-stage__boss-hud-stats">
            <span>我的伤害 {{ myBossDamage }}</span>
            <span>Boss 榜 {{ bossLeaderboard.length }} 人</span>
            <span>掉落池 {{ bossDropPool.length }} 件</span>
            <span v-if="displayedRecentRewards.length > 0">最近掉落 {{ recentRewardTitle }}</span>
          </div>
          <div class="vote-stage__boss-note">
            <span>只有对 Boss 造成超过 1% 生命值的伤害，才有资格掉落装备。</span>
            <span v-if="boss" class="boss-drop-link">
              <button type="button" @click="openBossDropPool">
                点击查看 Boss 掉落池
              </button>
              <span>{{ bossDropPool.length }} 件掉落物</span>
            </span>
          </div>
        </section>

        <div v-if="loading" class="feedback-panel">
          <p>正在把现场按钮搬上来...</p>
        </div>

        <div v-else-if="buttons.length === 0" class="feedback-panel">
          <p>还没有按钮上墙，先加一个再回来看看。</p>
        </div>

        <div v-else>
          <div v-if="displayedButtons.length === 0" class="feedback-panel">
            <p>还没有可用按钮。</p>
          </div>

          <div v-else class="button-grid">
            <button
                v-for="button in displayedButtons"
                :key="button.key"
                class="vote-card"
                :class="{
              'vote-card--image': button.imagePath,
              'vote-card--pending': pendingKeys.has(button.key),
              'vote-card--critical': Boolean(criticalBursts[button.key]),
              'vote-card--locked': !isLoggedIn,
            }"
                type="button"
                :disabled="pendingKeys.has(button.key) || !isLoggedIn"
                :aria-label="`${button.label}，当前 ${button.count} 票`"
                @pointerdown="handlePressStart(button.key, $event)"
                @pointerup="handlePressEnd(button.key, $event)"
                @pointercancel="handlePressCancel(button.key)"
                @click="clickButton(button.key)"
            >
              <span class="vote-card__shine"></span>
              <span
                  v-if="criticalBursts[button.key]"
                  :key="criticalBursts[button.key].nonce"
                  class="vote-card__burst"
                  aria-hidden="true"
              ></span>
              <span
                  v-if="criticalBursts[button.key]"
                  :key="`${criticalBursts[button.key].nonce}-label`"
                  class="vote-card__critical-text"
                  aria-hidden="true"
              >
              {{ criticalBursts[button.key].label }}
            </span>
              <span class="vote-card__badge">
              {{
                  !isLoggedIn
                      ? '先登录'
                      : pendingKeys.has(button.key)
                          ? '正在记票'
                          : boss?.status === 'active'
                              ? `拍一下 +${effectiveIncrement} 并打 Boss`
                              : `拍一下 +${effectiveIncrement}`
                }}
            </span>

              <img
                  v-if="button.imagePath"
                  class="vote-card__image"
                  :src="button.imagePath"
                  :alt="button.imageAlt || button.label"
              />
              <strong v-else class="vote-card__label">{{ button.label }}</strong>

              <span class="vote-card__count">{{ button.count }}</span>
            </button>
          </div>
        </div>

        <section class="player-hud__auto battle-auto-panel">
          <div class="player-hud__section-head">
            <div>
              <p class="vote-stage__eyebrow">挂机</p>
              <strong>官方挂机托管</strong>
            </div>
            <span class="player-hud__pill" :class="{ 'player-hud__pill--active': autoClickEnabled }">
              {{ autoClickEnabled ? '运行中' : '未开启' }}
            </span>
          </div>

          <p class="player-hud__note">{{ autoClickStatus }}</p>

          <div class="player-hud__auto-meta">
            <span class="player-hud__auto-chip">目标：{{ autoClickTargetLabel }}</span>
            <span class="player-hud__auto-chip">{{ AUTO_CLICK_RATE_LABEL }}</span>
          </div>

          <button
              class="nickname-form__submit player-hud__auto-button"
              type="button"
              :disabled="!autoClickEnabled && !canStartAutoClick"
              @click="toggleAutoClick"
          >
            {{ autoClickEnabled ? '关闭挂机' : '开启挂机' }}
          </button>
        </section>
      </section>

      <section
          v-if="bossDropModalOpen"
          class="boss-drop-modal"
          aria-label="Boss 掉落池"
      >
        <div class="boss-drop-modal__backdrop" @click="closeBossDropPool"></div>
        <article class="boss-drop-modal__card">
          <div class="boss-drop-modal__head">
            <div>
              <p class="vote-stage__eyebrow">Boss 掉落池</p>
              <strong>{{ boss?.name || '当前 Boss' }}</strong>
            </div>
            <button class="nickname-form__ghost" type="button" @click="closeBossDropPool">关闭</button>
          </div>

          <div v-if="bossDropPool.length === 0" class="leaderboard-list leaderboard-list--empty">
            <p>当前 Boss 还没配置掉落池。</p>
          </div>

          <section v-if="bossLoot.length > 0" class="boss-drop-modal__section">
            <div class="boss-drop-modal__section-head">
              <span>装备掉落</span>
              <strong>{{ bossLoot.length }} 件</strong>
            </div>
            <div class="boss-drop-pool__grid">
              <article
                  v-for="item in bossLoot"
                  :key="item.itemId"
                  class="boss-drop-card boss-drop-card--detail"
              >
                  <span class="boss-drop-card__type">装备</span>
                  <strong>
                    <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                    <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                  </strong>
                <ul class="boss-drop-card__details">
                  <li>掉落概率：{{ formatDropRate(item.dropRatePercent) }}</li>
                  <li>稀有度：{{ formatRarityLabel(item.rarity) }}</li>
                  <li>部位：{{ item.slot || '未分类' }}</li>
                  <li>{{ formatEnhanceCap(item.enhanceCap) }}</li>
                  <li>{{ formatItemStats(item) }}</li>
                </ul>
              </article>
            </div>
          </section>

          <section v-if="bossHeroLoot.length > 0" class="boss-drop-modal__section">
            <div class="boss-drop-modal__section-head">
              <span>英雄掉落</span>
              <strong>{{ bossHeroLoot.length }} 位</strong>
            </div>
            <div class="boss-drop-pool__grid">
              <article
                  v-for="hero in bossHeroLoot"
                  :key="hero.heroId"
                  class="boss-drop-card boss-drop-card--detail"
              >
                  <span class="boss-drop-card__type">英雄</span>
                  <img
                      v-if="hero.imagePath"
                      class="boss-drop-card__avatar"
                      :src="hero.imagePath"
                      :alt="heroImageAlt(hero)"
                  />
                  <strong>{{ hero.heroName || hero.name || hero.heroId }}</strong>
                <ul class="boss-drop-card__details">
                  <li>掉落概率：{{ formatDropRate(hero.dropRatePercent) }}</li>
                  <li>{{ formatAwakenCap(hero.awakenCap) }}</li>
                  <li>{{ formatItemStats(hero) }}</li>
                  <li>{{ formatHeroTrait(hero) }}</li>
                </ul>
              </article>
            </div>
          </section>
        </article>
      </section>


      <aside class="social-panel social-panel--ranking">
        <section class="social-card leaderboard-card leaderboard-card--stacked">
          <section class="leaderboard-card__section">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">Boss 伤害榜</p>
              <strong>{{ bossLeaderboard.length || 0 }} 人</strong>
            </div>

            <ol v-if="bossLeaderboard.length > 0" class="leaderboard-list">
              <li
                  v-for="entry in bossLeaderboard"
                  :key="entry.nickname"
                  class="leaderboard-list__item"
                  :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }"
              >
                <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
                <span class="leaderboard-list__name">{{ entry.nickname }}</span>
                <strong class="leaderboard-list__count">{{ entry.damage }}</strong>
              </li>
            </ol>
            <div v-else class="leaderboard-list leaderboard-list--empty">
              <p>当前 Boss 还没人动手，或者正在休战。</p>
            </div>
          </section>

          <section class="leaderboard-card__section">
            <div class="social-card__head">
              <p class="vote-stage__eyebrow">实时排行榜</p>
              <strong>前 {{ leaderboard.length || 0 }} 名</strong>
            </div>

            <ol v-if="leaderboard.length > 0" class="leaderboard-list">
              <li
                  v-for="entry in leaderboard"
                  :key="entry.nickname"
                  class="leaderboard-list__item"
                  :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }"
              >
                <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
                <span class="leaderboard-list__name">{{ entry.nickname }}</span>
                <strong class="leaderboard-list__count">{{ entry.clickCount }}</strong>
              </li>
            </ol>
            <div v-else class="leaderboard-list leaderboard-list--empty">
              <p>还没人上榜，等你来开张。</p>
            </div>
          </section>
        </section>
      </aside>

    </section>
</template>
