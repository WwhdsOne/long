<script setup>
defineProps({
  shopItems: { type: Array, required: true },
  shopItemForm: { type: Object, required: true },
  loadingShopItems: { type: Boolean, default: false },
  saving: { type: Boolean, default: false },
  openNewShopItem: { type: Function, required: true },
  saveShopItem: { type: Function, required: true },
  editShopItem: { type: Function, required: true },
  deleteShopItem: { type: Function, required: true },
  uploadShopImage: { type: Function, required: true },
  uploadShopPreviewImage: { type: Function, required: true },
  uploadShopCursorImage: { type: Function, required: true },
})
</script>

<template>
  <section class="admin-stack">
    <article class="admin-panel">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">商店商品</p>
        <strong>点击图标目录</strong>
      </div>
      <button class="nickname-form__ghost" type="button" @click="openNewShopItem">新建商品</button>

      <form class="admin-form" @submit.prevent="saveShopItem">
        <div class="nickname-form__group">
          <label for="itemId">Item ID:</label>
          <input v-model="shopItemForm.itemId" class="nickname-form__input" type="text" placeholder="itemId" id="itemId" />
        </div>
        <div class="nickname-form__group">
          <label for="title">标题:</label>
          <input v-model="shopItemForm.title" class="nickname-form__input" type="text" placeholder="标题" id="title" />
        </div>
        <div class="nickname-form__group">
          <label for="priceGold">金币价格:</label>
          <input v-model="shopItemForm.priceGold" class="nickname-form__input" type="number" min="0" placeholder="金币价格" id="priceGold" />
        </div>
        <div class="nickname-form__group">
          <label for="sortOrder">排序:</label>
          <input v-model="shopItemForm.sortOrder" class="nickname-form__input" type="number" placeholder="排序" id="sortOrder" />
        </div>
        <div class="nickname-form__group">
          <label for="imagePath">主图 URL:</label>
          <input v-model="shopItemForm.imagePath" class="nickname-form__input" type="text" placeholder="主图 URL" id="imagePath" />
        </div>
        <div class="nickname-form__group">
          <label for="previewImagePath">预览图 URL:</label>
          <input v-model="shopItemForm.previewImagePath" class="nickname-form__input" type="text" placeholder="预览图 URL" id="previewImagePath" />
        </div>
        <div class="nickname-form__group">
          <label for="battleClickCursorImagePath">战斗光标 URL:</label>
          <input v-model="shopItemForm.battleClickCursorImagePath" class="nickname-form__input" type="text" placeholder="战斗光标 URL" id="battleClickCursorImagePath" />
        </div>
        <div class="nickname-form__group">
          <label for="imageAlt">图片 alt:</label>
          <input v-model="shopItemForm.imageAlt" class="nickname-form__input" type="text" placeholder="图片 alt" id="imageAlt" />
        </div>
        <div class="nickname-form__group">
          <label for="description">描述:</label>
          <textarea v-model="shopItemForm.description" class="nickname-form__input" rows="4" placeholder="描述" id="description"></textarea>
        </div>
        <div class="nickname-form__group">
          <label for="active">
            <input v-model="shopItemForm.active" type="checkbox" id="active" /> 上架
          </label>
        </div>
        <div class="nickname-form__group">
          <label for="autoEquipOnPurchase">
            <input v-model="shopItemForm.autoEquipOnPurchase" type="checkbox" id="autoEquipOnPurchase" /> 购买后自动装备
          </label>
        </div>
        <div class="admin-toolbar__actions">
          <label class="nickname-form__ghost">
            上传主图
            <input hidden type="file" accept="image/*" @change="uploadShopImage" />
          </label>
          <label class="nickname-form__ghost">
            上传预览图
            <input hidden type="file" accept="image/*" @change="uploadShopPreviewImage" />
          </label>
          <label class="nickname-form__ghost">
            上传点击图标
            <input hidden type="file" accept="image/*" @change="uploadShopCursorImage" />
          </label>
          <button class="nickname-form__submit" type="submit" :disabled="saving">
            {{ saving ? '保存中...' : '保存商品' }}
          </button>
        </div>
      </form>
    </article>

    <article class="admin-panel">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">目录列表</p>
        <strong>当前商店商品</strong>
      </div>

      <p v-if="loadingShopItems" class="feedback-panel">商店商品加载中...</p>
      <div v-else class="admin-list">
        <article v-for="item in shopItems" :key="item.itemId" class="social-card">
          <div class="social-card__head">
            <strong>{{ item.title }}</strong>
            <span>{{ item.active ? '上架中' : '未上架' }}</span>
          </div>
          <p class="social-card__copy">ID: {{ item.itemId }}</p>
          <p class="social-card__copy">价格: {{ item.priceGold }} 金币</p>
          <p class="social-card__copy">排序: {{ item.sortOrder }}</p>
          <div class="admin-toolbar__actions">
            <button class="nickname-form__ghost" type="button" @click="editShopItem(item)">编辑</button>
            <button class="nickname-form__ghost" type="button" @click="deleteShopItem(item.itemId)">删除</button>
          </div>
        </article>
      </div>
    </article>
  </section>
</template>
