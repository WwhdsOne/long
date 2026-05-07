<script setup>
defineProps({
  buttonPage: {type: Object, required: true},
  buttonForm: {type: Object, required: true},
  uploadingImage: {type: Boolean, required: true},
  loadingButtons: {type: Boolean, required: true},
  saving: {type: Boolean, required: true},
  saveButton: {type: Function, required: true},
  uploadButtonImage: {type: Function, required: true},
  editButton: {type: Function, required: true},
  fetchButtonPage: {type: Function, required: true},
})
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">按钮配置</p>
          <strong>{{ buttonPage.total }} 个</strong>
        </div>

        <form class="admin-form" @submit.prevent="saveButton">
          <input v-model="buttonForm.slug" class="nickname-form__input" type="text" placeholder="唯一标识，如 feel"/>
          <input v-model="buttonForm.label" class="nickname-form__input" type="text" placeholder="前台显示的文字"/>
          <input v-model="buttonForm.sort" class="nickname-form__input" type="number"
                 placeholder="排序，数字小的排前面"/>
          <input v-model="buttonForm.tagsText" class="nickname-form__input" type="text"
                 placeholder="标签，逗号分隔，如 日常, 活动"/>
          <input v-model="buttonForm.imagePath" class="nickname-form__input" type="text"
                 placeholder="图片 URL（可选，可直接填 OSS/CDN 地址）"/>
          <input v-model="buttonForm.imageAlt" class="nickname-form__input" type="text" placeholder="图片说明（可选）"/>
          <label class="admin-upload">
            <span>或上传到 OSS</span>
            <input type="file" accept="image/*" :disabled="uploadingImage" @change="uploadButtonImage"/>
          </label>
          <p v-if="buttonForm.imagePath" class="admin-upload__result">
            当前图片地址：{{ buttonForm.imagePath }}
          </p>
          <img
              v-if="buttonForm.imagePath"
              class="admin-upload__preview"
              :src="buttonForm.imagePath"
              :alt="buttonForm.imageAlt || buttonForm.label || '按钮预览图'"
          />
          <p class="feedback">
            {{ uploadingImage ? '图片上传中...' : '如果 OSS 还没配置，也可以继续手填图片 URL。' }}
          </p>
          <label class="admin-check">
            <input v-model="buttonForm.enabled" type="checkbox"/>
            启用按钮
          </label>
          <button class="nickname-form__submit" type="submit" :disabled="saving">
            保存按钮
          </button>
        </form>
      </section>

      <section class="social-card">
        <div v-if="loadingButtons" class="feedback-panel">
          <p>按钮列表加载中...</p>
        </div>
        <ul class="inventory-list">
          <li v-for="button in buttonPage.items" :key="button.key" class="inventory-item">
            <div>
              <strong>{{ button.label }}</strong>
              <p>{{ button.key }} · sort {{ button.sort }} · {{ button.enabled ? '启用' : '停用' }}</p>
              <p>{{ Array.isArray(button.tags) && button.tags.length > 0 ? button.tags.join(' / ') : '未打标签' }}</p>
            </div>
            <button class="inventory-item__action" type="button" @click="editButton(button)">编辑</button>
          </li>
        </ul>
        <div class="admin-inline-actions" style="margin-top: 1rem;">
          <button
              class="nickname-form__ghost"
              type="button"
              :disabled="loadingButtons || buttonPage.page <= 1"
              @click="fetchButtonPage(buttonPage.page - 1)"
          >
            上一页
          </button>
          <span class="feedback">第 {{ buttonPage.page }} / {{ Math.max(buttonPage.totalPages, 1) }} 页</span>
          <button
              class="nickname-form__ghost"
              type="button"
              :disabled="loadingButtons || buttonPage.page >= buttonPage.totalPages"
              @click="fetchButtonPage(buttonPage.page + 1)"
          >
            下一页
          </button>
        </div>
      </section>
    </div>
  </div>
</template>
