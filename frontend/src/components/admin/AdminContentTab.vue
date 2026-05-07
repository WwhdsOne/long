<script setup>
defineProps({
  announcementForm: {type: Object, required: true},
  announcements: {type: Array, required: true},
  loadingAnnouncements: {type: Boolean, required: true},
  loadingMessages: {type: Boolean, required: true},
  messagePage: {type: Object, required: true},
  saving: {type: Boolean, required: true},
  deleteAnnouncement: {type: Function, required: true},
  deleteMessage: {type: Function, required: true},
  fetchMessages: {type: Function, required: true},
  formatTime: {type: Function, required: true},
  saveAnnouncement: {type: Function, required: true},
})
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">更新公告</p>
          <strong>{{ announcements.length }} 条</strong>
        </div>

        <form class="admin-form" @submit.prevent="saveAnnouncement">
          <input v-model="announcementForm.title" class="nickname-form__input" type="text" placeholder="公告标题"/>
          <textarea v-model="announcementForm.content" class="nickname-form__input admin-textarea" rows="5"
                    placeholder="公告正文，首次进入前台时会弹一次提醒"></textarea>
          <label class="admin-check">
            <input v-model="announcementForm.active" type="checkbox"/>
            设为生效公告
          </label>
          <button class="nickname-form__submit" type="submit" :disabled="saving">发布公告</button>
        </form>

        <div v-if="loadingAnnouncements" class="feedback-panel">
          <p>公告加载中...</p>
        </div>
        <ul v-else class="inventory-list" style="margin-top: 1rem;">
          <li v-for="item in announcements" :key="item.id" class="inventory-item inventory-item--stacked">
            <div>
              <strong>{{ item.title }}</strong>
              <p>{{ item.active ? '生效中' : '未生效' }} · {{ formatTime(item.publishedAt) }}</p>
              <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
            </div>
            <button class="nickname-form__ghost" type="button" @click="deleteAnnouncement(item.id)">删除</button>
          </li>
        </ul>
      </section>

      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">公共留言墙</p>
          <strong>{{ messagePage.items.length }} 条</strong>
        </div>

        <div v-if="loadingMessages" class="feedback-panel">
          <p>留言加载中...</p>
        </div>
        <ul v-else class="inventory-list">
          <li v-for="item in messagePage.items" :key="item.id" class="inventory-item inventory-item--stacked">
            <div>
              <strong>{{ item.nickname }}</strong>
              <p>{{ formatTime(item.createdAt) }}</p>
              <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
            </div>
            <button class="nickname-form__ghost" type="button" @click="deleteMessage(item.id)">删除</button>
          </li>
        </ul>

        <button
            v-if="messagePage.nextCursor"
            class="nickname-form__ghost"
            type="button"
            :disabled="loadingMessages"
            @click="fetchMessages(messagePage.nextCursor, true)"
        >
          加载更多留言
        </button>
      </section>
    </div>
  </div>
</template>
