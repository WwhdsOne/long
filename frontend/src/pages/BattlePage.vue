<script setup>
import {computed, ref} from 'vue'
import {usePublicPageState} from './publicPageState'

const {
  boss,
  bossLeaderboard,
  bossLoot,
  bossGoldRange,
  bossStoneRange,
  leaderboard,
  nickname,
  loading,
  errorMessage,
  pendingKeys,
  criticalBursts,
  totalVotes,
  isLoggedIn,
  myClicks,
  myRank,
  myBossDamage,
  myBossRank,
  effectiveIncrement,
  bossStatusLabel,
  bossProgress,
  displayedRecentRewards,
  recentRewardTitle,
  formatDropRate,
  formatRarityLabel,
  formatItemStats,
  equipmentNameParts,
  equipmentNameClass,
  afkSettlement,
  closeAfkSettlementModal,
  clickButton,
} = usePublicPageState()

const bossDropModalOpen = ref(false)

const bossZones = computed(() => {
  if (!boss.value?.parts || !Array.isArray(boss.value.parts)) return []
  const grid = Array.from({length: 5}, () => Array(5).fill(null))
  boss.value.parts.forEach((part) => {
    if (part.x >= 0 && part.x < 5 && part.y >= 0 && part.y < 5) {
      grid[part.y][part.x] = {
        ...part,
        healthPercent: getPartHealthPercent(part),
        zoneKey: `${part.x}-${part.y}`,
      }
    }
  })
  return grid
})

const bossPartCount = computed(() => {
  if (!boss.value?.parts || !Array.isArray(boss.value.parts)) return 0
  return boss.value.parts.length
})

const partTypeLabels = {
  soft: '软组织',
  heavy: '重甲',
  weak: '弱点',
}

const partTypeColors = {
  soft: '#4ade80',
  heavy: '#9ca3af',
  weak: '#ef4444',
}

const bossDropPool = computed(() =>
    bossLoot.value.map((item) => ({
      id: `equipment:${item.itemId}`,
      type: 'equipment',
      label: '装备',
      item,
    })),
)

function openBossDropPool() {
  bossDropModalOpen.value = true
}

function closeBossDropPool() {
  bossDropModalOpen.value = false
}

function getPartHealthPercent(part) {
  if (!part?.maxHp) return 0
  return Math.max(0, Math.min(100, (part.currentHp / part.maxHp) * 100))
}

function getBossZoneButtonKey(zone) {
  if (!zone) return ''
  return `boss-part:${zone.x}-${zone.y}`
}

// 纯点击
function clickBossZone(zone) {
  const key = getBossZoneButtonKey(zone)
  if (key) clickButton(key)
}

function isBossZoneDisabled(zone) {
  const key = getBossZoneButtonKey(zone)
  return !key || !isLoggedIn.value || pendingKeys.value.has(key)
}

function bossZoneAriaLabel(zone) {
  if (!zone) return '空 Boss 分区'
  const label = zone.displayName || partTypeLabels[zone.type] || zone.type
  return `${label} 分区，血量 ${zone.currentHp}/${zone.maxHp}`
}
</script>

<template>
  <section class="stats-band stats-band--wide" aria-label="实时统计">
    <article class="stats-band__card">
      <span class="stats-band__label">Boss 部位</span>
      <strong>{{ bossPartCount }}</strong>
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
      <span class="stats-band__label">Boss 排名</span>
      <strong>{{ isLoggedIn ? (myBossRank ? `#${myBossRank}` : '未上榜') : '先登录' }}</strong>
    </article>
  </section>

  <section class="stage-layout stage-layout--battle">
    <section class="vote-stage">

      <p v-if="errorMessage" class="feedback feedback--error">{{ errorMessage }}</p>

      <section class="vote-stage__boss-hud vote-stage__boss-hud--merged">
        <div class="vote-stage__boss-hud-head">
          <div>
            <div class="vote-stage__head">
              <div>
                <h1 class="vote-stage__worldBoss">世界 Boss 战场</h1>
                <p class="vote-stage__eyebrow">当前 Boss</p>
                <h2>{{ boss?.name || '等待 Boss 登场' }}</h2>
              </div>
            </div>
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
        <div v-if="loading" class="feedback-panel feedback-panel--compact">
          <p>正在加载 Boss 战场...</p>
        </div>
        <div v-else-if="!boss" class="feedback-panel feedback-panel--compact">
          <p>当前没有活动 Boss。</p>
        </div>
        <div v-else-if="bossZones.length === 0" class="feedback-panel feedback-panel--compact">
          <p>当前 Boss 尚未配置可攻击分区。</p>
        </div>
        <div v-else class="boss-part-grid">
          <div v-for="(row, yi) in bossZones" :key="yi" class="boss-part-grid__row">
            <button
                v-for="(zone, xi) in row"
                :key="yi + '-' + xi"
                class="boss-part-cell boss-zone-button"
                :class="{
                  'boss-part-cell--alive': zone?.alive,
                  'boss-part-cell--dead': zone && !zone.alive,
                  'boss-part-cell--soft': zone?.type === 'soft',
                  'boss-part-cell--heavy': zone?.type === 'heavy',
                  'boss-part-cell--weak': zone?.type === 'weak',
                  'boss-part-cell--low': zone?.alive && zone.healthPercent < 25,
                  'boss-zone-button--empty': !zone,
                  'boss-zone-button--pending': pendingKeys.has(getBossZoneButtonKey(zone)),
                  'boss-zone-button--critical': Boolean(criticalBursts[getBossZoneButtonKey(zone)]),
                }"
                :style="zone ? { '--part-color': partTypeColors[zone.type] || '#64748b' } : {}"
                type="button"
                :disabled="isBossZoneDisabled(zone)"
                :aria-label="bossZoneAriaLabel(zone)"
                @click="clickBossZone(zone)"
            >
              <template v-if="zone">
                <img
                    v-if="zone.imagePath"
                    class="boss-part-cell__image"
                    :src="zone.imagePath"
                    :alt="zone.displayName || partTypeLabels[zone.type] || zone.type"
                />
                <span
                    v-if="criticalBursts[getBossZoneButtonKey(zone)]"
                    :key="criticalBursts[getBossZoneButtonKey(zone)].nonce"
                    class="boss-zone-button__critical-text"
                    aria-hidden="true"
                >
                    {{ criticalBursts[getBossZoneButtonKey(zone)].label }}
                  </span>
                <div class="boss-part-cell__type">{{ partTypeLabels[zone.type] || zone.type }}</div>
                <strong class="boss-zone-button__label">{{
                    zone.displayName || partTypeLabels[zone.type] || zone.type
                  }}</strong>
                <div class="boss-part-cell__bar">
                    <span
                        class="boss-part-cell__fill"
                        :style="{ width: `${zone.healthPercent}%` }"
                    ></span>
                </div>
                <div class="boss-zone-button__meta">
                  <span>{{ zone.currentHp }}/{{ zone.maxHp }}</span>
                  <span>点击 +1</span>
                </div>
              </template>
              <span v-else class="boss-part-cell__empty"></span>
            </button>
          </div>
        </div>
        <div class="vote-stage__boss-hud-stats">
          <span>我的伤害 {{ myBossDamage }}</span>
          <span>Boss 榜 {{ bossLeaderboard.length }} 人</span>
          <span>掉落池 {{ bossDropPool.length }} 件</span>
          <span v-if="displayedRecentRewards.length > 0">最近掉落 {{ recentRewardTitle }}</span>
        </div>
        <div class="vote-stage__boss-note">
          <span>只有对 Boss 造成至少 1% 生命值的伤害，才有资格掉落装备与资源。</span>
          <span v-if="boss" class="boss-drop-link">
              <button type="button" @click="openBossDropPool">
                点击查看 Boss 掉落池
              </button>
              <span>{{ bossDropPool.length }} 件掉落物</span>
            </span>
        </div>
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
                <li>{{ formatItemStats(item) }}</li>
              </ul>
            </article>
          </div>
        </section>
        <section class="boss-drop-modal__section">
          <div class="boss-drop-modal__section-head">
            <span>资源掉落</span>
          </div>
          <div class="boss-drop-pool__grid">
            <article class="boss-drop-card boss-drop-card--detail">
              <span class="boss-drop-card__type">金币</span>
              <strong>{{ bossGoldRange.min }} ~ {{ bossGoldRange.max }}</strong>
              <ul class="boss-drop-card__details">
                <li>按击杀结算随机波动（向下取整）</li>
              </ul>
            </article>
            <article class="boss-drop-card boss-drop-card--detail">
              <span class="boss-drop-card__type">强化石</span>
              <strong>{{ bossStoneRange.min }} ~ {{ bossStoneRange.max }}</strong>
              <ul class="boss-drop-card__details">
                <li>按击杀结算随机波动（向下取整）</li>
              </ul>
            </article>
          </div>
        </section>
      </article>
    </section>

    <section v-if="afkSettlement" class="boss-drop-modal" aria-label="挂机结算">
      <div class="boss-drop-modal__backdrop" @click="closeAfkSettlementModal"></div>
      <article class="boss-drop-modal__card">
        <div class="boss-drop-modal__head">
          <div>
            <p class="vote-stage__eyebrow">挂机结算</p>
            <strong>离页挂机已结束</strong>
          </div>
          <button class="nickname-form__ghost" type="button" @click="closeAfkSettlementModal">关闭</button>
        </div>
        <div class="leaderboard-list">
          <p>击杀数：{{ afkSettlement.kills }}</p>
          <p>金币：+{{ afkSettlement.goldTotal }}</p>
          <p>强化石：+{{ afkSettlement.stoneTotal }}</p>
        </div>
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
