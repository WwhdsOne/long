<script setup>
import { formatRarityLabel, RARITY_OPTIONS } from '../../utils/rarity'

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
  uploadEquipmentImage: { type: Function, required: true },
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

          <fieldset class="admin-fieldset">
            <legend class="admin-fieldset__legend">图片</legend>
            <input v-model="equipmentForm.imagePath" class="nickname-form__input" type="text" placeholder="图片 URL（可直接填 OSS/CDN 地址）" />
            <input v-model="equipmentForm.imageAlt" class="nickname-form__input" type="text" placeholder="图片说明（可选）" />
            <label class="admin-upload">
              <span>上传到 OSS</span>
              <input type="file" accept="image/*" @change="uploadEquipmentImage" />
            </label>
            <img
              v-if="equipmentForm.imagePath"
              class="admin-upload__preview"
              :src="equipmentForm.imagePath"
              :alt="equipmentForm.imageAlt || equipmentForm.name || '装备预览图'"
            />
          </fieldset>

          <fieldset class="admin-fieldset">
            <legend class="admin-fieldset__legend">属性</legend>
            <input v-model="equipmentForm.attackPower" class="nickname-form__input" type="number" min="0" step="1" placeholder="攻击力" />
            <input v-model="equipmentForm.armorPenPercent" class="nickname-form__input" type="number" min="0" max="1" step="0.01" placeholder="破甲率 0~0.80" />
            <input v-model="equipmentForm.critDamageMultiplier" class="nickname-form__input" type="number" min="1" step="0.1" placeholder="暴击伤害倍率，默认 1.0" />
            <input v-model="equipmentForm.bossDamagePercent" class="nickname-form__input" type="number" min="0" max="10" step="0.01" placeholder="Boss 增伤百分比，如 0.30" />
            <input v-model="equipmentForm.partTypeDamageSoft" class="nickname-form__input" type="number" min="0" step="0.01" placeholder="软组织增伤，如 0.40" />
            <input v-model="equipmentForm.partTypeDamageHeavy" class="nickname-form__input" type="number" min="0" step="0.01" placeholder="重甲增伤，如 0.50" />
            <input v-model="equipmentForm.partTypeDamageWeak" class="nickname-form__input" type="number" min="0" step="0.01" placeholder="弱点增伤，如 0.30" />
            <select v-model="equipmentForm.talentAffinity" class="nickname-form__input">
              <option value="">通用（无天赋绑定）</option>
              <option value="normal">均衡攻势</option>
              <option value="armor">碎盾攻坚</option>
              <option value="crit">致命洞察</option>
            </select>
          </fieldset>
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
              <p>{{ item.itemId }} · {{ item.slot }} · {{ formatRarityLabel(item.rarity) }}</p>
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
