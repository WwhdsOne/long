<script setup>
defineProps({
  heroes: { type: Array, required: true },
  heroForm: { type: Object, required: true },
  uploadingImage: { type: Boolean, required: true },
  saving: { type: Boolean, required: true },
  formatItemStats: { type: Function, required: true },
  formatHeroTrait: { type: Function, required: true },
  heroImageAlt: { type: Function, required: true },
  saveHero: { type: Function, required: true },
  uploadHeroImage: { type: Function, required: true },
  editHero: { type: Function, required: true },
  deleteHero: { type: Function, required: true },
})
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">英雄模板</p>
          <strong>{{ heroes.length }} 位</strong>
        </div>

        <form class="admin-form" @submit.prevent="saveHero">
          <input v-model="heroForm.heroId" class="nickname-form__input" type="text" placeholder="唯一标识，如 spark-cat" />
          <input v-model="heroForm.name" class="nickname-form__input" type="text" placeholder="前台显示名称" />
          <input v-model="heroForm.imagePath" class="nickname-form__input" type="text" placeholder="头像 URL（可选）" />
          <input v-model="heroForm.imageAlt" class="nickname-form__input" type="text" placeholder="头像说明（可选）" />
          <label class="admin-upload">
            <span>或上传到 OSS（支持 webp）</span>
            <input type="file" accept="image/*" :disabled="uploadingImage" @change="uploadHeroImage" />
          </label>
          <p v-if="heroForm.imagePath" class="admin-upload__result">
            当前头像地址：{{ heroForm.imagePath }}
          </p>
          <img
            v-if="heroForm.imagePath"
            class="admin-upload__preview admin-upload__preview--avatar"
            :src="heroForm.imagePath"
            :alt="heroForm.imageAlt || heroForm.name || heroForm.heroId || '英雄头像预览'"
          />
          <p class="feedback">
            {{ uploadingImage ? '图片上传中...' : '如果 OSS 还没配置，也可以继续手填图片 URL。' }}
          </p>
          <input v-model="heroForm.bonusClicks" class="nickname-form__input" type="number" min="0" placeholder="点击加成" />
          <input v-model="heroForm.bonusCriticalChancePercent" class="nickname-form__input" type="number" min="0" max="100" placeholder="暴击率加成" />
          <input v-model="heroForm.bonusCriticalCount" class="nickname-form__input" type="number" min="0" placeholder="暴击额外加成" />
          <input v-model="heroForm.awakenCap" class="nickname-form__input" type="number" min="0" placeholder="觉醒上限（0 表示不设上限）" />
          <select v-model="heroForm.traitType" class="nickname-form__input">
            <option value="bonus_clicks">额外点击</option>
            <option value="critical_chance_percent">暴击率</option>
            <option value="critical_count_bonus">暴击额外</option>
            <option value="final_damage_percent">最终伤害百分比</option>
          </select>
          <input v-model="heroForm.traitValue" class="nickname-form__input" type="number" min="0" placeholder="被动数值" />
          <button class="nickname-form__submit" type="submit" :disabled="saving">
            保存英雄
          </button>
        </form>
      </section>

      <section class="social-card">
        <ul class="inventory-list">
          <li v-for="hero in heroes" :key="hero.heroId" class="inventory-item inventory-item--stacked">
            <div class="admin-entity">
              <img
                v-if="hero.imagePath"
                class="admin-entity__avatar"
                :src="hero.imagePath"
                :alt="heroImageAlt(hero)"
              />
              <div>
                <strong>{{ hero.name }}</strong>
                <p>{{ hero.heroId }}</p>
                <p>觉醒上限 {{ hero.awakenCap || '未设置' }}</p>
                <p>{{ formatItemStats(hero) }}</p>
                <p>{{ formatHeroTrait(hero) }}</p>
              </div>
            </div>
            <div class="admin-inline-actions">
              <button class="inventory-item__action" type="button" @click="editHero(hero)">编辑</button>
              <button class="nickname-form__ghost" type="button" @click="deleteHero(hero.heroId)">删除</button>
            </div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>
