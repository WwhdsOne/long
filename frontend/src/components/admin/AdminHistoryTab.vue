<script setup>
defineProps({
  bossHistoryPage: { type: Object, required: true },
  loadingHistory: { type: Boolean, required: true },
  formatItemStats: { type: Function, required: true },
  fetchBossHistory: { type: Function, required: true },
})
</script>

<template>
  <div class="admin-section">
    <div v-if="loadingHistory" class="feedback-panel">
      <p>加载历史 Boss...</p>
    </div>
    <div v-else-if="bossHistoryPage.items.length === 0" class="feedback-panel">
      <p>暂无历史 Boss 记录。</p>
    </div>
    <div v-else class="admin-grid">
      <section v-for="entry in bossHistoryPage.items" :key="entry.id" class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">{{ entry.status === 'defeated' ? '已击败' : entry.status }}</p>
          <strong>{{ entry.name }}</strong>
        </div>
        <p class="social-card__copy">
          ID: {{ entry.id }} · 血量 {{ entry.currentHp }}/{{ entry.maxHp }}
        </p>

        <div v-if="entry.loot.length > 0" style="margin-top: 0.5rem;">
          <p class="vote-stage__eyebrow">掉落池</p>
          <ul class="inventory-list">
            <li v-for="item in entry.loot" :key="item.itemId" class="inventory-item">
              <div>
                <strong>{{ item.itemName || item.itemId }}</strong>
              <p>{{ item.itemId }} · {{ item.slot }} · 掉落几率 {{ item.dropRatePercent }}%</p>
                <p>{{ formatItemStats(item) }}</p>
              </div>
            </li>
          </ul>
        </div>

        <div v-if="entry.damage.length > 0" style="margin-top: 0.5rem;">
          <p class="vote-stage__eyebrow">伤害榜</p>
          <ol class="leaderboard-list">
            <li v-for="d in entry.damage" :key="d.nickname" class="leaderboard-list__item">
              <span class="leaderboard-list__rank">#{{ d.rank }}</span>
              <span class="leaderboard-list__name">{{ d.nickname }}</span>
              <strong class="leaderboard-list__count">{{ d.damage }}</strong>
            </li>
          </ol>
        </div>
      </section>
      <div class="admin-inline-actions" style="grid-column: 1 / -1;">
        <button
          class="nickname-form__ghost"
          type="button"
          :disabled="loadingHistory || bossHistoryPage.page <= 1"
          @click="fetchBossHistory(bossHistoryPage.page - 1)"
        >
          上一页
        </button>
        <span class="feedback">第 {{ bossHistoryPage.page }} / {{ Math.max(bossHistoryPage.totalPages, 1) }} 页</span>
        <button
          class="nickname-form__ghost"
          type="button"
          :disabled="loadingHistory || bossHistoryPage.page >= bossHistoryPage.totalPages"
          @click="fetchBossHistory(bossHistoryPage.page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>
