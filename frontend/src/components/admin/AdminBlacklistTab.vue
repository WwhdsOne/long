<script setup>
defineProps({
  blacklistPage: {type: Object, required: true},
  loadingBlacklist: {type: Boolean, required: true},
  saving: {type: Boolean, required: true},
  fetchBlacklist: {type: Function, required: true},
  unblockBlacklistEntry: {type: Function, required: true},
  formatDuration: {type: Function, required: true},
  formatTime: {type: Function, required: true},
})
</script>

<template>
  <div class="admin-section">
    <section class="social-card">
      <div class="social-card__head">
        <div>
          <p class="vote-stage__eyebrow">账号风险名单</p>
          <strong>{{ blacklistPage.items.length }} 条</strong>
        </div>
        <button class="nickname-form__ghost" type="button" :disabled="loadingBlacklist" @click="fetchBlacklist()">
          刷新
        </button>
      </div>

      <div v-if="loadingBlacklist" class="feedback-panel">
        <p>风险名单加载中...</p>
      </div>
      <div v-else-if="blacklistPage.items.length === 0" class="feedback-panel">
        <p>当前没有有分账号。</p>
      </div>
      <ul v-else class="inventory-list">
        <li v-for="entry in blacklistPage.items" :key="entry.nickname" class="inventory-item inventory-item--stacked">
          <div>
            <strong>{{ entry.nickname }}</strong>
            <p>当前积分：{{ entry.score }}</p>
            <p v-if="entry.banUntil">封禁截止：{{ formatTime(entry.banUntil) }}</p>
          </div>
          <button
              class="nickname-form__ghost"
              type="button"
              :disabled="saving"
              @click="unblockBlacklistEntry(entry.nickname)"
          >
            清除风险状态
          </button>
        </li>
      </ul>
    </section>
  </div>
</template>
