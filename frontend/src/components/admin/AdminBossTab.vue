<script setup>
defineProps({
  adminState: { type: Object, required: true },
  bossCycleEnabled: { type: Boolean, required: true },
  bossForm: { type: Object, required: true },
  bossTemplates: { type: Array, required: true },
  equipmentOptions: { type: Array, required: true },
  heroLootRows: { type: Array, required: true },
  heroOptions: { type: Array, required: true },
  hasBoss: { type: Boolean, required: true },
  hasEquipmentTemplates: { type: Boolean, required: true },
  hasHeroTemplates: { type: Boolean, required: true },
  lootRows: { type: Array, required: true },
  saving: { type: Boolean, required: true },
  selectedBossTemplate: { type: Object, default: null },
  selectedBossTemplateId: { type: String, required: true },
  addHeroLootRow: { type: Function, required: true },
  addLootRow: { type: Function, required: true },
  deactivateBoss: { type: Function, required: true },
  deleteBossTemplate: { type: Function, required: true },
  disableBossCycle: { type: Function, required: true },
  editBossTemplate: { type: Function, required: true },
  enableBossCycle: { type: Function, required: true },
  findEquipmentTemplate: { type: Function, required: true },
  findHeroTemplate: { type: Function, required: true },
  formatHeroTrait: { type: Function, required: true },
  formatItemStats: { type: Function, required: true },
  heroImageAlt: { type: Function, required: true },
  removeHeroLootRow: { type: Function, required: true },
  removeLootRow: { type: Function, required: true },
  saveBossTemplate: { type: Function, required: true },
  saveHeroLoot: { type: Function, required: true },
  saveLoot: { type: Function, required: true },
  selectBossTemplate: { type: Function, required: true },
})
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">循环状态</p>
          <strong>{{ bossCycleEnabled ? '循环已开启' : '循环未开启' }}</strong>
        </div>

        <p class="social-card__copy">当前 Boss：{{ adminState.boss?.name || '暂无活动 Boss' }}</p>
        <div class="admin-cycle-pills">
          <span class="boss-stage__pill">
            {{ bossCycleEnabled ? '击败后会立即补下一只' : '击败后不会自动补位' }}
          </span>
          <span class="boss-stage__pill">
            {{ adminState.boss?.templateId ? `来源模板 ${adminState.boss.templateId}` : '当前没有绑定模板' }}
          </span>
        </div>

        <div v-if="hasBoss" class="admin-boss-summary">
          <p>实例 ID：{{ adminState.boss.id }}</p>
          <p>状态：{{ adminState.boss.status }} · 血量 {{ adminState.boss.currentHp }}/{{ adminState.boss.maxHp }}</p>
        </div>
        <p v-else class="feedback" style="margin-top: 0.75rem;">
          开启循环后，如果当前没有 Boss，会立刻从 Boss 池里随机刷出一只。
        </p>

        <div class="admin-inline-actions" style="margin-top: 1rem;">
          <button class="nickname-form__submit" type="button" :disabled="saving || bossCycleEnabled" @click="enableBossCycle">
            开启循环
          </button>
          <button class="nickname-form__ghost" type="button" :disabled="saving || !bossCycleEnabled" @click="disableBossCycle">
            停止循环
          </button>
          <button v-if="hasBoss" class="nickname-form__ghost" type="button" :disabled="saving" @click="deactivateBoss">
            {{ bossCycleEnabled ? '跳过当前 Boss' : '关闭当前 Boss' }}
          </button>
        </div>

        <div v-if="hasBoss && adminState.loot.length > 0" style="margin-top: 1rem;">
          <p class="vote-stage__eyebrow">当前实例掉落快照</p>
          <ul class="inventory-list">
            <li v-for="item in adminState.loot" :key="item.itemId" class="inventory-item">
              <div>
                <strong>{{ item.itemName || item.itemId }}</strong>
                <p>{{ item.itemId }} · {{ item.slot }} · 权重 {{ item.weight }}</p>
                <p>{{ formatItemStats(item) }}</p>
              </div>
            </li>
          </ul>
        </div>

        <div v-if="hasBoss && adminState.heroLoot.length > 0" style="margin-top: 1rem;">
          <p class="vote-stage__eyebrow">当前实例英雄快照</p>
          <ul class="inventory-list">
            <li v-for="hero in adminState.heroLoot" :key="hero.heroId" class="inventory-item">
              <div class="admin-entity">
                <img v-if="hero.imagePath" class="admin-entity__avatar" :src="hero.imagePath" :alt="heroImageAlt(hero)" />
                <div>
                  <strong>{{ hero.heroName || hero.heroId }}</strong>
                  <p>{{ hero.heroId }} · 权重 {{ hero.weight }}</p>
                  <p>{{ formatItemStats(hero) }}</p>
                  <p>{{ formatHeroTrait(hero) }}</p>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </section>

      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">Boss 池模板</p>
          <strong>{{ bossTemplates.length }} 只</strong>
        </div>

        <form class="admin-form" @submit.prevent="saveBossTemplate">
          <input v-model="bossForm.id" class="nickname-form__input" type="text" placeholder="模板 ID，如 dragon" />
          <input v-model="bossForm.name" class="nickname-form__input" type="text" placeholder="Boss 显示名称" />
          <input v-model="bossForm.maxHp" class="nickname-form__input" type="number" min="1" placeholder="总血量，玩家点击消耗" />
          <button class="nickname-form__submit" type="submit" :disabled="saving">保存 Boss 模板</button>
        </form>

        <ul class="inventory-list">
          <li v-for="entry in bossTemplates" :key="entry.id" class="inventory-item inventory-item--stacked">
            <div>
              <strong>{{ entry.name }}</strong>
              <p>{{ entry.id }} · 血量 {{ entry.maxHp }} · 装备 {{ entry.loot.length }} 件 · 英雄 {{ entry.heroLoot.length }} 位</p>
            </div>
            <div class="admin-inline-actions admin-inline-actions--stacked">
              <button class="inventory-item__action" type="button" @click="selectBossTemplate(entry.id)">编辑掉落</button>
              <button class="inventory-item__action" type="button" @click="editBossTemplate(entry)">编辑模板</button>
              <button class="nickname-form__ghost" type="button" @click="deleteBossTemplate(entry.id)">删除</button>
            </div>
          </li>
        </ul>
      </section>
    </div>

    <section class="social-card admin-section-card">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">模板掉落池</p>
        <strong>{{ selectedBossTemplate?.name || selectedBossTemplateId || '未选择模板' }}</strong>
      </div>

      <p class="feedback" style="margin-bottom: 0.75rem;">
        掉落池保存到模板上。Boss 刷出来时会复制一份到当前实例，所以你后面再改模板，不会改到场上的那只。
      </p>

      <p v-if="!hasEquipmentTemplates" class="feedback" style="margin-bottom: 0.75rem;">
        当前还没有装备模板，先去“装备”页创建装备，再回来配置掉落池。
      </p>

      <div class="admin-form admin-form--tight">
        <div v-for="(entry, index) in lootRows" :key="`${selectedBossTemplateId}-${index}-${entry.itemId}`" class="admin-inline-row">
          <div class="admin-loot-select">
            <select v-model="entry.itemId" class="nickname-form__input" :disabled="!hasEquipmentTemplates && !entry.itemId">
              <option value="">选择已有装备</option>
              <option v-if="entry.itemId && !findEquipmentTemplate(entry.itemId)" :value="entry.itemId">
                {{ entry.itemId }}（已删除的装备）
              </option>
              <option v-for="item in equipmentOptions" :key="item.itemId" :value="item.itemId">
                {{ item.name }} · {{ item.itemId }} · {{ item.slot }}
              </option>
            </select>
            <p v-if="findEquipmentTemplate(entry.itemId)" class="admin-loot-select__meta">
              {{ formatItemStats(findEquipmentTemplate(entry.itemId)) }}
            </p>
          </div>
          <input v-model="entry.weight" class="nickname-form__input" type="number" min="1" placeholder="掉率权重，越大越容易掉落" />
          <button class="nickname-form__ghost" type="button" @click="removeLootRow(index)">删</button>
        </div>
        <div class="admin-inline-actions">
          <button class="nickname-form__ghost" type="button" @click="addLootRow">加一行</button>
          <button class="nickname-form__submit" type="button" :disabled="saving" @click="saveLoot">保存模板掉落池</button>
        </div>
      </div>
    </section>

    <section class="social-card admin-section-card">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">模板英雄池</p>
        <strong>{{ selectedBossTemplate?.name || selectedBossTemplateId || '未选择模板' }}</strong>
      </div>

      <p class="feedback" style="margin-bottom: 0.75rem;">英雄池和装备池分开配置，Boss 被击败时会分别独立抽取。</p>

      <p v-if="!hasHeroTemplates" class="feedback" style="margin-bottom: 0.75rem;">
        当前还没有英雄模板，先去“英雄”页创建模板，再回来配置英雄池。
      </p>

      <div class="admin-form admin-form--tight">
        <div v-for="(entry, index) in heroLootRows" :key="`${selectedBossTemplateId}-hero-${index}-${entry.heroId}`" class="admin-inline-row">
          <div class="admin-loot-select">
            <select v-model="entry.heroId" class="nickname-form__input" :disabled="!hasHeroTemplates && !entry.heroId">
              <option value="">选择已有英雄</option>
              <option v-if="entry.heroId && !findHeroTemplate(entry.heroId)" :value="entry.heroId">
                {{ entry.heroId }}（已删除的英雄）
              </option>
              <option v-for="hero in heroOptions" :key="hero.heroId" :value="hero.heroId">
                {{ hero.name }} · {{ hero.heroId }}
              </option>
            </select>
            <p v-if="findHeroTemplate(entry.heroId)" class="admin-loot-select__meta">
              {{ formatItemStats(findHeroTemplate(entry.heroId)) }} · {{ formatHeroTrait(findHeroTemplate(entry.heroId)) }}
            </p>
          </div>
          <input v-model="entry.weight" class="nickname-form__input" type="number" min="1" placeholder="掉率权重，越大越容易招募" />
          <button class="nickname-form__ghost" type="button" @click="removeHeroLootRow(index)">删</button>
        </div>
        <div class="admin-inline-actions">
          <button class="nickname-form__ghost" type="button" @click="addHeroLootRow">加一行</button>
          <button class="nickname-form__submit" type="button" :disabled="saving" @click="saveHeroLoot">保存模板英雄池</button>
        </div>
      </div>
    </section>
  </div>
</template>
