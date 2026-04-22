<script setup>
import { RARITY_OPTIONS } from '../../utils/rarity'

defineProps({
  equipmentPage: { type: Object, required: true },
  equipmentForm: { type: Object, required: true },
  loadingEquipment: { type: Boolean, required: true },
  saving: { type: Boolean, required: true },
  formatItemStats: { type: Function, required: true },
  saveEquipment: { type: Function, required: true },
  editEquipment: { type: Function, required: true },
  deleteEquipment: { type: Function, required: true },
  fetchEquipmentPage: { type: Function, required: true },
})
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">装备模板</p>
          <strong>{{ equipmentPage.total }} 件</strong>
        </div>

        <form class="admin-form" @submit.prevent="saveEquipment">
          <input v-model="equipmentForm.itemId" class="nickname-form__input" type="text" placeholder="唯一标识，如 wood-sword" />
          <input v-model="equipmentForm.name" class="nickname-form__input" type="text" placeholder="前台显示的名称" />
          <select v-model="equipmentForm.slot" class="nickname-form__input">
            <option value="weapon">weapon</option>
            <option value="armor">armor</option>
            <option value="accessory">accessory</option>
          </select>
          <select v-model="equipmentForm.rarity" class="nickname-form__input">
            <option v-for="rarity in RARITY_OPTIONS" :key="rarity" :value="rarity">{{ rarity }}</option>
          </select>
          <input v-model="equipmentForm.bonusClicks" class="nickname-form__input" type="number" min="0" placeholder="每次点击额外加几票" />
          <input v-model="equipmentForm.bonusCriticalChancePercent" class="nickname-form__input" type="number" min="0" max="100" placeholder="暴击概率 +N%" />
          <input v-model="equipmentForm.bonusCriticalCount" class="nickname-form__input" type="number" min="0" placeholder="暴击时额外加几票" />
          <input v-model="equipmentForm.enhanceCap" class="nickname-form__input" type="number" min="0" placeholder="强化上限（0 表示不设上限）" />
          <button class="nickname-form__submit" type="submit" :disabled="saving">
            保存装备
          </button>
        </form>
      </section>

      <section class="social-card">
        <div v-if="loadingEquipment" class="feedback-panel">
          <p>装备列表加载中...</p>
        </div>
        <ul class="inventory-list">
          <li v-for="item in equipmentPage.items" :key="item.itemId" class="inventory-item">
            <div>
              <strong>{{ item.name }}</strong>
              <p>{{ item.itemId }} · {{ item.slot }} · {{ item.rarity || '普通' }}</p>
              <p>强化上限 {{ item.enhanceCap || '未设置' }}</p>
              <p>{{ formatItemStats(item) }}</p>
            </div>
            <div class="admin-inline-actions">
              <button class="inventory-item__action" type="button" @click="editEquipment(item)">编辑</button>
              <button class="nickname-form__ghost" type="button" @click="deleteEquipment(item.itemId)">删除</button>
            </div>
          </li>
        </ul>
        <div class="admin-inline-actions" style="margin-top: 1rem;">
          <button
            class="nickname-form__ghost"
            type="button"
            :disabled="loadingEquipment || equipmentPage.page <= 1"
            @click="fetchEquipmentPage(equipmentPage.page - 1)"
          >
            上一页
          </button>
          <span class="feedback">第 {{ equipmentPage.page }} / {{ Math.max(equipmentPage.totalPages, 1) }} 页</span>
          <button
            class="nickname-form__ghost"
            type="button"
            :disabled="loadingEquipment || equipmentPage.page >= equipmentPage.totalPages"
            @click="fetchEquipmentPage(equipmentPage.page + 1)"
          >
            下一页
          </button>
        </div>
      </section>
    </div>
  </div>
</template>
