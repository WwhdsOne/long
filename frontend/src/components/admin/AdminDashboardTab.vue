<script setup>
defineProps({
  adminState: { type: Object, required: true },
  loadingPlayers: { type: Boolean, required: true },
  playerPage: { type: Object, required: true },
  fetchPlayerPage: { type: Function, required: true },
  resetPlayerPassword: { type: Function, required: true },
  saving: { type: Boolean, required: true },
})
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">Boss 伤害榜</p>
          <strong>{{ adminState.bossLeaderboard.length }} 人</strong>
        </div>

        <ol class="leaderboard-list">
          <li v-for="entry in adminState.bossLeaderboard" :key="entry.nickname" class="leaderboard-list__item">
            <span class="leaderboard-list__rank">#{{ entry.rank }}</span>
            <span class="leaderboard-list__name">{{ entry.nickname }}</span>
            <strong class="leaderboard-list__count">{{ entry.damage }}</strong>
          </li>
        </ol>
      </section>

      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">玩家概览</p>
          <strong>{{ adminState.playerCount }} 人</strong>
        </div>

        <p class="social-card__copy">最近 24 小时活跃 {{ adminState.recentPlayerCount }} 人</p>

        <div v-if="loadingPlayers" class="feedback-panel">
          <p>玩家列表加载中...</p>
        </div>
        <ul v-else class="inventory-list">
          <li v-for="player in playerPage.items" :key="player.nickname" class="inventory-item inventory-item--stacked">
            <div>
              <strong>{{ player.nickname }}</strong>
              <p>累计点击 {{ player.clickCount }} · 背包 {{ player.inventory.length }} 件</p>
              <p>
                穿戴：
                {{ player.loadout.weapon?.name || '空武器' }} /
                {{ player.loadout.armor?.name || '空护甲' }} /
                {{ player.loadout.accessory?.name || '空饰品' }}
              </p>
            </div>
            <button class="nickname-form__ghost" type="button" :disabled="saving" @click="resetPlayerPassword(player.nickname)">
              重置密码
            </button>
          </li>
        </ul>

        <button
          v-if="playerPage.nextCursor"
          class="nickname-form__ghost"
          type="button"
          :disabled="loadingPlayers"
          @click="fetchPlayerPage(playerPage.nextCursor, true)"
        >
          加载更多玩家
        </button>
      </section>
    </div>
  </div>
</template>
