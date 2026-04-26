<script setup>
import AdminBossTab from '../components/admin/AdminBossTab.vue'
import AdminContentTab from '../components/admin/AdminContentTab.vue'
import AdminDashboardTab from '../components/admin/AdminDashboardTab.vue'
import AdminEquipmentTab from '../components/admin/AdminEquipmentTab.vue'
import AdminHistoryTab from '../components/admin/AdminHistoryTab.vue'
import { useAdminPage } from './admin/useAdminPage'
import {reactive} from "vue";

const admin = reactive(useAdminPage())
</script>

<template>
  <main class="page-shell admin-shell">
    <div class="page-shell__glow page-shell__glow--pink"></div>
    <div class="page-shell__glow page-shell__glow--blue"></div>
    <div class="page-shell__glow page-shell__glow--yellow"></div>

    <section class="hero">
      <div class="hero__copy">
        <p class="hero__eyebrow">Long Control Room</p>
        <h1>管理现场、Boss 与掉落。</h1>
        <p class="hero__lede">这里管理 Boss、装备、公告和留言，也能把素材图片直传到 OSS。</p>
      </div>

      <div class="hero__status">
        <span class="live-pill">
          <span class="live-pill__dot"></span>
          {{ admin.authenticated ? '后台已解锁' : '等待登录' }}
        </span>
        <a class="hero__admin-link" href="/">返回前台</a>
      </div>
    </section>

    <section v-if="admin.checkingSession.value" class="admin-card admin-card--single">
      <p class="feedback-panel">正在确认后台会话...{{admin.checkingSession.value}}</p>
    </section>

    <section v-else-if="!admin.authenticated" class="admin-card admin-card--single">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">后台登录</p>
        <strong>固定口令</strong>
      </div>

        <p class="social-card__copy">先输入后台账号口令，解锁 Boss、装备和内容配置。</p>
      <p v-if="admin.errorMessage" class="feedback feedback--error">{{ admin.errorMessage }}</p>

      <form class="admin-form" @submit.prevent="admin.login">
        <input v-model="admin.loginForm.username" class="nickname-form__input" type="text" placeholder="账号" />
        <input v-model="admin.loginForm.password" class="nickname-form__input" type="password" placeholder="口令" />
        <button class="nickname-form__submit" type="submit" :disabled="admin.saving">
          {{ admin.saving ? '正在解锁...' : '进入后台' }}
        </button>
      </form>
    </section>

    <section v-else class="admin-layout">
      <article class="admin-card admin-card--toolbar">
        <div>
          <p class="vote-stage__eyebrow">控制台</p>
          <strong>{{ admin.adminState.boss?.name || '暂无活动 Boss' }}</strong>
        </div>

        <div class="admin-toolbar__actions">
          <button class="nickname-form__ghost" type="button" @click="admin.refreshAll">刷新数据</button>
          <button class="nickname-form__ghost" type="button" @click="admin.logout">退出后台</button>
        </div>

        <p v-if="admin.errorMessage" class="feedback feedback--error">{{ admin.errorMessage }}</p>
        <p v-else-if="admin.successMessage" class="feedback">{{ admin.successMessage }}</p>
      </article>

      <article class="admin-card">
        <div class="admin-tabs">
          <button class="admin-tab" :class="{ 'admin-tab--active': admin.activeTab === 'boss' }" @click="admin.activeTab = 'boss'">Boss</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': admin.activeTab === 'equipment' }" @click="admin.activeTab = 'equipment'; admin.fetchEquipmentPage(admin.equipmentPage.page)">装备</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': admin.activeTab === 'content' }" @click="admin.activeTab = 'content'; admin.fetchAnnouncements(); admin.fetchMessages()">内容</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': admin.activeTab === 'history' }" @click="admin.activeTab = 'history'; admin.fetchBossHistory(admin.bossHistoryPage.page)">历史</button>
          <button class="admin-tab" :class="{ 'admin-tab--active': admin.activeTab === 'dashboard' }" @click="admin.activeTab = 'dashboard'">看板</button>
        </div>

        <div v-if="admin.loading" class="feedback-panel">
          <p>后台数据加载中...</p>
        </div>

        <AdminBossTab
          v-else-if="admin.activeTab === 'boss'"
          :admin-state="admin.adminState"
          :boss-cycle-enabled="admin.bossCycleEnabled"
          :boss-form="admin.bossForm"
          :boss-templates="admin.bossTemplates"
          :equipment-options="admin.equipmentOptions"
          :has-boss="admin.hasBoss"
          :has-equipment-templates="admin.hasEquipmentTemplates"
          :loot-rows="admin.lootRows"
          :saving="admin.saving"
          :selected-boss-template="admin.selectedBossTemplate"
          :selected-boss-template-id="admin.selectedBossTemplateId"
          :add-loot-row="admin.addLootRow"
          :deactivate-boss="admin.deactivateBoss"
          :delete-boss-template="admin.deleteBossTemplate"
          :disable-boss-cycle="admin.disableBossCycle"
          :edit-boss-template="admin.editBossTemplate"
          :enable-boss-cycle="admin.enableBossCycle"
          :find-equipment-template="admin.findEquipmentTemplate"
          :format-item-stats="admin.formatItemStats"
          :remove-loot-row="admin.removeLootRow"
          :save-boss-template="admin.saveBossTemplate"
          :save-loot="admin.saveLoot"
          :select-boss-template="admin.selectBossTemplate"
        />

        <AdminEquipmentTab
          v-else-if="admin.activeTab === 'equipment'"
          :equipment-page="admin.equipmentPage"
          :equipment-form="admin.equipmentForm"
          :equipment-prompt="admin.equipmentPrompt"
          :show-equipment-editor="admin.showEquipmentEditor"
          :loading-equipment="admin.loadingEquipment"
          :generating-equipment-draft="admin.generatingEquipmentDraft"
          :saving="admin.saving"
          :format-item-stats="admin.formatItemStats"
          :open-new-equipment="admin.openNewEquipment"
          :update-equipment-prompt="admin.updateEquipmentPrompt"
          :generate-equipment-draft="admin.generateEquipmentDraft"
          :save-equipment="admin.saveEquipment"
          :edit-equipment="admin.editEquipment"
          :delete-equipment="admin.deleteEquipment"
          :fetch-equipment-page="admin.fetchEquipmentPage"
          :upload-equipment-image="admin.uploadEquipmentImage"
        />
        <AdminContentTab
          v-else-if="admin.activeTab === 'content'"
          :announcement-form="admin.announcementForm"
          :announcements="admin.announcements"
          :loading-announcements="admin.loadingAnnouncements"
          :loading-messages="admin.loadingMessages"
          :message-page="admin.messagePage"
          :saving="admin.saving"
          :delete-announcement="admin.deleteAnnouncement"
          :delete-message="admin.deleteMessage"
          :fetch-messages="admin.fetchMessages"
          :format-time="admin.formatTime"
          :save-announcement="admin.saveAnnouncement"
        />

        <AdminHistoryTab
          v-else-if="admin.activeTab === 'history'"
          :boss-history-page="admin.bossHistoryPage"
          :loading-history="admin.loadingHistory"
          :format-item-stats="admin.formatItemStats"
          :fetch-boss-history="admin.fetchBossHistory"
        />

        <AdminDashboardTab
          v-else
          :admin-state="admin.adminState"
          :loading-players="admin.loadingPlayers"
          :player-page="admin.playerPage"
          :fetch-player-page="admin.fetchPlayerPage"
          :reset-player-password="admin.resetPlayerPassword"
          :saving="admin.saving"
        />
      </article>
    </section>
  </main>
</template>
