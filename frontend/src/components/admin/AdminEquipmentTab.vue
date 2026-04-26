<script setup>
import {formatRarityLabel, RARITY_OPTIONS} from '../../utils/rarity'
import {EQUIPMENT_SLOTS} from '../../utils/equipmentSlots'

defineProps({
  equipmentPage: {type: Object, required: true},
  equipmentForm: {type: Object, required: true},
  equipmentPrompt: {type: String, required: true},
  showEquipmentEditor: {type: Boolean, required: true},
  loadingEquipment: {type: Boolean, required: true},
  generatingEquipmentDraft: {type: Boolean, required: true},
  saving: {type: Boolean, required: true},
  formatItemStats: {type: Function, required: true},
  openNewEquipment: {type: Function, required: true},
  updateEquipmentPrompt: {type: Function, required: true},
  generateEquipmentDraft: {type: Function, required: true},
  saveEquipment: {type: Function, required: true},
  editEquipment: {type: Function, required: true},
  deleteEquipment: {type: Function, required: true},
  fetchEquipmentPage: {type: Function, required: true},
  uploadEquipmentImage: {type: Function, required: true},
})
</script>

<template>
  <div class="admin-section">
    <section class="social-card equipment-admin">
      <div class="social-card__head equipment-admin__head">
        <div>
          <p class="vote-stage__eyebrow">装备模板</p>
          <strong>{{ equipmentPage.total }} 件</strong>
        </div>
        <button class="nickname-form__submit equipment-admin__new" type="button" @click="openNewEquipment">
          新增装备
        </button>
      </div>

      <div v-if="showEquipmentEditor" class="equipment-editor">
        <div class="equipment-generator">
          <textarea
              :value="equipmentPrompt"
              class="nickname-form__input equipment-generator__textarea"
              rows="4"
              placeholder="描述装备定位、部位、稀有度和属性倾向，例如：做一把史诗武器，偏普攻流，强化软组织伤害。"
              @input="updateEquipmentPrompt($event.target.value)"
          ></textarea>
          <button
              class="nickname-form__ghost equipment-generator__button"
              type="button"
              :disabled="generatingEquipmentDraft"
              @click="generateEquipmentDraft"
          >
            生成草稿
          </button>
        </div>

        <form class="admin-form equipment-form" @submit.prevent="saveEquipment">
          <label class="admin-labeled-field">
            <span>标识:</span>
            <input v-model="equipmentForm.itemId" class="nickname-form__input" type="text" placeholder="wood-sword"/>
          </label>
          <label class="admin-labeled-field">
            <span>名称:</span>
            <input v-model="equipmentForm.name" class="nickname-form__input" type="text" placeholder="前台显示名称"/>
          </label>
          <label class="admin-labeled-field">
            <span>装备描述:</span>
            <textarea
                v-model="equipmentForm.description"
                class="nickname-form__input"
                rows="4"
                placeholder="描述装备背景故事"
            ></textarea>
          </label>
          <label class="admin-labeled-field">
            <span>部位:</span>
            <select v-model="equipmentForm.slot" class="nickname-form__input">
              <option v-for="slot in EQUIPMENT_SLOTS" :key="slot.value" :value="slot.value">
                {{ slot.label }}（{{ slot.value }}）
              </option>
            </select>
          </label>
          <label class="admin-labeled-field">
            <span>稀有度:</span>
            <select v-model="equipmentForm.rarity" class="nickname-form__input">
              <option v-for="rarity in RARITY_OPTIONS" :key="rarity" :value="rarity">{{ rarity }}</option>
            </select>
          </label>

          <fieldset class="admin-fieldset equipment-form__fieldset">
            <legend class="admin-fieldset__legend">图片</legend>
            <label class="admin-labeled-field">
              <span>地址:</span>
              <input v-model="equipmentForm.imagePath" class="nickname-form__input" type="text" placeholder="图片 URL"/>
            </label>
            <label class="admin-labeled-field">
              <span>说明:</span>
              <input v-model="equipmentForm.imageAlt" class="nickname-form__input" type="text" placeholder="图片说明"/>
            </label>
            <label class="admin-upload">
              <span>上传到 OSS</span>
              <input type="file" accept="image/*" @change="uploadEquipmentImage"/>
            </label>
            <img
                v-if="equipmentForm.imagePath"
                class="admin-upload__preview"
                :src="equipmentForm.imagePath"
                :alt="equipmentForm.imageAlt || equipmentForm.name || '装备预览图'"
            />
          </fieldset>

          <fieldset class="admin-fieldset equipment-form__fieldset">
            <legend class="admin-fieldset__legend">属性</legend>
            <label class="admin-labeled-field">
              <span>攻击力:</span>
              <input v-model="equipmentForm.attackPower" class="nickname-form__input" type="number" min="0" step="1"/>
            </label>
            <label class="admin-labeled-field">
              <span>破甲:</span>
              <input v-model="equipmentForm.armorPenPercent" class="nickname-form__input" type="number" min="0"
                     max="0.8" step="0.01"/>
            </label>
            <label class="admin-labeled-field">
              <span>暴击率:</span>
              <label class="admin-labeled-field">
                <input v-model="equipmentForm.critRate" class="nickname-form__input" type="number" min="0" max="0.35"
                       step="0.01"/>
              </label>
            </label>
            <label class="admin-labeled-field">
              <span>暴伤:</span>
              <input v-model="equipmentForm.critDamageMultiplier" class="nickname-form__input" type="number" min="0"
                     step="0.1"/>
            </label>
            <label class="admin-labeled-field">
              <span>软组织:</span>
              <input v-model="equipmentForm.partTypeDamageSoft" class="nickname-form__input" type="number" min="0"
                     step="0.01"/>
            </label>
            <label class="admin-labeled-field">
              <span>重甲:</span>
              <input v-model="equipmentForm.partTypeDamageHeavy" class="nickname-form__input" type="number" min="0"
                     step="0.01"/>
            </label>
            <label class="admin-labeled-field">
              <span>弱点:</span>
              <input v-model="equipmentForm.partTypeDamageWeak" class="nickname-form__input" type="number" min="0"
                     step="0.01"/>
            </label>
            <label class="admin-labeled-field">
              <span>天赋:</span>
              <select v-model="equipmentForm.talentAffinity" class="nickname-form__input">
                <option value="">通用（无天赋绑定）</option>
                <option value="normal">均衡攻势</option>
                <option value="armor">碎盾攻坚</option>
                <option value="crit">致命洞察</option>
              </select>
            </label>
          </fieldset>
          <button class="nickname-form__submit equipment-form__save" type="submit" :disabled="saving">
            保存装备
          </button>
        </form>
      </div>

      <div v-if="loadingEquipment" class="feedback-panel">
        <p>装备列表加载中...</p>
      </div>
      <ul class="inventory-list inventory-list--equipment-grid">
        <li v-for="item in equipmentPage.items" :key="item.itemId" class="inventory-item equipment-card">
          <div class="equipment-card__main">
            <strong>{{ item.name }}</strong>
            <p>{{ item.itemId }}</p>
            <dl class="equipment-card__meta">
              <div>
                <dt>部位</dt>
                <dd>{{ item.slot }}</dd>
              </div>
              <div>
                <dt>稀有度</dt>
                <dd>{{ formatRarityLabel(item.rarity) }}</dd>
              </div>
              <div>
                <dt>天赋</dt>
                <dd>{{ item.talentAffinity || '通用' }}</dd>
              </div>
            </dl>
            <p class="equipment-card__stats">{{ formatItemStats(item) || '无主要属性' }}</p>
          </div>
          <div class="admin-inline-actions equipment-card__actions">
            <button class="inventory-item__action" type="button" @click="editEquipment(item)">编辑</button>
            <button class="nickname-form__ghost" type="button" @click="deleteEquipment(item.itemId)">删除</button>
          </div>
        </li>
      </ul>
      <div class="admin-inline-actions equipment-admin__pager">
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
</template>
