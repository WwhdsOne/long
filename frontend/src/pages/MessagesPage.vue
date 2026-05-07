<script setup>
import {usePublicPageState} from './publicPageState'

const {
  leaderboard,
  latestAnnouncement,
  nickname,
  nicknameDraft,
  passwordDraft,
  messages,
  messageNextCursor,
  loadingMessages,
  postingMessage,
  messageDraft,
  messageError,
  isLoggedIn,
  formatTime,
  loadMessages,
  submitMessage,
  submitNickname,
  resetNickname,
} = usePublicPageState()
</script>

<template>
  <section class="stage-layout stage-layout--messages stage-layout--single">
    <aside class="player-hud player-hud--page">
      <section class="player-hud__shell">
        <div class="player-hud__content messages-page__grid">
          <section class="player-hud__panel messages-page__feed">
            <div class="player-hud__section-head">
              <p class="vote-stage__eyebrow">公共留言墙</p>
              <strong>{{ messages.length }} 条</strong>
            </div>

            <form class="admin-form player-hud__message-form" @submit.prevent="submitMessage">
              <textarea v-model="messageDraft" class="nickname-form__input admin-textarea" rows="4" maxlength="200"
                        placeholder="说点什么，所有人都能看到。"></textarea>
              <button class="nickname-form__submit" type="submit" :disabled="postingMessage || !isLoggedIn">
                {{ postingMessage ? '发送中...' : '发送留言' }}
              </button>
            </form>

            <p v-if="messageError" class="feedback feedback--error">{{ messageError }}</p>
            <div v-if="loadingMessages" class="leaderboard-list leaderboard-list--empty"><p>留言加载中...</p></div>
            <div v-else-if="messages.length === 0" class="leaderboard-list leaderboard-list--empty"><p>
              还没有留言，先写第一条。</p></div>
            <ul v-else class="history-list">
              <li v-for="item in messages" :key="item.id" class="history-item">
                <div class="history-item__head"><strong>{{ item.nickname }}</strong><span>{{
                    formatTime(item.createdAt)
                  }}</span></div>
                <p class="history-item__content history-item__content--multiline">{{ item.content }}</p>
              </li>
            </ul>
            <button v-if="messageNextCursor" class="nickname-form__ghost player-hud__retry" type="button"
                    :disabled="loadingMessages" @click="loadMessages(messageNextCursor, true)">加载更多
            </button>
          </section>

          <aside class="messages-page__side">
            <section class="player-hud__info-block">
              <div class="player-hud__mini-head"><span>最新公告</span><strong>{{
                  latestAnnouncement?.title || '暂无'
                }}</strong></div>
              <p class="player-hud__note player-hud__note--multiline">{{
                  latestAnnouncement?.content || '当前还没有新的站内公告。'
                }}</p>
            </section>
            <section class="player-hud__info-block">
              <div class="player-hud__mini-head"><span>规则</span><strong>公开留言</strong></div>
              <p class="player-hud__note player-hud__note--multiline">留言按时间倒序展示；发送前需要登录当前账号。</p>
            </section>
            <section class="player-hud__info-block">
              <div class="player-hud__mini-head"><span>在线玩家</span><strong>{{ leaderboard.length }} 人</strong></div>
              <ol v-if="leaderboard.length > 0" class="leaderboard-list">
                <li v-for="entry in leaderboard" :key="entry.nickname" class="leaderboard-list__item"
                    :class="{ 'leaderboard-list__item--me': entry.nickname === nickname }">
                  <span class="leaderboard-list__rank">#{{ entry.rank }}</span><span
                    class="leaderboard-list__name">{{ entry.nickname }}</span><strong
                    class="leaderboard-list__count">{{ entry.clickCount }}</strong>
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
