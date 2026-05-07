<script setup>
import {reactive, ref} from 'vue'

import {EQUIPMENT_SLOTS} from '../../utils/equipmentSlots'

const props = defineProps({
  adminState: {type: Object, required: true},
  equipmentOptions: {type: Array, required: true},
  grantPlayerEquipment: {type: Function, required: true},
  loadingPlayers: {type: Boolean, required: true},
  playerPage: {type: Object, required: true},
  fetchPlayerPage: {type: Function, required: true},
  resetPlayerPassword: {type: Function, required: true},
  saving: {type: Boolean, required: true},
})

const expandedNickname = ref('')
const grantDrafts = reactive({})

function formatPlayerLoadout(loadout) {
  return EQUIPMENT_SLOTS
      .map((slot) => loadout?.[slot.value]?.name || `空${slot.label}`)
      .join(' / ')
}

function togglePlayerDetail(nickname) {
  expandedNickname.value = expandedNickname.value === nickname ? '' : nickname
}

function grantDraftFor(nickname) {
  if (!grantDrafts[nickname]) {
    grantDrafts[nickname] = {
      itemId: props.equipmentOptions[0]?.itemId || '',
      quantity: 1,
    }
  }
  return grantDrafts[nickname]
}

function inventorySummary(inventory) {
  return Array.isArray(inventory) && inventory.length > 0
      ? inventory.map((item) => item?.name || item?.itemId || '未命名装备').join(' / ')
      : '背包为空'
}

async function submitGrant(player) {
  const grantDraft = grantDraftFor(player.nickname)
  await props.grantPlayerEquipment(player.nickname, {
    itemId: grantDraft.itemId,
    quantity: Number(grantDraft.quantity || 0),
  })
}
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
                {{ formatPlayerLoadout(player.loadout) }}
              </p>
              <p>{{ inventorySummary(player.inventory) }}</p>
            </div>
            <div class="player-detail-card__actions">
              <button class="nickname-form__ghost" type="button" @click="togglePlayerDetail(player.nickname)">
                {{ expandedNickname === player.nickname ? '收起详情' : '查看详情' }}
              </button>
              <button class="nickname-form__ghost" type="button" :disabled="saving"
                      @click="resetPlayerPassword(player.nickname)">
                重置密码
              </button>
            </div>
            <div v-if="expandedNickname === player.nickname" class="player-detail-card">
              <div class="player-detail-card__section">
                <p class="vote-stage__eyebrow">玩家详情</p>
                <p>背包明细：{{ inventorySummary(player.inventory) }}</p>
              </div>

              <form class="admin-form grant-form" @submit.prevent="submitGrant(player)">
                <label>
                  <span>装备模板</span>
                  <select v-model="grantDraftFor(player.nickname).itemId" class="nickname-form__input">
                    <option value="">选择装备模板</option>
                    <option v-for="item in equipmentOptions" :key="item.itemId" :value="item.itemId">
                      {{ item.name }}（{{ item.itemId }}）
                    </option>
                  </select>
                </label>
                <label>
                  <span>数量</span>
                  <input v-model.number="grantDraftFor(player.nickname).quantity" class="nickname-form__input"
                         type="number" min="1" step="1"/>
                </label>
                <button class="nickname-form__submit" type="submit" :disabled="saving || !equipmentOptions.length">
                  发装备
                </button>
              </form>
            </div>
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
