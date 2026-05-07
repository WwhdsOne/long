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
          <p class="vote-stage__eyebrow">限流黑名单</p>
          <strong>{{ blacklistPage.items.length }} 条</strong>
        </div>
        <button class="nickname-form__ghost" type="button" :disabled="loadingBlacklist" @click="fetchBlacklist()">
          刷新
        </button>
      </div>

      <div v-if="loadingBlacklist" class="feedback-panel">
        <p>黑名单加载中...</p>
      </div>
      <div v-else-if="blacklistPage.items.length === 0" class="feedback-panel">
        <p>当前没有封禁中的昵称。</p>
      </div>
      <ul v-else class="inventory-list">
        <li v-for="entry in blacklistPage.items" :key="entry.clientId" class="inventory-item inventory-item--stacked">
          <div>
            <strong>{{ entry.nickname }}</strong>
            <p v-if="entry.clientId.startsWith('ip:')">IP：{{ entry.clientId.slice(3) }}</p>
            <p>封禁开始：{{ formatTime(entry.blockedAt) }}</p>
            <p>封禁结束：{{ formatTime(entry.blockedUntil) }}</p>
            <p>剩余时间：{{ formatDuration(entry.remainingSeconds) }}</p>
          </div>
          <button
              class="nickname-form__ghost"
              type="button"
              :disabled="saving"
              @click="unblockBlacklistEntry(entry.clientId, entry.nickname)"
          >
            手动解封
          </button>
        </li>
      </ul>
    </section>
  </div>
</template>
